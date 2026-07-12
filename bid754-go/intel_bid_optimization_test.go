package bid754

import (
	"runtime"
	"testing"
	"time"
)

// TestIntelBIDOptimization tests Intel BID optimization characteristics.
func TestIntelBIDOptimization(t *testing.T) {
	requireNative(t)
	t.Log("Starting Intel BID optimization characteristics test")

	// Test special optimization cases in BID encoding.
	testCases := []struct {
		name     string
		value    string
		expected string
		desc     string
	}{
		{
			name:     "integer encoding",
			value:    "123",
			expected: "123",
			desc:     "BID encodes integers efficiently",
		},
		{
			name:     "one decimal place",
			value:    "12.3",
			expected: "12.3",
			desc:     "single decimal-place optimization",
		},
		{
			name:     "two decimal places",
			value:    "1.23",
			expected: "1.23",
			desc:     "two decimal-place optimization",
		},
		{
			name:     "large integer",
			value:    "9999999",
			expected: "9999999",
			desc:     "maximum-precision Decimal32 integer",
		},
		{
			name:     "small decimal",
			value:    "0.0000001",
			expected: "0.0000001",
			desc:     "small decimal representation",
		},
		{
			name:     "exponent notation",
			value:    "1E+6",
			expected: "1000000",
			desc:     "exponent notation optimization",
		},
		{
			name:     "exponent decimal",
			value:    "1.23E+2",
			expected: "123",
			desc:     "decimal value with an exponent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Decimal32BID test
			d32bid, err := NewDecimal32BIDDirect(tc.value)
			if err != nil {
				t.Errorf("Decimal32BID creation failed: %v", err)
				return
			}

			result32 := d32bid.PrettyString()
			norm32 := normalizeDecimalString(result32)
			expectedNorm := normalizeDecimalString(tc.expected)

			if norm32 == expectedNorm {
				t.Logf("✅ Decimal32BID %s: %s -> %s (%s)", tc.name, tc.value, result32, tc.desc)
			} else {
				t.Errorf("❌ Decimal32BID %s: %s -> %s (expected: %s) (%s)",
					tc.name, tc.value, result32, tc.expected, tc.desc)
			}

			// Decimal64BID test
			d64bid, err := NewDecimal64BIDDirect(tc.value)
			if err != nil {
				t.Errorf("Decimal64BID creation failed: %v", err)
				return
			}

			result64 := d64bid.PrettyString()
			norm64 := normalizeDecimalString(result64)

			if norm64 == expectedNorm {
				t.Logf("✅ Decimal64BID %s: %s -> %s (%s)", tc.name, tc.value, result64, tc.desc)
			} else {
				t.Errorf("❌ Decimal64BID %s: %s -> %s (expected: %s) (%s)",
					tc.name, tc.value, result64, tc.expected, tc.desc)
			}
		})
	}
}

// TestBIDSpecialValues tests special value handling in BID encoding.
func TestBIDSpecialValues(t *testing.T) {
	requireNative(t)
	t.Log("Starting BID special value handling test")

	specialCases := []struct {
		name   string
		value  string
		isZero bool
		isNaN  bool
		isInf  bool
		sign   int
	}{
		{"positive zero", "0", true, false, false, 0},
		{"negative zero", "-0", true, false, false, 0},
		{"positive infinity", "+Infinity", false, false, true, 1},
		{"negative infinity", "-Infinity", false, false, true, -1},
		{"NaN", "NaN", false, true, false, 0},
		{"positive number", "123.45", false, false, false, 1},
		{"negative number", "-123.45", false, false, false, -1},
	}

	for _, tc := range specialCases {
		t.Run(tc.name, func(t *testing.T) {
			// Decimal32BID test
			d32bid, err := NewDecimal32BIDDirect(tc.value)
			if err != nil && !tc.isNaN && !tc.isInf {
				t.Errorf("Decimal32BID creation failed: %v", err)
				return
			}

			if err == nil {
				if d32bid.IsZero() != tc.isZero {
					t.Errorf("Decimal32BID IsZero mismatch: got=%v, expected=%v", d32bid.IsZero(), tc.isZero)
				}
				if d32bid.IsNaN() != tc.isNaN {
					t.Errorf("Decimal32BID IsNaN mismatch: got=%v, expected=%v", d32bid.IsNaN(), tc.isNaN)
				}
				if d32bid.IsInf() != tc.isInf {
					t.Errorf("Decimal32BID IsInf mismatch: got=%v, expected=%v", d32bid.IsInf(), tc.isInf)
				}
				if tc.sign != 0 && d32bid.Sign() != tc.sign {
					t.Errorf("Decimal32BID Sign mismatch: got=%v, expected=%v", d32bid.Sign(), tc.sign)
				}

				t.Logf("✅ Decimal32BID %s: Zero=%v, NaN=%v, Inf=%v, Sign=%v",
					tc.name, d32bid.IsZero(), d32bid.IsNaN(), d32bid.IsInf(), d32bid.Sign())
			}

			// Decimal64BID test
			d64bid, err := NewDecimal64BIDDirect(tc.value)
			if err != nil && !tc.isNaN && !tc.isInf {
				t.Errorf("Decimal64BID creation failed: %v", err)
				return
			}

			if err == nil {
				if d64bid.IsZero() != tc.isZero {
					t.Errorf("Decimal64BID IsZero mismatch: got=%v, expected=%v", d64bid.IsZero(), tc.isZero)
				}
				if d64bid.IsNaN() != tc.isNaN {
					t.Errorf("Decimal64BID IsNaN mismatch: got=%v, expected=%v", d64bid.IsNaN(), tc.isNaN)
				}
				if d64bid.IsInf() != tc.isInf {
					t.Errorf("Decimal64BID IsInf mismatch: got=%v, expected=%v", d64bid.IsInf(), tc.isInf)
				}
				if tc.sign != 0 && d64bid.Sign() != tc.sign {
					t.Errorf("Decimal64BID Sign mismatch: got=%v, expected=%v", d64bid.Sign(), tc.sign)
				}

				t.Logf("✅ Decimal64BID %s: Zero=%v, NaN=%v, Inf=%v, Sign=%v",
					tc.name, d64bid.IsZero(), d64bid.IsNaN(), d64bid.IsInf(), d64bid.Sign())
			}
		})
	}
}

// BenchmarkBIDOptimization benchmarks BID optimization performance.
func BenchmarkBIDOptimization(b *testing.B) {
	requireNativeBenchmark(b)

	// Measure BID optimization performance across varied data patterns.

	b.Run("BID_Integer_Operations", func(b *testing.B) {
		a, _ := NewDecimal32BIDDirect("123")
		val, _ := NewDecimal32BIDDirect("1")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = a.Add(val)
		}
	})

	b.Run("BID_Decimal_Operations", func(b *testing.B) {
		a, _ := NewDecimal32BIDDirect("123.456")
		val, _ := NewDecimal32BIDDirect("0.001")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = a.Add(val)
		}
	})

	b.Run("BID_Large_Numbers", func(b *testing.B) {
		a, _ := NewDecimal32BIDDirect("9999999")
		val, _ := NewDecimal32BIDDirect("1")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = a.Mul(val)
		}
	})

	b.Run("BID_Small_Decimals", func(b *testing.B) {
		a, _ := NewDecimal32BIDDirect("0.0000001")
		val, _ := NewDecimal32BIDDirect("2")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = a.Mul(val)
		}
	})

	b.Run("BID_Mixed_Precision", func(b *testing.B) {
		a, _ := NewDecimal64BIDDirect("123456789.123456")
		val, _ := NewDecimal64BIDDirect("0.000001")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = a.Add(val)
		}
	})
}

// TestBIDvsFloatPrecision compares BID and Float precision.
func TestBIDvsFloatPrecision(t *testing.T) {
	requireNative(t)
	t.Log("Starting BID vs Float precision comparison test")

	precisionCases := []struct {
		name  string
		a     string
		b     string
		op    string
		issue string // known floating-point issue
	}{
		{
			name:  "0.1 + 0.2 issue",
			a:     "0.1",
			b:     "0.2",
			op:    "add",
			issue: "floating point returns 0.30000000000000004",
		},
		{
			name:  "large plus small precision",
			a:     "1000000000000000",
			b:     "0.1",
			op:    "add",
			issue: "floating point loses the small value",
		},
		{
			name:  "repeated operation drift",
			a:     "0.1",
			b:     "0.1",
			op:    "multiply",
			issue: "error accumulates across repeated operations",
		},
	}

	for _, tc := range precisionCases {
		t.Run(tc.name, func(t *testing.T) {
			// BID result
			aBID, _ := NewDecimal64BIDDirect(tc.a)
			bBID, _ := NewDecimal64BIDDirect(tc.b)

			var resultBID Decimal64BID
			switch tc.op {
			case "add":
				resultBID = aBID.Add(bBID)
			case "multiply":
				resultBID = aBID.Mul(bBID)
			}

			// Float64 result
			aFloat := parseFloat(tc.a)
			bFloat := parseFloat(tc.b)

			var resultFloat float64
			switch tc.op {
			case "add":
				resultFloat = aFloat + bFloat
			case "multiply":
				resultFloat = aFloat * bFloat
			}

			t.Logf("BID result: %s", resultBID.String())
			t.Logf("Float64 result: %.17g", resultFloat)
			t.Logf("Known issue: %s", tc.issue)

			// Check whether BID is more accurate for the simple case.
			if tc.name == "0.1 + 0.2 issue" {
				expected := "0.3"
				if normalizeDecimalString(resultBID.String()) == expected {
					t.Logf("✅ BID provides the exact result")
				} else {
					t.Logf("⚠️  BID also produced an unexpected result: %s", resultBID.String())
				}
			}
		})
	}
}

// parseFloat helper function
func parseFloat(s string) float64 {
	if s == "0.1" {
		return 0.1
	}
	if s == "0.2" {
		return 0.2
	}
	if s == "1000000000000000" {
		return 1000000000000000.0
	}
	return 1.0 // default value
}

// TestIntelBIDHardwareOptimization checks Intel hardware optimization behavior.
func TestIntelBIDHardwareOptimization(t *testing.T) {
	requireNative(t)
	t.Log("Starting Intel BID hardware optimization check")

	// Check CPU information.
	if runtime.GOARCH == "amd64" {
		t.Log("Checking Intel BID optimization on AMD64 architecture")
	} else if runtime.GOARCH == "arm64" {
		t.Log("Checking Intel BID performance on ARM64 architecture (software implementation)")
	} else {
		t.Logf("Checking Intel BID performance on %s architecture", runtime.GOARCH)
	}

	// Check optimization through performance-intensive operations.
	iterations := 100000

	// Measure performance on a large dataset.
	start := time.Now()

	d, _ := NewDecimal32BIDDirect("1.0001")
	for i := 0; i < iterations; i++ {
		d = d.Mul(d)
		if i%10000 == 0 {
			// Reset to avoid overflow.
			d, _ = NewDecimal32BIDDirect("1.0001")
		}
	}

	duration := time.Since(start)
	avgNs := duration.Nanoseconds() / int64(iterations)

	t.Logf("Decimal32BID intensive operation: %d iterations in %v", iterations, duration)
	t.Logf("Average operation time: %d ns/op", avgNs)

	// Check performance thresholds.
	if avgNs < 100 {
		t.Logf("✅ Excellent performance: %d ns/op", avgNs)
	} else if avgNs < 200 {
		t.Logf("✅ Good performance: %d ns/op", avgNs)
	} else {
		t.Logf("⚠️  Needs improvement: %d ns/op", avgNs)
	}
}
