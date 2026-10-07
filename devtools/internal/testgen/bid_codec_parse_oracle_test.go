package testgen

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestBidCodecParseOracleSelection(t *testing.T) {
	rows, err := bidCodecParseOracleRows()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) <= len(bidCodecRoundedInputs()) {
		t.Fatal("oracle selection mismatch")
	}
	anchors := loadVerificationAnchors(t)
	if len(rows) != anchors.BidCodecParseOracleTuples {
		t.Errorf("oracle tuples=%d, anchored=%d", len(rows), anchors.BidCodecParseOracleTuples)
	}
	counts := map[string]int{}
	seen := map[struct {
		input       string
		width, mode int
	}]bool{}
	modes := map[int]map[int]bool{}
	for _, row := range rows {
		key := struct {
			input       string
			width, mode int
		}{row.Input, row.Width, row.Mode}
		if seen[key] {
			t.Fatalf("duplicate oracle tuple %+v", key)
		}
		seen[key] = true
		counts[fmt.Sprintf("bid%d", row.Width)]++
		if modes[row.Width] == nil {
			modes[row.Width] = map[int]bool{}
		}
		modes[row.Width][row.Mode] = true
	}
	for _, width := range []int{32, 64, 128} {
		if counts[fmt.Sprintf("bid%d", width)] != anchors.BidCodecParseOracleByWidth[fmt.Sprintf("bid%d", width)] {
			t.Errorf("d%d oracle count=%d differs from independent anchor=%d", width, counts[fmt.Sprintf("bid%d", width)], anchors.BidCodecParseOracleByWidth[fmt.Sprintf("bid%d", width)])
		}
		if len(modes[width]) != 5 {
			t.Fatalf("d%d missing modes", width)
		}
	}
	hash := sha256.New()
	for _, row := range rows {
		fmt.Fprintf(hash, "%d\t%s\t%d\t%016x\t%016x\t%02x\n", row.Width, strconv.Quote(row.Input), row.Mode, row.Lo, row.Hi, row.Flags)
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != anchors.BidCodecParseOracleSHA256 {
		t.Errorf("parse oracle tuple hash=%s, anchored=%s", got, anchors.BidCodecParseOracleSHA256)
	}
	t.Logf("pinned Intel C with registered IEEE parse corrections tuples=%d by_width=%v tuple_sha256=%x", len(rows), counts, hash.Sum(nil))
}

func TestBidCodecPublicParseGeneratedOutputs(t *testing.T) {
	files, err := GenerateBidCodecVectorTestOutputs()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join("..", "..")
	for _, path := range []string{bidCodecVectorsGoFullTestPath, bidCodecVectorsRustFullParseTestPath} {
		full := filepath.Join(root, path)
		got, err := os.ReadFile(full)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, files[path]) {
			t.Fatalf("%s differs from codec generator", path)
		}
	}
}

func TestBidCodecParseOracleRejectsUnpinnedSources(t *testing.T) {
	source, err := filepath.Abs("../../third_party/intel_dfp")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("archive_checksum", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "IntelRDFPMathLib20U4.tar.gz"), []byte("corrupt"), 0600); err != nil {
			t.Fatal(err)
		}
		err := bidCodecVerifyIntelParseSources(dir)
		if err == nil || !strings.Contains(err.Error(), "archive checksum mismatch") {
			t.Fatalf("want checksum failure, got %v", err)
		}
	})
	t.Run("extracted_source", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Symlink(filepath.Join(source, "IntelRDFPMathLib20U4.tar.gz"), filepath.Join(dir, "IntelRDFPMathLib20U4.tar.gz")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(source, "TESTS"), filepath.Join(dir, "TESTS")); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(dir, "LIBRARY/src")
		if err := os.MkdirAll(target, 0755); err != nil {
			t.Fatal(err)
		}
		files, err := os.ReadDir(filepath.Join(source, "LIBRARY/src"))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			path := filepath.Join(target, file.Name())
			if file.Name() == "bid32_string.c" {
				if err := os.WriteFile(path, []byte("altered"), 0600); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Symlink(filepath.Join(source, "LIBRARY/src", file.Name()), path); err != nil {
				t.Fatal(err)
			}
		}
		err = bidCodecVerifyIntelParseSources(dir)
		if err == nil || !strings.Contains(err.Error(), "source differs from pinned archive: LIBRARY/src/bid32_string.c") {
			t.Fatalf("want source mismatch, got %v", err)
		}
	})
	for _, name := range []string{"readtest.h", "readtest.in"} {
		t.Run("extracted_"+name, func(t *testing.T) {
			dir := t.TempDir()
			for _, path := range []string{"IntelRDFPMathLib20U4.tar.gz", "LIBRARY"} {
				if err := os.Symlink(filepath.Join(source, path), filepath.Join(dir, path)); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Mkdir(filepath.Join(dir, "TESTS"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{"readtest.h", "readtest.in"} {
				target := filepath.Join(dir, "TESTS", path)
				if path == name {
					if err := os.WriteFile(target, []byte("altered"), 0600); err != nil {
						t.Fatal(err)
					}
				} else if err := os.Symlink(filepath.Join(source, "TESTS", path), target); err != nil {
					t.Fatal(err)
				}
			}
			err := bidCodecVerifyIntelParseSources(dir)
			if err == nil || !strings.Contains(err.Error(), "source differs from pinned archive: TESTS/"+name) {
				t.Fatalf("want official readtest mismatch, got %v", err)
			}
		})
	}
}

func TestBidCodecParseOracleOfficialModePairs(t *testing.T) {
	rows, err := bidCodecParseOracleRows()
	if err != nil {
		t.Fatal(err)
	}
	sourceCounts := map[int]int{}
	classes := map[int]map[string]int{}
	coverage := map[int]uint16{}
	for _, item := range bidCodecParseOracleCache.Selection {
		sourceCounts[item.Width] += item.SourceRows
		if classes[item.Width] == nil {
			classes[item.Width] = map[string]int{}
		}
		classes[item.Width][item.Classification]++
		if !item.Selected {
			continue
		}
		if item.Classification != "public_finite" {
			t.Fatalf("selected excluded official input: %+v", item)
		}
		group := make([]bidCodecParseExpectation, 5)
		count := 0
		for _, row := range rows {
			if row.Width == item.Width && row.Input == item.Input {
				group[row.Mode] = row
				count++
			}
		}
		if count != 5 {
			t.Fatalf("official row has %d modes: %+v", count, item)
		}
		mask := bidCodecParsePairMask(group)
		coverage[item.Width] |= mask
		t.Logf("selected d%d official_line=%d input=%q pair_mask=%03x", item.Width, item.Line, item.Input, mask)
		for _, row := range group {
			t.Logf("selected_tuple %d\t%s\t%d\t%016x\t%016x\t%02x", row.Width, strconv.Quote(row.Input), row.Mode, row.Lo, row.Hi, row.Flags)
		}
	}
	for _, width := range []int{32, 64, 128} {
		t.Logf("d%d official_source_rows=%d unique_classes=%v pair_mask=%03x", width, sourceCounts[width], classes[width], coverage[width])
		if coverage[width] != 0x3ff {
			t.Errorf("d%d not all 10 mode pairs separated", width)
		}
	}
}

func TestBidCodecParseOracleIEEEUnderflowAndZero(t *testing.T) {
	type expectation struct {
		input string
		width int
		mode  int
		lo    uint64
		hi    uint64
		flags uint32
	}
	rows := []expectation{
		{"1e-102", 32, 2, 1, 0, 0x30},
		{"-1e-102", 32, 1, 0x80000001, 0, 0x30},
		{"1e-102", 32, 0, 0, 0, 0x30},
		{"1e-399", 64, 2, 1, 0, 0x30},
		{"-1e-399", 64, 1, 0x8000000000000001, 0, 0x30},
		{"0." + strings.Repeat("0", 100) + "14999999", 32, 0, 1, 0, 0x30},
		{"14999999e-108", 32, 0, 1, 0, 0x30},
		{"-0." + strings.Repeat("0", 100) + "14999999", 32, 0, 0x80000001, 0, 0x30},
		{"0." + strings.Repeat("0", 894) + "12345678901234567", 64, 0, 0, 0, 0x30},
		{"-0." + strings.Repeat("0", 894) + "12345678901234567", 64, 1, 0x8000000000000001, 0, 0x30},
		{"0e-6211", 128, 2, 0, 0, 0},
		{"-0e-6211", 128, 1, 0, 0x8000000000000000, 0},
	}
	request := make([]bidCodecParseExpectation, len(rows))
	for i, row := range rows {
		request[i] = bidCodecParseExpectation{Input: row.input, Width: row.width, Mode: row.mode}
	}
	got, err := bidCodecRunParseOracle(request)
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range rows {
		if got[i].Lo != want.lo || got[i].Hi != want.hi || got[i].Flags != want.flags {
			t.Errorf("d%d %q mode=%d: got=%016x:%016x flags=%02x, want=%016x:%016x flags=%02x", want.width, want.input, want.mode, got[i].Hi, got[i].Lo, got[i].Flags, want.hi, want.lo, want.flags)
		}
	}
	for _, input := range []string{"0e-6211", "-0e-6211", "0e-399", "-0e-399"} {
		width := 128
		if strings.Contains(input, "399") {
			width = 32
		}
		if bidCodecPublicFiniteCohortFits(input, width) || !bidCodecFiniteLiteralIsZero(input) {
			t.Errorf("d%d %q: out-of-range zero cohort selected for public parse", width, input)
		}
	}
	for _, row := range bidCodecRoundedInputs() {
		if bidCodecFiniteLiteralIsZero(row.Input) && !bidCodecPublicFiniteCohortFits(row.Input, row.Width) {
			t.Errorf("d%d %q: invalid zero cohort retained in public rounded oracle", row.Width, row.Input)
		}
	}
	wrongSign := bidCodecParseExpectation{Input: "-0e-6211", Width: 128, Mode: 0}
	if err := bidCodecApplyTinyParseDeviation(&wrongSign); err != nil {
		t.Fatal(err)
	}
	if wrongSign.Lo != 0 || wrongSign.Hi != 0x8000000000000000 || wrongSign.Flags != 0 {
		t.Errorf("negative written zero corrected to %016x:%016x flags=%02x", wrongSign.Hi, wrongSign.Lo, wrongSign.Flags)
	}
}

func TestNumericBoundaryReadtestSources(t *testing.T) {
	root := filepath.Join("..", "..")
	manifest, err := LoadManifest(filepath.Join(root, "testgen_manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, read := range manifest.ReadTests {
		if !strings.HasPrefix(read.Name, "numeric_boundary_") && read.Name != "bid32_from_string_underflow_boundary_cdiverge" {
			continue
		}
		if read.CompareGroup != "CMP_FUZZYSTATUS" || strings.HasSuffix(read.Name, "_cdiverge") != (read.NativeCompareSkipReason != "") {
			t.Errorf("readtest %q has incorrect comparator or C skip contract", read.Name)
		}
		cases, skips, err := parseReadtestSubset(filepath.Join(root, read.Source), read)
		if err != nil {
			t.Errorf("readtest %q: %v", read.Name, err)
			continue
		}
		if len(cases) == 0 || len(skips) != 0 {
			t.Errorf("readtest %q parsed=%d skips=%v", read.Name, len(cases), skips)
		}
		total += len(cases)
	}
	if total != 4888 {
		t.Errorf("numeric boundary rows parsed=%d, want 4888", total)
	}
}

func TestTinyParseDeviationPreservesExactCohorts(t *testing.T) {
	for _, tc := range []struct {
		width  int
		input  string
		lo, hi uint64
	}{
		{32, "1e-100", 0x00800001, 0},
		{64, "1e-397", 0x0020000000000001, 0},
		{128, "1e-6175", 1, 0x0002000000000000},
	} {
		for mode := 0; mode < 5; mode++ {
			row := bidCodecParseExpectation{Width: tc.width, Input: tc.input, Mode: mode, Lo: tc.lo, Hi: tc.hi}
			want := row
			if err := bidCodecApplyTinyParseDeviation(&row); err != nil {
				t.Fatal(err)
			}
			if row != want {
				t.Errorf("d%d mode%d: exact cohort changed: got %+v want %+v", tc.width, mode, row, want)
			}
		}
	}
}
