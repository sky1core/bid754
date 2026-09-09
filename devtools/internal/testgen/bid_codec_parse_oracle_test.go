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
	t.Logf("independent Intel C parse tuples=%d by_width=%v tuple_sha256=%x", len(rows), counts, hash.Sum(nil))
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
