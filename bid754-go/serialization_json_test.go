package bid754_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	bid754 "github.com/sky1core/bid754/bid754-go"
)

type rawJSONRecord struct {
	D32  uint32
	D64  uint64
	D128 [16]byte
}

type decimalJSONRecord struct {
	D32  bid754.Decimal32BID
	D64  bid754.Decimal64BID
	D128 bid754.Decimal128BID
}

func jsonRawPatterns() []struct {
	name string
	raw  rawJSONRecord
} {
	patterns := []struct {
		name        string
		b32         uint32
		b64, lo, hi uint64
	}{
		{"all_zero_bits", 0, 0, 0, 0},
		{"positive_zero", 0x32800000, 0x31c0000000000000, 0, 0x3040000000000000},
		{"negative_zero", 0xb2800000, 0xb1c0000000000000, 0, 0xb040000000000000},
		{"cohort_one", 0x32800001, 0x31c0000000000001, 1, 0x3040000000000000},
		{"cohort_ten", 0x3200000a, 0x31a000000000000a, 10, 0x303e000000000000},
		{"quiet_nan_payload", 0x7c00002a, 0x7c0000000000002a, 42, 0x7c00000000000000},
		{"signed_signaling_nan_payload", 0xfe00002a, 0xfe0000000000002a, 42, 0xfe00000000000000},
		{"infinity", 0x78000000, 0x7800000000000000, 0, 0x7800000000000000},
		{"all_one_bits", 0xffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		{"byte_order", 0x01234567, 0x0123456789abcdef, 0x0706050403020100, 0x0f0e0d0c0b0a0908},
	}
	var out []struct {
		name string
		raw  rawJSONRecord
	}
	for _, p := range patterns {
		raw := rawJSONRecord{D32: p.b32, D64: p.b64}
		binary.LittleEndian.PutUint64(raw.D128[:8], p.lo)
		binary.LittleEndian.PutUint64(raw.D128[8:], p.hi)
		out = append(out, struct {
			name string
			raw  rawJSONRecord
		}{p.name, raw})
	}
	return out
}

func TestJSONRawRepresentationRoundtrip(t *testing.T) {
	for _, tc := range jsonRawPatterns() {
		t.Run(tc.name, func(t *testing.T) {
			want := decimalJSONRecord{bid754.Decimal32BIDFromBits(tc.raw.D32), bid754.Decimal64BIDFromBits(tc.raw.D64), bid754.Decimal128BIDFromBytes(tc.raw.D128)}
			wire, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			legacy, err := json.Marshal(tc.raw)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(wire, legacy) {
				t.Fatalf("JSON=%s, legacy raw=%s", wire, legacy)
			}
			got := decimalJSONRecord{bid754.Decimal32BIDFromBits(9), bid754.Decimal64BIDFromBits(9), bid754.Decimal128BIDFromBytes([16]byte{9})}
			if err := json.Unmarshal(legacy, &got); err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatal("raw representation changed")
			}
			for i, pair := range []struct{ value, raw any }{
				{want.D32, tc.raw.D32}, {want.D64, tc.raw.D64}, {want.D128, tc.raw.D128},
				{&want.D32, tc.raw.D32}, {&want.D64, tc.raw.D64}, {&want.D128, tc.raw.D128},
				{[]bid754.Decimal32BID{want.D32}, []uint32{tc.raw.D32}},
				{[]bid754.Decimal64BID{want.D64}, []uint64{tc.raw.D64}},
				{[]bid754.Decimal128BID{want.D128}, [][16]byte{tc.raw.D128}},
				{map[string]bid754.Decimal32BID{"v": want.D32}, map[string]uint32{"v": tc.raw.D32}},
				{map[string]bid754.Decimal64BID{"v": want.D64}, map[string]uint64{"v": tc.raw.D64}},
				{map[string]bid754.Decimal128BID{"v": want.D128}, map[string][16]byte{"v": tc.raw.D128}},
			} {
				actual, err := json.Marshal(pair.value)
				if err != nil {
					t.Fatal(err)
				}
				expected, err := json.Marshal(pair.raw)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(actual, expected) {
					t.Fatalf("container %d: JSON=%s, raw=%s", i, actual, expected)
				}
			}
		})
	}
}

func checkJSONReject[T comparable](t *testing.T, initial T, inputs []string) {
	t.Helper()
	for i, input := range inputs {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			for _, direct := range []bool{false, true} {
				got := initial
				var err error
				if direct {
					err = any(&got).(json.Unmarshaler).UnmarshalJSON([]byte(input))
				} else {
					err = json.Unmarshal([]byte(input), &got)
				}
				if err == nil {
					t.Fatalf("accepted %q (direct=%v)", input, direct)
				}
				if got != initial {
					t.Fatalf("receiver changed on %q (direct=%v)", input, direct)
				}
			}
		})
	}
}

func TestJSONRawRejectRetainsReceiver(t *testing.T) {
	invalid := []string{"", " ", "{}", "true", "false", `"null"`, "-1", "-0", "1.0", "1e0", "NaN", "+1", "01", "1 2", "null 1", "\vnull"}
	t.Run("d32", func(t *testing.T) {
		checkJSONReject(t, bid754.Decimal32BIDFromBits(0xfe00002a), append(append([]string{}, invalid...), "[]", "[1]", "4294967296", "18446744073709551615"))
	})
	t.Run("d64", func(t *testing.T) {
		checkJSONReject(t, bid754.Decimal64BIDFromBits(0xfe0000000000002a), append(append([]string{}, invalid...), "[]", "[1]", "18446744073709551616", "999999999999999999999999999999999999"))
	})
	t.Run("d128", func(t *testing.T) {
		cases := append(append([]string{}, invalid...), "0", "[]", "[1]", "["+strings.Repeat("0,", 14)+"0]", "["+strings.Repeat("0,", 16)+"0]", `"AAAAAAAAAAAAAAAAAAAAAA=="`)
		for _, bad := range []string{"null", "-1", "-0", "256", "1.5", "1.0", "1e0", `"1"`, "true", "{}", "[]", "18446744073709551616"} {
			for _, pos := range []int{0, 7, 15} {
				elems := strings.Split(strings.Repeat("0,", 15)+"0", ",")
				elems[pos] = bad
				cases = append(cases, "["+strings.Join(elems, ",")+"]")
			}
		}
		cases = append(cases, "["+strings.Repeat("0,", 15)+"0,]", "["+strings.Repeat("0,", 15)+"0] null")
		checkJSONReject(t, bid754.Decimal128BIDFromBytes([16]byte{42, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 254}), cases)
	})
}

func checkJSONNull[T comparable](t *testing.T, initial T) {
	t.Helper()
	for _, input := range []string{"null", " \t\r\nnull \t\r\n"} {
		got := initial
		if err := json.Unmarshal([]byte(input), &got); err != nil {
			t.Fatal(err)
		}
		if got != initial {
			t.Fatal("null changed receiver")
		}
		if err := any(&got).(json.Unmarshaler).UnmarshalJSON([]byte(input)); err != nil {
			t.Fatal(err)
		}
		if got != initial {
			t.Fatal("direct null changed receiver")
		}
	}
	var ptr *T
	if err := any(ptr).(json.Unmarshaler).UnmarshalJSON([]byte("null")); err == nil {
		t.Fatal("nil receiver accepted null")
	}
	if err := any(ptr).(json.Unmarshaler).UnmarshalJSON([]byte("0")); err == nil {
		t.Fatal("nil receiver accepted value")
	}
	wire, err := json.Marshal(ptr)
	if err != nil || string(wire) != "null" {
		t.Fatalf("marshal nil=%s, %v", wire, err)
	}
	if err := json.Unmarshal([]byte("0"), ptr); err == nil {
		t.Fatal("Unmarshal accepted nil target")
	}
}

func TestJSONNullAndNilReceiver(t *testing.T) {
	t.Run("d32", func(t *testing.T) { checkJSONNull(t, bid754.Decimal32BIDFromBits(42)) })
	t.Run("d64", func(t *testing.T) { checkJSONNull(t, bid754.Decimal64BIDFromBits(42)) })
	t.Run("d128", func(t *testing.T) { checkJSONNull(t, bid754.Decimal128BIDFromBytes([16]byte{42})) })
}

func TestJSONRawTextRejectsUnicodeReplacement(t *testing.T) {
	initial := bid754.Decimal128BIDFromBytes([16]byte{42})
	for _, input := range []string{`"\ud800abcdefghijklm"`, `"\udc00abcdefghijklm"`, `"\ud800\u0041abcdefghijkl"`, "\"\xffabcdefghijklm\""} {
		value := initial
		if err := json.Unmarshal([]byte(input), &value); err == nil || value != initial {
			t.Fatalf("lossy Unicode accepted: %q, %v", input, err)
		}
	}
	var value bid754.Decimal128BID
	if err := json.Unmarshal([]byte(`"\ud83d\ude00\ud83d\ude00\ud83d\ude00\ud83d\ude00"`), &value); err != nil {
		t.Fatal(err)
	}
	raw := value.ToBytes()
	if string(raw[:]) != "😀😀😀😀" {
		t.Fatalf("surrogate pairs changed raw bytes: %x", raw)
	}
	var scalar bid754.Decimal64BID
	if err := json.Unmarshal([]byte(`"4\u0032"`), &scalar); err != nil || scalar.ToUint64() != 42 {
		t.Fatalf("escaped raw integer: %v", err)
	}
	for _, input := range []string{`"-1"`, `"1e0"`, `"1.0"`, `"18446744073709551616"`} {
		if err := json.Unmarshal([]byte(input), &scalar); err == nil || scalar.ToUint64() != 42 {
			t.Fatalf("invalid raw text changed receiver: %q, %v", input, err)
		}
	}
}
