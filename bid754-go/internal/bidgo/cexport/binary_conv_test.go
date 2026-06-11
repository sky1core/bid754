package main

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func parseBracketHexUint64(s string, hexPattern *regexp.Regexp) (uint64, bool) {
	matches := hexPattern.FindStringSubmatch(s)
	if matches == nil {
		return 0, false
	}
	value, err := strconv.ParseUint(matches[1], 16, 64)
	if err != nil {
		return 0, false
	}
	return value, true
}

func parseBracketHexUint128(s string, hexPattern *regexp.Regexp) (uint64, uint64, bool) {
	matches := hexPattern.FindStringSubmatch(s)
	if matches == nil {
		return 0, 0, false
	}
	hex := matches[1]
	if len(hex) > 32 {
		return 0, 0, false
	}
	if len(hex) < 32 {
		hex = strings.Repeat("0", 32-len(hex)) + hex
	}
	hi, err := strconv.ParseUint(hex[:16], 16, 64)
	if err != nil {
		return 0, 0, false
	}
	lo, err := strconv.ParseUint(hex[16:], 16, 64)
	if err != nil {
		return 0, 0, false
	}
	return hi, lo, true
}

func testBinary32Readtest(t *testing.T, prefix string, fn func(uint64, int) (uint32, uint32)) {
	testFile := "../../third_party/intel_dfp/TESTS/readtest.in"

	file, err := os.Open(testFile)
	if err != nil {
		t.Skipf("Intel test file not found: %v", err)
		return
	}
	defer file.Close()

	hexPattern := regexp.MustCompile(`^\[([0-9a-fA-F]+)\]$`)

	passed := 0
	failed := 0
	skipped := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, prefix+" ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 5 {
			skipped++
			continue
		}

		rndMode, err := strconv.Atoi(parts[1])
		if err != nil {
			skipped++
			continue
		}
		a, ok1 := parseCexportBid64Input(parts[2], hexPattern)
		expected, ok2 := parseBracketHexUint64(parts[3], hexPattern)
		expectedFlags, ok3 := parseCexportFlags(parts[4])
		if !ok1 || !ok2 || !ok3 {
			skipped++
			continue
		}

		result, actualFlags := fn(a, rndMode)
		if uint64(result) == expected && actualFlags == expectedFlags {
			passed++
		} else {
			failed++
			if failed <= 10 {
				t.Errorf("%s -> got=[%08x]/%02x want=[%08x]/%02x", line, result, actualFlags, uint32(expected), expectedFlags)
			}
		}
	}

	t.Logf("%s: %d passed, %d failed, %d skipped", prefix, passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("%s: %d tests failed", prefix, failed)
	}
}

func testBinary64Readtest(t *testing.T, prefix string, fn func(uint64, int) (uint64, uint32)) {
	testFile := "../../third_party/intel_dfp/TESTS/readtest.in"

	file, err := os.Open(testFile)
	if err != nil {
		t.Skipf("Intel test file not found: %v", err)
		return
	}
	defer file.Close()

	hexPattern := regexp.MustCompile(`^\[([0-9a-fA-F]+)\]$`)

	passed := 0
	failed := 0
	skipped := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, prefix+" ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 5 {
			skipped++
			continue
		}

		rndMode, err := strconv.Atoi(parts[1])
		if err != nil {
			skipped++
			continue
		}
		a, ok1 := parseCexportBid64Input(parts[2], hexPattern)
		expected, ok2 := parseBracketHexUint64(parts[3], hexPattern)
		expectedFlags, ok3 := parseCexportFlags(parts[4])
		if !ok1 || !ok2 || !ok3 {
			skipped++
			continue
		}

		result, actualFlags := fn(a, rndMode)
		if result == expected && actualFlags == expectedFlags {
			passed++
		} else {
			failed++
			if failed <= 10 {
				t.Errorf("%s -> got=[%016x]/%02x want=[%016x]/%02x", line, result, actualFlags, expected, expectedFlags)
			}
		}
	}

	t.Logf("%s: %d passed, %d failed, %d skipped", prefix, passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("%s: %d tests failed", prefix, failed)
	}
}

func testBinary128Readtest(t *testing.T, prefix string, fn func(uint64, int) (uint64, uint64, uint32)) {
	testFile := "../../third_party/intel_dfp/TESTS/readtest.in"

	file, err := os.Open(testFile)
	if err != nil {
		t.Skipf("Intel test file not found: %v", err)
		return
	}
	defer file.Close()

	hexPattern := regexp.MustCompile(`^\[([0-9a-fA-F]+)\]$`)

	passed := 0
	failed := 0
	skipped := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, prefix+" ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 5 {
			skipped++
			continue
		}

		rndMode, err := strconv.Atoi(parts[1])
		if err != nil {
			skipped++
			continue
		}
		a, ok1 := parseCexportBid64Input(parts[2], hexPattern)
		expectedHi, expectedLo, ok2 := parseBracketHexUint128(parts[3], hexPattern)
		expectedFlags, ok3 := parseCexportFlags(parts[4])
		if !ok1 || !ok2 || !ok3 {
			skipped++
			continue
		}

		resultHi, resultLo, actualFlags := fn(a, rndMode)
		if resultHi == expectedHi && resultLo == expectedLo && actualFlags == expectedFlags {
			passed++
		} else {
			failed++
			if failed <= 10 {
				t.Errorf("%s -> got=[%016x%016x]/%02x want=[%016x%016x]/%02x", line, resultHi, resultLo, actualFlags, expectedHi, expectedLo, expectedFlags)
			}
		}
	}

	t.Logf("%s: %d passed, %d failed, %d skipped", prefix, passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("%s: %d tests failed", prefix, failed)
	}
}

func TestBid64ToBinary32IntelReadtest(t *testing.T) {
	testBinary32Readtest(t, "bid64_to_binary32", func(x uint64, rndMode int) (uint32, uint32) {
		return bid64ToBinary32(x, rndMode)
	})
}

func TestBid64ToBinary64IntelReadtest(t *testing.T) {
	testBinary64Readtest(t, "bid64_to_binary64", func(x uint64, rndMode int) (uint64, uint32) {
		return bid64ToBinary64(x, rndMode)
	})
}

func TestBid64ToBinary128IntelReadtest(t *testing.T) {
	testBinary128Readtest(t, "bid64_to_binary128", func(x uint64, rndMode int) (uint64, uint64, uint32) {
		return bid64ToBinary128(x, rndMode)
	})
}
