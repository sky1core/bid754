package bidgo

import "testing"

var sink64 uint64
var sink32 uint32
var sink128 BID_UINT128
var sinkString string

func BenchmarkFairBID32(b *testing.B) {
	x, _ := Bid32FromStringRaw("123.456", 0)
	y, _ := Bid32FromStringRaw("789.012", 0)

	b.Run("add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink32, _ = Bid32AddWithFlags(x, y, 0)
		}
	})
	b.Run("mul", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink32, _ = Bid32MulWithFlags(x, y, 0)
		}
	})
	b.Run("div", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink32, _ = Bid32DivWithFlags(x, y, 0)
		}
	})
	b.Run("parse", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink32, _ = Bid32FromStringRaw("123.456", 0)
		}
	})
	b.Run("to_string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sinkString = Bid32ToStringRaw(x)
		}
	})
}

func BenchmarkFairBID64(b *testing.B) {
	x, _ := Bid64FromString("123456789.123456789", 0)
	y, _ := Bid64FromString("987654321.987654321", 0)

	b.Run("add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink64, _ = Bid64AddWithFlags(x, y, 0)
		}
	})
	b.Run("mul", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink64, _ = Bid64MulWithFlags(x, y, 0)
		}
	})
	b.Run("div", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink64, _ = Bid64DivWithFlags(x, y, 0)
		}
	})
	b.Run("parse", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink64, _ = Bid64FromString("123456789.123456789", 0)
		}
	})
	b.Run("to_string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sinkString = Bid64ToString(x)
		}
	})
}

func BenchmarkFairBID128(b *testing.B) {
	x, _ := Bid128FromString("12345678901234567890.12345678901234", 0)
	y, _ := Bid128FromString("98765432109876543210.98765432109876", 0)

	b.Run("add", func(b *testing.B) {
		var pfpsf uint32
		for i := 0; i < b.N; i++ {
			pfpsf = 0
			sink128 = Bid128Add(x, y, 0, &pfpsf)
		}
	})
	b.Run("mul", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink128, _ = Bid128Mul(x, y, 0)
		}
	})
	b.Run("div", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink128, _ = Bid128Div(x, y, 0)
		}
	})
	b.Run("parse", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink128, _ = Bid128FromString("12345678901234567890.12345678901234", 0)
		}
	})
	b.Run("to_string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sinkString = Bid128ToString(x)
		}
	})
}
