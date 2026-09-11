package bid754

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand/v2"
	"strconv"
	"strings"
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
