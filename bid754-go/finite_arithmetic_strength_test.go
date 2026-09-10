package bid754

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

func TestFiniteArithmeticOracleStrength(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		killed := make(map[[2]int]bool)
		for _, family := range decimalprobe.Families() {
			for _, negative := range []bool{false, true} {
				for mode, name := range finiteModes {
					s, err := decimalprobe.Generate(family, width, name, 0, 754, 40, negative)
					if err != nil {
						t.Fatal(err)
					}
					want, err := decimalprobe.Validate(s)
					if err != nil {
						t.Fatal(err)
					}
					raw, flags, err := finitePublic(s.Case)
					if err != nil {
						t.Fatal(err)
					}
					if err := decimalref.Compare(width, want, raw, flags); err != nil {
						t.Fatalf("baseline %v: %v", s, err)
					}
					if err := decimalref.Compare(width, want, raw, flags^0x20); err == nil {
						t.Fatalf("inexact flag corruption escaped for %v", s)
					}
					for other, otherName := range finiteModes {
						if mode == other {
							continue
						}
						misrouted := s.Case
						misrouted.Mode = otherName
						raw, flags, err := finitePublic(misrouted)
						if err != nil {
							t.Fatal(err)
						}
						if decimalref.Compare(width, want, raw, flags) != nil {
							killed[[2]int{mode, other}] = true
						}
					}
					if family == "fma_cancellation" {
						mul := s.Case
						mul.Op, mul.Operands = "mul", s.Case.Operands[:2]
						product, productFlags, err := finitePublic(mul)
						if err != nil {
							t.Fatal(err)
						}
						add := s.Case
						add.Op, add.Operands = "add", []string{product, s.Case.Operands[2]}
						unfused, addFlags, err := finitePublic(add)
						if err != nil {
							t.Fatal(err)
						}
						if decimalref.Compare(width, want, unfused, productFlags|addFlags) == nil {
							t.Fatalf("unfused dispatch escaped: %+v", s)
						}
					}
				}
			}
		}
		if len(killed) != 20 {
			t.Fatalf("decimal%d mode substitutions detected=%d/20: %v", width, len(killed), killed)
		}
		t.Logf("FINITE-ORACLE-STRENGTH width=%d mode-directions=%d/20", width, len(killed))
	}
}

func TestFiniteArithmeticSemanticShrinkAndReplay(t *testing.T) {
	s, err := decimalprobe.Generate("add_tie", 32, "nearest_even", 0, 567890, 40, false)
	if err != nil {
		t.Fatal(err)
	}
	failure := func(sample decimalprobe.Sample) (bool, error) {
		want, err := decimalprobe.Validate(sample)
		if err != nil {
			return false, err
		}
		baseline, flags, err := finitePublic(sample.Case)
		if err != nil {
			return false, err
		}
		if err := decimalref.Compare(sample.Case.Width, want, baseline, flags); err != nil {
			return false, fmt.Errorf("baseline failed: %w", err)
		}
		changed := sample.Case
		changed.Mode = "nearest_away"
		raw, flags, err := finitePublic(changed)
		if err != nil {
			return false, err
		}
		return decimalref.Compare(sample.Case.Width, want, raw, flags) != nil, nil
	}
	shrunk, stats, err := decimalprobe.Shrink(s, 128, failure)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Accepted == 0 || reflect.DeepEqual(s, shrunk) {
		t.Fatalf("failed to simplify a real mode-routing counterexample: %+v", stats)
	}
	data, err := json.Marshal(shrunk)
	if err != nil {
		t.Fatal(err)
	}
	var replay decimalprobe.Sample
	if err := json.Unmarshal(data, &replay); err != nil {
		t.Fatal(err)
	}
	if fails, err := failure(replay); err != nil || !fails {
		t.Fatalf("raw counterexample did not replay: fails=%v error=%v", fails, err)
	}
	want, err := decimalprobe.Validate(replay)
	if err != nil || want.Rounding != "tie" {
		t.Fatalf("shrink lost midpoint relation: %v %+v", err, want)
	}
	if _, stats, err := decimalprobe.Shrink(s, 1, failure); err != nil || !stats.Exhausted || stats.Attempts != 1 {
		t.Fatalf("shrink budget not explicit: %+v %v", stats, err)
	}
	t.Logf("FINITE-SHRINK attempts=%d accepted=%d exhausted=%v raw=%s", stats.Attempts, stats.Accepted, stats.Exhausted, data)
}
