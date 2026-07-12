package bidgo

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func parseBracketHexUint64Readtest(s string, hexPattern *regexp.Regexp) (uint64, bool) {
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

func testBid64Binary32Readtest(t *testing.T, prefix string, fn func(uint64, int) (uint32, uint32)) {
	t.Helper()

	file, err := os.Open("../../../devtools/third_party/intel_dfp/TESTS/readtest.in")
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
		a, ok1 := parseBid64ReadtestInput(parts[2], hexPattern)
		expected, ok2 := parseBracketHexUint64Readtest(parts[3], hexPattern)
		expectedFlags, ok3 := parseBidReadtestFlags(parts[4])
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
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

	t.Logf("%s: %d passed, %d failed, %d skipped", prefix, passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("%s: %d tests failed", prefix, failed)
	}
}

func testBid64Binary64Readtest(t *testing.T, prefix string, fn func(uint64, int) (uint64, uint32)) {
	t.Helper()

	file, err := os.Open("../../../devtools/third_party/intel_dfp/TESTS/readtest.in")
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
		a, ok1 := parseBid64ReadtestInput(parts[2], hexPattern)
		expected, ok2 := parseBracketHexUint64Readtest(parts[3], hexPattern)
		expectedFlags, ok3 := parseBidReadtestFlags(parts[4])
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
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

	t.Logf("%s: %d passed, %d failed, %d skipped", prefix, passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("%s: %d tests failed", prefix, failed)
	}
}

func testBid64Binary128Readtest(t *testing.T, prefix string, fn func(uint64, int) (BID_UINT128, uint32)) {
	t.Helper()

	file, err := os.Open("../../../devtools/third_party/intel_dfp/TESTS/readtest.in")
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
		a, ok1 := parseBid64ReadtestInput(parts[2], hexPattern)
		expected, ok2 := parseBid128HexLocal(parts[3], hexPattern)
		expectedFlags, ok3 := parseBidReadtestFlags(parts[4])
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
				t.Errorf("%s -> got=[%016x%016x]/%02x want=[%016x%016x]/%02x",
					line, result.w[1], result.w[0], actualFlags, expected.w[1], expected.w[0], expectedFlags)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

	t.Logf("%s: %d passed, %d failed, %d skipped", prefix, passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("%s: %d tests failed", prefix, failed)
	}
}

func testBid128Binary128Readtest(t *testing.T, prefix string, fn func(BID_UINT128, int) (BID_UINT128, uint32)) {
	t.Helper()

	file, err := os.Open("../../../devtools/third_party/intel_dfp/TESTS/readtest.in")
	if err != nil {
		t.Skipf("Intel test file not found: %v", err)
		return
	}
	defer file.Close()

	hexPattern := regexp.MustCompile(`^\[([0-9a-fA-F,]+)\]$`)

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
		a, ok1 := parseBid128BinaryReadtestInput(parts[2], hexPattern)
		expected, ok2 := parseBid128BinaryReadtestInput(parts[3], hexPattern)
		expectedFlags, ok3 := parseBidReadtestFlags(parts[4])
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
				t.Errorf("%s -> got=[%016x%016x]/%02x want=[%016x%016x]/%02x",
					line, result.w[1], result.w[0], actualFlags, expected.w[1], expected.w[0], expectedFlags)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

	t.Logf("%s: %d passed, %d failed, %d skipped", prefix, passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("%s: %d tests failed", prefix, failed)
	}
}

func TestBid64ToBinary32IntelReadtest(t *testing.T) {
	testBid64Binary32Readtest(t, "bid64_to_binary32", func(x uint64, rndMode int) (uint32, uint32) {
		return Bid64ToBinary32(x, rndMode)
	})
}

func parseBid128BinaryReadtestInput(s string, hexPattern *regexp.Regexp) (BID_UINT128, bool) {
	return parseBid128HexLocal(strings.ReplaceAll(s, ",", ""), hexPattern)
}

func TestBid64ToBinary64IntelReadtest(t *testing.T) {
	testBid64Binary64Readtest(t, "bid64_to_binary64", func(x uint64, rndMode int) (uint64, uint32) {
		return Bid64ToBinary64(x, rndMode)
	})
}

func TestBid64ToBinary128IntelReadtest(t *testing.T) {
	testBid64Binary128Readtest(t, "bid64_to_binary128", func(x uint64, rndMode int) (BID_UINT128, uint32) {
		return Bid64ToBinary128(x, rndMode)
	})
}

func TestBid128ToBinary128IntelReadtest(t *testing.T) {
	testBid128Binary128Readtest(t, "bid128_to_binary128", func(x BID_UINT128, rndMode int) (BID_UINT128, uint32) {
		return Bid128ToBinary128(x, rndMode)
	})
}
