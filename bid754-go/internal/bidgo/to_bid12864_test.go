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

func TestBid64ToBid128IntelReadtest(t *testing.T) {
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
		if !strings.HasPrefix(line, "bid64_to_bid128 ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 5 {
			skipped++
			continue
		}

		a, ok1 := parseBid64InputLocal(parts[2], hexPattern)
		expected, ok2 := parseBid128HexLocal(parts[3], hexPattern)
		expectedFlags, ok3 := parseFlagsLocal(parts[4])
		if !ok1 || !ok2 || !ok3 {
			skipped++
			continue
		}

		result, actualFlags := Bid64ToBid128(a)
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

	t.Logf("bid64_to_bid128: %d passed, %d failed, %d skipped", passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("bid64_to_bid128: %d tests failed", failed)
	}
}

func parseFlagsLocal(s string) (uint32, bool) {
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, false
	}
	return uint32(v), true
}

func parseBid64HexLocal(s string, hexPattern *regexp.Regexp) (uint64, bool) {
	matches := hexPattern.FindStringSubmatch(s)
	if matches == nil {
		return 0, false
	}

	hexStr := matches[1]
	for len(hexStr) < 16 {
		hexStr = "0" + hexStr
	}

	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, false
	}

	var val uint64
	for _, b := range bytes {
		val = (val << 8) | uint64(b)
	}
	return val, true
}

func parseBid64InputLocal(s string, hexPattern *regexp.Regexp) (uint64, bool) {
	if matches := hexPattern.FindStringSubmatch(s); matches != nil {
		return parseBid64HexLocal(s, hexPattern)
	}
	val, _ := Bid64FromString(s, 0)
	return val, true
}

func parseBid128HexLocal(s string, hexPattern *regexp.Regexp) (BID_UINT128, bool) {
	var val BID_UINT128

	matches := hexPattern.FindStringSubmatch(s)
	if matches == nil {
		return val, false
	}

	hexStr := matches[1]
	for len(hexStr) < 32 {
		hexStr = "0" + hexStr
	}

	bytes, err := hex.DecodeString(hexStr)
	if err != nil || len(bytes) != 16 {
		return val, false
	}

	for _, b := range bytes[:8] {
		val.w[1] = (val.w[1] << 8) | uint64(b)
	}
	for _, b := range bytes[8:] {
		val.w[0] = (val.w[0] << 8) | uint64(b)
	}
	return val, true
}
