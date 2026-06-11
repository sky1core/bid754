//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

import "testing"

func BenchmarkAlignedBID32(b *testing.B) {
	requireNativeBenchmark(b)

	x, err := NewDecimal32BIDDirect("123.456")
	if err != nil {
		b.Fatalf("NewDecimal32BIDDirect(x): %v", err)
	}
	y, err := NewDecimal32BIDDirect("789.012")
	if err != nil {
		b.Fatalf("NewDecimal32BIDDirect(y): %v", err)
	}

	b.Run("add", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = x.Add(y)
		}
	})
	b.Run("mul", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = x.Mul(y)
		}
	})
	b.Run("div", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = x.Div(y)
		}
	})
	b.Run("parse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = NewDecimal32BIDDirect("123.456")
		}
	})
	b.Run("to_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = x.String()
		}
	})
}

func BenchmarkAlignedBID64(b *testing.B) {
	requireNativeBenchmark(b)

	x, err := NewDecimal64BIDDirect("123456789.123456789")
	if err != nil {
		b.Fatalf("NewDecimal64BIDDirect(x): %v", err)
	}
	y, err := NewDecimal64BIDDirect("987654321.987654321")
	if err != nil {
		b.Fatalf("NewDecimal64BIDDirect(y): %v", err)
	}

	b.Run("add", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = x.Add(y)
		}
	})
	b.Run("mul", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = x.Mul(y)
		}
	})
	b.Run("div", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = x.Div(y)
		}
	})
	b.Run("parse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = NewDecimal64BIDDirect("123456789.123456789")
		}
	})
	b.Run("to_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = x.String()
		}
	})
}

func BenchmarkAlignedBID128(b *testing.B) {
	requireNativeBenchmark(b)

	x, err := NewDecimal128BIDDirect("12345678901234567890.12345678901234")
	if err != nil {
		b.Fatalf("NewDecimal128BIDDirect(x): %v", err)
	}
	y, err := NewDecimal128BIDDirect("98765432109876543210.98765432109876")
	if err != nil {
		b.Fatalf("NewDecimal128BIDDirect(y): %v", err)
	}

	b.Run("add", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = x.Add(y)
		}
	})
	b.Run("mul", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = x.Mul(y)
		}
	})
	b.Run("div", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = x.Div(y)
		}
	})
	b.Run("parse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = NewDecimal128BIDDirect("12345678901234567890.12345678901234")
		}
	})
	b.Run("to_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = x.String()
		}
	})
}
