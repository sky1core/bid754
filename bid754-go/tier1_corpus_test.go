package bid754

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
	"github.com/sky1core/bid754/bid754-go/internal/tier1ref"
)

var tier1Ops = []string{"rem", "fmod", "scaleb", "minnum", "maxnum", "quiet_equal", "quiet_not_equal", "quiet_greater", "quiet_greater_equal", "quiet_greater_unordered", "quiet_less", "quiet_less_equal", "quiet_less_unordered", "quiet_not_greater", "quiet_not_less", "quiet_ordered", "quiet_unordered", "convert", "from_int", "from_uint", "to_int", "to_uint", "to_int_exact", "to_uint_exact"}

const tier1InputSize = 45

func tier1Targets(op string, width int) []int {
	switch {
	case op == "convert":
		var targets []int
		for _, target := range []int{32, 64, 128} {
			if width != target {
				targets = append(targets, target)
			}
		}
		return targets
	case strings.HasPrefix(op, "from_"):
		return []int{32, 64}
	case strings.HasPrefix(op, "to_"):
		return []int{8, 16, 32, 64}
	default:
		return []int{0}
	}
}

func tier1Raw(width int, hi, lo uint64) string {
	switch width {
	case 32:
		return fmt.Sprintf("%08x", uint32(lo))
	case 64:
		return fmt.Sprintf("%016x", lo)
	default:
		return fmt.Sprintf("%016x:%016x", hi, lo)
	}
}

func tier1Finite(width int, coeff *big.Int, exponent int, negative bool) (string, error) {
	return decimalref.Encode(width, decimalref.Decimal{Kind: "finite", Negative: negative, Coeff: coeff, Exp: exponent})
}

func tier1Sample(data []byte) (tier1ref.Case, error) {
	if len(data) != tier1InputSize {
		return tier1ref.Case{}, fmt.Errorf("input requires %d bytes", tier1InputSize)
	}
	if data[4]&0x80 != 0 {
		cases, err := tier1RoundingBoundaryCases()
		if err != nil {
			return tier1ref.Case{}, err
		}
		c := cases[binary.LittleEndian.Uint64(data[5:13])%uint64(len(cases))]
		c.Mode = finiteModes[data[2]%5]
		mask := binary.LittleEndian.Uint64(data[13:21])
		if c.Op == "from_int" {
			n, err := strconv.ParseInt(c.Param, 10, c.Target)
			if err != nil {
				return c, err
			}
			n = int64(uint64(n) ^ mask)
			if c.Target == 32 {
				n = int64(int32(n))
			}
			c.Param = strconv.FormatInt(n, 10)
		} else if c.Op == "from_uint" {
			n, err := strconv.ParseUint(c.Param, 10, c.Target)
			if err != nil {
				return c, err
			}
			n ^= mask
			if c.Target == 32 {
				n = uint64(uint32(n))
			}
			c.Param = strconv.FormatUint(n, 10)
		} else {
			hiText, loText, wide := strings.Cut(c.Operands[0], ":")
			if !wide {
				loText, hiText = hiText, "0"
			}
			hi, err := strconv.ParseUint(hiText, 16, 64)
			if err != nil {
				return c, err
			}
			lo, err := strconv.ParseUint(loText, 16, 64)
			if err != nil {
				return c, err
			}
			hi ^= binary.LittleEndian.Uint64(data[21:29])
			c.Operands = []string{tier1Raw(c.Width, hi, lo^mask)}
		}
		return c, tier1ref.Validate(c)
	}
	width := []int{32, 64, 128}[data[1]%3]
	op := tier1Ops[int(data[0])%len(tier1Ops)]
	targets := tier1Targets(op, width)
	c := tier1ref.Case{Width: width, Op: op, Mode: finiteModes[data[2]%5], Target: targets[int(data[3])%len(targets)]}
	param := binary.LittleEndian.Uint64(data[5:13])
	hi, lo := binary.LittleEndian.Uint64(data[13:21]), binary.LittleEndian.Uint64(data[21:29])
	yhi, ylo := binary.LittleEndian.Uint64(data[29:37]), binary.LittleEndian.Uint64(data[37:45])
	if strings.HasPrefix(op, "from_") {
		switch data[4] % 8 {
		case 2:
			param = 0
		case 3:
			param = 1
		case 4:
			param = ^uint64(0)
		case 5:
			param = uint64(1) << uint(c.Target-1)
		case 6:
			param = (uint64(1) << uint(c.Target-1)) - 1
		case 7:
			param = 10000005
		}
		if op == "from_int" {
			i := int64(param)
			if c.Target == 32 {
				i = int64(int32(i))
			}
			c.Param = strconv.FormatInt(i, 10)
		} else {
			if c.Target == 32 {
				param = uint64(uint32(param))
			}
			c.Param = strconv.FormatUint(param, 10)
		}
		return c, tier1ref.Validate(c)
	}
	p, _ := decimalref.ParametersFor(width)
	power := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(p.Precision)), nil)
	finiteRaw := func(a, b uint64) (string, error) {
		n := new(big.Int).Lsh(new(big.Int).SetUint64(a), 64)
		n.Or(n, new(big.Int).SetUint64(b))
		n.Mod(n, power)
		e := p.MinExp + int(a%uint64(p.MaxExp-p.MinExp+1))
		return tier1Finite(width, n, e, a>>63 != 0)
	}
	x, y := tier1Raw(width, hi, lo), tier1Raw(width, yhi, ylo)
	var err error
	lane := data[4] % 8
	if lane != 0 {
		x, err = finiteRaw(hi, lo)
		if err != nil {
			return c, err
		}
		y, err = finiteRaw(yhi, ylo)
		if err != nil {
			return c, err
		}
	}
	switch lane {
	case 2:
		x, err = tier1Finite(width, big.NewInt(int64(2*(lo%100000)+1)), -1, hi>>63 != 0)
		y, _ = tier1Finite(width, big.NewInt(1), 0, yhi>>63 != 0)
	case 3:
		x, err = tier1Finite(width, big.NewInt(0), p.MinExp+int(lo%uint64(p.MaxExp-p.MinExp+1)), hi>>63 != 0)
		y, _ = tier1Finite(width, big.NewInt(0), p.MinExp, yhi>>63 != 0)
	case 4:
		x, err = tier1Finite(width, big.NewInt(1), p.MinExp, hi>>63 != 0)
		y, _ = tier1Finite(width, new(big.Int).Sub(power, big.NewInt(1)), p.MaxExp, yhi>>63 != 0)
	case 5:
		bits := c.Target
		if bits == 0 || bits == 128 {
			bits = 64
		}
		n := new(big.Int).Lsh(big.NewInt(1), uint(bits-1))
		if strings.Contains(op, "uint") {
			n.Lsh(n, 1)
		}
		n.Mul(n, big.NewInt(10))
		n.Add(n, big.NewInt(int64(lo%21)-10))
		exponent := -1
		for n.Cmp(power) >= 0 {
			n.Quo(n, big.NewInt(10))
			exponent++
		}
		x, err = tier1Finite(width, n, exponent, hi>>63 != 0)
		y, _ = tier1Finite(width, big.NewInt(2), 0, false)
	case 6:
		coeff := big.NewInt(int64(lo%99999 + 1))
		x, err = tier1Finite(width, coeff, -1, hi>>63 != 0)
		y, _ = tier1Finite(width, new(big.Int).Mul(coeff, big.NewInt(10)), -2, hi>>63 != 0)
	case 7:
		prefixes := []uint64{0x7800000000000000, 0xf800000000000000, 0x7c00000000000000, 0x7e00000000000000}
		a, b := prefixes[hi%4], prefixes[yhi%4]
		switch width {
		case 32:
			x = tier1Raw(width, 0, a>>32|lo%1000000)
			y = tier1Raw(width, 0, b>>32|ylo%1000000)
		case 64:
			x = tier1Raw(width, 0, a|lo%1000000000000000)
			y = tier1Raw(width, 0, b|ylo%1000000000000000)
		case 128:
			x = tier1Raw(width, a, lo)
			y = tier1Raw(width, b, ylo)
		}
	}
	if err != nil {
		return c, err
	}
	c.Operands = []string{x}
	if op == "scaleb" {
		delta := int64(param)
		if lane != 0 && lane != 7 {
			delta = int64(param%uint64(2*(p.MaxExp-p.MinExp+2))) - int64(p.MaxExp-p.MinExp+2)
		}
		c.Param = strconv.FormatInt(delta, 10)
	} else if op != "convert" && !strings.HasPrefix(op, "to_") {
		c.Operands = append(c.Operands, y)
	}
	return c, tier1ref.Validate(c)
}

func tier1BoundaryPower(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

func tier1BoundaryCoefficients(precision, shift int) []*big.Int {
	low, high := tier1BoundaryPower(precision-1), tier1BoundaryPower(precision)
	pattern := new(big.Int).Mul(big.NewInt(1234566), tier1BoundaryPower(precision-7))
	retained := []*big.Int{
		new(big.Int).Sub(low, big.NewInt(1)), low, new(big.Int).Add(low, big.NewInt(1)),
		pattern, new(big.Int).Add(pattern, big.NewInt(1)),
		new(big.Int).Sub(high, big.NewInt(2)), new(big.Int).Sub(high, big.NewInt(1)),
	}
	scale := tier1BoundaryPower(shift)
	half := new(big.Int).Quo(scale, big.NewInt(2))
	var out []*big.Int
	for _, q := range retained {
		base := new(big.Int).Mul(q, scale)
		out = append(out, new(big.Int).Set(base))
		for delta := int64(-2); delta <= 2; delta++ {
			n := new(big.Int).Add(base, half)
			out = append(out, n.Add(n, big.NewInt(delta)))
		}
	}
	return out
}

func tier1BoundaryIntegers(bits int, unsigned bool) []*big.Int {
	lo := new(big.Int)
	hi := new(big.Int).Lsh(big.NewInt(1), uint(bits))
	if !unsigned {
		hi.Rsh(hi, 1)
		lo.Neg(hi)
	}
	hi.Sub(hi, big.NewInt(1))
	var out []*big.Int
	seen := map[string]bool{}
	add := func(n *big.Int) {
		if n.Cmp(lo) >= 0 && n.Cmp(hi) <= 0 && !seen[n.String()] {
			seen[n.String()] = true
			out = append(out, new(big.Int).Set(n))
		}
	}
	for _, limit := range []*big.Int{lo, hi, big.NewInt(0)} {
		for delta := int64(-2); delta <= 2; delta++ {
			add(new(big.Int).Add(limit, big.NewInt(delta)))
		}
	}
	for digits := 1; digits <= len(hi.String()); digits++ {
		for delta := int64(-1); delta <= 1; delta++ {
			n := new(big.Int).Add(tier1BoundaryPower(digits), big.NewInt(delta))
			add(n)
			add(new(big.Int).Neg(n))
		}
	}
	for _, precision := range []int{7, 16} {
		for shift := 1; shift <= len(hi.String())-precision; shift++ {
			for _, n := range tier1BoundaryCoefficients(precision, shift) {
				add(n)
				add(new(big.Int).Neg(n))
			}
		}
	}
	return out
}

var tier1RoundingBoundaryCases = sync.OnceValues(func() ([]tier1ref.Case, error) {
	var out []tier1ref.Case
	seen := map[string]bool{}
	add := func(c tier1ref.Case) error {
		if err := tier1ref.Validate(c); err != nil {
			return err
		}
		key := fmt.Sprintf("%+v", c)
		if !seen[key] {
			seen[key] = true
			out = append(out, c)
		}
		return nil
	}
	for _, width := range []int{32, 64, 128} {
		for _, bits := range []int{32, 64} {
			for _, op := range []string{"from_int", "from_uint"} {
				for _, n := range tier1BoundaryIntegers(bits, op == "from_uint") {
					for _, mode := range finiteModes {
						if err := add(tier1ref.Case{Width: width, Target: bits, Op: op, Mode: mode, Param: n.String()}); err != nil {
							return nil, err
						}
					}
				}
			}
		}
	}
	for _, pair := range [][2]int{{64, 32}, {128, 32}, {128, 64}} {
		src, _ := decimalref.ParametersFor(pair[0])
		dst, _ := decimalref.ParametersFor(pair[1])
		gap := src.Precision - dst.Precision
		shifts := map[int]bool{}
		for _, shift := range []int{1, 2, gap - 1, gap, 16 - dst.Precision, 17 - dst.Precision} {
			if shift < 1 || shift > gap || shifts[shift] {
				continue
			}
			shifts[shift] = true
			for _, coeff := range tier1BoundaryCoefficients(dst.Precision, shift) {
				for _, exp := range []int{0, -shift, dst.MinExp - shift, dst.MinExp - shift - 1, dst.MaxExp - shift} {
					for _, negative := range []bool{false, true} {
						raw, err := tier1Finite(pair[0], coeff, exp, negative)
						if err != nil {
							return nil, err
						}
						for _, mode := range finiteModes {
							if err := add(tier1ref.Case{Width: pair[0], Target: pair[1], Op: "convert", Mode: mode, Operands: []string{raw}}); err != nil {
								return nil, err
							}
						}
					}
				}
			}
		}
	}
	return out, nil
})

func tier1BoundaryFuzzSeedIndices(cases []tier1ref.Case) []int {
	counts := map[string]int{}
	var selected []int
	for i, c := range cases {
		key := fmt.Sprintf("%s/%d/%d/%s", c.Op, c.Width, c.Target, c.Mode)
		include := counts[key]%64 == 0
		counts[key]++
		if c.Op == "from_int" || c.Op == "from_uint" {
			switch strings.TrimPrefix(c.Param, "-") {
			case "12345665000000001", "12345664999999999", "12345675000000001", "12345674999999999":
				include = true
			}
		}
		if include {
			selected = append(selected, i)
		}
	}
	return selected
}

func tier1Corpus(seed uint64, lanes int, yield func([]byte) error) error {
	rng := rand.New(rand.NewPCG(seed, seed^0x754dec))
	for w, width := range []int{32, 64, 128} {
		for op, name := range tier1Ops {
			for target := range tier1Targets(name, width) {
				for mode := range finiteModes {
					for lane := 0; lane < lanes; lane++ {
						data := make([]byte, tier1InputSize)
						data[0], data[1], data[2], data[3], data[4] = byte(op), byte(w), byte(mode), byte(target), byte(lane)
						for off := 5; off < len(data); off += 8 {
							binary.LittleEndian.PutUint64(data[off:off+8], rng.Uint64())
						}
						if err := yield(data); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func TestTier1CorpusContract(t *testing.T) {
	cells := map[string]int{}
	for _, seed := range []uint64{754, 2019, 0xdec1} {
		err := tier1Corpus(seed, 8, func(data []byte) error {
			c, err := tier1Sample(data)
			if err != nil {
				return err
			}
			_, err = tier1ref.Evaluate(c)
			if err != nil {
				return fmt.Errorf("%+v: %w", c, err)
			}
			cells[fmt.Sprintf("%d/%s/%s/%d", c.Width, c.Op, c.Mode, c.Target)]++
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(cells) != 585 {
		t.Fatalf("cells=%d want585", len(cells))
	}
	for cell, count := range cells {
		if count != 24 {
			t.Fatalf("%s samples=%d", cell, count)
		}
	}
}
