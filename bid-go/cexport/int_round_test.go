package main

import (
	"bufio"
	bidgo "github.com/sky1core/bid754/bid-go"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func parseCexportBid64Input(s string, hexPattern *regexp.Regexp) (uint64, bool) {
	if matches := hexPattern.FindStringSubmatch(s); matches != nil {
		x, err := strconv.ParseUint(matches[1], 16, 64)
		if err != nil {
			return 0, false
		}
		return x, true
	}
	x, _ := bidgo.Bid64FromString(s, 0)
	return x, true
}

func parseCexportFlags(s string) (uint32, bool) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, false
	}
	return uint32(v), true
}

func testIntRoundReadtest(t *testing.T, prefix string, fn func(uint64, int) (int64, uint32)) {
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
		if strings.Contains(line, "longintsize=32") && bidSizeofLong() != 4 {
			skipped++
			continue
		}
		if strings.Contains(line, "longintsize=64") && bidSizeofLong() != 8 {
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
		a, ok1 := parseCexportBid64Input(parts[2], hexPattern)
		expected, err := strconv.ParseInt(parts[3], 10, 64)
		expectedFlags, ok3 := parseCexportFlags(parts[4])
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

	t.Logf("%s: %d passed, %d failed, %d skipped", prefix, passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("%s: %d tests failed", prefix, failed)
	}
}

func TestBid64LlrintIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_llrint", func(x uint64, rndMode int) (int64, uint32) {
		return bid64Llrint(x, rndMode)
	})
}

func TestBid64LrintIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_lrint", func(x uint64, rndMode int) (int64, uint32) {
		return bid64Lrint(x, rndMode)
	})
}

func TestBid64LlroundIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_llround", func(x uint64, _ int) (int64, uint32) {
		return bid64Llround(x)
	})
}

func TestBid64LroundIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_lround", func(x uint64, _ int) (int64, uint32) {
		return bid64Lround(x)
	})
}

func TestBid64ToInt32RnintIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int32_rnint", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt32Rnint(x)
	})
}

func TestBid64ToInt32RnintaIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int32_rninta", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt32Rninta(x)
	})
}

func TestBid64ToInt32FloorIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int32_floor", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt32Floor(x)
	})
}

func TestBid64ToInt32CeilIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int32_ceil", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt32Ceil(x)
	})
}

func TestBid64ToInt32IntIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int32_int", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt32Int(x)
	})
}

func TestBid64ToInt32XrnintIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int32_xrnint", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt32Xrnint(x)
	})
}

func TestBid64ToInt32XrnintaIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int32_xrninta", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt32Xrninta(x)
	})
}

func TestBid64ToInt32XfloorIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int32_xfloor", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt32Xfloor(x)
	})
}

func TestBid64ToInt32XceilIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int32_xceil", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt32Xceil(x)
	})
}

func TestBid64ToInt32XintIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int32_xint", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt32Xint(x)
	})
}

func testUintRoundReadtest(t *testing.T, prefix string, fn func(uint64, int) (uint64, uint32)) {
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
		expected, err := strconv.ParseUint(parts[3], 10, 64)
		expectedFlags, ok3 := parseCexportFlags(parts[4])
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

	t.Logf("%s: %d passed, %d failed, %d skipped", prefix, passed, failed, skipped)
	if failed > 0 {
		t.Fatalf("%s: %d tests failed", prefix, failed)
	}
}

func TestBid64ToInt64RnintIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int64_rnint", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt64Rnint(x)
	})
}

func TestBid64ToInt64RnintaIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int64_rninta", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt64Rninta(x)
	})
}

func TestBid64ToInt64FloorIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int64_floor", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt64Floor(x)
	})
}

func TestBid64ToInt64CeilIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int64_ceil", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt64Ceil(x)
	})
}

func TestBid64ToInt64IntIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int64_int", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt64Int(x)
	})
}

func TestBid64ToInt64XrnintIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int64_xrnint", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt64Xrnint(x)
	})
}

func TestBid64ToInt64XrnintaIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int64_xrninta", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt64Xrninta(x)
	})
}

func TestBid64ToInt64XfloorIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int64_xfloor", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt64Xfloor(x)
	})
}

func TestBid64ToInt64XceilIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int64_xceil", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt64Xceil(x)
	})
}

func TestBid64ToInt64XintIntelReadtest(t *testing.T) {
	testIntRoundReadtest(t, "bid64_to_int64_xint", func(x uint64, _ int) (int64, uint32) {
		return bid64ToInt64Xint(x)
	})
}

func TestBid64ToUint64RnintIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint64_rnint", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint64Rnint(x)
	})
}

func TestBid64ToUint64RnintaIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint64_rninta", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint64Rninta(x)
	})
}

func TestBid64ToUint64FloorIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint64_floor", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint64Floor(x)
	})
}

func TestBid64ToUint64CeilIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint64_ceil", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint64Ceil(x)
	})
}

func TestBid64ToUint64IntIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint64_int", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint64Int(x)
	})
}

func TestBid64ToUint64XrnintIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint64_xrnint", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint64Xrnint(x)
	})
}

func TestBid64ToUint64XrnintaIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint64_xrninta", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint64Xrninta(x)
	})
}

func TestBid64ToUint64XfloorIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint64_xfloor", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint64Xfloor(x)
	})
}

func TestBid64ToUint64XceilIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint64_xceil", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint64Xceil(x)
	})
}

func TestBid64ToUint64XintIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint64_xint", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint64Xint(x)
	})
}

func TestBid64ToUint32RnintIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint32_rnint", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint32Rnint(x)
	})
}

func TestBid64ToUint32RnintaIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint32_rninta", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint32Rninta(x)
	})
}

func TestBid64ToUint32FloorIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint32_floor", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint32Floor(x)
	})
}

func TestBid64ToUint32CeilIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint32_ceil", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint32Ceil(x)
	})
}

func TestBid64ToUint32IntIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint32_int", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint32Int(x)
	})
}

func TestBid64ToUint32XrnintIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint32_xrnint", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint32Xrnint(x)
	})
}

func TestBid64ToUint32XrnintaIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint32_xrninta", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint32Xrninta(x)
	})
}

func TestBid64ToUint32XfloorIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint32_xfloor", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint32Xfloor(x)
	})
}

func TestBid64ToUint32XceilIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint32_xceil", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint32Xceil(x)
	})
}

func TestBid64ToUint32XintIntelReadtest(t *testing.T) {
	testUintRoundReadtest(t, "bid64_to_uint32_xint", func(x uint64, _ int) (uint64, uint32) {
		return bid64ToUint32Xint(x)
	})
}

func TestBid64Int8IntelReadtest(t *testing.T) {
	testCases := []struct {
		name   string
		prefix string
		fn     func(uint64) (int64, uint32)
	}{
		{"rnint", "bid64_to_int8_rnint", bid64ToInt8Rnint},
		{"rninta", "bid64_to_int8_rninta", bid64ToInt8Rninta},
		{"floor", "bid64_to_int8_floor", bid64ToInt8Floor},
		{"ceil", "bid64_to_int8_ceil", bid64ToInt8Ceil},
		{"int", "bid64_to_int8_int", bid64ToInt8Int},
		{"xrnint", "bid64_to_int8_xrnint", bid64ToInt8Xrnint},
		{"xrninta", "bid64_to_int8_xrninta", bid64ToInt8Xrninta},
		{"xfloor", "bid64_to_int8_xfloor", bid64ToInt8Xfloor},
		{"xceil", "bid64_to_int8_xceil", bid64ToInt8Xceil},
		{"xint", "bid64_to_int8_xint", bid64ToInt8Xint},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testIntRoundReadtest(t, tc.prefix, func(x uint64, _ int) (int64, uint32) {
				return tc.fn(x)
			})
		})
	}
}

func TestBid64Int16IntelReadtest(t *testing.T) {
	testCases := []struct {
		name   string
		prefix string
		fn     func(uint64) (int64, uint32)
	}{
		{"rnint", "bid64_to_int16_rnint", bid64ToInt16Rnint},
		{"rninta", "bid64_to_int16_rninta", bid64ToInt16Rninta},
		{"floor", "bid64_to_int16_floor", bid64ToInt16Floor},
		{"ceil", "bid64_to_int16_ceil", bid64ToInt16Ceil},
		{"int", "bid64_to_int16_int", bid64ToInt16Int},
		{"xrnint", "bid64_to_int16_xrnint", bid64ToInt16Xrnint},
		{"xrninta", "bid64_to_int16_xrninta", bid64ToInt16Xrninta},
		{"xfloor", "bid64_to_int16_xfloor", bid64ToInt16Xfloor},
		{"xceil", "bid64_to_int16_xceil", bid64ToInt16Xceil},
		{"xint", "bid64_to_int16_xint", bid64ToInt16Xint},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testIntRoundReadtest(t, tc.prefix, func(x uint64, _ int) (int64, uint32) {
				return tc.fn(x)
			})
		})
	}
}

func TestBid64Uint8IntelReadtest(t *testing.T) {
	testCases := []struct {
		name   string
		prefix string
		fn     func(uint64) (uint64, uint32)
	}{
		{"rnint", "bid64_to_uint8_rnint", bid64ToUint8Rnint},
		{"rninta", "bid64_to_uint8_rninta", bid64ToUint8Rninta},
		{"floor", "bid64_to_uint8_floor", bid64ToUint8Floor},
		{"ceil", "bid64_to_uint8_ceil", bid64ToUint8Ceil},
		{"int", "bid64_to_uint8_int", bid64ToUint8Int},
		{"xrnint", "bid64_to_uint8_xrnint", bid64ToUint8Xrnint},
		{"xrninta", "bid64_to_uint8_xrninta", bid64ToUint8Xrninta},
		{"xfloor", "bid64_to_uint8_xfloor", bid64ToUint8Xfloor},
		{"xceil", "bid64_to_uint8_xceil", bid64ToUint8Xceil},
		{"xint", "bid64_to_uint8_xint", bid64ToUint8Xint},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testUintRoundReadtest(t, tc.prefix, func(x uint64, _ int) (uint64, uint32) {
				return tc.fn(x)
			})
		})
	}
}

func TestBid64Uint16IntelReadtest(t *testing.T) {
	testCases := []struct {
		name   string
		prefix string
		fn     func(uint64) (uint64, uint32)
	}{
		{"rnint", "bid64_to_uint16_rnint", bid64ToUint16Rnint},
		{"rninta", "bid64_to_uint16_rninta", bid64ToUint16Rninta},
		{"floor", "bid64_to_uint16_floor", bid64ToUint16Floor},
		{"ceil", "bid64_to_uint16_ceil", bid64ToUint16Ceil},
		{"int", "bid64_to_uint16_int", bid64ToUint16Int},
		{"xrnint", "bid64_to_uint16_xrnint", bid64ToUint16Xrnint},
		{"xrninta", "bid64_to_uint16_xrninta", bid64ToUint16Xrninta},
		{"xfloor", "bid64_to_uint16_xfloor", bid64ToUint16Xfloor},
		{"xceil", "bid64_to_uint16_xceil", bid64ToUint16Xceil},
		{"xint", "bid64_to_uint16_xint", bid64ToUint16Xint},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testUintRoundReadtest(t, tc.prefix, func(x uint64, _ int) (uint64, uint32) {
				return tc.fn(x)
			})
		})
	}
}
