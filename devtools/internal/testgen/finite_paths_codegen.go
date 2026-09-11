package testgen

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

const (
	finitePathsGoPath   = "../bid754-go/generated_finite_public_paths_test.go"
	finitePathsRustPath = "../bid754-rs/examples/finite_probe.rs"
)

func GenerateFinitePathsOutputs() (map[string][]byte, error) {
	goSrc := genmarker.Line("testgen") + "\n\n" + finitePathsGoBody
	formatted, err := format.Source([]byte(goSrc))
	if err != nil {
		return nil, fmt.Errorf("format finite public paths Go output: %w", err)
	}
	rustSrc := genmarker.Line("testgen") + "\n" + finitePathsRustBody
	outputs, err := GenerateBigDecimalOutputs()
	if err != nil {
		return nil, err
	}
	outputs[finitePathsGoPath] = formatted
	outputs[finitePathsRustPath] = []byte(rustSrc)
	return outputs, nil
}

func WriteFinitePathsOutputs(repoRoot string) error {
	files, err := GenerateFinitePathsOutputs()
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated finite path output %q: %w", fullPath, err)
		}
	}
	return nil
}

const finitePathsGoBody = `package bid754

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"
	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

func finiteMapPublicFlags(flags ExceptionFlags) (uint32, error) {
	raw := uint64(flags)
	const publicIEEE = uint64(0x1f)
	if raw&^publicIEEE != 0 {
		return 0, fmt.Errorf("public flags %#x carry bits outside the IEEE arithmetic set", raw)
	}
	var native uint32
	if raw&0x01 != 0 {
		native |= 0x20
	}
	if raw&0x02 != 0 {
		native |= 0x10
	}
	if raw&0x04 != 0 {
		native |= 0x08
	}
	if raw&0x08 != 0 {
		native |= 0x04
	}
	if raw&0x10 != 0 {
		native |= 0x01
	}
	return native, nil
}

func finitePublicMode(mode string) (RoundingMode, bool) {
	switch mode {
	case "nearest_even":
		return RoundNearestEven, true
	case "nearest_away":
		return RoundNearestAway, true
	case "toward_zero":
		return RoundTowardZero, true
	case "toward_positive":
		return RoundTowardPositive, true
	case "toward_negative":
		return RoundTowardNegative, true
	default:
		return 0, false
	}
}

func finitePortRounding(mode string) (int, bool) {
	switch mode {
	case "nearest_even":
		return 0, true
	case "toward_negative":
		return 1, true
	case "toward_positive":
		return 2, true
	case "toward_zero":
		return 3, true
	case "nearest_away":
		return 4, true
	default:
		return 0, false
	}
}

func finiteHex32(u uint32) string  { return fmt.Sprintf("%08x", u) }
func finiteHex64(u uint64) string  { return fmt.Sprintf("%016x", u) }
func finiteHex128(hi, lo uint64) string {
	return fmt.Sprintf("%016x:%016x", hi, lo)
}

func finiteDec128Hex(v Decimal128BID) string {
	b := v.ToBytes()
	return finiteHex128(binary.LittleEndian.Uint64(b[8:]), binary.LittleEndian.Uint64(b[:8]))
}

type finiteOperand128 struct {
	Dec Decimal128BID
	Hi  uint64
	Lo  uint64
}

func (o finiteOperand128) bidgo() bidgo.BID_UINT128 {
	return bidgo.Bid128FromWords(o.Hi, o.Lo)
}

func finiteParse32(operands []string) ([]Decimal32BID, error) {
	out := make([]Decimal32BID, 0, len(operands))
	for _, raw := range operands {
		x, err := strconv.ParseUint(raw, 16, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid decimal32 raw image %q: %w", raw, err)
		}
		out = append(out, Decimal32BIDFromBits(uint32(x)))
	}
	return out, nil
}

func finiteParse64(operands []string) ([]Decimal64BID, error) {
	out := make([]Decimal64BID, 0, len(operands))
	for _, raw := range operands {
		x, err := strconv.ParseUint(raw, 16, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid decimal64 raw image %q: %w", raw, err)
		}
		out = append(out, Decimal64BIDFromBits(x))
	}
	return out, nil
}

func finiteParse128(operands []string) ([]finiteOperand128, error) {
	out := make([]finiteOperand128, 0, len(operands))
	for _, raw := range operands {
		parts := strings.Split(raw, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid decimal128 raw image %q", raw)
		}
		hi, err := strconv.ParseUint(parts[0], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid decimal128 hi in %q: %w", raw, err)
		}
		lo, err := strconv.ParseUint(parts[1], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid decimal128 lo in %q: %w", raw, err)
		}
		var b [16]byte
		binary.LittleEndian.PutUint64(b[:8], lo)
		binary.LittleEndian.PutUint64(b[8:], hi)
		out = append(out, finiteOperand128{Dec: Decimal128BIDFromBytes(b), Hi: hi, Lo: lo})
	}
	return out, nil
}

func finitePublicPaths(c decimalref.Case) ([]finiteObservation, error) {
	if c.Width != 32 && c.Width != 64 && c.Width != 128 { return nil, fmt.Errorf("unsupported decimal width %d", c.Width) }
	for _, raw := range c.Operands {
		length := c.Width/4
		if c.Width == 128 { length++ }
		if len(raw) != length { return nil, fmt.Errorf("raw image must have %d characters", length) }
		for i := range raw {
			if c.Width == 128 && i == 16 {
				if raw[i] != ':' { return nil, fmt.Errorf("raw image requires hi:lo separator") }
			} else if !((raw[i] >= '0' && raw[i] <= '9') || (raw[i] >= 'a' && raw[i] <= 'f')) { return nil, fmt.Errorf("raw image must be padded lowercase hex") }
		}
	}
	mode, ok := finitePublicMode(c.Mode)
	if !ok {
		return nil, fmt.Errorf("unsupported rounding mode %q", c.Mode)
	}
	portRnd, ok := finitePortRounding(c.Mode)
	if !ok {
		return nil, fmt.Errorf("unsupported rounding mode %q", c.Mode)
	}
	nearestEven := c.Mode == "nearest_even"
	switch c.Width {
	case 32:
		return finitePublicPaths32(c, mode, portRnd, nearestEven)
	case 64:
		return finitePublicPaths64(c, mode, portRnd, nearestEven)
	case 128:
		return finitePublicPaths128(c, mode, portRnd, nearestEven)
	default:
		return nil, fmt.Errorf("unsupported decimal width %d", c.Width)
	}
}

func finiteWantOperands(op string) (int, error) {
	switch op {
	case "add", "sub", "mul", "div", "quantize":
		return 2, nil
	case "fma":
		return 3, nil
	default:
		return 0, fmt.Errorf("unsupported operation %q", op)
	}
}

func finitePublicPaths32(c decimalref.Case, mode RoundingMode, portRnd int, nearestEven bool) ([]finiteObservation, error) {
	want, err := finiteWantOperands(c.Op)
	if err != nil {
		return nil, err
	}
	if len(c.Operands) != want {
		return nil, fmt.Errorf("op %q requires %d operands, got %d", c.Op, want, len(c.Operands))
	}
	ops, err := finiteParse32(c.Operands)
	if err != nil {
		return nil, err
	}
	x, y := ops[0], ops[1]
	var obs []finiteObservation
	mapped := func(path string, v Decimal32BID, f ExceptionFlags) error {
		native, err := finiteMapPublicFlags(f)
		if err != nil {
			return err
		}
		obs = append(obs, finiteObservation{Path: path, Bits: finiteHex32(v.ToUint32()), Flags: native, HasFlags: true})
		return nil
	}
	port := func(bits, native uint32) {
		obs = append(obs, finiteObservation{Path: "go/port", Bits: finiteHex32(bits), Flags: native, HasFlags: true})
	}
	value := func(v Decimal32BID) {
		obs = append(obs, finiteObservation{Path: "go/public/value", Bits: finiteHex32(v.ToUint32()), HasFlags: false})
	}

	switch c.Op {
	case "add":
		mv, mf := x.AddWithMode(y, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		pb, pf := bidgo.Bid32AddWithFlags(x.ToUint32(), y.ToUint32(), portRnd)
		port(pb, pf)
		if nearestEven {
			value(x.Add(y))
			fv, ff := x.AddWithFlags(y)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "sub":
		mv, mf := x.SubWithMode(y, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		pb, pf := bidgo.Bid32SubWithFlags(x.ToUint32(), y.ToUint32(), portRnd)
		port(pb, pf)
		if nearestEven {
			value(x.Sub(y))
			fv, ff := x.SubWithFlags(y)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "mul":
		mv, mf := x.MulWithMode(y, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		pb, pf := bidgo.Bid32MulWithFlags(x.ToUint32(), y.ToUint32(), portRnd)
		port(pb, pf)
		if nearestEven {
			value(x.Mul(y))
			fv, ff := x.MulWithFlags(y)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "div":
		mv, mf := x.DivWithMode(y, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		pb, pf := bidgo.Bid32DivWithFlags(x.ToUint32(), y.ToUint32(), portRnd)
		port(pb, pf)
		if nearestEven {
			value(x.Div(y))
			fv, ff := x.DivWithFlags(y)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "quantize":
		mv, mf := x.QuantizeWithMode(y, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		pb, pf := bidgo.Bid32Quantize(x.ToUint32(), y.ToUint32(), portRnd)
		port(pb, pf)
		if nearestEven {
			value(x.Quantize(y))
			fv, ff := x.QuantizeWithFlags(y)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "fma":
		z := ops[2]
		mv, mf := x.FMAWithMode(y, z, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		pb, pf := bidgo.Bid32Fma(x.ToUint32(), y.ToUint32(), z.ToUint32(), portRnd)
		port(pb, pf)
		if nearestEven {
			fv, ff := x.FMA(y, z)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	}
	return obs, nil
}

func finitePublicPaths64(c decimalref.Case, mode RoundingMode, portRnd int, nearestEven bool) ([]finiteObservation, error) {
	want, err := finiteWantOperands(c.Op)
	if err != nil {
		return nil, err
	}
	if len(c.Operands) != want {
		return nil, fmt.Errorf("op %q requires %d operands, got %d", c.Op, want, len(c.Operands))
	}
	ops, err := finiteParse64(c.Operands)
	if err != nil {
		return nil, err
	}
	x, y := ops[0], ops[1]
	var obs []finiteObservation
	mapped := func(path string, v Decimal64BID, f ExceptionFlags) error {
		native, err := finiteMapPublicFlags(f)
		if err != nil {
			return err
		}
		obs = append(obs, finiteObservation{Path: path, Bits: finiteHex64(v.ToUint64()), Flags: native, HasFlags: true})
		return nil
	}
	port := func(bits uint64, native uint32) {
		obs = append(obs, finiteObservation{Path: "go/port", Bits: finiteHex64(bits), Flags: native, HasFlags: true})
	}
	value := func(v Decimal64BID) {
		obs = append(obs, finiteObservation{Path: "go/public/value", Bits: finiteHex64(v.ToUint64()), HasFlags: false})
	}

	switch c.Op {
	case "add":
		mv, mf := x.AddWithMode(y, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		pb, pf := bidgo.Bid64AddWithFlags(x.ToUint64(), y.ToUint64(), portRnd)
		port(pb, pf)
		if nearestEven {
			value(x.Add(y))
			fv, ff := x.AddWithFlags(y)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "sub":
		mv, mf := x.SubWithMode(y, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		pb, pf := bidgo.Bid64SubWithFlags(x.ToUint64(), y.ToUint64(), portRnd)
		port(pb, pf)
		if nearestEven {
			value(x.Sub(y))
			fv, ff := x.SubWithFlags(y)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "mul":
		mv, mf := x.MulWithMode(y, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		pb, pf := bidgo.Bid64MulWithFlags(x.ToUint64(), y.ToUint64(), portRnd)
		port(pb, pf)
		if nearestEven {
			value(x.Mul(y))
			fv, ff := x.MulWithFlags(y)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "div":
		mv, mf := x.DivWithMode(y, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		pb, pf := bidgo.Bid64DivWithFlags(x.ToUint64(), y.ToUint64(), portRnd)
		port(pb, pf)
		if nearestEven {
			value(x.Div(y))
			fv, ff := x.DivWithFlags(y)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "quantize":
		mv, mf := x.QuantizeWithMode(y, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		pb, pf := bidgo.Bid64Quantize(x.ToUint64(), y.ToUint64(), portRnd)
		port(pb, pf)
		if nearestEven {
			value(x.Quantize(y))
			fv, ff := x.QuantizeWithFlags(y)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "fma":
		z := ops[2]
		mv, mf := x.FMAWithMode(y, z, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		pb, pf := bidgo.Bid64Fma(x.ToUint64(), y.ToUint64(), z.ToUint64(), portRnd)
		port(pb, pf)
		if nearestEven {
			fv, ff := x.FMA(y, z)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	}
	return obs, nil
}

func finitePublicPaths128(c decimalref.Case, mode RoundingMode, portRnd int, nearestEven bool) ([]finiteObservation, error) {
	want, err := finiteWantOperands(c.Op)
	if err != nil {
		return nil, err
	}
	if len(c.Operands) != want {
		return nil, fmt.Errorf("op %q requires %d operands, got %d", c.Op, want, len(c.Operands))
	}
	ops, err := finiteParse128(c.Operands)
	if err != nil {
		return nil, err
	}
	x, y := ops[0], ops[1]
	var obs []finiteObservation
	mapped := func(path string, v Decimal128BID, f ExceptionFlags) error {
		native, err := finiteMapPublicFlags(f)
		if err != nil {
			return err
		}
		obs = append(obs, finiteObservation{Path: path, Bits: finiteDec128Hex(v), Flags: native, HasFlags: true})
		return nil
	}
	port := func(r bidgo.BID_UINT128, native uint32) {
		hi, lo := bidgo.Bid128Words(r)
		obs = append(obs, finiteObservation{Path: "go/port", Bits: finiteHex128(hi, lo), Flags: native, HasFlags: true})
	}
	value := func(v Decimal128BID) {
		obs = append(obs, finiteObservation{Path: "go/public/value", Bits: finiteDec128Hex(v), HasFlags: false})
	}

	switch c.Op {
	case "add":
		mv, mf := x.Dec.AddWithMode(y.Dec, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		var pf uint32
		r := bidgo.Bid128Add(x.bidgo(), y.bidgo(), portRnd, &pf)
		port(r, pf)
		if nearestEven {
			value(x.Dec.Add(y.Dec))
			fv, ff := x.Dec.AddWithFlags(y.Dec)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "sub":
		mv, mf := x.Dec.SubWithMode(y.Dec, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		var pf uint32
		r := bidgo.Bid128Sub(x.bidgo(), y.bidgo(), portRnd, &pf)
		port(r, pf)
		if nearestEven {
			value(x.Dec.Sub(y.Dec))
			fv, ff := x.Dec.SubWithFlags(y.Dec)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "mul":
		mv, mf := x.Dec.MulWithMode(y.Dec, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		r, pf := bidgo.Bid128Mul(x.bidgo(), y.bidgo(), portRnd)
		port(r, pf)
		if nearestEven {
			value(x.Dec.Mul(y.Dec))
			fv, ff := x.Dec.MulWithFlags(y.Dec)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "div":
		mv, mf := x.Dec.DivWithMode(y.Dec, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		r, pf := bidgo.Bid128Div(x.bidgo(), y.bidgo(), portRnd)
		port(r, pf)
		if nearestEven {
			value(x.Dec.Div(y.Dec))
			fv, ff := x.Dec.DivWithFlags(y.Dec)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "quantize":
		mv, mf := x.Dec.QuantizeWithMode(y.Dec, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		r, pf := bidgo.Bid128Quantize(x.bidgo(), y.bidgo(), portRnd)
		port(r, pf)
		if nearestEven {
			value(x.Dec.Quantize(y.Dec))
			fv, ff := x.Dec.QuantizeWithFlags(y.Dec)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	case "fma":
		z := ops[2]
		mv, mf := x.Dec.FMAWithMode(y.Dec, z.Dec, mode)
		if err := mapped("go/public/mode", mv, mf); err != nil {
			return nil, err
		}
		r, pf := bidgo.Bid128Fma(x.bidgo(), y.bidgo(), z.bidgo(), portRnd)
		port(r, pf)
		if nearestEven {
			fv, ff := x.Dec.FMA(y.Dec, z.Dec)
			if err := mapped("go/public/flags", fv, ff); err != nil {
				return nil, err
			}
		}
	}
	return obs, nil
}
`

const finitePathsRustBody = `

use std::io::{self, BufRead, Write};

use bid754::{Decimal128, Decimal32, Decimal64, ExceptionFlags, RoundingMode};
use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct Case {
    width: u32,
    op: String,
    mode: String,
    operands: Vec<String>,
}

#[derive(Serialize)]
struct Observation {
    path: String,
    bits: String,
    flags: u32,
    has_flags: bool,
}

fn public_mode(mode: &str) -> Option<RoundingMode> {
    Some(match mode {
        "nearest_even" => RoundingMode::NearestEven,
        "nearest_away" => RoundingMode::NearestAway,
        "toward_zero" => RoundingMode::TowardZero,
        "toward_positive" => RoundingMode::TowardPositive,
        "toward_negative" => RoundingMode::TowardNegative,
        _ => return None,
    })
}

fn port_rounding(mode: &str) -> Option<i64> {
    Some(match mode {
        "nearest_even" => 0,
        "toward_negative" => 1,
        "toward_positive" => 2,
        "toward_zero" => 3,
        "nearest_away" => 4,
        _ => return None,
    })
}

fn map_public_flags(f: ExceptionFlags) -> Result<u32, String> {
    let raw = f.bits();
    if raw & !0x1f != 0 {
        return Err(format!("public flags {raw:#x} carry bits outside the IEEE arithmetic set"));
    }
    let mut native = 0u32;
    if raw & 0x01 != 0 {
        native |= 0x20;
    }
    if raw & 0x02 != 0 {
        native |= 0x10;
    }
    if raw & 0x04 != 0 {
        native |= 0x08;
    }
    if raw & 0x08 != 0 {
        native |= 0x04;
    }
    if raw & 0x10 != 0 {
        native |= 0x01;
    }
    Ok(native)
}

fn hex32(v: u32) -> String {
    format!("{v:08x}")
}

fn hex64(v: u64) -> String {
    format!("{v:016x}")
}

fn hex128(hi: u64, lo: u64) -> String {
    format!("{hi:016x}:{lo:016x}")
}

fn dec128_hex(v: Decimal128) -> String {
    let b = v.to_le_bytes();
    let mut lo = [0u8; 8];
    let mut hi = [0u8; 8];
    lo.copy_from_slice(&b[0..8]);
    hi.copy_from_slice(&b[8..16]);
    hex128(u64::from_le_bytes(hi), u64::from_le_bytes(lo))
}

fn bid128(hi: u64, lo: u64) -> bid754::gen_types::BID_UINT128 {
    bid754::gen_types::BID_UINT128 { lo, hi }
}

fn push_value(obs: &mut Vec<Observation>, path: &str, bits: String) {
    obs.push(Observation {
        path: path.to_string(),
        bits,
        flags: 0,
        has_flags: false,
    });
}

fn push_mapped(
    obs: &mut Vec<Observation>,
    path: &str,
    bits: String,
    f: ExceptionFlags,
) -> Result<(), String> {
    let native = map_public_flags(f)?;
    obs.push(Observation {
        path: path.to_string(),
        bits,
        flags: native,
        has_flags: true,
    });
    Ok(())
}

fn push_port(obs: &mut Vec<Observation>, bits: String, native: u32) {
    obs.push(Observation {
        path: "rust/port".to_string(),
        bits,
        flags: native,
        has_flags: true,
    });
}

fn want_operands(op: &str) -> Result<usize, String> {
    match op {
        "add" | "sub" | "mul" | "div" | "quantize" => Ok(2),
        "fma" => Ok(3),
        _ => Err(format!("unsupported operation {op:?}")),
    }
}

fn observe(case: &Case) -> Result<Vec<Observation>, String> {
    let length = match case.width { 32 => 8, 64 => 16, 128 => 33, other => return Err(format!("unsupported decimal width {other}")) };
    for raw in &case.operands {
        if raw.len() != length { return Err(format!("raw image must have {length} characters")); }
        for (i, b) in raw.bytes().enumerate() {
            if case.width == 128 && i == 16 {
                if b != b':' { return Err("raw image requires hi:lo separator".into()); }
            } else if !b.is_ascii_digit() && !(b'a'..=b'f').contains(&b) { return Err("raw image must be padded lowercase hex".into()); }
        }
    }
    let mode = public_mode(&case.mode)
        .ok_or_else(|| format!("unsupported rounding mode {:?}", case.mode))?;
    let rnd = port_rounding(&case.mode)
        .ok_or_else(|| format!("unsupported rounding mode {:?}", case.mode))?;
    let nearest_even = case.mode == "nearest_even";
    match case.width {
        32 => observe32(case, mode, rnd, nearest_even),
        64 => observe64(case, mode, rnd, nearest_even),
        128 => observe128(case, mode, rnd, nearest_even),
        other => Err(format!("unsupported decimal width {other}")),
    }
}

fn observe32(
    case: &Case,
    mode: RoundingMode,
    rnd: i64,
    nearest_even: bool,
) -> Result<Vec<Observation>, String> {
    let want = want_operands(&case.op)?;
    if case.operands.len() != want {
        return Err(format!(
            "op {:?} requires {} operands, got {}",
            case.op,
            want,
            case.operands.len()
        ));
    }
    let mut raws = Vec::with_capacity(case.operands.len());
    for raw in &case.operands {
        raws.push(
            u32::from_str_radix(raw, 16)
                .map_err(|e| format!("invalid decimal32 raw image {raw:?}: {e}"))?,
        );
    }
    let xb = raws[0];
    let yb = raws[1];
    let x = Decimal32::from_bits(xb);
    let y = Decimal32::from_bits(yb);
    let mut obs = Vec::new();

    match case.op.as_str() {
        "add" => {
            let (mv, mf) = x.add_with_mode(y, mode);
            push_mapped(&mut obs, "rust/public/mode", hex32(mv.to_bits()), mf)?;
            let (pb, pf) = bid754::generated::bid32_status::bid32_add_with_flags(xb, yb, rnd);
            push_port(&mut obs, hex32(pb), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", hex32(x.add(y).to_bits()));
                let (fv, ff) = x.add_with_flags(y);
                push_mapped(&mut obs, "rust/public/flags", hex32(fv.to_bits()), ff)?;
            }
        }
        "sub" => {
            let (mv, mf) = x.sub_with_mode(y, mode);
            push_mapped(&mut obs, "rust/public/mode", hex32(mv.to_bits()), mf)?;
            let (pb, pf) = bid754::generated::bid32_status::bid32_sub_with_flags(xb, yb, rnd);
            push_port(&mut obs, hex32(pb), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", hex32(x.sub(y).to_bits()));
                let (fv, ff) = x.sub_with_flags(y);
                push_mapped(&mut obs, "rust/public/flags", hex32(fv.to_bits()), ff)?;
            }
        }
        "mul" => {
            let (mv, mf) = x.mul_with_mode(y, mode);
            push_mapped(&mut obs, "rust/public/mode", hex32(mv.to_bits()), mf)?;
            let (pb, pf) = bid754::generated::bid32_status::bid32_mul_with_flags(xb, yb, rnd);
            push_port(&mut obs, hex32(pb), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", hex32(x.mul(y).to_bits()));
                let (fv, ff) = x.mul_with_flags(y);
                push_mapped(&mut obs, "rust/public/flags", hex32(fv.to_bits()), ff)?;
            }
        }
        "div" => {
            let (mv, mf) = x.div_with_mode(y, mode);
            push_mapped(&mut obs, "rust/public/mode", hex32(mv.to_bits()), mf)?;
            let (pb, pf) = bid754::generated::bid32_status::bid32_div_with_flags(xb, yb, rnd);
            push_port(&mut obs, hex32(pb), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", hex32(x.div(y).to_bits()));
                let (fv, ff) = x.div_with_flags(y);
                push_mapped(&mut obs, "rust/public/flags", hex32(fv.to_bits()), ff)?;
            }
        }
        "quantize" => {
            let (mv, mf) = x.quantize_with_mode(y, mode);
            push_mapped(&mut obs, "rust/public/mode", hex32(mv.to_bits()), mf)?;
            let (pb, pf) = bid754::generated::bid32_quantize::bid32_quantize(xb, yb, rnd);
            push_port(&mut obs, hex32(pb), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", hex32(x.quantize(y).to_bits()));
                let (fv, ff) = x.quantize_with_flags(y);
                push_mapped(&mut obs, "rust/public/flags", hex32(fv.to_bits()), ff)?;
            }
        }
        "fma" => {
            let zb = raws[2];
            let z = Decimal32::from_bits(zb);
            let (mv, mf) = x.fma_with_mode(y, z, mode);
            push_mapped(&mut obs, "rust/public/mode", hex32(mv.to_bits()), mf)?;
            let (pb, pf) = bid754::generated::bid32_fma::bid32_fma(xb, yb, zb, rnd);
            push_port(&mut obs, hex32(pb), pf);
            if nearest_even {
                let (fv, ff) = x.fma(y, z);
                push_mapped(&mut obs, "rust/public/flags", hex32(fv.to_bits()), ff)?;
            }
        }
        other => return Err(format!("unsupported operation {other:?}")),
    }
    Ok(obs)
}

fn observe64(
    case: &Case,
    mode: RoundingMode,
    rnd: i64,
    nearest_even: bool,
) -> Result<Vec<Observation>, String> {
    let want = want_operands(&case.op)?;
    if case.operands.len() != want {
        return Err(format!(
            "op {:?} requires {} operands, got {}",
            case.op,
            want,
            case.operands.len()
        ));
    }
    let mut raws = Vec::with_capacity(case.operands.len());
    for raw in &case.operands {
        raws.push(
            u64::from_str_radix(raw, 16)
                .map_err(|e| format!("invalid decimal64 raw image {raw:?}: {e}"))?,
        );
    }
    let xb = raws[0];
    let yb = raws[1];
    let x = Decimal64::from_bits(xb);
    let y = Decimal64::from_bits(yb);
    let mut obs = Vec::new();

    match case.op.as_str() {
        "add" => {
            let (mv, mf) = x.add_with_mode(y, mode);
            push_mapped(&mut obs, "rust/public/mode", hex64(mv.to_bits()), mf)?;
            let (pb, pf) = bid754::generated::add64::bid64_add_with_flags(xb, yb, rnd);
            push_port(&mut obs, hex64(pb), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", hex64(x.add(y).to_bits()));
                let (fv, ff) = x.add_with_flags(y);
                push_mapped(&mut obs, "rust/public/flags", hex64(fv.to_bits()), ff)?;
            }
        }
        "sub" => {
            let (mv, mf) = x.sub_with_mode(y, mode);
            push_mapped(&mut obs, "rust/public/mode", hex64(mv.to_bits()), mf)?;
            let (pb, pf) = bid754::generated::add64::bid64_sub_with_flags(xb, yb, rnd);
            push_port(&mut obs, hex64(pb), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", hex64(x.sub(y).to_bits()));
                let (fv, ff) = x.sub_with_flags(y);
                push_mapped(&mut obs, "rust/public/flags", hex64(fv.to_bits()), ff)?;
            }
        }
        "mul" => {
            let (mv, mf) = x.mul_with_mode(y, mode);
            push_mapped(&mut obs, "rust/public/mode", hex64(mv.to_bits()), mf)?;
            let (pb, pf) = bid754::generated::mul64::bid64_mul_with_flags(xb, yb, rnd);
            push_port(&mut obs, hex64(pb), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", hex64(x.mul(y).to_bits()));
                let (fv, ff) = x.mul_with_flags(y);
                push_mapped(&mut obs, "rust/public/flags", hex64(fv.to_bits()), ff)?;
            }
        }
        "div" => {
            let (mv, mf) = x.div_with_mode(y, mode);
            push_mapped(&mut obs, "rust/public/mode", hex64(mv.to_bits()), mf)?;
            let (pb, pf) = bid754::generated::div64::bid64_div_with_flags(xb, yb, rnd);
            push_port(&mut obs, hex64(pb), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", hex64(x.div(y).to_bits()));
                let (fv, ff) = x.div_with_flags(y);
                push_mapped(&mut obs, "rust/public/flags", hex64(fv.to_bits()), ff)?;
            }
        }
        "quantize" => {
            let (mv, mf) = x.quantize_with_mode(y, mode);
            push_mapped(&mut obs, "rust/public/mode", hex64(mv.to_bits()), mf)?;
            let (pb, pf) = bid754::generated::quantize64::bid64_quantize(xb, yb, rnd);
            push_port(&mut obs, hex64(pb), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", hex64(x.quantize(y).to_bits()));
                let (fv, ff) = x.quantize_with_flags(y);
                push_mapped(&mut obs, "rust/public/flags", hex64(fv.to_bits()), ff)?;
            }
        }
        "fma" => {
            let zb = raws[2];
            let z = Decimal64::from_bits(zb);
            let (mv, mf) = x.fma_with_mode(y, z, mode);
            push_mapped(&mut obs, "rust/public/mode", hex64(mv.to_bits()), mf)?;
            let (pb, pf) = bid754::generated::fma64::bid64_fma(xb, yb, zb, rnd);
            push_port(&mut obs, hex64(pb), pf);
            if nearest_even {
                let (fv, ff) = x.fma(y, z);
                push_mapped(&mut obs, "rust/public/flags", hex64(fv.to_bits()), ff)?;
            }
        }
        other => return Err(format!("unsupported operation {other:?}")),
    }
    Ok(obs)
}

struct Operand128 {
    dec: Decimal128,
    hi: u64,
    lo: u64,
}

fn parse128(operands: &[String]) -> Result<Vec<Operand128>, String> {
    let mut out = Vec::with_capacity(operands.len());
    for raw in operands {
        let parts: Vec<&str> = raw.split(':').collect();
        if parts.len() != 2 {
            return Err(format!("invalid decimal128 raw image {raw:?}"));
        }
        let hi = u64::from_str_radix(parts[0], 16)
            .map_err(|e| format!("invalid decimal128 hi in {raw:?}: {e}"))?;
        let lo = u64::from_str_radix(parts[1], 16)
            .map_err(|e| format!("invalid decimal128 lo in {raw:?}: {e}"))?;
        let mut bytes = [0u8; 16];
        bytes[0..8].copy_from_slice(&lo.to_le_bytes());
        bytes[8..16].copy_from_slice(&hi.to_le_bytes());
        out.push(Operand128 {
            dec: Decimal128::from_le_bytes(bytes),
            hi,
            lo,
        });
    }
    Ok(out)
}

fn observe128(
    case: &Case,
    mode: RoundingMode,
    rnd: i64,
    nearest_even: bool,
) -> Result<Vec<Observation>, String> {
    let want = want_operands(&case.op)?;
    if case.operands.len() != want {
        return Err(format!(
            "op {:?} requires {} operands, got {}",
            case.op,
            want,
            case.operands.len()
        ));
    }
    let ops = parse128(&case.operands)?;
    let x = &ops[0];
    let y = &ops[1];
    let mut obs = Vec::new();

    match case.op.as_str() {
        "add" => {
            let (mv, mf) = x.dec.add_with_mode(y.dec, mode);
            push_mapped(&mut obs, "rust/public/mode", dec128_hex(mv), mf)?;
            let mut pf = 0u32;
            let r = bid754::generated::bid128_add::bid128_add(
                bid128(x.hi, x.lo),
                bid128(y.hi, y.lo),
                rnd,
                &mut pf,
            );
            push_port(&mut obs, hex128(r.hi, r.lo), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", dec128_hex(x.dec.add(y.dec)));
                let (fv, ff) = x.dec.add_with_flags(y.dec);
                push_mapped(&mut obs, "rust/public/flags", dec128_hex(fv), ff)?;
            }
        }
        "sub" => {
            let (mv, mf) = x.dec.sub_with_mode(y.dec, mode);
            push_mapped(&mut obs, "rust/public/mode", dec128_hex(mv), mf)?;
            let mut pf = 0u32;
            let r = bid754::generated::bid128_add::bid128_sub(
                bid128(x.hi, x.lo),
                bid128(y.hi, y.lo),
                rnd,
                &mut pf,
            );
            push_port(&mut obs, hex128(r.hi, r.lo), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", dec128_hex(x.dec.sub(y.dec)));
                let (fv, ff) = x.dec.sub_with_flags(y.dec);
                push_mapped(&mut obs, "rust/public/flags", dec128_hex(fv), ff)?;
            }
        }
        "mul" => {
            let (mv, mf) = x.dec.mul_with_mode(y.dec, mode);
            push_mapped(&mut obs, "rust/public/mode", dec128_hex(mv), mf)?;
            let (r, pf) = bid754::generated::bid128_mul::bid128_mul(
                bid128(x.hi, x.lo),
                bid128(y.hi, y.lo),
                rnd,
            );
            push_port(&mut obs, hex128(r.hi, r.lo), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", dec128_hex(x.dec.mul(y.dec)));
                let (fv, ff) = x.dec.mul_with_flags(y.dec);
                push_mapped(&mut obs, "rust/public/flags", dec128_hex(fv), ff)?;
            }
        }
        "div" => {
            let (mv, mf) = x.dec.div_with_mode(y.dec, mode);
            push_mapped(&mut obs, "rust/public/mode", dec128_hex(mv), mf)?;
            let (r, pf) = bid754::generated::bid128_div::bid128_div(
                bid128(x.hi, x.lo),
                bid128(y.hi, y.lo),
                rnd,
            );
            push_port(&mut obs, hex128(r.hi, r.lo), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", dec128_hex(x.dec.div(y.dec)));
                let (fv, ff) = x.dec.div_with_flags(y.dec);
                push_mapped(&mut obs, "rust/public/flags", dec128_hex(fv), ff)?;
            }
        }
        "quantize" => {
            let (mv, mf) = x.dec.quantize_with_mode(y.dec, mode);
            push_mapped(&mut obs, "rust/public/mode", dec128_hex(mv), mf)?;
            let (r, pf) = bid754::generated::bid128_quantize::bid128_quantize(
                bid128(x.hi, x.lo),
                bid128(y.hi, y.lo),
                rnd,
            );
            push_port(&mut obs, hex128(r.hi, r.lo), pf);
            if nearest_even {
                push_value(&mut obs, "rust/public/value", dec128_hex(x.dec.quantize(y.dec)));
                let (fv, ff) = x.dec.quantize_with_flags(y.dec);
                push_mapped(&mut obs, "rust/public/flags", dec128_hex(fv), ff)?;
            }
        }
        "fma" => {
            let z = &ops[2];
            let (mv, mf) = x.dec.fma_with_mode(y.dec, z.dec, mode);
            push_mapped(&mut obs, "rust/public/mode", dec128_hex(mv), mf)?;
            let (r, pf) = bid754::generated::bid128_fma::bid128_fma(
                bid128(x.hi, x.lo),
                bid128(y.hi, y.lo),
                bid128(z.hi, z.lo),
                rnd,
            );
            push_port(&mut obs, hex128(r.hi, r.lo), pf);
            if nearest_even {
                let (fv, ff) = x.dec.fma(y.dec, z.dec);
                push_mapped(&mut obs, "rust/public/flags", dec128_hex(fv), ff)?;
            }
        }
        other => return Err(format!("unsupported operation {other:?}")),
    }
    Ok(obs)
}

fn run() -> Result<(), String> {
    let stdin = io::stdin();
    let stdout = io::stdout();
    let mut out = stdout.lock();
    for line in stdin.lock().lines() {
        let line = line.map_err(|e| e.to_string())?;
        if line.trim().is_empty() { return Err("empty case record".into()); }
        let case: Case = serde_json::from_str(&line).map_err(|e| format!("parse case: {e}"))?;
        let obs = observe(&case)?;
        let encoded = serde_json::to_string(&obs).map_err(|e| e.to_string())?;
        writeln!(out, "{encoded}").map_err(|e| e.to_string())?;
    }
    Ok(())
}

fn main() {
    if let Err(e) = run() {
        eprintln!("finite_probe: {e}");
        std::process::exit(1);
    }
}
`
