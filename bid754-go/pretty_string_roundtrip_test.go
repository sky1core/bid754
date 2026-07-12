package bid754

import (
	"encoding/binary"
	"math/rand"
	"testing"
)

func TestPrettyStringReparsesExpandedIntegerRegressions(t *testing.T) {
	t.Run("decimal32", func(t *testing.T) {
		d := Decimal32BID(0x399de940)
		if got, want := d.PrettyString(), "1.960256e20"; got != want {
			t.Fatalf("PrettyString() = %q, want %q", got, want)
		}
		assertPrettyStringRoundTrips32(t, d)
	})
	t.Run("decimal64", func(t *testing.T) {
		d := Decimal64BID(0x6ca28d5e8c49fba1)
		if got, want := d.PrettyString(), "9725586428263329E+6"; got != want {
			t.Fatalf("PrettyString() = %q, want %q", got, want)
		}
		assertPrettyStringRoundTrips64(t, d)
	})
	t.Run("decimal128", func(t *testing.T) {
		d := Decimal128BID{
			0x66, 0x37, 0x20, 0x1e, 0x0d, 0xb7, 0x8c, 0xa2,
			0xf9, 0x65, 0xcf, 0x37, 0xda, 0xb9, 0x5b, 0xb0,
		}
		if got, want := d.PrettyString(), "-8961831647043245083745335122081638E+13"; got != want {
			t.Fatalf("PrettyString() = %q, want %q", got, want)
		}
		assertPrettyStringRoundTrips128(t, d)
	})

	// Plain expansion remains useful when the resulting spelling itself fits
	// the width; exponent notation is selected only for an unrepresentable
	// written cohort.
	d, err := NewDecimal32("123E+2")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := d.PrettyString(), "12300"; got != want {
		t.Fatalf("representable plain expansion = %q, want %q", got, want)
	}
}

func TestPrettyStringReparsesDeterministicFiniteBitSamples(t *testing.T) {
	// A fixed seed makes this secondary metamorphic sweep reproducible. Exact
	// regressions above pin all three failures that originally exposed the bug.
	rng := rand.New(rand.NewSource(754))
	for i := 0; i < 50_000; i++ {
		assertPrettyStringRoundTrips32(t, Decimal32BID(rng.Uint32()))
		assertPrettyStringRoundTrips64(t, Decimal64BID(rng.Uint64()))

		var raw [16]byte
		binary.LittleEndian.PutUint64(raw[:8], rng.Uint64())
		binary.LittleEndian.PutUint64(raw[8:], rng.Uint64())
		assertPrettyStringRoundTrips128(t, Decimal128BID(raw))
	}
}

func assertPrettyStringRoundTrips32(t *testing.T, d Decimal32BID) {
	t.Helper()
	if !d.IsFinite() {
		return
	}
	pretty := d.PrettyString()
	reparsed, err := NewDecimal32(pretty)
	if err != nil {
		t.Fatalf("Decimal32 bits=%08x raw=%q PrettyString=%q is not parseable: %v", d.ToUint32(), d.String(), pretty, err)
	}
	equal, flags := d.QuietEqual(reparsed)
	if !equal || flags != 0 {
		t.Fatalf("Decimal32 bits=%08x raw=%q PrettyString=%q reparsed as %08x/%q (equal=%v flags=%v)", d.ToUint32(), d.String(), pretty, reparsed.ToUint32(), reparsed.String(), equal, flags)
	}
}

func assertPrettyStringRoundTrips64(t *testing.T, d Decimal64BID) {
	t.Helper()
	if !d.IsFinite() {
		return
	}
	pretty := d.PrettyString()
	reparsed, err := NewDecimal64(pretty)
	if err != nil {
		t.Fatalf("Decimal64 bits=%016x raw=%q PrettyString=%q is not parseable: %v", d.ToUint64(), d.String(), pretty, err)
	}
	equal, flags := d.QuietEqual(reparsed)
	if !equal || flags != 0 {
		t.Fatalf("Decimal64 bits=%016x raw=%q PrettyString=%q reparsed as %016x/%q (equal=%v flags=%v)", d.ToUint64(), d.String(), pretty, reparsed.ToUint64(), reparsed.String(), equal, flags)
	}
}

func assertPrettyStringRoundTrips128(t *testing.T, d Decimal128BID) {
	t.Helper()
	if !d.IsFinite() {
		return
	}
	pretty := d.PrettyString()
	reparsed, err := NewDecimal128(pretty)
	if err != nil {
		t.Fatalf("Decimal128 bits=%x raw=%q PrettyString=%q is not parseable: %v", d.ToBytes(), d.String(), pretty, err)
	}
	equal, flags := d.QuietEqual(reparsed)
	if !equal || flags != 0 {
		t.Fatalf("Decimal128 bits=%x raw=%q PrettyString=%q reparsed as %x/%q (equal=%v flags=%v)", d.ToBytes(), d.String(), pretty, reparsed.ToBytes(), reparsed.String(), equal, flags)
	}
}
