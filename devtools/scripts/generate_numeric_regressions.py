#!/usr/bin/env python3
import argparse
import collections
import decimal
import hashlib
import json
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile


ROOT = Path(__file__).resolve().parents[2]
INTEL = ROOT / "devtools/third_party/intel_dfp"
DATA = ROOT / "devtools/testdata"
OBS = DATA / "readtest_numeric_c_observations.json"
ARCHIVE_SHA = "1df86132e7a31fd74d784fee1c679b21a088f73a8ec979cfaf784c200392e125"
ROUNDINGS = [decimal.ROUND_HALF_EVEN, decimal.ROUND_FLOOR, decimal.ROUND_CEILING,
             decimal.ROUND_DOWN, decimal.ROUND_HALF_UP]
EXACT = decimal.Context(prec=120, Emin=-999999, Emax=999999)
PARAMS = {32: (7, -95, 96, -101, 101), 64: (16, -383, 384, -398, 398),
          128: (34, -6143, 6144, -6176, 6176)}
REASONS = {
    "carry_underflow": "INTEL-BID-013: pinned Intel BID C from_string normalizes a delayed directed-rounding carry before testing tininess and omits Underflow; IEEE 754 section 7.5 requires Underflow|Inexact for the inexact tiny input",
    "directed_underflow": "pinned Intel BID C from_string forces nearest rounding in tiny-input packing instead of the requested mode; IEEE 754 requires the correctly rounded signed result with Underflow|Inexact",
    "fixed_point_underflow": "pinned Intel BID C bid32_from_string passes a pre-rounded coefficient to a nearest-only packer in the no-exponent path, causing double rounding, loss of requested rounding mode, or loss of original tininess; IEEE 754 requires one correctly directed rounding of the exact input with Underflow|Inexact",
    "long_fraction": "INTEL-BID-009: pinned Intel BID C bid64_from_string passes a tiny no-exponent input through an overflow-only packer; IEEE 754 requires rounding the exact input and raising Underflow|Inexact",
    "zero_cohort": "pinned Intel BID C from_string raises Underflow|Inexact or produces a minimum nonzero coefficient for exact written zero outside the finite cohort range; IEEE 754 requires exact signed zero without flags in the raw parser",
    "threshold": "pinned Intel BID C bid128_scalbn/scalbln/ldexp compares only the upper coefficient limb at the 10^33 normalization threshold and falsely overflows an exactly representable finite result",
    "premature_overflow": "pinned Intel BID C bid128_fma overflows a near-maximum finite subtraction before rounding; IEEE 754 requires maximum finite with Inexact and no Overflow",
    "missing_overflow": "pinned Intel BID C bid128_fma packs an out-of-range exponent after exact subtraction without raising Overflow|Inexact; IEEE 754 requires the mode-dependent overflow result and flags",
    "overflow_sign": "pinned Intel BID C bid128_fma reads the overflow sign before assembling the negative result; IEEE 754 requires a negative directed overflow result with Overflow|Inexact",
    "quantum_steering": "INTEL-BID-006: pinned Intel BID C bid32_quantum tests a 64-bit steering mask against a 32-bit operand and extracts the wrong finite exponent; bid754 uses the encoded Decimal32 exponent",
    "sticky_inexact": "INTEL-BID-012: pinned Intel BID C bid128_scalbn/scalbln treats prior Inexact as current-operation Inexact and raises Underflow for an exact subnormal; IEEE 754 sections 7.5 and 7.6 require no new Underflow and retention of existing status",
}


def verify_archive():
    archive = INTEL / "IntelRDFPMathLib20U4.tar.gz"
    if hashlib.sha256(archive.read_bytes()).hexdigest() != ARCHIVE_SHA:
        raise ValueError("pinned Intel archive checksum mismatch")
    names = ["bid32_string.c", "bid64_string.c", "bid128_string.c",
             "bid128_scalb.c", "bid128_scalbl.c", "bid128_ldexp.c",
             "bid128_fma.c", "bid32_next.c", "bid32_quantumd.c"]
    with tarfile.open(archive, "r:gz") as tar:
        for name in names:
            path = f"LIBRARY/src/{name}"
            member = tar.extractfile(path)
            if member is None or member.read() != (INTEL / path).read_bytes():
                raise ValueError(f"extracted Intel source differs: {path}")


def raw(width, value):
    sign, digits, exp = value.as_tuple()
    signbit = sign << (width - 1)
    if value.is_infinite():
        return f"{signbit | {32: 0x78000000, 64: 0x7800000000000000, 128: 0x78000000000000000000000000000000}[width]:0{width // 4}x}"
    coeff = int("".join(map(str, digits)))
    _, _, _, minq, bias = PARAMS[width]
    if exp < minq or exp > {32: 90, 64: 369, 128: 6111}[width]:
        raise ValueError(f"unencodable decimal cohort: {value}")
    biased = exp + bias
    if width == 32:
        bits = ((biased << 23) | coeff) if coeff < 0x800000 else (0x60000000 | biased << 21 | (coeff & 0x1fffff))
    elif width == 64:
        bits = ((biased << 53) | coeff) if coeff < 0x20000000000000 else (0x6000000000000000 | biased << 51 | (coeff & 0x7ffffffffffff))
    else:
        bits = biased << 113 | coeff
    return f"{signbit | bits:0{width // 4}x}"


def flags(ctx, exact, width):
    inexact = bool(ctx.flags[decimal.Inexact])
    bits = 0x20 if inexact else 0
    if ctx.flags[decimal.Overflow]:
        bits |= 0x08
    if inexact and exact.is_finite() and exact != 0 and exact.copy_abs() < decimal.Decimal(f"1e{PARAMS[width][1]}"):
        bits |= 0x10
    return bits


def rounded(width, mode, exact, operation=None):
    precision, emin, emax, _, _ = PARAMS[width]
    ctx = decimal.Context(prec=precision, Emin=emin, Emax=emax,
                          rounding=ROUNDINGS[mode], clamp=1)
    for signal in ctx.traps:
        ctx.traps[signal] = False
    if operation is None:
        result = ctx.create_decimal(exact)
    else:
        result = operation(ctx)
    return raw(width, result), flags(ctx, exact, width)


def parser_carry_rows():
    rows = []
    for width in (32, 64):
        precision, _, _, minq, _ = PARAMS[width]
        maxq = {32: 90, 64: 369}[width]
        for coeff in (10**precision - 2, 10**precision - 1):
            for tail in ("0", "01", "001", "1", "5", "9"):
                for quantum in (minq - 2, minq - 1, minq, 0, maxq):
                    for sign in ("", "-"):
                        exponent = quantum - len(tail)
                        literal = f"{sign}{coeff}{tail}e{exponent}"
                        exact = decimal.Decimal(literal)
                        for spelling in (literal, format(exact, "f")):
                            for mode in range(5):
                                want, status = rounded(width, mode, exact)
                                rows.append((f"bid{width}_from_string", mode,
                                             (spelling,), want, status))
    return rows


def make_rows():
    rows = []

    def add(fn, mode, args, expected, status):
        rows.append((fn, mode, tuple(args), expected, status))

    for width, minq in [(32, -101), (64, -398)]:
        for sign in ("", "-"):
            for exp in (minq, minq - 1, minq - 2):
                literal = f"{sign}1e{exp}"
                for mode in range(5):
                    want, status = rounded(width, mode, decimal.Decimal(literal))
                    add(f"bid{width}_from_string", mode, [literal], want, status)

    for digits in ("14999998", "14999999", "15000000", "15000001"):
        for sign in ("", "-"):
            for literal in (f"{sign}0.{('0' * 100)}{digits}", f"{sign}{digits}e-108"):
                for mode in range(5):
                    want, status = rounded(32, mode, decimal.Decimal(literal))
                    add("bid32_from_string", mode, [literal], want, status)

    for zeros in (893, 894, 895):
        for sign in ("", "-"):
            literal = f"{sign}0.{('0' * zeros)}12345678901234567"
            for mode in range(5):
                want, status = rounded(64, mode, decimal.Decimal(literal))
                add("bid64_from_string", mode, [literal], want, status)

    for width, exponents in [(32, (-101, -102, -108, -109, -398, -399, -10000)),
                             (64, (-398, -399, -414, -415, -10000)),
                             (128, (-6176, -6177, -6210, -6211, -6212, -10000))]:
        for sign in ("", "-"):
            for exp in exponents:
                literal = f"{sign}0e{exp}"
                for mode in range(5):
                    want, status = rounded(width, mode, decimal.Decimal(literal))
                    add(f"bid{width}_from_string", mode, [literal], want, status)

    threshold = 10 ** 33
    limb = threshold >> 64 << 64
    for coeff in (threshold - 1, threshold, threshold + 1,
                  limb - 1, limb, limb + 1):
        for exp, shift in ((6110, 1), (6110, 2), (6111, 1)):
            for sign in (1, -1):
                x = decimal.Decimal(f"{sign * coeff}e{exp}")
                xraw = raw(128, x)
                for mode in range(5):
                    exact = EXACT.scaleb(x, shift)
                    want, status = rounded(128, mode, exact,
                                           lambda ctx: ctx.scaleb(x, shift))
                    for fn in ("bid128_scalbn", "bid128_scalbln", "bid128_ldexp"):
                        add(fn, mode, [xraw, str(shift)], want, status)

    for xcoeff, xexp in ((1, 35), (10, 34)):
        x = decimal.Decimal(f"{xcoeff}e{xexp}")
        for exp in (6110, 6111):
            for subtract in range(4, 12):
                for sign in (1, -1):
                    y = decimal.Decimal(f"{sign}e{exp}")
                    z = decimal.Decimal(f"{-sign * subtract}e{exp}")
                    args = [raw(128, x), raw(128, y), raw(128, z)]
                    exact = EXACT.fma(x, y, z)
                    for mode in range(5):
                        want, status = rounded(128, mode, exact,
                                               lambda ctx: ctx.fma(x, y, z))
                        add("bid128_fma", mode, args, want, status)

    for fn in ("bid128_scalbn", "bid128_scalbln"):
        for sign in (1, -1):
            for coeff in (0, 1, 9, 10, 11, 20, 100):
                x = decimal.Decimal(f"{'-' if sign < 0 else ''}{coeff}e-6176")
                for shift in (-1, -2):
                    for mode in range(5):
                        exact = EXACT.scaleb(x, shift)
                        want, status = rounded(128, mode, exact,
                                               lambda ctx: ctx.scaleb(x, shift))
                        add(fn, mode, [raw(128, x), str(shift), "32"], want, status | 0x20)
            x = decimal.Decimal(f"{sign * (10**34 - 1)}e6111")
            for mode in range(5):
                exact = EXACT.scaleb(x, 1)
                want, status = rounded(128, mode, exact,
                                       lambda ctx: ctx.scaleb(x, 1))
                add(fn, mode, [raw(128, x), "1", "32"], want, status | 0x20)
            for signaling in (False, True):
                bits = ((1 if sign < 0 else 0) << 127) | ((0x7e if signaling else 0x7c) << 120) | 23
                want = bits & ~(1 << 121)
                for mode in range(5):
                    add(fn, mode, [f"{bits:032x}", "-1", "32"], f"{want:032x}", 0x21 if signaling else 0x20)

    for sign in (0, 0x80000000):
        for coeff in (999998, 999999, 1000000, 1000001, 1000002):
            for outward in (False, True):
                x = sign | coeff
                toward = sign | 0x78000000 if outward else (sign ^ 0x80000000) | 0x78000000
                target = coeff + (1 if outward else -1)
                status = 0x30 if target < 1000000 else 0
                add("bid32_nextafter", 0, [f"{x:08x}", f"{toward:08x}"],
                    f"{sign | target:08x}", status)
        for coeff in (1, 7, 100001, 765432):
            x = sign | (101 << 23) | coeff
            y = sign | (100 << 23) | coeff * 10
            for first, second in ((x, y), (y, x)):
                add("bid32_nextafter", 0, [f"{first:08x}", f"{second:08x}"],
                    f"{first:08x}", 0)
    for x, y, want in ((0x32800000, 0x80000000, 0xb2800000),
                       (0x78001234, 0x78005678, 0x78000000),
                       (0xf8001234, 0xf8005678, 0xf8000000),
                       (0x6cbfffff, 0x00000000, 0x32800000)):
        add("bid32_nextafter", 0, [f"{x:08x}", f"{y:08x}"], f"{want:08x}", 0)

    for biased in range(192):
        for sign in (0, 0x80000000):
            want = f"{(biased << 23) | 1:08x}"
            for bits in ((biased << 23) | 1,
                         0x60000000 | (biased << 21)):
                add("bid32_quantum", 0, [f"{sign | bits:08x}"], want, 0)
    for biased in (0, 1, 63, 64, 100, 101, 127, 128, 190, 191):
        for sign in (0, 0x80000000):
            want = f"{(biased << 23) | 1:08x}"
            for bits in ((biased << 23), (biased << 23) | 0x7fffff,
                         0x60000000 | (biased << 21) | 1,
                         0x60000000 | (biased << 21) | (9999999 - 0x800000),
                         0x60000000 | (biased << 21) | (10000000 - 0x800000),
                         0x60000000 | (biased << 21) | 0x1fffff):
                add("bid32_quantum", 0, [f"{sign | bits:08x}"], want, 0)
    for bits in (0x78000000, 0xf8000000, 0x78001234, 0xf8001234,
                 0x7c000001, 0xfc000001, 0x7e000123, 0xfe000123):
        add("bid32_quantum", 0, [f"{bits:08x}"], f"{bits & 0x7fffffff:08x}", 0)
    rows.extend(parser_carry_rows())
    return rows


C_PROBE = r'''
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "bid_conf.h"
#include "bid_functions.h"
static BID_UINT128 dec128(const char *s) {
  BID_UINT128 x; unsigned long long hi=0,lo=0;
  sscanf(s,"%16llx%16llx",&hi,&lo); x.w[0]=lo; x.w[1]=hi; return x;
}
int main(void) {
  char line[8192];
  while (fgets(line,sizeof line,stdin)) {
    char *save=0,*fn=strtok_r(line,"\t\n",&save),*ms=strtok_r(0,"\t\n",&save);
    char *a=strtok_r(0,"\t\n",&save),*b=strtok_r(0,"\t\n",&save),*c=strtok_r(0,"\t\n",&save);
    if (!fn||!ms||!a) return 2;
    int mode=atoi(ms); unsigned f=0;
    if (!strcmp(fn,"bid32_from_string")) {
      BID_UINT32 v=bid32_from_string(a,mode,&f); printf("%08x\t%02x\n",v,f);
    } else if (!strcmp(fn,"bid64_from_string")) {
      BID_UINT64 v=bid64_from_string(a,mode,&f); printf("%016llx\t%02x\n",(unsigned long long)v,f);
    } else if (!strcmp(fn,"bid128_from_string")) {
      BID_UINT128 v=bid128_from_string(a,mode,&f);
      printf("%016llx%016llx\t%02x\n",(unsigned long long)v.w[1],(unsigned long long)v.w[0],f);
    } else if (!strcmp(fn,"bid128_scalbn") || !strcmp(fn,"bid128_scalbln") || !strcmp(fn,"bid128_ldexp")) {
      BID_UINT128 v;
      if (c) f=strtoul(c,0,10);
      if (!strcmp(fn,"bid128_scalbn")) v=bid128_scalbn(dec128(a),atoi(b),mode,&f);
      else if (!strcmp(fn,"bid128_ldexp")) v=bid128_ldexp(dec128(a),atoi(b),mode,&f);
      else v=bid128_scalbln(dec128(a),atol(b),mode,&f);
      printf("%016llx%016llx\t%02x\n",(unsigned long long)v.w[1],(unsigned long long)v.w[0],f);
    } else if (!strcmp(fn,"bid128_fma")) {
      BID_UINT128 v=bid128_fma(dec128(a),dec128(b),dec128(c),mode,&f);
      printf("%016llx%016llx\t%02x\n",(unsigned long long)v.w[1],(unsigned long long)v.w[0],f);
    } else if (!strcmp(fn,"bid32_nextafter")) {
      BID_UINT32 v=bid32_nextafter(strtoul(a,0,16),strtoul(b,0,16),&f);
      printf("%08x\t%02x\n",v,f);
    } else if (!strcmp(fn,"bid32_quantum")) {
      BID_UINT32 v=bid32_quantum(strtoul(a,0,16),&f);
      printf("%08x\t%02x\n",v,f);
    } else return 3;
  }
  return 0;
}
'''


def capture_native(rows):
    subprocess.run(["bash", str(ROOT / "devtools/scripts/setup_generation_inputs.sh"),
                    "verify-intel"], check=True)
    lib = INTEL / "lib/libbid.a"
    stamp = (INTEL / "lib/.libbid.build-flags").read_text()
    for required in ("CALL_BY_REF=0", "GLOBAL_RND=0", "GLOBAL_FLAGS=0",
                     "UNCHANGED_BINARY_FLAGS=0", "CFLAGS_OPT=-O3 -ffp-contract=off"):
        if required not in stamp:
            raise ValueError(f"native library build stamp lacks {required}")
    with tempfile.TemporaryDirectory(prefix="bid754-numeric-native-") as temp:
        source = Path(temp) / "probe.c"
        binary = Path(temp) / "probe"
        source.write_text(C_PROBE)
        subprocess.run(["cc", "-O2", "-DBID_SIZE_LONG=8",
                        "-DDECIMAL_CALL_BY_REFERENCE=0", "-DDECIMAL_GLOBAL_ROUNDING=0",
                        "-DDECIMAL_GLOBAL_EXCEPTION_FLAGS=0", "-I", str(INTEL / "src"),
                        str(source), str(lib), "-lm", "-o", str(binary)], check=True)
        payload = "".join("\t".join([fn, str(mode), *args]) + "\n"
                          for fn, mode, args, _, _ in rows)
        result = subprocess.run([str(binary)], input=payload, text=True,
                                capture_output=True, check=True)
        observed = [line.split("\t") for line in result.stdout.splitlines()]
        if len(observed) != len(rows):
            raise ValueError(f"native probe returned {len(observed)} of {len(rows)} rows")
        return observed


def key(row):
    fn, mode, args, _, _ = row
    return json.dumps([fn, mode, *args], separators=(",", ":"))


def write_or_check(path, content, check):
    if check:
        if not path.is_file() or path.read_text() != content:
            raise ValueError(f"numeric regression artifact differs: {path.relative_to(ROOT)}")
    else:
        path.write_text(content)


def validate_quantum_row(row):
    fn, mode, args, want, status = row
    if fn != "bid32_quantum":
        return
    if mode != 0 or len(args) != 1 or not re.fullmatch(r"[0-9a-f]{8}", args[0]) or status != 0:
        raise ValueError(f"malformed bid32_quantum regression shape: {row}")
    bits = int(args[0], 16)
    if bits & 0x78000000 == 0x78000000:
        expected = bits & 0x7fffffff
    elif bits & 0x60000000 == 0x60000000:
        expected = (((bits >> 21) & 0xff) << 23) | 1
    else:
        expected = (((bits >> 23) & 0xff) << 23) | 1
    if want != f"{expected:08x}":
        raise ValueError(f"malformed bid32_quantum expected result: {row}")


def emit(rows, observed, check=False):
    manifest_path = ROOT / "devtools/testgen_manifest.json"
    manifest = json.loads(manifest_path.read_text())
    manifest["readtests"] = [b for b in manifest["readtests"]
                             if not b["name"].startswith("numeric_boundary_")
                             and b["name"] != "bid32_from_string_underflow_boundary_cdiverge"]
    blocks = collections.defaultdict(list)
    counts = collections.Counter()
    for row in rows:
        fn, mode, args, want, status = row
        validate_quantum_row(row)
        actual, actual_status = observed[key(row)]
        if not re.fullmatch(r"[0-9a-f]{8,32}", actual) or not re.fullmatch(r"[0-9a-f]{2}", actual_status):
            raise ValueError(f"malformed pinned-C observation: {row}")
        width = int(fn.split("_")[0][3:])
        match = actual_status == f"{status:02x}" and actual == want
        if fn == "bid32_quantum":
            bits = int(args[0], 16)
            steering_finite = bits & 0x60000000 == 0x60000000 and bits & 0x78000000 != 0x78000000
            if match == steering_finite:
                raise ValueError(f"unexpected INTEL-BID-006 C classification: {row}")
        kind = "cmatch" if match else "cdiverge"
        counts[(fn, kind)] += 1
        issue = ""
        if not match:
            if fn.endswith("from_string"):
                digits = args[0].lower().split("e")[0].lstrip("+-").replace(".", "").lstrip("0")
                precision = PARAMS[width][0]
                delayed_carry = (len(digits) > precision + 1 and
                                 digits[:precision] == "9" * precision and
                                 digits[precision] == "0" and
                                 mode == (1 if args[0].startswith("-") else 2))
                if delayed_carry and actual == want and actual_status == "20" and status == 0x30:
                    issue = "carry_underflow"
                elif decimal.Decimal(args[0]).is_zero():
                    issue = "zero_cohort"
                elif fn == "bid32_from_string" and "e" not in args[0].lower():
                    issue = "fixed_point_underflow"
                elif fn == "bid64_from_string" and len(args[0]) > 100:
                    issue = "long_fraction"
                else:
                    issue = "directed_underflow"
            elif fn == "bid32_quantum":
                issue = "quantum_steering"
            elif fn in ("bid128_scalbn", "bid128_scalbln", "bid128_ldexp"):
                issue = "threshold"
            elif status == 0x20 and actual_status == "28":
                issue = "premature_overflow"
            elif actual_status == "00":
                issue = "missing_overflow"
            else:
                issue = "overflow_sign"
        initial = int(args[2]) if fn in ("bid128_scalbn", "bid128_scalbln") and len(args) == 3 else 0
        if initial:
            if not match:
                if actual != want or int(actual_status, 16) != (status | 0x10) or status & 0x10:
                    raise ValueError(f"unexpected INTEL-BID-012 C observation: {row}")
                issue = "sticky_inexact"
            else:
                issue = "sticky_status"
        blocks[(fn, issue, kind, initial)].append(row)
    for (fn, issue, kind, initial), group in sorted(blocks.items()):
        width = int(fn.split("_")[0][3:])
        stem = f"numeric_boundary_{fn}_{issue + '_' if issue else ''}{kind}"
        filename = f"readtest_{stem}.in"
        source = DATA / filename
        oracle = "independent BID successor/cohort raw-bit model" if fn == "bid32_nextafter" else "independent constructed BID exponent model" if fn == "bid32_quantum" else "independent Python decimal arithmetic"
        scope = "Decimal32 optional quantum boundary" if fn == "bid32_quantum" else "IEEE 754 numeric boundary"
        lines = [f"-- {scope}; {oracle}; pinned Intel v20U4 direct C probe {kind}."]
        for _, mode, args, want, status in group:
            if initial:
                args = args[:2]
            operands = [f"[{arg}]" if fn in ("bid128_scalbn", "bid128_scalbln", "bid128_ldexp", "bid128_fma", "bid32_nextafter", "bid32_quantum") and
                        (fn not in ("bid128_scalbn", "bid128_scalbln", "bid128_ldexp") or i == 0) else arg
                        for i, arg in enumerate(args)]
            lines.append(" ".join([fn, str(mode), *operands, f"[{want}]", f"{status:02x}"]))
        write_or_check(source, "\n".join(lines) + "\n", check)
        if fn.endswith("from_string"):
            inputs, kind_name = [f"OP_DEC{width}"], "from_string"
        elif fn in ("bid128_scalbn", "bid128_scalbln", "bid128_ldexp"):
            inputs, kind_name = ["OP_DEC128", "OP_LINT" if fn == "bid128_scalbln" else "OP_INT32"], "binary_op"
        elif fn == "bid128_fma":
            inputs, kind_name = ["OP_DEC128"] * 3, "ternary_op"
        elif fn == "bid32_quantum":
            inputs, kind_name = ["OP_DEC32"], "unary_op"
        else:
            inputs, kind_name = ["OP_DEC32"] * 2, "binary_op"
        block = {"name": stem,
                 "group": f"decimal{width}_ieee754_regressions", "format": f"decimal{width}",
                 "header": "third_party/intel_dfp/TESTS/readtest.h",
                 "source": f"testdata/{filename}", "function": fn, "kind": kind_name,
                 "output_type": f"OP_DEC{width}", "input_types": inputs,
                 "compare_group": "CMP_FUZZYSTATUS"}
        if kind == "cdiverge":
            block["native_compare_skip_reason"] = REASONS[issue]
        if initial:
            block["initial_status"] = initial
        block["statuses"] = sorted({f"{row[4]:02x}" for row in group})
        block["rounding_modes"] = sorted({row[1] for row in group})
        manifest["readtests"].append(block)
    original = manifest_path.read_text()
    first = re.search(r'\n    \{\n      "name": "numeric_boundary_', original)
    close = '\n  ],\n  "readtest_profiles"'
    end = original.index(close)
    prefix = original[:first.start() if first else end].rstrip()
    if prefix.endswith(','):
        prefix = prefix[:-1]
    generated = ',\n' + ',\n'.join(
        '\n'.join('    ' + line for line in json.dumps(block, indent=2).splitlines())
        for block in manifest["readtests"] if block["name"].startswith("numeric_boundary_")
    )
    write_or_check(manifest_path, prefix + generated + original[end:], check)
    for fn in sorted({r[0] for r in rows}):
        print(fn, "cmatch", counts[(fn, "cmatch")], "cdiverge", counts[(fn, "cdiverge")])


def main():
    parser = argparse.ArgumentParser()
    modes = parser.add_mutually_exclusive_group()
    modes.add_argument("--capture-native", action="store_true")
    modes.add_argument("--check", action="store_true")
    args = parser.parse_args()
    verify_archive()
    rows = make_rows()
    if args.capture_native:
        native = capture_native(rows)
        observations = {key(row): result for row, result in zip(rows, native)}
        OBS.write_text(json.dumps({"archive_sha256": ARCHIVE_SHA, "rows": observations},
                                  sort_keys=True, indent=2) + "\n")
    saved = json.loads(OBS.read_text())
    if saved["archive_sha256"] != ARCHIVE_SHA or set(saved["rows"]) != {key(row) for row in rows}:
        raise ValueError("pinned C observations do not cover the independent input batch")
    emit(rows, saved["rows"], check=args.check)


if __name__ == "__main__":
    main()
