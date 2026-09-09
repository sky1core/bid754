package testgen

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type bidCodecParseExpectation struct {
	Input  string
	Width  int
	Mode   int
	Lo, Hi uint64
	Flags  uint32
}

func WriteBidCodecPublicParseOracleOutputs(repoRoot string) error {
	files, err := GenerateBidCodecVectorTestOutputs()
	if err != nil {
		return err
	}
	for _, path := range []string{bidCodecVectorsGoFullTestPath, bidCodecVectorsRustFullParseTestPath} {
		if err := os.WriteFile(filepath.Join(repoRoot, path), files[path], 0644); err != nil {
			return err
		}
	}
	return nil
}

func bidCodecRoundedInputs() []bidCodecParseExpectation {
	var rows []bidCodecParseExpectation
	add := func(input string, classes [3]string) {
		for i, width := range []int{32, 64, 128} {
			if classes[i] != "rounded" {
				continue
			}
			for mode := 0; mode < 5; mode++ {
				rows = append(rows, bidCodecParseExpectation{Input: input, Width: width, Mode: mode})
			}
		}
	}
	for _, row := range bidCodecGoFullFromStringRecords() {
		class := bidCodecGoFullFromStringClasses[*row.Input]
		add(*row.Input, [3]string{class, class, class})
	}
	for _, row := range bidCodecGoFullStringVectorRecords() {
		add(row.Input, bidCodecGoFullStringVectorClasses[row.Input])
	}
	return rows
}

var bidCodecParseOracleCache struct {
	sync.Once
	Rows      []bidCodecParseExpectation
	Selection []bidCodecOfficialParseSelection
	Err       error
}

func bidCodecParseOracleRows() ([]bidCodecParseExpectation, error) {
	bidCodecParseOracleCache.Do(func() {
		bidCodecParseOracleCache.Rows, bidCodecParseOracleCache.Selection, bidCodecParseOracleCache.Err = bidCodecBuildParseOracleRows()
	})
	return bidCodecParseOracleCache.Rows, bidCodecParseOracleCache.Err
}

type bidCodecOfficialParseSelection struct {
	Input                   string
	Width, Line, SourceRows int
	Classification          string
	Selected                bool
}

var bidCodecPublicFinitePattern = regexp.MustCompile(`^[ \t]*[+-]?([0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE]([+-]?[0-9]+))?$`)
var bidCodecPublicSpecialPattern = regexp.MustCompile(`(?i)^[ \t]*[+-]?(?:inf(?:inity)?|s?nan[0-9]*)$`)

func bidCodecPublicFiniteCohortFits(input string, width int) bool {
	m := bidCodecPublicFinitePattern.FindStringSubmatch(input)
	if m == nil {
		return false
	}
	precision, minQuantum, maxQuantum := 7, int64(-101), int64(90)
	switch width {
	case 64:
		precision, minQuantum, maxQuantum = 16, -398, 369
	case 128:
		precision, minQuantum, maxQuantum = 34, -6176, 6111
	}
	digits := strings.ReplaceAll(m[1], ".", "")
	if len(strings.TrimLeft(digits, "0")) > precision {
		return false
	}
	quantum := new(big.Int)
	if m[2] != "" {
		quantum.SetString(m[2], 10)
	}
	if point := strings.IndexByte(m[1], '.'); point >= 0 {
		quantum.Sub(quantum, big.NewInt(int64(len(m[1])-point-1)))
	}
	return quantum.Cmp(big.NewInt(minQuantum)) >= 0 && quantum.Cmp(big.NewInt(maxQuantum)) <= 0
}

func bidCodecOfficialParseInputs() ([]bidCodecParseExpectation, []bidCodecOfficialParseSelection, error) {
	dir, err := bidCodecParseIntelDir()
	if err != nil {
		return nil, nil, err
	}
	if err := bidCodecVerifyIntelParseSources(dir); err != nil {
		return nil, nil, err
	}
	specs, err := parseReadtestFunctionSpecs(filepath.Join(dir, "TESTS/readtest.h"))
	if err != nil {
		return nil, nil, err
	}
	var rows []bidCodecParseExpectation
	var inventory []bidCodecOfficialParseSelection
	for _, width := range []int{32, 64, 128} {
		name, token := fmt.Sprintf("bid%d_from_string", width), fmt.Sprintf("OP_DEC%d", width)
		var signatures []readtestFunctionSpec
		for _, spec := range specs {
			if spec.Name == name {
				signatures = append(signatures, spec)
			}
		}
		if len(signatures) != 1 || signatures[0].Output != token || len(signatures[0].Inputs) != 1 || signatures[0].Inputs[0] != token {
			return nil, nil, fmt.Errorf("codec parse oracle: unexpected official signature for %s", name)
		}
		cases, skips, err := parseReadtestSubset(filepath.Join(dir, "TESTS/readtest.in"), ReadTestSpec{
			Function: name, Kind: "from_string", OutputType: token, InputTypes: signatures[0].Inputs,
		})
		if err != nil {
			return nil, nil, err
		}
		if len(skips) != 0 || len(cases) == 0 {
			return nil, nil, fmt.Errorf("codec parse oracle: %s cases=%d unclassified source skips=%v", name, len(cases), skips)
		}
		seen := map[string]int{}
		for _, tc := range cases {
			input := tc.Operands[0]
			if index, ok := seen[input]; ok {
				inventory[index].SourceRows++
				continue
			}
			seen[input] = len(inventory)
			class := "public_finite"
			if bidCodecPublicSpecialPattern.MatchString(input) {
				class = "special_outside_finite_selection"
			} else if !bidCodecPublicFinitePattern.MatchString(input) {
				class = "public_grammar_reject_c_legacy"
			}
			inventory = append(inventory, bidCodecOfficialParseSelection{Input: input, Width: width, Line: tc.Line, SourceRows: 1, Classification: class})
			if class != "public_finite" {
				continue
			}
			for mode := 0; mode < 5; mode++ {
				rows = append(rows, bidCodecParseExpectation{Input: input, Width: width, Mode: mode})
			}
		}
	}
	return rows, inventory, nil
}

func bidCodecParsePairMask(rows []bidCodecParseExpectation) uint16 {
	var mask uint16
	bit := uint(0)
	for a := 0; a < 5; a++ {
		for b := a + 1; b < 5; b++ {
			if rows[a].Lo != rows[b].Lo || rows[a].Hi != rows[b].Hi || rows[a].Flags != rows[b].Flags {
				mask |= 1 << bit
			}
			bit++
		}
	}
	return mask
}

func bidCodecBuildParseOracleRows() ([]bidCodecParseExpectation, []bidCodecOfficialParseSelection, error) {
	candidates, inventory, err := bidCodecOfficialParseInputs()
	if err != nil {
		return nil, nil, err
	}
	rounded := bidCodecRoundedInputs()
	all, err := bidCodecRunParseOracle(append(rounded, candidates...))
	if err != nil {
		return nil, nil, err
	}
	rows := all[:len(rounded):len(rounded)]
	for _, row := range rows {
		if row.Flags == 0 {
			return nil, nil, fmt.Errorf("codec rounded oracle unexpectedly exact: %+v", row)
		}
	}
	candidates = all[len(rounded):]
	groups := map[struct {
		width int
		input string
	}][]bidCodecParseExpectation{}
	for i := 0; i < len(candidates); i += 5 {
		group := candidates[i : i+5]
		groups[struct {
			width int
			input string
		}{group[0].Width, group[0].Input}] = group
	}
	for i := range inventory {
		item := &inventory[i]
		if item.Classification != "public_finite" {
			continue
		}
		group := groups[struct {
			width int
			input string
		}{item.Width, item.Input}]
		for _, row := range group {
			if row.Flags == 0 && !bidCodecPublicFiniteCohortFits(item.Input, item.Width) {
				item.Classification = "public_exact_cohort_reject_c_legacy"
				break
			}
		}
	}
	for _, width := range []int{32, 64, 128} {
		var covered uint16
		for i := range inventory {
			item := &inventory[i]
			if item.Width != width || item.Classification != "public_finite" {
				continue
			}
			group := groups[struct {
				width int
				input string
			}{item.Width, item.Input}]
			mask := bidCodecParsePairMask(group)
			if mask & ^covered == 0 {
				continue
			}
			covered |= mask
			item.Selected = true
			duplicate := false
			for _, row := range rounded {
				if row.Width == width && row.Input == item.Input {
					duplicate = true
					break
				}
			}
			if !duplicate {
				rows = append(rows, group...)
			}
		}
		if covered != 0x3ff {
			return nil, inventory, fmt.Errorf("codec official parse d%d: mode-pair mask=%03x want=3ff", width, covered)
		}
	}
	return rows, inventory, nil
}

func bidCodecParseIntelDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "third_party/intel_dfp/LIBRARY/src/bid32_string.c")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("codec parse oracle: pinned Intel source unavailable; provision generation inputs")
		}
		dir = parent
	}
	return filepath.Join(dir, "third_party/intel_dfp"), nil
}

func bidCodecRunParseOracle(rows []bidCodecParseExpectation) ([]bidCodecParseExpectation, error) {
	dir, err := bidCodecParseIntelDir()
	if err != nil {
		return nil, err
	}
	if err := bidCodecVerifyIntelParseSources(dir); err != nil {
		return nil, err
	}
	tmp, err := os.MkdirTemp("", "bidcodec-parse-oracle-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	var source strings.Builder
	source.WriteString("#include <stdio.h>\n#include \"bid_conf.h\"\n#include \"bid_functions.h\"\nint main(void) {\n")
	for _, row := range rows {
		fmt.Fprintf(&source, "{ unsigned f=0; char s[]=%s; ", strconv.Quote(row.Input))
		switch row.Width {
		case 32, 64:
			fmt.Fprintf(&source, "BID_UINT%d v=bid%d_from_string(s,%d,&f); printf(\"%%llx 0 %%x\\n\",(unsigned long long)v,f);", row.Width, row.Width, row.Mode)
		case 128:
			fmt.Fprintf(&source, "BID_UINT128 v=bid128_from_string(s,%d,&f); printf(\"%%llx %%llx %%x\\n\",(unsigned long long)v.w[0],(unsigned long long)v.w[1],f);", row.Mode)
		default:
			return nil, fmt.Errorf("codec parse oracle: unsupported width %d", row.Width)
		}
		source.WriteString("}\n")
	}
	source.WriteString("return 0; }\n")
	sourcePath := filepath.Join(tmp, "oracle.c")
	if err := os.WriteFile(sourcePath, []byte(source.String()), 0600); err != nil {
		return nil, err
	}
	src := filepath.Join(dir, "LIBRARY/src")
	binary := filepath.Join(tmp, "oracle")
	args := []string{"-O2", "-DBID_SIZE_LONG=8", "-DDECIMAL_CALL_BY_REFERENCE=0", "-DDECIMAL_GLOBAL_ROUNDING=0", "-DDECIMAL_GLOBAL_EXCEPTION_FLAGS=0", "-I", src, sourcePath}
	for _, name := range []string{"bid32_string.c", "bid64_string.c", "bid128_string.c", "bid_decimal_data.c", "bid128_2_str_tables.c", "bid128.c"} {
		args = append(args, filepath.Join(src, name))
	}
	args = append(args, "-o", binary)
	if out, err := exec.Command("cc", args...).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("codec parse oracle compile: %w\n%s", err, out)
	}
	out, err := exec.Command(binary).Output()
	if err != nil {
		return nil, fmt.Errorf("codec parse oracle execute: %w", err)
	}
	lines := bytes.Split(bytes.TrimSpace(out), []byte{'\n'})
	if len(lines) != len(rows) {
		return nil, fmt.Errorf("codec parse oracle returned %d rows, want %d", len(lines), len(rows))
	}
	for i, line := range lines {
		if n, err := fmt.Sscanf(string(line), "%x %x %x", &rows[i].Lo, &rows[i].Hi, &rows[i].Flags); err != nil || n != 3 {
			return nil, fmt.Errorf("codec parse oracle row %d: malformed result %q", i, line)
		}
		if rows[i].Flags & ^uint32(0x38) != 0 {
			return nil, fmt.Errorf("codec parse row %d: unexpected Intel flags %#x", i, rows[i].Flags)
		}
		bidCodecApplyDirectedOverflowDeviation(&rows[i])
	}
	return rows, nil
}

func bidCodecVerifyIntelParseSources(dir string) error {
	archive, err := os.ReadFile(filepath.Join(dir, "IntelRDFPMathLib20U4.tar.gz"))
	if err != nil {
		return fmt.Errorf("codec parse oracle pinned archive: %w", err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(archive)) != "1df86132e7a31fd74d784fee1c679b21a088f73a8ec979cfaf784c200392e125" {
		return fmt.Errorf("codec parse oracle: Intel v20U4 archive checksum mismatch")
	}
	compressed, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return err
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	verified := 0
	verifiedReadtest := 0
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		isReadtest := header.Name == "TESTS/readtest.h" || header.Name == "TESTS/readtest.in"
		if !header.FileInfo().Mode().IsRegular() || (!strings.HasPrefix(header.Name, "LIBRARY/src/") && !isReadtest) {
			continue
		}
		expected, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		actual, err := os.ReadFile(filepath.Join(dir, header.Name))
		if err != nil {
			return fmt.Errorf("codec parse oracle source %s: %w", header.Name, err)
		}
		if !bytes.Equal(actual, expected) {
			return fmt.Errorf("codec parse oracle source differs from pinned archive: %s", header.Name)
		}
		verified++
		if isReadtest {
			verifiedReadtest++
		}
	}
	if verified == 0 || verifiedReadtest != 2 {
		return fmt.Errorf("codec parse oracle: pinned archive lacks required source/readtest files")
	}
	return nil
}

func bidCodecApplyDirectedOverflowDeviation(row *bidCodecParseExpectation) {
	if row.Width == 128 || strings.ContainsAny(row.Input, "eE") || row.Flags&0x08 == 0 {
		return
	}
	negative := strings.HasPrefix(strings.TrimLeft(row.Input, " \t"), "-")
	if row.Mode != 3 && !(row.Mode == 1 && !negative) && !(row.Mode == 2 && negative) {
		return
	}
	precision, exponent := 7, int32(90)
	if row.Width == 64 {
		precision, exponent = 16, 369
	}
	coeff := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision)), nil), big.NewInt(1))
	c := bidCodecRefComponents{Sign: negative, Kind: bidCodecRefNormal, Coefficient: coeff, Exponent: exponent}
	if row.Width == 32 {
		row.Lo = uint64(refEncode32(c))
	} else {
		row.Lo = refEncode64(c)
	}
}

func bidCodecParseExpectationLiterals(rows []bidCodecParseExpectation, rust bool) (string, string) {
	var inputs, values strings.Builder
	indices := map[string]int{}
	for _, row := range rows {
		index, ok := indices[row.Input]
		if !ok {
			index = len(indices)
			indices[row.Input] = index
			literal := strconv.Quote(row.Input)
			if rust {
				literal = rustStringLiteral(row.Input)
			}
			fmt.Fprintf(&inputs, "\n    %s,", literal)
		}
		if rust {
			fmt.Fprintf(&values, "\n    ParseExpectation { input: PARSE_INPUTS[%d], width: %d, mode: %d, lo: 0x%x, hi: 0x%x, flags: 0x%x },", index, row.Width, row.Mode, row.Lo, row.Hi, row.Flags)
		} else {
			fmt.Fprintf(&values, "\n    {goFullParseInputs[%d], %d, %d, 0x%x, 0x%x, 0x%x},", index, row.Width, row.Mode, row.Lo, row.Hi, row.Flags)
		}
	}
	return inputs.String(), values.String()
}
