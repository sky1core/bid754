package bid754

import "testing"

func TestValueTypesPreserveEveryRawBit(t *testing.T) {
	t.Run("decimal32", func(t *testing.T) {
		for bit := 0; bit < 32; bit++ {
			want := uint32(1) << bit
			if got := Decimal32BID(want).ToUint32(); got != want {
				t.Fatalf("bit %d: got %#08x, want %#08x", bit, got, want)
			}
		}
	})

	t.Run("decimal64", func(t *testing.T) {
		for bit := 0; bit < 64; bit++ {
			want := uint64(1) << bit
			if got := Decimal64BID(want).ToUint64(); got != want {
				t.Fatalf("bit %d: got %#016x, want %#016x", bit, got, want)
			}
		}
	})

	t.Run("decimal128", func(t *testing.T) {
		for bit := 0; bit < 128; bit++ {
			var want [16]byte
			want[bit/8] = byte(1 << (bit % 8))
			if got := Decimal128BID(want).ToBytes(); got != want {
				t.Fatalf("bit %d: got %x, want %x", bit, got, want)
			}
		}
	})
}
