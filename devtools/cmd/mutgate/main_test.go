package main

import "testing"

func TestIntLitDelta(t *testing.T) {
	cases := []struct {
		lit   string
		delta int64
		want  string
		ok    bool
	}{
		{"5", 1, "6", true},
		{"5", -1, "4", true},
		{"0", -1, "(-1)", true},
		{"0", 1, "1", true},
		{"0x1f", 1, "0x20", true},
		{"0x0", -1, "(-1)", true},
		{"0xffffffffffffffff", 1, "", false},
		{"0Xff", -1, "0Xfe", true},
		{"0b101", 1, "0b110", true},
		{"0o17", 1, "0o20", true},
		{"017", 1, "020", true},
		{"1_000", 1, "1001", true},
		{"9999999", 1, "10000000", true},
	}
	for _, c := range cases {
		got, ok := intLitDelta(c.lit, c.delta)
		if ok != c.ok || got != c.want {
			t.Errorf("intLitDelta(%q,%d) = %q,%v want %q,%v", c.lit, c.delta, got, ok, c.want, c.ok)
		}
	}
}

func TestParseStrata(t *testing.T) {
	q, order, err := parseStrata("aor=5,cmp=4")
	if err != nil || q["aor"] != 5 || q["cmp"] != 4 || len(order) != 2 || order[0] != "aor" {
		t.Fatalf("parseStrata: q=%v order=%v err=%v", q, order, err)
	}
	if _, _, err := parseStrata("bogus"); err == nil {
		t.Fatal("parseStrata should reject entries without '='")
	}
	q, order, err = parseStrata("")
	if q != nil || order != nil || err != nil {
		t.Fatalf("empty strata should be nil,nil,nil; got %v %v %v", q, order, err)
	}
}
