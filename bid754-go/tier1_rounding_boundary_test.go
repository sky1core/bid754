package bid754

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
	"github.com/sky1core/bid754/bid754-go/internal/tier1ref"
)

func tier1BoundaryGroups(t *testing.T) ([]string, map[string][]tier1ref.Case) {
	t.Helper()
	cases, err := tier1RoundingBoundaryCases()
	if err != nil {
		t.Fatal(err)
	}
	groups := map[string][]tier1ref.Case{}
	for _, c := range cases {
		key := fmt.Sprintf("%s/%d/%d/%s", c.Op, c.Width, c.Target, c.Mode)
		groups[key] = append(groups[key], c)
	}
	var keys []string
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys, groups
}

func TestTier1RoundingBoundaryContract(t *testing.T) {
	cases, err := tier1RoundingBoundaryCases()
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	witnesses := map[string]bool{}
	for i, c := range cases {
		if _, err := tier1ref.Evaluate(c); err != nil {
			t.Fatalf("%+v: %v", c, err)
		}
		counts[fmt.Sprintf("%s/%d/%d", c.Op, c.Width, c.Target)]++
		if c.Op == "from_int" || c.Op == "from_uint" {
			witnesses[fmt.Sprintf("%s/%d/%d/%s/%s", c.Op, c.Width, c.Target, c.Mode, c.Param)] = true
			if c.Width == 32 && c.Op == "from_int" && (c.Param == "12345665000000001" && c.Mode == "nearest_even" || c.Param == "12345664999999999" && c.Mode == "nearest_away") {
				t.Logf("boundary index=%d mode=%s input=%s", i, c.Mode, c.Param)
			}
		} else {
			d, err := decimalref.Decode(c.Width, c.Operands[0])
			if err != nil {
				t.Fatal(err)
			}
			witnesses[fmt.Sprintf("convert/%d/%d/%s/%t/%sE%d", c.Width, c.Target, c.Mode, d.Negative, d.Coeff, d.Exp)] = true
		}
		data := make([]byte, tier1InputSize)
		data[4] = 0x80
		binary.LittleEndian.PutUint64(data[5:13], uint64(i))
		for mode, name := range finiteModes {
			data[2] = byte(mode)
			got, err := tier1Sample(data)
			want := c
			want.Mode = name
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("boundary mutation %d/%s got=%+v want=%+v err=%v", i, name, got, want, err)
			}
		}
	}
	for _, mode := range finiteModes {
		for _, negative := range []bool{false, true} {
			for _, n := range []string{"12345665000000001", "12345664999999999", "12345675000000001", "12345674999999999"} {
				key := fmt.Sprintf("convert/128/32/%s/%t/%sE0", mode, negative, n)
				if !witnesses[key] {
					t.Errorf("missing narrowing witness %s", key)
				}
			}
		}
	}
	for _, width := range []int{32, 64, 128} {
		for _, row := range []struct {
			op          string
			bits, count int
		}{{"from_int", 32, 1435}, {"from_uint", 32, 720}, {"from_int", 64, 6475}, {"from_uint", 64, 3660}} {
			key := fmt.Sprintf("%s/%d/%d", row.op, width, row.bits)
			if counts[key] != row.count {
				t.Errorf("%s cases=%d want%d", key, counts[key], row.count)
			}
		}
	}
	for key, want := range map[string]int{"convert/64/32": 8400, "convert/128/32": 12600, "convert/128/64": 8400} {
		if counts[key] != want {
			t.Errorf("%s cases=%d want%d", key, counts[key], want)
		}
	}
	if len(cases) != 66270 {
		t.Errorf("boundary cases=%d want66270", len(cases))
	}
	for _, op := range []string{"from_int", "from_uint"} {
		for _, width := range []int{32, 64, 128} {
			for _, sign := range []string{"", "-"} {
				if op == "from_uint" && sign != "" {
					continue
				}
				for _, mode := range finiteModes {
					for _, n := range []string{"12345665000000001", "12345664999999999", "12345675000000001", "12345674999999999"} {
						key := fmt.Sprintf("%s/%d/64/%s/%s%s", op, width, mode, sign, n)
						if !witnesses[key] {
							t.Errorf("missing boundary witness %s", key)
						}
					}
				}
			}
		}
	}
	keys, _ := tier1BoundaryGroups(t)
	if len(keys) != 75 {
		t.Errorf("boundary cells=%d want75", len(keys))
	}
	t.Logf("TIER1-ROUNDING-BOUNDARY cases=%d cells=%d counts=%v", len(cases), len(keys), counts)
}

func TestTier1RoundingBoundaryGo(t *testing.T) {
	keys, groups := tier1BoundaryGroups(t)
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			failures, observations := 0, 0
			for _, c := range groups[key] {
				want, err := tier1ref.Evaluate(c)
				if err != nil {
					t.Fatal(err)
				}
				obs, err := tier1PublicPaths(c)
				if err != nil {
					t.Fatalf("%+v: %v", c, err)
				}
				paths := map[string]bool{}
				for _, path := range tier1ExpectedPaths(c, "go") {
					paths[path] = true
				}
				for _, o := range obs {
					if !paths[o.Path] {
						t.Fatalf("unexpected/duplicate path %s", o.Path)
					}
					delete(paths, o.Path)
					observations++
					requiresFlags := !(c.Op != "convert" && (c.Width == 128 || c.Width == 64 && c.Target == 32))
					if requiresFlags && !o.HasFlags {
						t.Fatalf("missing flags %s", o.Path)
					}
					if err := tier1ref.Compare(c, want, o); err != nil {
						if failures < 2 {
							t.Errorf("case=%+v expected=%+v observation=%+v: %v", c, want, o, err)
						}
						failures++
					}
				}
				if len(paths) != 0 {
					t.Fatalf("missing paths %v", paths)
				}
			}
			t.Logf("cases=%d observations=%d failures=%d", len(groups[key]), observations, failures)
		})
	}
}

func TestTier1BoundaryMutation(t *testing.T) {
	cases, err := tier1RoundingBoundaryCases()
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range []string{"from_int", "from_uint", "convert"} {
		found := false
		for i, c := range cases {
			if c.Op != op || c.Mode != "nearest_even" {
				continue
			}
			if op == "convert" {
				if c.Width != 128 || c.Target != 32 {
					continue
				}
				d, err := decimalref.Decode(c.Width, c.Operands[0])
				if err != nil {
					t.Fatal(err)
				}
				if d.Negative || d.Exp != 0 || d.Coeff.String() != "12345665000000000" {
					continue
				}
			} else if c.Width != 32 || c.Target != 64 || c.Param != "12345665000000000" {
				continue
			}
			data := make([]byte, tier1InputSize)
			data[4], data[13] = 0x80, 1
			binary.LittleEndian.PutUint64(data[5:13], uint64(i))
			mutated, err := tier1Sample(data)
			if err != nil {
				t.Fatal(err)
			}
			before, err := tier1ref.Evaluate(c)
			if err != nil {
				t.Fatal(err)
			}
			after, err := tier1ref.Evaluate(mutated)
			if err != nil {
				t.Fatal(err)
			}
			if before.Decimal.Coeff.String() != "1234566" || after.Decimal.Coeff.String() != "1234567" || before.Decimal.Exp != 10 || after.Decimal.Exp != 10 {
				t.Fatalf("%s mutation failed to cross the rounding midpoint: before=%+v after=%+v", op, before, after)
			}
			if !reflect.DeepEqual(cases[i], c) {
				t.Fatal("mutation changed the shared seed corpus")
			}
			found = true
			break
		}
		if !found {
			t.Fatalf("missing midpoint seed for %s", op)
		}
	}
}

func TestTier1BoundaryFuzzSeeds(t *testing.T) {
	cases, err := tier1RoundingBoundaryCases()
	if err != nil {
		t.Fatal(err)
	}
	selected := tier1BoundaryFuzzSeedIndices(cases)
	if len(selected) == 0 || len(selected) > 2000 {
		t.Fatalf("boundary fuzz seeds=%d want 1..2000", len(selected))
	}
	cells, witnesses := map[string]bool{}, map[string]bool{}
	for _, i := range selected {
		c := cases[i]
		key := fmt.Sprintf("%s/%d/%d/%s", c.Op, c.Width, c.Target, c.Mode)
		cells[key] = true
		witnesses[key+"/"+c.Param] = true
	}
	keys, _ := tier1BoundaryGroups(t)
	for _, key := range keys {
		if !cells[key] {
			t.Errorf("missing fuzz seed cell %s", key)
		}
	}
	for _, op := range []string{"from_int", "from_uint"} {
		for _, mode := range finiteModes {
			for _, n := range []string{"12345665000000001", "12345664999999999", "12345675000000001", "12345674999999999"} {
				key := fmt.Sprintf("%s/32/64/%s/%s", op, mode, n)
				if !witnesses[key] {
					t.Errorf("missing fuzz witness %s", key)
				}
			}
		}
	}
	t.Logf("TIER1-BOUNDARY-FUZZ seeds=%d cells=%d", len(selected), len(cells))
}

func TestTier1RoundingBoundaryBigDecimal(t *testing.T) {
	configured, err := tier1Configured()
	if err != nil {
		t.Fatal(err)
	}
	if !configured {
		t.Skip("requires Tier1 Java and generated Rust comparison processes")
	}
	s, err := tier1Start()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.close(); err != nil {
			t.Error(err)
		}
	})
	keys, groups := tier1BoundaryGroups(t)
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			failures, checked, observations := 0, 0, 0
			excluded := map[string]int{}
			for _, c := range groups[key] {
				want, obs, java, err := s.execute(c)
				if err != nil {
					t.Fatalf("case=%+v: %v", c, err)
				}
				observations += len(obs)
				if java.Status == "ok" {
					checked++
				} else {
					excluded[java.Reason]++
				}
				if err := tier1Check(c, want, obs, java); err != nil {
					if failures < 2 {
						t.Errorf("case=%+v expected=%+v observations=%+v java=%+v: %v", c, want, obs, java, err)
					}
					failures++
				}
			}
			if checked == 0 {
				t.Error("no independent Java numeric comparisons")
			}
			t.Logf("cases=%d checked=%d excluded=%v observations=%d failures=%d", len(groups[key]), checked, excluded, observations, failures)
		})
	}
}
