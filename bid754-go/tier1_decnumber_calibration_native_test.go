//go:build cgo && bid754_native && bid754_decnumber_diff

package bid754

import (
	"fmt"
	"math/big"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
	"github.com/sky1core/bid754/bid754-go/internal/tier1ref"
)

func tier1DecnumberWidth(bits int) decnumberDiffWidthParams {
	switch bits {
	case 32:
		return decnumberDiffWidth32
	case 64:
		return decnumberDiffWidth64
	case 128:
		return decnumberDiffWidth128
	default:
		panic("invalid calibration width")
	}
}

func tier1DecnumberRaw(bits int, value decnumberDiffTriple) (string, error) {
	lo, hi, ok := decnumberDiffEncode(tier1DecnumberWidth(bits), value)
	if !ok {
		return "", fmt.Errorf("cannot encode decimal%d %s", bits, value.key())
	}
	return tier1Raw(bits, hi, lo), nil
}

func tier1DecnumberDecode(bits int, raw string) (decnumberDiffTriple, error) {
	var lo, hi uint64
	var err error
	if bits == 128 {
		parts := strings.Split(raw, ":")
		if len(parts) != 2 || len(parts[0]) != 16 || len(parts[1]) != 16 {
			return decnumberDiffTriple{}, fmt.Errorf("invalid decimal128 bits")
		}
		hi, err = strconv.ParseUint(parts[0], 16, 64)
		if err == nil {
			lo, err = strconv.ParseUint(parts[1], 16, 64)
		}
	} else {
		lo, err = strconv.ParseUint(raw, 16, bits)
	}
	if err != nil {
		return decnumberDiffTriple{}, err
	}
	return decnumberDiffDecode(tier1DecnumberWidth(bits), lo, hi), nil
}

func tier1DecnumberRun(code, width, mode, arity int, a, b string) (decnumberDiffTriple, uint32, error) {
	r := decnumberDiffRunDn(code, width, mode, a, b, "", arity)
	flags, unknown := decnumberDiffProjectStatus(r.rawStatus)
	if r.rc != 0 || r.parseStatus != 0 || unknown != 0 {
		return r.triple, flags, fmt.Errorf("decNumber op=%d width=%d mode=%d input=%s,%s rc=%d parse=%x status=%x unknown=%x", code, width, mode, a, b, r.rc, r.parseStatus, r.rawStatus, unknown)
	}
	return r.triple, flags, nil
}

func tier1DecnumberOracle(c tier1ref.Case) (tier1ref.Observation, error) {
	if err := tier1ref.Validate(c); err != nil {
		return tier1ref.Observation{}, err
	}
	mode := map[string]int{"nearest_even": 0, "toward_negative": 1, "toward_positive": 2, "toward_zero": 3, "nearest_away": 4}[c.Mode]
	width, arity := c.Width, len(c.Operands)
	var text [2]string
	for i, raw := range c.Operands {
		d, err := tier1DecnumberDecode(c.Width, raw)
		if err != nil {
			return tier1ref.Observation{}, err
		}
		text[i] = d.dnOperandString()
	}
	code := -1
	switch c.Op {
	case "rem":
		code = 7
	case "fmod":
		code = 8
	case "scaleb":
		code, arity, text[1] = 9, 2, c.Param
	case "minnum":
		code = 10
	case "maxnum":
		code = 11
	case "convert":
		code, width, arity, text[1] = 2, c.Target, 2, "1"
	case "from_int", "from_uint":
		code, arity, text[0], text[1] = 2, 2, c.Param, "1"
	case "to_int_exact", "to_uint_exact":
		code = 13
	case "to_int", "to_uint":
		code = 14
	default:
		if strings.HasPrefix(c.Op, "quiet_") {
			code = 12
		}
	}
	d, flags, err := tier1DecnumberRun(code, width, mode, arity, text[0], text[1])
	if err != nil {
		return tier1ref.Observation{}, err
	}
	o := tier1ref.Observation{Path: "decnumber", Kind: "decimal", Width: width, Flags: flags, HasFlags: true}
	if code == 12 {
		mask := map[string]uint8{
			"quiet_equal": 2, "quiet_not_equal": 13, "quiet_greater": 4,
			"quiet_greater_equal": 6, "quiet_greater_unordered": 12,
			"quiet_less": 1, "quiet_less_equal": 3, "quiet_less_unordered": 9,
			"quiet_not_greater": 11, "quiet_not_less": 14,
			"quiet_ordered": 7, "quiet_unordered": 8,
		}[c.Op]
		outcome := uint8(8)
		if d.kind == decnumberDiffKindFinite {
			outcome = 2
			if d.coeff.Sign() != 0 {
				if d.coeff.Cmp(big.NewInt(1)) != 0 || d.exp != 0 {
					return o, fmt.Errorf("invalid decNumber ordering result %s", d.key())
				}
				outcome = 4
				if d.sign {
					outcome = 1
				}
			}
		} else if d.kind != decnumberDiffKindQNaN {
			return o, fmt.Errorf("invalid decNumber ordering class")
		}
		o.Kind, o.Width, o.Value = "bool", 0, strconv.FormatBool(mask&outcome != 0)
		return o, nil
	}
	if code == 13 || code == 14 {
		o.Kind, o.Width = "integer", c.Target
		unsigned := strings.Contains(c.Op, "uint")
		limit := new(big.Int).Lsh(big.NewInt(1), uint(c.Target))
		low, high := new(big.Int), new(big.Int).Sub(limit, big.NewInt(1))
		if !unsigned {
			low.Neg(new(big.Int).Rsh(new(big.Int).Set(limit), 1))
			high.Rsh(high, 1)
		}
		n := new(big.Int)
		invalid := d.kind != decnumberDiffKindFinite
		if !invalid {
			if d.exp < 0 {
				return o, fmt.Errorf("decNumber integral result has fractional quantum")
			}
			n.Mul(d.coeff, decnumberDiffPow10(int(d.exp)))
			if d.sign {
				n.Neg(n)
			}
			invalid = n.Cmp(low) < 0 || n.Cmp(high) > 0
		}
		if invalid {
			n.Rsh(limit, 1)
			if !unsigned {
				n.Neg(n)
			}
			o.Flags = decnumberDiffFlagInvalid
		}
		o.Value = n.String()
		return o, nil
	}
	o.Value, err = tier1DecnumberRaw(width, d)
	return o, err
}

func tier1DecnumberResult(width int, raw string) (decnumberDiffTriple, error) {
	d, err := tier1DecnumberDecode(width, raw)
	if err != nil {
		return d, err
	}
	canonical, err := tier1DecnumberRaw(width, d)
	if err != nil {
		return d, err
	}
	if d.kind == decnumberDiffKindQNaN && len(raw) == len(canonical) {
		bits, _ := new(big.Int).SetString(strings.ReplaceAll(raw, ":", ""), 16)
		header, _ := new(big.Int).SetString(strings.ReplaceAll(canonical, ":", ""), 16)
		payload := bits.Xor(bits, header)
		if payload.Cmp(decnumberDiffPow10(tier1DecnumberWidth(width).p-1)) < 0 {
			return d, nil
		}
	} else if d.kind != decnumberDiffKindSNaN && canonical == raw {
		return d, nil
	}
	return d, fmt.Errorf("noncanonical or signaling result %s", raw)
}

func tier1DecnumberCompare(c tier1ref.Case, expected, actual tier1ref.Observation) error {
	valueOnly := strings.HasPrefix(c.Op, "from_") && (c.Width == 128 || c.Width == 64 && c.Target == 32)
	if !actual.HasFlags && (!valueOnly || actual.Flags != 0) {
		return fmt.Errorf("missing or inconsistent flags: %+v", actual)
	}
	if expected.Kind != actual.Kind || expected.Width != actual.Width || actual.HasFlags && expected.Flags != actual.Flags {
		return fmt.Errorf("kind/width/flags differ: expected=%+v actual=%+v", expected, actual)
	}
	if expected.Kind != "decimal" {
		if expected.Value != actual.Value {
			return fmt.Errorf("value differs: expected=%+v actual=%+v", expected, actual)
		}
		return nil
	}
	a, err := tier1DecnumberResult(expected.Width, expected.Value)
	if err != nil {
		return err
	}
	b, err := tier1DecnumberResult(actual.Width, actual.Value)
	if err != nil {
		return err
	}
	if decnumberDiffCompare(a, 0, b, 0) {
		return nil
	}
	if (c.Op == "minnum" || c.Op == "maxnum") && a.kind == decnumberDiffKindFinite && b.kind == decnumberDiffKindFinite {
		x, _ := tier1DecnumberDecode(c.Width, c.Operands[0])
		y, _ := tier1DecnumberDecode(c.Width, c.Operands[1])
		order, flags, err := tier1DecnumberRun(12, c.Width, 0, 2, x.dnOperandString(), y.dnOperandString())
		if err != nil {
			return err
		}
		if flags == 0 && order.kind == decnumberDiffKindFinite && order.coeff.Sign() == 0 {
			isInput := func(d decnumberDiffTriple) bool {
				return decnumberDiffCompare(d, 0, x, 0) || decnumberDiffCompare(d, 0, y, 0)
			}
			if isInput(a) && isInput(b) {
				return nil
			}
		}
	}
	return fmt.Errorf("decimal differs: expected=%s actual=%s", a.key(), b.key())
}

func tier1DecnumberCorpus(yield func(tier1ref.Case) error) error {
	for _, seed := range []uint64{10754, 12019, 67025} {
		rng := rand.New(rand.NewPCG(seed, seed^0xdec368))
		for _, width := range []int{32, 64, 128} {
			p := tier1DecnumberWidth(width)
			for _, op := range tier1Ops {
				for _, target := range tier1Targets(op, width) {
					for _, mode := range finiteModes {
						for lane := 0; lane < 8; lane++ {
							c := tier1ref.Case{Width: width, Op: op, Mode: mode, Target: target}
							if strings.HasPrefix(op, "from_") {
								v := rng.Uint64()
								if lane == 1 {
									v = 0
								}
								if lane == 5 {
									v = ^uint64(0)
								}
								if target == 32 {
									v = uint64(uint32(v))
								}
								c.Param = strconv.FormatUint(v, 10)
								if op == "from_int" {
									i := int64(v)
									if target == 32 {
										i = int64(int32(v))
									}
									c.Param = strconv.FormatInt(i, 10)
								}
							} else {
								x := decnumberDiffTriple{kind: decnumberDiffKindFinite, sign: rng.Uint64()&1 != 0, coeff: big.NewInt(int64(rng.Uint64()%9999 + 1)), exp: int32(rng.Uint64()%5) - 2}
								y := decnumberDiffTriple{kind: decnumberDiffKindFinite, sign: rng.Uint64()&1 != 0, coeff: big.NewInt(int64(rng.Uint64()%99 + 1)), exp: x.exp}
								switch lane {
								case 1:
									x.coeff.SetInt64(0)
								case 2:
									x.kind = decnumberDiffKindInf
								case 3:
									x.kind = decnumberDiffKindQNaN
								case 4:
									x.kind = decnumberDiffKindSNaN
								case 5:
									x.coeff, x.exp = new(big.Int).Set(p.maxCoeff), p.maxExp
								case 6:
									x.coeff, x.exp = big.NewInt(15), p.minExp
								case 7:
									x.coeff, x.exp = big.NewInt(15), -1
								}
								if op == "rem" || op == "fmod" {
									y.exp = x.exp
								}
								if (op == "minnum" || op == "maxnum") && lane == 7 {
									y = decnumberDiffTriple{kind: decnumberDiffKindFinite, sign: x.sign, coeff: big.NewInt(150), exp: -2}
								}
								if op == "scaleb" {
									limit := 2 * (int(p.maxExp) + 2*p.p - 1)
									c.Param = strconv.Itoa(rng.IntN(2*limit+1) - limit)
									if lane == 5 {
										c.Param = "1"
									}
									if lane == 6 || lane == 7 {
										c.Param = "-1"
									}
								}
								values := []decnumberDiffTriple{x}
								if op != "scaleb" && op != "convert" && !strings.HasPrefix(op, "to_") {
									values = append(values, y)
									if seed == 12019 {
										values[0], values[1] = values[1], values[0]
									}
								}
								for _, d := range values {
									raw, err := tier1DecnumberRaw(width, d)
									if err != nil {
										return err
									}
									c.Operands = append(c.Operands, raw)
								}
							}
							if err := yield(c); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func TestTier1ReferenceDecnumberCalibration(t *testing.T) {
	if version := decnumberDiffCVersion(); version != "decNumber 3.68" {
		t.Fatalf("unexpected oracle %q", version)
	}
	cells, counts, flags := map[string]int{}, map[int]int{}, map[int]uint32{}
	err := tier1DecnumberCorpus(func(c tier1ref.Case) error {
		dn, err := tier1DecnumberOracle(c)
		if err != nil {
			return fmt.Errorf("case=%+v: %w", c, err)
		}
		model, err := tier1ref.Evaluate(c)
		if err != nil {
			return err
		}
		modelObs := tier1ref.Observation{Path: "tier1ref", Kind: model.Kind, Width: model.Width, Value: model.Value, Flags: model.Flags, HasFlags: true}
		if model.Kind == "decimal" {
			modelObs.Value, err = decimalref.Encode(model.Width, model.Decimal)
			if err != nil {
				return err
			}
		}
		observations, err := tier1PublicPaths(c)
		if err != nil {
			return err
		}
		paths := map[string]bool{}
		for _, path := range tier1ExpectedPaths(c, "go") {
			paths[path] = true
		}
		for _, actual := range observations {
			if !paths[actual.Path] {
				return fmt.Errorf("unexpected or duplicate path %s", actual.Path)
			}
			delete(paths, actual.Path)
		}
		if len(paths) != 0 {
			return fmt.Errorf("missing paths: %v", paths)
		}
		observations = append(observations, modelObs)
		for _, actual := range observations {
			if err := tier1DecnumberCompare(c, dn, actual); err != nil {
				return fmt.Errorf("case=%+v path=%s: %w", c, actual.Path, err)
			}
		}
		cells[fmt.Sprintf("%d/%s/%s/%d", c.Width, c.Op, c.Mode, c.Target)]++
		counts[c.Width]++
		flags[c.Width] |= dn.Flags
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(cells) != 585 {
		t.Fatalf("cells=%d want585", len(cells))
	}
	for cell, count := range cells {
		if count != 24 {
			t.Fatalf("%s samples=%d want24", cell, count)
		}
	}
	for _, width := range []int{32, 64, 128} {
		if counts[width] != 4680 || flags[width] != 0x39 {
			t.Fatalf("width=%d cases=%d flags=%x", width, counts[width], flags[width])
		}
		t.Logf("TIER1-DECNUMBER width=%d cases=%d cells=195 flags=%02x excluded=0", width, counts[width], flags[width])
	}
}

func TestTier1DecnumberRoutingWitnesses(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		for _, row := range []struct {
			code       int
			a, b, want string
			arity      int
			flags      uint32
		}{
			{7, "7", "2", "-1E0", 2, 0}, {8, "7", "2", "1E0", 2, 0},
			{9, "125E-2", "-1", "125E-3", 2, 0}, {10, "7", "2", "2E0", 2, 0},
			{11, "7", "2", "7E0", 2, 0}, {12, "-1", "2", "-1E0", 2, 0},
			{13, "15E-1", "", "2E0", 1, 0x20}, {14, "15E-1", "", "2E0", 1, 0},
			{2, "-0E-3", "1", "-0E-3", 2, 0},
		} {
			d, flags, err := tier1DecnumberRun(row.code, width, 0, row.arity, row.a, row.b)
			if err != nil || flags != row.flags || d.key() != row.want {
				t.Fatalf("width=%d op=%d got=%s flags=%x err=%v", width, row.code, d.key(), flags, err)
			}
		}
	}
}

func TestTier1DecnumberComparatorStrength(t *testing.T) {
	for _, row := range []struct {
		width     int
		good, bad string
	}{
		{32, "32800000", "6cb89680"},
		{64, "31c0000000000000", "6c7386f26fc10000"},
		{128, "3040000000000000:0000000000000000", "3041ed09bead87c0:378d8e6400000000"},
		{32, "78000000", "78000001"},
		{64, "7800000000000000", "7800000000000001"},
		{128, "7800000000000000:0000000000000000", "7800000000000000:0000000000000001"},
		{32, "7c000000", "7c0f4240"},
		{64, "7c00000000000000", "7c038d7ea4c68000"},
		{128, "7c00000000000000:0000000000000000", "7c00314dc6448d93:38c15b0a00000000"},
	} {
		c := tier1ref.Case{Width: row.width, Op: "scaleb", Mode: "nearest_even", Operands: []string{row.good}, Param: "0"}
		want := tier1ref.Observation{Kind: "decimal", Width: row.width, Value: row.good, HasFlags: true}
		bad := want
		bad.Value = row.bad
		if err := tier1DecnumberCompare(c, want, want); err != nil {
			t.Fatalf("canonical result rejected: %v", err)
		}
		if tier1DecnumberCompare(c, want, bad) == nil {
			t.Errorf("accepted noncanonical result %s", row.bad)
		}
		if strings.HasPrefix(row.good, "7c") {
			header, _ := new(big.Int).SetString(strings.ReplaceAll(row.good, ":", ""), 16)
			limit := decnumberDiffPow10(tier1DecnumberWidth(row.width).p - 1)
			for _, payload := range []*big.Int{big.NewInt(1), new(big.Int).Sub(limit, big.NewInt(1))} {
				for _, sign := range []uint{0, 1} {
					bits := new(big.Int).Or(header, payload)
					bits.SetBit(bits, row.width-1, sign)
					raw := fmt.Sprintf("%0*x", row.width/4, bits)
					if row.width == 128 {
						raw = raw[:16] + ":" + raw[16:]
					}
					valid := want
					valid.Value = raw
					if err := tier1DecnumberCompare(c, want, valid); err != nil {
						t.Fatalf("canonical NaN payload/sign rejected: %v", err)
					}
				}
			}
			for _, prefix := range []string{"7d", "7e"} {
				bad.Value = prefix + row.good[2:]
				if tier1DecnumberCompare(c, want, bad) == nil {
					t.Errorf("accepted reserved/signaling NaN %s", bad.Value)
				}
			}
		}
	}
	c := tier1ref.Case{Width: 32, Op: "scaleb", Mode: "nearest_even", Operands: []string{"32800001"}, Param: "0"}
	want := tier1ref.Observation{Path: "decnumber", Kind: "decimal", Width: 32, Value: "32800001", HasFlags: true}
	if err := tier1DecnumberCompare(c, want, want); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*tier1ref.Observation){
		func(o *tier1ref.Observation) { o.Flags = 1 },
		func(o *tier1ref.Observation) { o.HasFlags = false },
		func(o *tier1ref.Observation) { o.Width = 64 },
		func(o *tier1ref.Observation) { o.Value = "32800002" },
		func(o *tier1ref.Observation) { o.Value = "b2800001" },
		func(o *tier1ref.Observation) { o.Value = "3200000a" },
		func(o *tier1ref.Observation) { o.Value = "7c000000" },
	} {
		bad := want
		mutate(&bad)
		if tier1DecnumberCompare(c, want, bad) == nil {
			t.Fatalf("accepted bad observation %+v", bad)
		}
	}
	c.Op, c.Param, c.Operands = "minnum", "", []string{"32800001", "3200000a"}
	other := want
	other.Value = c.Operands[1]
	if err := tier1DecnumberCompare(c, want, other); err != nil {
		t.Fatalf("legal equal operand rejected: %v", err)
	}
	other.Value = "31800064"
	if tier1DecnumberCompare(c, want, other) == nil {
		t.Fatal("accepted third min/max cohort")
	}
}
