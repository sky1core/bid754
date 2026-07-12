package bidgo

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func testBid64RoundIntegralUnaryIntelReadtest(t *testing.T, prefix string, fn func(uint64) (uint64, uint32)) {
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
		if !strings.HasPrefix(line, prefix+" ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 5 {
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

		result, actualFlags := fn(a)
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

	t.Logf("%s: %d passed, %d failed, %d skipped", prefix, passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("%s: %d tests failed", prefix, failed)
	}
}

func TestBid64RoundIntegralExactIntelReadtest(t *testing.T) {
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
		if !strings.HasPrefix(line, "bid64_round_integral_exact ") {
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

		result, actualFlags := Bid64RoundIntegralExact(a, rndMode)
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

	t.Logf("bid64_round_integral_exact: %d passed, %d failed, %d skipped", passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("bid64_round_integral_exact: %d tests failed", failed)
	}
}

func TestBid64NearbyIntIntelReadtest(t *testing.T) {
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
		if !strings.HasPrefix(line, "bid64_nearbyint ") {
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

		result, actualFlags := Bid64NearbyInt(a, rndMode)
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

	t.Logf("bid64_nearbyint: %d passed, %d failed, %d skipped", passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("bid64_nearbyint: %d tests failed", failed)
	}
}

func TestBid64RoundIntegralNearestEvenIntelReadtest(t *testing.T) {
	testBid64RoundIntegralUnaryIntelReadtest(t, "bid64_round_integral_nearest_even", Bid64RoundIntegralNearestEven)
}

func TestBid64RoundIntegralNearestAwayIntelReadtest(t *testing.T) {
	testBid64RoundIntegralUnaryIntelReadtest(t, "bid64_round_integral_nearest_away", Bid64RoundIntegralNearestAway)
}

func TestBid64RoundIntegralNegativeIntelReadtest(t *testing.T) {
	testBid64RoundIntegralUnaryIntelReadtest(t, "bid64_round_integral_negative", Bid64RoundIntegralNegative)
}

func TestBid64RoundIntegralPositiveIntelReadtest(t *testing.T) {
	testBid64RoundIntegralUnaryIntelReadtest(t, "bid64_round_integral_positive", Bid64RoundIntegralPositive)
}
