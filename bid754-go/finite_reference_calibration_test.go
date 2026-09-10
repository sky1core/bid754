package bid754

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

type finiteReferenceAnchors struct {
	Readtest       map[int]int `json:"readtest_by_width"`
	DecimalText    int         `json:"readtest_excluded_decimal_text"`
	NonfiniteInput int         `json:"readtest_excluded_nonfinite_input"`
	Decnumber      map[int]int `json:"decnumber_by_width"`
}

func loadFiniteReferenceAnchors(t *testing.T) finiteReferenceAnchors {
	t.Helper()
	data, err := os.ReadFile("../devtools/verification_anchors.json")
	if err != nil {
		t.Fatal(err)
	}
	var anchors struct {
		Finite finiteReferenceAnchors `json:"finite_reference_calibration"`
	}
	if err := json.Unmarshal(data, &anchors); err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{32, 64, 128} {
		if anchors.Finite.Readtest[width] <= 0 || anchors.Finite.Decnumber[width] <= 0 {
			t.Fatalf("missing finite reference anchors for decimal%d", width)
		}
	}
	return anchors.Finite
}

func TestFiniteReferenceReadtestCalibration(t *testing.T) {
	anchors := loadFiniteReferenceAnchors(t)
	modes := [...]string{"nearest_even", "toward_negative", "toward_positive", "toward_zero", "nearest_away"}
	checked, excluded := map[int]int{}, map[string]int{}
	for _, row := range goportLoadGeneratedReadSpec(t).ReadCases {
		width, op, found := 0, "", false
		for _, candidateWidth := range []int{32, 64, 128} {
			for _, candidateOp := range []string{"add", "sub", "mul", "div", "fma", "quantize"} {
				if row.Function == fmt.Sprintf("bid%d_%s", candidateWidth, candidateOp) {
					width, op, found = candidateWidth, candidateOp, true
				}
			}
		}
		if !found {
			continue
		}
		c := decimalref.Case{Width: width, Op: op}
		if row.Rounding < 0 || row.Rounding >= len(modes) {
			t.Fatalf("invalid rounding in %s", row.ID)
		}
		c.Mode = modes[row.Rounding]
		values := append(append([]string(nil), row.Operands...), row.Expected)
		reason := ""
		for _, text := range values {
			if !strings.HasPrefix(text, "[") || !strings.HasSuffix(text, "]") {
				reason = "decimal_text"
				break
			}
		}
		if reason != "" {
			excluded[reason]++
			continue
		}
		var expected string
		for i, text := range values {
			raw := strings.ToLower(text[1 : len(text)-1])
			raw = strings.ReplaceAll(raw, ",", ":")
			raw = strings.ReplaceAll(raw, " ", "")
			if width == 128 && len(raw) == 32 {
				raw = raw[:16] + ":" + raw[16:]
			}
			d, err := decimalref.Decode(width, raw)
			if err != nil {
				t.Fatalf("%s raw %q: %v", row.ID, raw, err)
			}
			if i < len(row.Operands) {
				if d.Kind != "finite" {
					reason = "nonfinite_input"
					break
				}
				c.Operands = append(c.Operands, raw)
			} else {
				expected = raw
			}
		}
		if reason != "" {
			excluded[reason]++
			continue
		}
		want, err := decimalref.Evaluate(c)
		if err != nil {
			t.Fatalf("%s: %v", row.ID, err)
		}
		flags, err := strconv.ParseUint(row.Status, 16, 32)
		if err != nil {
			t.Fatal(err)
		}
		if err := decimalref.CompareQuantum(width, want, expected, uint32(flags)); err != nil {
			t.Fatalf("official row %s: %v sample=%+v expected=%s/%s reference=%+v", row.ID, err, c, expected, row.Status, want)
		}
		checked[width]++
	}
	for width, count := range anchors.Readtest {
		if checked[width] != count {
			t.Fatalf("official raw rows calibrated for decimal%d: got %d want %d", width, checked[width], count)
		}
		t.Logf("FINITE-REFERENCE-READTEST width=%d cases=%d", width, checked[width])
	}
	if excluded["decimal_text"] != anchors.DecimalText || excluded["nonfinite_input"] != anchors.NonfiniteInput {
		t.Fatalf("official calibration exclusion inventory changed: %v", excluded)
	}
	t.Logf("FINITE-REFERENCE-READTEST exclusions=%v", excluded)
}
