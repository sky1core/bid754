package bidgo

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func parseBid64ReadtestInput(s string, hexPattern *regexp.Regexp) (uint64, bool) {
	if matches := hexPattern.FindStringSubmatch(s); matches != nil {
		x, err := strconv.ParseUint(matches[1], 16, 64)
		if err != nil {
			return 0, false
		}
		return x, true
	}
	x, _ := Bid64FromString(s, 0)
	return x, true
}

func parseBidReadtestFlags(s string) (uint32, bool) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, false
	}
	return uint32(v), true
}

func parseReadtestInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseReadtestUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

func parseReadtestUint32(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	return uint32(v), err
}

func parseReadtestUint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 10, 8)
	return uint8(v), err
}

func parseReadtestUint16(s string) (uint16, error) {
	v, err := strconv.ParseUint(s, 10, 16)
	return uint16(v), err
}

func testUint16Readtest(t *testing.T, prefix string, fn func(uint64) (uint16, uint32)) {
	t.Helper()

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
		if !strings.HasPrefix(line, prefix+" ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 5 {
			skipped++
			continue
		}

		a, ok1 := parseBid64ReadtestInput(parts[2], hexPattern)
		expected, err := parseReadtestUint16(parts[3])
		expectedFlags, ok3 := parseBidReadtestFlags(parts[4])
		if !ok1 || err != nil || !ok3 {
			skipped++
			continue
		}

		result, actualFlags := fn(a)
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

func testUint8Readtest(t *testing.T, prefix string, fn func(uint64) (uint8, uint32)) {
	t.Helper()

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
		if !strings.HasPrefix(line, prefix+" ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 5 {
			skipped++
			continue
		}

		a, ok1 := parseBid64ReadtestInput(parts[2], hexPattern)
		expected, err := parseReadtestUint8(parts[3])
		expectedFlags, ok3 := parseBidReadtestFlags(parts[4])
		if !ok1 || err != nil || !ok3 {
			skipped++
			continue
		}

		result, actualFlags := fn(a)
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

func testUint32Readtest(t *testing.T, prefix string, fn func(uint64) (uint32, uint32)) {
	t.Helper()

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
		if !strings.HasPrefix(line, prefix+" ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 5 {
			skipped++
			continue
		}

		a, ok1 := parseBid64ReadtestInput(parts[2], hexPattern)
		expected, err := parseReadtestUint32(parts[3])
		expectedFlags, ok3 := parseBidReadtestFlags(parts[4])
		if !ok1 || err != nil || !ok3 {
			skipped++
			continue
		}

		result, actualFlags := fn(a)
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

func testUint64Readtest(t *testing.T, prefix string, fn func(uint64) (uint64, uint32)) {
	t.Helper()

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
		if !strings.HasPrefix(line, prefix+" ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 5 {
			skipped++
			continue
		}

		a, ok1 := parseBid64ReadtestInput(parts[2], hexPattern)
		expected, err := parseReadtestUint64(parts[3])
		expectedFlags, ok3 := parseBidReadtestFlags(parts[4])
		if !ok1 || err != nil || !ok3 {
			skipped++
			continue
		}

		result, actualFlags := fn(a)
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

func testSignedIntReadtest[T ~int8 | ~int16 | ~int32 | ~int64](t *testing.T, prefix string, parse func(string) (int64, error), fn func(uint64) (T, uint32)) {
	t.Helper()

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
		if !strings.HasPrefix(line, prefix+" ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 5 {
			skipped++
			continue
		}

		a, ok1 := parseBid64ReadtestInput(parts[2], hexPattern)
		expected, err := parse(parts[3])
		expectedFlags, ok3 := parseBidReadtestFlags(parts[4])
		if !ok1 || err != nil || !ok3 {
			skipped++
			continue
		}

		result, actualFlags := fn(a)
		if int64(result) == expected && actualFlags == expectedFlags {
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

func TestBid64ToInt32RnintIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int32_rnint", parseReadtestInt64, func(x uint64) (int32, uint32) {
		return Bid64ToInt32Rnint(x)
	})
}

func TestBid64ToInt32XrnintIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int32_xrnint", parseReadtestInt64, func(x uint64) (int32, uint32) {
		return Bid64ToInt32Xrnint(x)
	})
}

func TestBid64ToInt8RnintIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int8_rnint", parseReadtestInt64, func(x uint64) (int8, uint32) {
		return Bid64ToInt8Rnint(x)
	})
}

func TestBid64ToInt8XrnintIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int8_xrnint", parseReadtestInt64, func(x uint64) (int8, uint32) {
		return Bid64ToInt8Xrnint(x)
	})
}

func TestBid64ToInt16RnintIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int16_rnint", parseReadtestInt64, func(x uint64) (int16, uint32) {
		return Bid64ToInt16Rnint(x)
	})
}

func TestBid64ToInt16XrnintIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int16_xrnint", parseReadtestInt64, func(x uint64) (int16, uint32) {
		return Bid64ToInt16Xrnint(x)
	})
}

func TestBid64ToInt32RnintaIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int32_rninta", parseReadtestInt64, func(x uint64) (int32, uint32) {
		return Bid64ToInt32Rninta(x)
	})
}

func TestBid64ToInt32XrnintaIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int32_xrninta", parseReadtestInt64, func(x uint64) (int32, uint32) {
		return Bid64ToInt32Xrninta(x)
	})
}

func TestBid64ToInt8RnintaIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int8_rninta", parseReadtestInt64, func(x uint64) (int8, uint32) {
		return Bid64ToInt8Rninta(x)
	})
}

func TestBid64ToInt8XrnintaIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int8_xrninta", parseReadtestInt64, func(x uint64) (int8, uint32) {
		return Bid64ToInt8Xrninta(x)
	})
}

func TestBid64ToInt16RnintaIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int16_rninta", parseReadtestInt64, func(x uint64) (int16, uint32) {
		return Bid64ToInt16Rninta(x)
	})
}

func TestBid64ToInt16XrnintaIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int16_xrninta", parseReadtestInt64, func(x uint64) (int16, uint32) {
		return Bid64ToInt16Xrninta(x)
	})
}

func TestBid64ToInt32IntIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int32_int", parseReadtestInt64, func(x uint64) (int32, uint32) {
		return Bid64ToInt32Int(x)
	})
}

func TestBid64ToInt32XintIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int32_xint", parseReadtestInt64, func(x uint64) (int32, uint32) {
		return Bid64ToInt32Xint(x)
	})
}

func TestBid64ToInt8IntIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int8_int", parseReadtestInt64, func(x uint64) (int8, uint32) {
		return Bid64ToInt8Int(x)
	})
}

func TestBid64ToInt8XintIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int8_xint", parseReadtestInt64, func(x uint64) (int8, uint32) {
		return Bid64ToInt8Xint(x)
	})
}

func TestBid64ToInt16IntIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int16_int", parseReadtestInt64, func(x uint64) (int16, uint32) {
		return Bid64ToInt16Int(x)
	})
}

func TestBid64ToInt16XintIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int16_xint", parseReadtestInt64, func(x uint64) (int16, uint32) {
		return Bid64ToInt16Xint(x)
	})
}

func TestBid64ToInt32FloorIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int32_floor", parseReadtestInt64, func(x uint64) (int32, uint32) {
		return Bid64ToInt32Floor(x)
	})
}

func TestBid64ToInt32XfloorIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int32_xfloor", parseReadtestInt64, func(x uint64) (int32, uint32) {
		return Bid64ToInt32Xfloor(x)
	})
}

func TestBid64ToInt8FloorIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int8_floor", parseReadtestInt64, func(x uint64) (int8, uint32) {
		return Bid64ToInt8Floor(x)
	})
}

func TestBid64ToInt8XfloorIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int8_xfloor", parseReadtestInt64, func(x uint64) (int8, uint32) {
		return Bid64ToInt8Xfloor(x)
	})
}

func TestBid64ToInt16FloorIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int16_floor", parseReadtestInt64, func(x uint64) (int16, uint32) {
		return Bid64ToInt16Floor(x)
	})
}

func TestBid64ToInt16XfloorIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int16_xfloor", parseReadtestInt64, func(x uint64) (int16, uint32) {
		return Bid64ToInt16Xfloor(x)
	})
}

func TestBid64ToInt32CeilIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int32_ceil", parseReadtestInt64, func(x uint64) (int32, uint32) {
		return Bid64ToInt32Ceil(x)
	})
}

func TestBid64ToInt32XceilIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int32_xceil", parseReadtestInt64, func(x uint64) (int32, uint32) {
		return Bid64ToInt32Xceil(x)
	})
}

func TestBid64ToInt8CeilIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int8_ceil", parseReadtestInt64, func(x uint64) (int8, uint32) {
		return Bid64ToInt8Ceil(x)
	})
}

func TestBid64ToInt8XceilIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int8_xceil", parseReadtestInt64, func(x uint64) (int8, uint32) {
		return Bid64ToInt8Xceil(x)
	})
}

func TestBid64ToInt16CeilIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int16_ceil", parseReadtestInt64, func(x uint64) (int16, uint32) {
		return Bid64ToInt16Ceil(x)
	})
}

func TestBid64ToInt16XceilIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int16_xceil", parseReadtestInt64, func(x uint64) (int16, uint32) {
		return Bid64ToInt16Xceil(x)
	})
}

func TestBid64ToInt64RnintIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int64_rnint", parseReadtestInt64, func(x uint64) (int64, uint32) {
		return Bid64ToInt64Rnint(x)
	})
}

func TestBid64ToInt64XrnintIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int64_xrnint", parseReadtestInt64, func(x uint64) (int64, uint32) {
		return Bid64ToInt64Xrnint(x)
	})
}

func TestBid64ToInt64RnintaIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int64_rninta", parseReadtestInt64, func(x uint64) (int64, uint32) {
		return Bid64ToInt64Rninta(x)
	})
}

func TestBid64ToInt64XrnintaIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int64_xrninta", parseReadtestInt64, func(x uint64) (int64, uint32) {
		return Bid64ToInt64Xrninta(x)
	})
}

func TestBid64ToInt64IntIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int64_int", parseReadtestInt64, func(x uint64) (int64, uint32) {
		return Bid64ToInt64Int(x)
	})
}

func TestBid64ToInt64XintIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int64_xint", parseReadtestInt64, func(x uint64) (int64, uint32) {
		return Bid64ToInt64Xint(x)
	})
}

func TestBid64ToInt64FloorIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int64_floor", parseReadtestInt64, func(x uint64) (int64, uint32) {
		return Bid64ToInt64Floor(x)
	})
}

func TestBid64ToInt64XfloorIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int64_xfloor", parseReadtestInt64, func(x uint64) (int64, uint32) {
		return Bid64ToInt64Xfloor(x)
	})
}

func TestBid64ToInt64CeilIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int64_ceil", parseReadtestInt64, func(x uint64) (int64, uint32) {
		return Bid64ToInt64Ceil(x)
	})
}

func TestBid64ToInt64XceilIntelReadtest(t *testing.T) {
	testSignedIntReadtest(t, "bid64_to_int64_xceil", parseReadtestInt64, func(x uint64) (int64, uint32) {
		return Bid64ToInt64Xceil(x)
	})
}

func TestBid64ToUint64RnintIntelReadtest(t *testing.T) {
	testUint64Readtest(t, "bid64_to_uint64_rnint", func(x uint64) (uint64, uint32) {
		return Bid64ToUint64Rnint(x)
	})
}

func TestBid64ToUint64XrnintIntelReadtest(t *testing.T) {
	testUint64Readtest(t, "bid64_to_uint64_xrnint", func(x uint64) (uint64, uint32) {
		return Bid64ToUint64Xrnint(x)
	})
}

func TestBid64ToUint64RnintaIntelReadtest(t *testing.T) {
	testUint64Readtest(t, "bid64_to_uint64_rninta", func(x uint64) (uint64, uint32) {
		return Bid64ToUint64Rninta(x)
	})
}

func TestBid64ToUint64XrnintaIntelReadtest(t *testing.T) {
	testUint64Readtest(t, "bid64_to_uint64_xrninta", func(x uint64) (uint64, uint32) {
		return Bid64ToUint64Xrninta(x)
	})
}

func TestBid64ToUint64IntIntelReadtest(t *testing.T) {
	testUint64Readtest(t, "bid64_to_uint64_int", func(x uint64) (uint64, uint32) {
		return Bid64ToUint64Int(x)
	})
}

func TestBid64ToUint64XintIntelReadtest(t *testing.T) {
	testUint64Readtest(t, "bid64_to_uint64_xint", func(x uint64) (uint64, uint32) {
		return Bid64ToUint64Xint(x)
	})
}

func TestBid64ToUint64FloorIntelReadtest(t *testing.T) {
	testUint64Readtest(t, "bid64_to_uint64_floor", func(x uint64) (uint64, uint32) {
		return Bid64ToUint64Floor(x)
	})
}

func TestBid64ToUint64XfloorIntelReadtest(t *testing.T) {
	testUint64Readtest(t, "bid64_to_uint64_xfloor", func(x uint64) (uint64, uint32) {
		return Bid64ToUint64Xfloor(x)
	})
}

func TestBid64ToUint64CeilIntelReadtest(t *testing.T) {
	testUint64Readtest(t, "bid64_to_uint64_ceil", func(x uint64) (uint64, uint32) {
		return Bid64ToUint64Ceil(x)
	})
}

func TestBid64ToUint64XceilIntelReadtest(t *testing.T) {
	testUint64Readtest(t, "bid64_to_uint64_xceil", func(x uint64) (uint64, uint32) {
		return Bid64ToUint64Xceil(x)
	})
}

func TestBid64ToUint32RnintIntelReadtest(t *testing.T) {
	testUint32Readtest(t, "bid64_to_uint32_rnint", func(x uint64) (uint32, uint32) {
		return Bid64ToUint32Rnint(x)
	})
}

func TestBid64ToUint32XrnintIntelReadtest(t *testing.T) {
	testUint32Readtest(t, "bid64_to_uint32_xrnint", func(x uint64) (uint32, uint32) {
		return Bid64ToUint32Xrnint(x)
	})
}

func TestBid64ToUint32RnintaIntelReadtest(t *testing.T) {
	testUint32Readtest(t, "bid64_to_uint32_rninta", func(x uint64) (uint32, uint32) {
		return Bid64ToUint32Rninta(x)
	})
}

func TestBid64ToUint32XrnintaIntelReadtest(t *testing.T) {
	testUint32Readtest(t, "bid64_to_uint32_xrninta", func(x uint64) (uint32, uint32) {
		return Bid64ToUint32Xrninta(x)
	})
}

func TestBid64ToUint32IntIntelReadtest(t *testing.T) {
	testUint32Readtest(t, "bid64_to_uint32_int", func(x uint64) (uint32, uint32) {
		return Bid64ToUint32Int(x)
	})
}

func TestBid64ToUint32XintIntelReadtest(t *testing.T) {
	testUint32Readtest(t, "bid64_to_uint32_xint", func(x uint64) (uint32, uint32) {
		return Bid64ToUint32Xint(x)
	})
}

func TestBid64ToUint32FloorIntelReadtest(t *testing.T) {
	testUint32Readtest(t, "bid64_to_uint32_floor", func(x uint64) (uint32, uint32) {
		return Bid64ToUint32Floor(x)
	})
}

func TestBid64ToUint32XfloorIntelReadtest(t *testing.T) {
	testUint32Readtest(t, "bid64_to_uint32_xfloor", func(x uint64) (uint32, uint32) {
		return Bid64ToUint32Xfloor(x)
	})
}

func TestBid64ToUint32CeilIntelReadtest(t *testing.T) {
	testUint32Readtest(t, "bid64_to_uint32_ceil", func(x uint64) (uint32, uint32) {
		return Bid64ToUint32Ceil(x)
	})
}

func TestBid64ToUint32XceilIntelReadtest(t *testing.T) {
	testUint32Readtest(t, "bid64_to_uint32_xceil", func(x uint64) (uint32, uint32) {
		return Bid64ToUint32Xceil(x)
	})
}

func TestBid64ToUint8RnintIntelReadtest(t *testing.T) {
	testUint8Readtest(t, "bid64_to_uint8_rnint", func(x uint64) (uint8, uint32) {
		return Bid64ToUint8Rnint(x)
	})
}

func TestBid64ToUint8XrnintIntelReadtest(t *testing.T) {
	testUint8Readtest(t, "bid64_to_uint8_xrnint", func(x uint64) (uint8, uint32) {
		return Bid64ToUint8Xrnint(x)
	})
}

func TestBid64ToUint8RnintaIntelReadtest(t *testing.T) {
	testUint8Readtest(t, "bid64_to_uint8_rninta", func(x uint64) (uint8, uint32) {
		return Bid64ToUint8Rninta(x)
	})
}

func TestBid64ToUint8XrnintaIntelReadtest(t *testing.T) {
	testUint8Readtest(t, "bid64_to_uint8_xrninta", func(x uint64) (uint8, uint32) {
		return Bid64ToUint8Xrninta(x)
	})
}

func TestBid64ToUint8IntIntelReadtest(t *testing.T) {
	testUint8Readtest(t, "bid64_to_uint8_int", func(x uint64) (uint8, uint32) {
		return Bid64ToUint8Int(x)
	})
}

func TestBid64ToUint8XintIntelReadtest(t *testing.T) {
	testUint8Readtest(t, "bid64_to_uint8_xint", func(x uint64) (uint8, uint32) {
		return Bid64ToUint8Xint(x)
	})
}

func TestBid64ToUint8FloorIntelReadtest(t *testing.T) {
	testUint8Readtest(t, "bid64_to_uint8_floor", func(x uint64) (uint8, uint32) {
		return Bid64ToUint8Floor(x)
	})
}

func TestBid64ToUint8XfloorIntelReadtest(t *testing.T) {
	testUint8Readtest(t, "bid64_to_uint8_xfloor", func(x uint64) (uint8, uint32) {
		return Bid64ToUint8Xfloor(x)
	})
}

func TestBid64ToUint8CeilIntelReadtest(t *testing.T) {
	testUint8Readtest(t, "bid64_to_uint8_ceil", func(x uint64) (uint8, uint32) {
		return Bid64ToUint8Ceil(x)
	})
}

func TestBid64ToUint8XceilIntelReadtest(t *testing.T) {
	testUint8Readtest(t, "bid64_to_uint8_xceil", func(x uint64) (uint8, uint32) {
		return Bid64ToUint8Xceil(x)
	})
}

func TestBid64ToUint16RnintIntelReadtest(t *testing.T) {
	testUint16Readtest(t, "bid64_to_uint16_rnint", func(x uint64) (uint16, uint32) {
		return Bid64ToUint16Rnint(x)
	})
}

func TestBid64ToUint16XrnintIntelReadtest(t *testing.T) {
	testUint16Readtest(t, "bid64_to_uint16_xrnint", func(x uint64) (uint16, uint32) {
		return Bid64ToUint16Xrnint(x)
	})
}

func TestBid64ToUint16RnintaIntelReadtest(t *testing.T) {
	testUint16Readtest(t, "bid64_to_uint16_rninta", func(x uint64) (uint16, uint32) {
		return Bid64ToUint16Rninta(x)
	})
}

func TestBid64ToUint16XrnintaIntelReadtest(t *testing.T) {
	testUint16Readtest(t, "bid64_to_uint16_xrninta", func(x uint64) (uint16, uint32) {
		return Bid64ToUint16Xrninta(x)
	})
}

func TestBid64ToUint16IntIntelReadtest(t *testing.T) {
	testUint16Readtest(t, "bid64_to_uint16_int", func(x uint64) (uint16, uint32) {
		return Bid64ToUint16Int(x)
	})
}

func TestBid64ToUint16XintIntelReadtest(t *testing.T) {
	testUint16Readtest(t, "bid64_to_uint16_xint", func(x uint64) (uint16, uint32) {
		return Bid64ToUint16Xint(x)
	})
}

func TestBid64ToUint16FloorIntelReadtest(t *testing.T) {
	testUint16Readtest(t, "bid64_to_uint16_floor", func(x uint64) (uint16, uint32) {
		return Bid64ToUint16Floor(x)
	})
}

func TestBid64ToUint16CeilIntelReadtest(t *testing.T) {
	testUint16Readtest(t, "bid64_to_uint16_ceil", func(x uint64) (uint16, uint32) {
		return Bid64ToUint16Ceil(x)
	})
}

func TestBid64ToUint16XfloorIntelReadtest(t *testing.T) {
	testUint16Readtest(t, "bid64_to_uint16_xfloor", func(x uint64) (uint16, uint32) {
		return Bid64ToUint16Xfloor(x)
	})
}

func TestBid64ToUint16XceilIntelReadtest(t *testing.T) {
	testUint16Readtest(t, "bid64_to_uint16_xceil", func(x uint64) (uint16, uint32) {
		return Bid64ToUint16Xceil(x)
	})
}
