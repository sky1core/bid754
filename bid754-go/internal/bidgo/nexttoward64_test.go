package bidgo

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestBid64NextTowardIntelReadtest(t *testing.T) {
	file, err := os.Open("../third_party/intel_dfp/TESTS/readtest.in")
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
		if !strings.HasPrefix(line, "bid64_nexttoward ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 6 {
			skipped++
			continue
		}

		x, ok1 := parseBid64InputLocal(parts[2], hexPattern)
		y, ok2 := parseBid128HexLocal(parts[3], hexPattern)
		expected, ok3 := parseBid64HexLocal(parts[4], hexPattern)
		expectedFlags, ok4 := parseBidReadtestFlags(parts[5])
		if !ok1 || !ok2 || !ok3 || !ok4 {
			skipped++
			continue
		}

		result, actualFlags := Bid64NextToward(x, y)
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

	t.Logf("bid64_nexttoward: %d passed, %d failed, %d skipped", passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("bid64_nexttoward: %d tests failed", failed)
	}
}
