package bidgo

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func testBid64IntRoundReadtest(t *testing.T, prefix string, fn func(uint64, int) (int64, uint32)) {
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
		if strings.Contains(line, "longintsize=32") && bidSizeLong() != 4 {
			skipped++
			continue
		}
		if strings.Contains(line, "longintsize=64") && bidSizeLong() != 8 {
			skipped++
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
		expected, err := parseReadtestInt64(parts[3])
		expectedFlags, ok3 := parseBidReadtestFlags(parts[4])
		if !ok1 || err != nil || !ok3 {
			skipped++
			continue
		}

		result, actualFlags := fn(a, rndMode)
		if result == expected && actualFlags == expectedFlags {
			passed++
		} else {
			failed++
			if failed <= 10 {
				t.Errorf("%s -> got=%d/%02x want=%d/%02x", line, result, actualFlags, expected, expectedFlags)
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

func TestBid64LlrintIntelReadtest(t *testing.T) {
	testBid64IntRoundReadtest(t, "bid64_llrint", func(x uint64, rndMode int) (int64, uint32) {
		return Bid64Llrint(x, rndMode)
	})
}

func TestBid64LrintIntelReadtest(t *testing.T) {
	testBid64IntRoundReadtest(t, "bid64_lrint", func(x uint64, rndMode int) (int64, uint32) {
		return Bid64Lrint(x, rndMode)
	})
}

func TestBid64LlroundIntelReadtest(t *testing.T) {
	testBid64IntRoundReadtest(t, "bid64_llround", func(x uint64, _ int) (int64, uint32) {
		return Bid64Llround(x)
	})
}

func TestBid64LroundIntelReadtest(t *testing.T) {
	testBid64IntRoundReadtest(t, "bid64_lround", func(x uint64, _ int) (int64, uint32) {
		return Bid64Lround(x)
	})
}
