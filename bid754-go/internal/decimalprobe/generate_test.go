package decimalprobe

import (
	"errors"
	"math/big"
	"math/rand"
	"reflect"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

func TestRelationCoverage(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		counts := map[string]int{}
		rng := rand.New(rand.NewSource(9102026))
		for _, family := range Families() {
			for _, mode := range []string{"nearest_even", "nearest_away", "toward_zero", "toward_positive", "toward_negative"} {
				for i := 0; i < 16; i++ {
					s, err := Generate(family, width, mode, rng.Uint64(), rng.Uint64(), int32(rng.Uint32()), i%2 == 1)
					if err != nil {
						t.Fatalf("%s decimal%d %s: %v", family, width, mode, err)
					}
					if _, err := Validate(s); err != nil {
						t.Fatal(err)
					}
					counts[family]++
				}
			}
		}
		if len(counts) != 20 {
			t.Fatalf("family coverage %v", counts)
		}
		for family, count := range counts {
			if count != 80 {
				t.Fatalf("%s reached=%d want=80", family, count)
			}
		}
		t.Logf("RELATION-COVERAGE width=%d families=20 reached=1600 oracle_completed=1600", width)
	}
}

func TestCarryBoundaryRequiresActualCarry(t *testing.T) {
	falseBoundary := Sample{Version: 1, Family: "carry_boundary", Case: decimalref.Case{Width: 32, Op: "add", Mode: "nearest_even", Operands: []string{"6c98967f", "320f4246"}}}
	if _, err := Validate(falseBoundary); err == nil {
		t.Fatal("accepted a midpoint between two seven-digit results as a carry boundary")
	}
	x := decimal(big.NewInt(9999999), 0, false)
	for _, y := range []decimalref.Decimal{
		decimal(big.NewInt(5), -1, true),
		decimal(big.NewInt(5), 0, false),
	} {
		xraw, _ := decimalref.Encode(32, x)
		yraw, _ := decimalref.Encode(32, y)
		s := Sample{Version: 1, Family: "carry_boundary", Case: decimalref.Case{Width: 32, Op: "add", Mode: "nearest_even", Operands: []string{xraw, yraw}}}
		if _, err := Validate(s); err == nil {
			t.Fatalf("accepted false carry boundary: %+v", s)
		}
	}
}

func TestShrinkPreservesNonzeroResultSign(t *testing.T) {
	bits := func(n int64) string {
		raw, err := decimalref.Encode(32, decimal(big.NewInt(n), 0, false))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	s := Sample{Version: 1, Family: "uniform-finite", Case: decimalref.Case{Width: 32, Op: "sub", Mode: "nearest_even", Operands: []string{bits(3), bits(2)}}}
	_, _, err := Shrink(s, 32, func(candidate Sample) (bool, error) {
		r, err := Validate(candidate)
		if err != nil {
			return false, err
		}
		if r.Value.Negative {
			t.Fatalf("shrinker admitted a result sign change: %+v", candidate)
		}
		return true, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	cancellation, err := Generate("cancellation", 32, "nearest_even", 0, 54, 101, false)
	if err != nil {
		t.Fatal(err)
	}
	shrunk, _, err := Shrink(cancellation, 2, func(candidate Sample) (bool, error) {
		r, err := Validate(candidate)
		if err != nil {
			return false, err
		}
		if r.Value.Negative {
			t.Fatalf("deep cancellation changed result sign: %+v", candidate)
		}
		return true, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if r, err := Validate(shrunk); err != nil || r.Value.Negative {
		t.Fatalf("invalid reduced cancellation: %+v %v", shrunk, err)
	}
}

func TestRelationValidationRejectsFalseClaims(t *testing.T) {
	s, err := Generate("add_tie", 32, "nearest_even", 0, 754, 30, false)
	if err != nil {
		t.Fatal(err)
	}
	for name, change := range map[string]func(*Sample){
		"false midpoint label": func(s *Sample) { s.Family = "add_below" },
		"false operation":      func(s *Sample) { s.Case.Op = "sub" },
		"unknown version":      func(s *Sample) { s.Version = 2 },
		"unknown mode":         func(s *Sample) { s.Case.Mode = "nearest" },
		"noncanonical input":   func(s *Sample) { s.Case.Operands[0] = "6cbfffff" },
		"nonfinite input":      func(s *Sample) { s.Case.Operands[0] = "78000000" },
		"incomplete input":     func(s *Sample) { s.Case.Operands = s.Case.Operands[:1] },
	} {
		t.Run(name, func(t *testing.T) {
			bad := s
			bad.Case.Operands = append([]string(nil), s.Case.Operands...)
			change(&bad)
			if _, err := Validate(bad); err == nil {
				t.Fatalf("accepted invalid sample %+v", bad)
			}
		})
	}
}

func TestShrinkRequiresReproducedFailure(t *testing.T) {
	s, err := Generate("add_tie", 32, "nearest_even", 0, 567890, 40, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := Shrink(s, 10, func(Sample) (bool, error) { return false, nil }); err == nil {
		t.Fatal("accepted passing baseline")
	}
	sentinel := errors.New("target execution failed")
	if _, _, err := Shrink(s, 10, func(Sample) (bool, error) { return false, sentinel }); !errors.Is(err, sentinel) {
		t.Fatalf("execution failure was treated as detection: %v", err)
	}
	original := s
	original.Case.Operands = append([]string(nil), s.Case.Operands...)
	if _, _, err := Shrink(s, 10, func(Sample) (bool, error) { return true, nil }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s, original) {
		t.Fatal("shrinking modified original raw evidence")
	}
}
