"""Standalone BID codec benchmark leg (Python, stdlib-only).

Operands come from the hand-pinned shared contract
``bid754-codec-go/testdata/codec_benchmark_operands.json``, consumed by all
four codec benchmark legs (Go ``testing.B``, Rust criterion, the JS script,
and this script). Setup re-checks the contract's exactness round trips
(``encode(decode(bits)) == bits``, ``to_string(decode(bits)) ==
decimal_string``, ``encode(from_string(string)) == bits``) and raises on any
violation, so the run rejects invalid or inexact operands instead of timing
them. Every row rotates operands with an ``i % n`` counter and feeds results
into an observable sink printed at the end, matching the other legs.
Interpreted-leg caveat: each timed iteration pays a fixed harness dispatch
cost (one callable call plus one sink call), uniform across rows, so
within-leg comparisons are sound but absolute ns/op includes that constant;
prefer the compiled legs for cross-language absolute comparisons.
Benchmark infrastructure, not a regular verification domain.

Run from the ``bid754-codec-py`` package directory:
``python3 benchmarks/bench_runner.py [samples]``.
"""

from __future__ import annotations

import json
import re
import sys
import time
from dataclasses import dataclass
from pathlib import Path

_PACKAGE_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(_PACKAGE_ROOT))

from bid_codec import (  # noqa: E402
    Components,
    Kind,
    decode32,
    decode64,
    decode128,
    encode32,
    encode64,
    encode128,
    from_string,
    to_string,
)

_OPERAND_PATH = (
    _PACKAGE_ROOT.parent / "bid754-codec-go" / "testdata" / "codec_benchmark_operands.json"
)

_HEX_RE = re.compile(r"^[0-9a-f]+$")


def _parse_hex_word(width: str, index: int, field: str, hex_text: object, bits: int) -> int:
    digits = bits // 4
    if (
        not isinstance(hex_text, str)
        or len(hex_text) != digits
        or not _HEX_RE.fullmatch(hex_text)
    ):
        raise ValueError(
            f"{width}[{index}].{field}: hex {hex_text!r} must be exactly {digits} lowercase hex digits"
        )
    return int(hex_text, 16)


@dataclass
class _WordOperand:
    """Single-word operand (decimal32 and decimal64)."""

    bits: int
    comp: Components
    string: str


@dataclass
class _Operand128:
    lo: int
    hi: int
    comp: Components
    string: str


def _load_operands() -> tuple[list[_WordOperand], list[_WordOperand], list[_Operand128]]:
    file = json.loads(_OPERAND_PATH.read_text(encoding="utf-8"))
    if file["format_version"] != 1:
        raise ValueError(f"benchmark operand format_version {file['format_version']}, want 1")
    for width in ("decimal32", "decimal64", "decimal128"):
        if not file.get(width):
            raise ValueError(f"benchmark operand contract requires non-empty {width} entries")

    operands32: list[_WordOperand] = []
    for i, entry in enumerate(file["decimal32"]):
        if entry.get("hex_hi") is not None:
            raise ValueError(f"decimal32[{i}]: hex_hi is only valid for decimal128 entries")
        bits = _parse_hex_word("decimal32", i, "hex", entry["hex"], 32)
        comp = decode32(bits)
        if comp.kind is not Kind.NORMAL:
            raise ValueError(f"decimal32[{i}] must decode NORMAL")
        if encode32(comp) != bits:
            raise ValueError(f"decimal32[{i}] operand is not canonical")
        if to_string(comp) != entry["decimal_string"]:
            raise ValueError(f"decimal32[{i}] canonical string mismatch")
        if encode32(from_string(entry["decimal_string"])) != bits:
            raise ValueError(f"decimal32[{i}] string operand is not exactly representable")
        operands32.append(_WordOperand(bits=bits, comp=comp, string=entry["decimal_string"]))

    operands64: list[_WordOperand] = []
    for i, entry in enumerate(file["decimal64"]):
        if entry.get("hex_hi") is not None:
            raise ValueError(f"decimal64[{i}]: hex_hi is only valid for decimal128 entries")
        bits = _parse_hex_word("decimal64", i, "hex", entry["hex"], 64)
        comp = decode64(bits)
        if comp.kind is not Kind.NORMAL:
            raise ValueError(f"decimal64[{i}] must decode NORMAL")
        if encode64(comp) != bits:
            raise ValueError(f"decimal64[{i}] operand is not canonical")
        if to_string(comp) != entry["decimal_string"]:
            raise ValueError(f"decimal64[{i}] canonical string mismatch")
        if encode64(from_string(entry["decimal_string"])) != bits:
            raise ValueError(f"decimal64[{i}] string operand is not exactly representable")
        operands64.append(_WordOperand(bits=bits, comp=comp, string=entry["decimal_string"]))

    operands128: list[_Operand128] = []
    for i, entry in enumerate(file["decimal128"]):
        lo = _parse_hex_word("decimal128", i, "hex", entry["hex"], 64)
        hi = _parse_hex_word("decimal128", i, "hex_hi", entry.get("hex_hi"), 64)
        comp = decode128(lo, hi)
        if comp.kind is not Kind.NORMAL:
            raise ValueError(f"decimal128[{i}] must decode NORMAL")
        if encode128(comp) != (lo, hi):
            raise ValueError(f"decimal128[{i}] operand is not canonical")
        if to_string(comp) != entry["decimal_string"]:
            raise ValueError(f"decimal128[{i}] canonical string mismatch")
        if encode128(from_string(entry["decimal_string"])) != (lo, hi):
            raise ValueError(f"decimal128[{i}] string operand is not exactly representable")
        operands128.append(_Operand128(lo=lo, hi=hi, comp=comp, string=entry["decimal_string"]))

    return operands32, operands64, operands128


_SINK = 0


def _sink_components(comp: Components) -> None:
    global _SINK
    _SINK ^= comp.coefficient ^ comp.payload ^ comp.exponent ^ comp.kind.value ^ int(comp.sign)


def _sink_string(s: str) -> None:
    global _SINK
    _SINK ^= len(s)


def _sink_int(v: int) -> None:
    global _SINK
    _SINK ^= v


_TARGET_SAMPLE_NS = 100_000_000  # ~100ms per timed sample
_CALIBRATION_ITERS = 2_000


def _run_batch(fn, iters: int) -> int:
    start = time.perf_counter_ns()
    for i in range(iters):
        fn(i)
    return time.perf_counter_ns() - start


def _bench_row(name: str, samples: int, fn) -> None:
    calibrated = max(_run_batch(fn, _CALIBRATION_ITERS), 1)
    iters = (_TARGET_SAMPLE_NS * _CALIBRATION_ITERS) // calibrated
    iters = max(1_000, min(5_000_000, iters))
    measured: list[float] = []
    for _ in range(samples):
        elapsed = _run_batch(fn, iters)
        measured.append(elapsed / iters)
    median = sorted(measured)[len(measured) // 2]
    rendered = ",".join(f"{v:.1f}" for v in measured)
    print(f"BENCH {name} ns_op_median={median:.1f} iters={iters} samples_ns_op=[{rendered}]")


def main() -> None:
    samples = int(sys.argv[1]) if len(sys.argv) > 1 else 5
    operands32, operands64, operands128 = _load_operands()
    n32, n64, n128 = len(operands32), len(operands64), len(operands128)
    print(
        f"BENCH-LEG codec-py python={sys.version_info.major}.{sys.version_info.minor}."
        f"{sys.version_info.micro} samples={samples} operands={n32}/{n64}/{n128}"
    )

    _bench_row(
        "codec_bid32/decode", samples, lambda i: _sink_components(decode32(operands32[i % n32].bits))
    )
    _bench_row(
        "codec_bid32/encode", samples, lambda i: _sink_int(encode32(operands32[i % n32].comp))
    )
    _bench_row(
        "codec_bid32/to_string", samples, lambda i: _sink_string(to_string(operands32[i % n32].comp))
    )
    _bench_row(
        "codec_bid32/from_string",
        samples,
        lambda i: _sink_components(from_string(operands32[i % n32].string)),
    )

    _bench_row(
        "codec_bid64/decode", samples, lambda i: _sink_components(decode64(operands64[i % n64].bits))
    )
    _bench_row(
        "codec_bid64/encode", samples, lambda i: _sink_int(encode64(operands64[i % n64].comp))
    )
    _bench_row(
        "codec_bid64/to_string", samples, lambda i: _sink_string(to_string(operands64[i % n64].comp))
    )
    _bench_row(
        "codec_bid64/from_string",
        samples,
        lambda i: _sink_components(from_string(operands64[i % n64].string)),
    )

    _bench_row(
        "codec_bid128/decode",
        samples,
        lambda i: _sink_components(decode128(operands128[i % n128].lo, operands128[i % n128].hi)),
    )

    def _encode128_row(i: int) -> None:
        lo, hi = encode128(operands128[i % n128].comp)
        _sink_int(lo ^ hi)

    _bench_row("codec_bid128/encode", samples, _encode128_row)
    _bench_row(
        "codec_bid128/to_string",
        samples,
        lambda i: _sink_string(to_string(operands128[i % n128].comp)),
    )
    _bench_row(
        "codec_bid128/from_string",
        samples,
        lambda i: _sink_components(from_string(operands128[i % n128].string)),
    )

    print(f"BENCH-SINK {_SINK}")


if __name__ == "__main__":
    main()
