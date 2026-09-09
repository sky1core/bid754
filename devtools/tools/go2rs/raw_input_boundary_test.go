package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var rawScaleFunctions = []struct {
	file, goName, rustName string
	width                  int
	flagsPointer           bool
}{
	{"bid32_scalb", "Bid32Scalbn", "bid32_scalbn", 32, false},
	{"bid32_status", "Bid32ScalbnWithFlags", "bid32_scalbn_with_flags", 32, false},
	{"bid32_misc", "Bid32Ldexp", "bid32_ldexp", 32, false},
	{"bid32_status", "Bid32LdexpWithFlags", "bid32_ldexp_with_flags", 32, false},
	{"scalb64", "Bid64Scalbn", "bid64_scalbn", 64, false},
	{"scalb64", "Bid64Ldexp", "bid64_ldexp", 64, false},
	{"bid128_misc", "Bid128Scalbn", "bid128_scalbn", 128, true},
	{"bid128_ldexp", "Bid128Ldexp", "bid128_ldexp", 128, false},
}

func TestI32BoundaryContractRejectsInvalidShapes(t *testing.T) {
	for _, indexes := range [][]int{{-1}, {3}, {0}, {2}, {1, 1}} {
		c := roundingContract{params: []string{"u64", "i64", "i64"}, modes: []int{2}, maxMode: 4, i32Params: indexes}
		if _, err := c.boundary("bid64_scalbn", "x: u64, delta: i64, mode: i64", "(u64, u32)"); err == nil {
			t.Fatalf("accepted invalid i32 parameter indexes %v", indexes)
		}
	}
	c := roundingContracts["bid128_scalbn"]
	for _, params := range []string{
		"x: BID_UINT128, delta: i32, mode: i64, flags: &mut u32",
		"x: BID_UINT128, delta: i64, mode: i64",
		"delta: i64, x: BID_UINT128, mode: i64, flags: &mut u32",
	} {
		if _, err := c.boundary("bid128_scalbn", params, "BID_UINT128"); err == nil {
			t.Fatalf("accepted changed full signature %s", params)
		}
	}
	c = roundingContracts["bid64_scalbn"]
	b, err := c.boundary("bid64_scalbn", "x: u64, delta: i64, mode: i64", "u64")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.wrapper(); err == nil {
		t.Fatal("accepted an i32 boundary without its flags failure channel")
	}
}

func rawScaleBoundaryProbe(t *testing.T) string {
	t.Helper()
	var probe strings.Builder
	probe.WriteString(rawScaleExpectations)
	for _, entry := range rawScaleFunctions {
		input, result := "x as u32", "got as u128"
		typeImport := ""
		if entry.width == 64 {
			input = "x as u64"
		} else if entry.width == 128 {
			input = "BID_UINT128 { lo: x as u64, hi: (x >> 64) as u64 }"
			typeImport = "use bid754::gen_types::BID_UINT128;"
			result = "((got.hi as u128) << 64) | got.lo as u128"
		}
		call := fmt.Sprintf("bid754::generated::%s::%s(%s, n, mode", entry.file, entry.rustName, input)
		if entry.flagsPointer {
			call = "let mut flags = initial; let got = " + call + ", &mut flags);"
		} else {
			call = "let (got, flags) = " + call + ");"
		}
		fmt.Fprintf(&probe, `
#[test]
fn raw_%s_i32_input_contract() {
    %s
    let width = %d;
    let (bias, shift, _, sign, inf, nan, _) = scale_layout(width);
    let one = ((bias as u128) << shift) | 1;
    let zero = (bias as u128) << shift;
    let snan = nan | (2u128 << (width - 8));
    let initials: &[u32] = %s;
    for &initial in initials {
        for mode in 0..=4 {
            for n in [i32::MIN as i64 - 1, i32::MAX as i64 + 1, -(1i64 << 32), 1i64 << 32, i64::MIN, i64::MAX] {
                for x in [one, one | sign, zero, zero | sign, 0, inf, inf | sign, nan, nan | sign, snan, snan | sign, sign - 1, (6u128 << (width - 4)) | 1] {
                    %s
                    assert_eq!((%s, flags), (nan, initial | 1), "reject n={n} mode={mode} x={x:x}");
                }
            }
            for n in [i32::MIN as i64, -1, 0, 1, i32::MAX as i64] {
                for x in [one, one | sign, zero, zero | sign, inf, inf | sign, nan, nan | sign, snan, snan | sign] {
                    let (expected, expected_flags) = scale_expected(width, x, n, mode);
                    %s
                    assert_eq!((%s, flags), (expected, initial | expected_flags), "valid n={n} mode={mode} x={x:x}");
                }
            }
        }
    }
}
`, entry.rustName, typeImport, entry.width, map[bool]string{false: "&[0]", true: "&[0, 0x04, 0x20, 0x80000000, u32::MAX]"}[entry.flagsPointer], call, result, call, result)
	}
	probe.WriteString(rawScaleLongProbe)
	return probe.String()
}

const rawScaleExpectations = `
fn scale_layout(width: u32) -> (i64, u32, i64, u128, u128, u128, u128) {
    let (bias, shift, max_exp, max_finite) = match width {
        32 => (101, 23, 191, 0x77f8967fu128),
        64 => (398, 53, 767, 0x77fb86f26fc0ffffu128),
        128 => (6176, 113, 12287, 0x5fffed09bead87c0378d8e63ffffffffu128),
        _ => unreachable!(),
    };
    (bias, shift, max_exp, 1u128 << (width - 1), 0x78u128 << (width - 8), 0x7cu128 << (width - 8), max_finite)
}

fn scale_expected(width: u32, x: u128, n: i64, mode: i64) -> (u128, u32) {
    let (bias, shift, max_exp, sign_bit, inf, nan, max_finite) = scale_layout(width);
    let sign = x & sign_bit;
    let magnitude = x & !sign_bit;
    if magnitude >= nan {
        return (nan | sign, if magnitude & (2u128 << (width - 8)) != 0 { 1 } else { 0 });
    }
    if magnitude == inf { return (x, 0); }
    if magnitude == (bias as u128) << shift {
        return (sign | ((bias + n).clamp(0, max_exp) as u128) << shift, 0);
    }
    if n == i32::MIN as i64 {
        let away = (sign == 0 && mode == 2) || (sign != 0 && mode == 1);
        return (sign | if away { 1 } else { 0 }, 0x30);
    }
    if n == i32::MAX as i64 {
        let finite = mode == 3 || (sign == 0 && mode == 1) || (sign != 0 && mode == 2);
        return (sign | if finite { max_finite } else { inf }, 0x28);
    }
    (sign | (((bias + n) as u128) << shift) | 1, 0)
}
`

const rawScaleLongProbe = `
#[test]
fn long_scale_and_public_scaleb_keep_i64_domain() {
    use bid754::{Decimal32, Decimal64, Decimal128, RoundingMode, ExceptionFlags};
    use bid754::generated::prelude::{BID_UINT128, bid32_scalbln, bid32_scalbln_with_flags, bid64_scalbln, bid128_scalbln};
    let one32 = 0x32800001u32;
    let one64 = 0x31c0000000000001u64;
    let one128 = BID_UINT128 { lo: 1, hi: 0x3040000000000000 };
    for n in [1i64 << 32, i64::MAX, -(1i64 << 32), i64::MIN] {
        for (mode, public_mode) in [(0, RoundingMode::NearestEven), (1, RoundingMode::TowardNegative), (2, RoundingMode::TowardPositive), (3, RoundingMode::TowardZero), (4, RoundingMode::NearestAway)] {
            let bounded = if n > 0 { i32::MAX as i64 } else { i32::MIN as i64 };
            let (expected32, flags32) = scale_expected(32, one32 as u128, bounded, mode);
            let (expected64, flags64) = scale_expected(64, one64 as u128, bounded, mode);
            let (expected128, flags128) = scale_expected(128, 0x30400000000000000000000000000001, bounded, mode);
            assert_eq!(bid32_scalbln(one32, n, mode), (expected32 as u32, flags32));
            assert_eq!(bid32_scalbln_with_flags(one32, n, mode), (expected32 as u32, flags32));
            assert_eq!(bid64_scalbln(one64, n, mode), (expected64 as u64, flags64));
            let mut flags = 0x04;
            let got = bid128_scalbln(one128, n, mode, &mut flags);
            assert_eq!((((got.hi as u128) << 64) | got.lo as u128, flags), (expected128, 0x04 | flags128));
            let expected_flags = ExceptionFlags::INEXACT | if n > 0 { ExceptionFlags::OVERFLOW } else { ExceptionFlags::UNDERFLOW };
            let (got, flags) = Decimal32::ONE.scaleb_with_mode(n, public_mode);
            assert_eq!((got.to_bits(), flags), (expected32 as u32, expected_flags));
            let (got, flags) = Decimal64::ONE.scaleb_with_mode(n, public_mode);
            assert_eq!((got.to_bits(), flags), (expected64 as u64, expected_flags));
            let (got, flags) = Decimal128::ONE.scaleb_with_mode(n, public_mode);
            assert_eq!((u128::from_le_bytes(got.to_le_bytes()), flags), (expected128, expected_flags));
            if mode == 0 {
                let (got, flags) = Decimal32::ONE.scaleb(n);
                assert_eq!((got.to_bits(), flags), (expected32 as u32, expected_flags));
                let (got, flags) = Decimal64::ONE.scaleb(n);
                assert_eq!((got.to_bits(), flags), (expected64 as u64, expected_flags));
                let (got, flags) = Decimal128::ONE.scaleb(n);
                assert_eq!((u128::from_le_bytes(got.to_le_bytes()), flags), (expected128, expected_flags));
            }
        }
    }
}
`

func checkInternalHelperPrivacy(t *testing.T, root, probePath string, run func(...string) ([]byte, error)) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "bid754-rs/src/generated/prelude.rs"))
	if err != nil {
		t.Fatal(err)
	}
	functions := regexp.MustCompile(`(?m)^pub(?:\(crate\))? fn (go_\w+)`).FindAllStringSubmatch(string(data), -1)
	if len(functions) != 31 {
		t.Fatalf("prelude helper census changed: got %d, want 31", len(functions))
	}
	var probe strings.Builder
	for i, fn := range functions {
		fmt.Fprintf(&probe, "use bid754::generated::prelude::%s as private_helper_%d;\n", fn[1], i)
	}
	probe.WriteString("use bid754::generated::inline_round64::bid_normalize as direct_normalize;\nuse bid754::generated::prelude::bid_normalize as prelude_normalize;\nfn main() {}\n")
	if err := os.WriteFile(probePath, []byte(probe.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := run("check", "--test", "rounding_boundary")
	if err == nil || strings.Count(string(output), "error[E0603]") != len(functions)+2 || strings.Count(string(output), "\nerror[") != len(functions)+2 {
		t.Fatalf("expected %d private helper errors: %v\n%s", len(functions)+2, err, output)
	}
	t.Logf("%d prelude helpers and both normalize access paths rejected by Rust privacy", len(functions))
}

func checkPublicCInt32Inventory(t *testing.T, repo string, paths []string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repo, "devtools/generated/json/intel_dfp_symbols.json"))
	if err != nil {
		t.Fatal(err)
	}
	var inventory struct {
		Symbols []struct {
			Name       string   `json:"name"`
			Parameters []string `json:"parameters"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(data, &inventory); err != nil {
		t.Fatal(err)
	}
	cParams := make(map[string][]string)
	for _, symbol := range inventory.Symbols {
		cParams[strings.TrimPrefix(symbol.Name, "__")] = symbol.Parameters
	}
	widened, typed := 0, 0
	seen := make(map[string]bool)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, sig := range rustPortSignature.FindAllStringSubmatch(string(data), -1) {
			if sig[1] != "pub " {
				continue
			}
			name := sig[2]
			params, ok := cParams[strings.TrimSuffix(name, "_with_flags")]
			if !ok {
				continue
			}
			fields := strings.Split(sig[3], ", ")
			for i, param := range params {
				if !regexp.MustCompile(`^int\s+\w+$`).MatchString(param) {
					continue
				}
				if i >= len(fields) {
					t.Fatalf("public %s lost C int parameter %d", name, i)
				}
				_, typ, ok := strings.Cut(fields[i], ": ")
				if !ok {
					t.Fatalf("unrecognized Rust parameter %s", fields[i])
				}
				if typ == "i32" {
					typed++
					continue
				}
				if typ != "i64" {
					t.Fatalf("public %s C int parameter %d has unexpected type %s", name, i, typ)
				}
				c, ok := roundingContracts[name]
				if !ok {
					t.Fatalf("public %s widens C int parameter %d without a boundary contract", name, i)
				}
				if _, err := c.boundary(name, sig[3], sig[4]); err != nil {
					t.Fatal(err)
				}
				guarded := false
				for _, index := range c.i32Params {
					guarded = guarded || index == i
				}
				if !guarded {
					t.Fatalf("public %s C int parameter %d lacks its i32 guard", name, i)
				}
				widened++
				seen[name] = true
			}
		}
	}
	if widened != 8 || typed != 3 {
		t.Fatalf("public signed C int census: widened=%d typed=%d; want 8 and 3", widened, typed)
	}
	for name, c := range roundingContracts {
		if len(c.i32Params) != 0 && !seen[name] {
			t.Fatalf("i32 contract %s has no public C int predecessor", name)
		}
	}
	t.Logf("C signature inventory: %d guarded widened inputs, %d typed i32 inputs", widened, typed)
}
