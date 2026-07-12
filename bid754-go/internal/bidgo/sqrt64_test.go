package bidgo

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestBid64SqrtIntelReadtest(t *testing.T) {
	testFile := "../../../devtools/third_party/intel_dfp/TESTS/readtest.in"

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
		if !strings.HasPrefix(line, "bid64_sqrt ") {
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
		a, ok1 := parseBid64InputLocal(parts[2], hexPattern)
		expected, ok2 := parseBid64InputLocal(parts[3], hexPattern)
		expectedFlags, ok3 := parseFlagsLocal(parts[4])
		if !ok1 || !ok2 || !ok3 {
			skipped++
			continue
		}

		result, actualFlags := Bid64Sqrt(a, rndMode)
		if result == expected && actualFlags == expectedFlags {
			passed++
		} else {
			failed++
			if failed <= 10 {
				t.Errorf("%s -> got=[%016x]/%02x want=[%016x]/%02x",
					line, result, actualFlags, expected, expectedFlags)
			}
		}
	}

	t.Logf("bid64_sqrt: %d passed, %d failed, %d skipped", passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("bid64_sqrt: %d tests failed", failed)
	}
}
