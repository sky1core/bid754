package bidgo

import (
	"bufio"
	"encoding/hex"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestBid64ToBid32IntelReadtest(t *testing.T) {
	testFile := "../third_party/intel_dfp/TESTS/readtest.in"

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
		if !strings.HasPrefix(line, "bid64_to_bid32 ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 5 {
			skipped++
			continue
		}

		roundingMode, err := strconv.Atoi(parts[1])
		if err != nil || roundingMode < 0 || roundingMode > 5 {
			skipped++
			continue
		}

		a, ok1 := parseBid64InputLocal(parts[2], hexPattern)
		expected, ok2 := parseBid32HexLocal(parts[3], hexPattern)
		expectedFlags, ok3 := parseFlagsLocal(parts[4])
		if !ok1 || !ok2 || !ok3 {
			skipped++
			continue
		}

		result, actualFlags := Bid64ToBid32(a, roundingMode)
		if result == expected && actualFlags == expectedFlags {
			passed++
		} else {
			failed++
			if failed <= 10 {
				t.Errorf("%s -> got=[%08x]/%02x want=[%08x]/%02x",
					line, result, actualFlags, expected, expectedFlags)
			}
		}
	}

	t.Logf("bid64_to_bid32: %d passed, %d failed, %d skipped", passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("bid64_to_bid32: %d tests failed", failed)
	}
}

func parseBid32HexLocal(s string, hexPattern *regexp.Regexp) (uint32, bool) {
	matches := hexPattern.FindStringSubmatch(s)
	if matches == nil {
		return 0, false
	}

	hexStr := matches[1]
	for len(hexStr) < 8 {
		hexStr = "0" + hexStr
	}

	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, false
	}

	var val uint32
	for _, b := range bytes {
		val = (val << 8) | uint32(b)
	}
	return val, true
}
