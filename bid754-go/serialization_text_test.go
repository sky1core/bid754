package bid754_test

import (
	"bytes"
	"encoding"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	bid754 "github.com/sky1core/bid754/bid754-go"
)

type serializationLegacy32 uint32
type serializationLegacy64 uint64
type serializationLegacy128 [16]byte

func serializationText128(text string) bid754.Decimal128BID {
	if len(text) != 16 {
		panic("serialization fixture must contain 16 bytes")
	}
	var raw [16]byte
	copy(raw[:], text)
	return bid754.Decimal128BIDFromBytes(raw)
}

func checkRawText[T comparable](t *testing.T, want T, text string) {
	t.Helper()
	for _, value := range []any{want, &want} {
		wire, err := value.(encoding.TextMarshaler).MarshalText()
		if err != nil || string(wire) != text {
			t.Fatalf("MarshalText=%q, %v; want %q", wire, err, text)
		}
		var got T
		if err := any(&got).(encoding.TextUnmarshaler).UnmarshalText(wire); err != nil || got != want {
			t.Fatalf("UnmarshalText changed raw bits: %v", err)
		}
		wire[0] ^= 1
		if got != want {
			t.Fatal("UnmarshalText retained input storage")
		}
		again, err := value.(encoding.TextMarshaler).MarshalText()
		if err != nil || string(again) != text {
			t.Fatal("MarshalText exposed receiver storage")
		}
	}
}

func TestSerializationRawTextContracts(t *testing.T) {
	for _, tc := range jsonRawPatterns() {
		t.Run(tc.name, func(t *testing.T) {
			checkRawText(t, bid754.Decimal32BIDFromBits(tc.raw.D32), fmt.Sprint(tc.raw.D32))
			checkRawText(t, bid754.Decimal64BIDFromBits(tc.raw.D64), fmt.Sprint(tc.raw.D64))
		})
	}
	for _, text := range []string{"abcdefghijklmnop", "\t\n<&>\"'abcdefghi", "éééééééé", "😀😀😀😀", "�abcdefghijklm", "]]>abcdefghijklm", "\ud7ff\ue000\ufffdabcdefg", "\U00010000\U0010ffffabcdefgh", "                "} {
		checkRawText(t, serializationText128(text), text)
	}
	t.Log("raw text method count: 6 (MarshalText and UnmarshalText for each of 3 widths)")
}

func checkRawTextReject[T comparable](t *testing.T, initial T, inputs []string) {
	t.Helper()
	for i, input := range inputs {
		got := initial
		err := any(&got).(encoding.TextUnmarshaler).UnmarshalText([]byte(input))
		if err == nil || got != initial {
			t.Fatalf("input %d: error=%v, receiver changed=%v", i, err, got != initial)
		}
		if len(err.Error()) > 256 {
			t.Fatalf("input %d: error has %d bytes", i, len(err.Error()))
		}
	}
	var ptr *T
	for _, text := range [][]byte{nil, []byte("0"), []byte("abcdefghijklmnop")} {
		if err := any(ptr).(encoding.TextUnmarshaler).UnmarshalText(text); err == nil {
			t.Fatal("nil UnmarshalText receiver accepted input")
		}
	}
}

func TestSerializationRawTextRejectAtomicity(t *testing.T) {
	invalid := []string{"", " ", "-1", "-0", "+1", "1.0", "1e0", "0x1", "1_0", "1 0", "null", "NaN", "１２", "1\x00", "1\xff", strings.Repeat("9", 1<<20), strings.Repeat("0", 1<<20) + "x"}
	checkRawTextReject(t, bid754.Decimal32BIDFromBits(42), append(append([]string{}, invalid...), "4294967296"))
	checkRawTextReject(t, bid754.Decimal64BIDFromBits(42), append(append([]string{}, invalid...), "18446744073709551616"))
	invalid128 := []string{"", strings.Repeat("a", 15), strings.Repeat("a", 17), strings.Repeat("a", 1<<20), strings.Repeat("é", 16), "YWJjZGVmZ2hpamtsbW5vcA==", "6162636465666768696a6b6c6d6e6f70"}
	for _, bad := range []string{"\x00", "\x01", "\v", "\f", "\xff", "\x80", "\xed\xa0\x80", "\xef\xbf\xbe", "\xef\xbf\xbf", "\xf4\x90\x80\x80"} {
		text := bad + strings.Repeat("a", 16-len(bad))
		invalid128 = append(invalid128, text)
		if _, err := serializationText128(text).MarshalText(); err == nil {
			t.Fatalf("MarshalText accepted %x", text)
		}
	}
	checkRawTextReject(t, serializationText128("abcdefghijklmnop"), invalid128)
}

func TestSerializationRawTextNumericLegacyInputs(t *testing.T) {
	for _, tc := range []struct {
		text string
		raw  uint64
	}{
		{"0", 0}, {"000", 0}, {"00042", 42}, {" \t\n42\r ", 42}, {"\u00a042\u2003", 42},
		{strings.Repeat("0", 1<<20) + "42", 42}, {strings.Repeat("0", 1<<20), 0},
	} {
		var d32 bid754.Decimal32BID
		var d64 bid754.Decimal64BID
		if err := d32.UnmarshalText([]byte(tc.text)); err != nil || d32.ToUint32() != uint32(tc.raw) {
			t.Fatalf("d32 text length %d: %v", len(tc.text), err)
		}
		if err := d64.UnmarshalText([]byte(tc.text)); err != nil || d64.ToUint64() != tc.raw {
			t.Fatalf("d64 text length %d: %v", len(tc.text), err)
		}
		var old32 serializationLegacy32
		var old64 serializationLegacy64
		wire := []byte("<r>" + tc.text + "</r>")
		if err := xml.Unmarshal(wire, &old32); err != nil || uint32(old32) != d32.ToUint32() {
			t.Fatalf("legacy d32 text length %d: %v", len(tc.text), err)
		}
		if err := xml.Unmarshal(wire, &old64); err != nil || uint64(old64) != d64.ToUint64() {
			t.Fatalf("legacy d64 text length %d: %v", len(tc.text), err)
		}
	}
}

func checkSerializationJSONMap[K comparable, L comparable](t *testing.T, key K, legacyKey L) {
	t.Helper()
	want := map[K]string{key: "raw"}
	wire, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := json.Marshal(map[L]string{legacyKey: "raw"})
	if err != nil || !bytes.Equal(wire, legacy) {
		t.Fatalf("map keys: current=%s legacy=%s, %v", wire, legacy, err)
	}
	got := map[K]string{key: "retained"}
	if err := json.Unmarshal(legacy, &got); err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("map key roundtrip: %v, %v", got, err)
	}
	var old map[L]string
	if err := json.Unmarshal(legacy, &old); err != nil || old[legacyKey] != "raw" {
		t.Fatalf("legacy map key roundtrip: %v", err)
	}
	values := map[string]K{"raw": key}
	wire, err = json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err = json.Marshal(map[string]L{"raw": legacyKey})
	if err != nil || !bytes.Equal(wire, legacy) {
		t.Fatalf("map values: current=%s legacy=%s, %v", wire, legacy, err)
	}
	var decoded map[string]K
	if err := json.Unmarshal(wire, &decoded); err != nil || !reflect.DeepEqual(decoded, values) {
		t.Fatalf("map value roundtrip: %v", err)
	}
}

func TestSerializationJSONMapKeysAndValues(t *testing.T) {
	for _, tc := range jsonRawPatterns() {
		t.Run(tc.name, func(t *testing.T) {
			checkSerializationJSONMap(t, bid754.Decimal32BIDFromBits(tc.raw.D32), serializationLegacy32(tc.raw.D32))
			checkSerializationJSONMap(t, bid754.Decimal64BIDFromBits(tc.raw.D64), serializationLegacy64(tc.raw.D64))
			value := bid754.Decimal128BIDFromBytes(tc.raw.D128)
			want := map[string]bid754.Decimal128BID{"raw": value}
			wire, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			old, err := json.Marshal(map[string]serializationLegacy128{"raw": serializationLegacy128(tc.raw.D128)})
			if err != nil || !bytes.Equal(wire, old) {
				t.Fatalf("d128 map value: new=%s legacy=%s, %v", wire, old, err)
			}
			var got map[string]bid754.Decimal128BID
			if err := json.Unmarshal(old, &got); err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("d128 map value roundtrip: %v", err)
			}
		})
	}
	for _, text := range []string{"abcdefghijklmnop", "\t\n<&>\"'abcdefghi", "éééééééé", "😀😀😀😀"} {
		key := serializationText128(text)
		want := map[bid754.Decimal128BID]string{key: "raw"}
		wire, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := json.Marshal(map[string]string{text: "raw"})
		if err != nil || !bytes.Equal(wire, expected) {
			t.Fatalf("d128 key=%s, expected=%s, %v", wire, expected, err)
		}
		got := map[bid754.Decimal128BID]string{key: "retained"}
		if err := json.Unmarshal(wire, &got); err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("d128 map key roundtrip: %v", err)
		}
		if _, err := json.Marshal(map[serializationLegacy128]string{serializationLegacy128(key.ToBytes()): "raw"}); err == nil {
			t.Fatal("legacy array map key unexpectedly supported")
		}
	}
	t.Log("legacy [16]byte JSON map keys are unsupported; new d128 keys use validated raw UTF-8 text")
}

func serializationXMLField[T any](value T, tag string) any {
	switch tag {
	case "value":
		return struct {
			XMLName xml.Name `xml:"r"`
			V       T        `xml:"v"`
		}{V: value}
	case "attr":
		return struct {
			XMLName xml.Name `xml:"r"`
			V       T        `xml:"v,attr"`
		}{V: value}
	case "omitempty":
		return struct {
			XMLName xml.Name `xml:"r"`
			V       T        `xml:"v,omitempty"`
		}{V: value}
	case "attr,omitempty":
		return struct {
			XMLName xml.Name `xml:"r"`
			V       T        `xml:"v,attr,omitempty"`
		}{V: value}
	case "ignored":
		return struct {
			XMLName xml.Name `xml:"r"`
			V       T        `xml:"-"`
		}{V: value}
	case "chardata":
		return struct {
			XMLName xml.Name `xml:"r"`
			V       T        `xml:",chardata"`
		}{V: value}
	case "cdata":
		return struct {
			XMLName xml.Name `xml:"r"`
			V       T        `xml:",cdata"`
		}{V: value}
	default:
		panic("unknown fixture tag")
	}
}

func checkSerializationXMLFields[T comparable, L comparable](t *testing.T, value T, legacy L, array bool) {
	t.Helper()
	for _, tag := range []string{"value", "attr", "chardata", "cdata"} {
		t.Run(tag, func(t *testing.T) {
			current := serializationXMLField(value, tag)
			baseline := serializationXMLField(legacy, tag)
			wire, err := xml.Marshal(current)
			if err != nil {
				t.Fatal(err)
			}
			old, err := xml.Marshal(baseline)
			if err != nil {
				t.Fatal(err)
			}
			if array && (tag == "chardata" || tag == "cdata") {
				if string(old) != "<r></r>" || bytes.Equal(wire, old) {
					t.Fatalf("legacy array text difference: new=%q, legacy=%q", wire, old)
				}
				t.Logf("legacy [16]byte XML %s drops content; new=%q", tag, wire)
			} else if !bytes.Equal(wire, old) {
				t.Fatalf("new=%q, legacy=%q", wire, old)
			}
			got := reflect.New(reflect.TypeOf(current))
			if err := xml.Unmarshal(wire, got.Interface()); err != nil {
				t.Fatal(err)
			}
			if got.Elem().FieldByName("V").Interface() != any(value) {
				t.Fatal("XML changed raw bits")
			}
			legacyOut := reflect.New(reflect.TypeOf(baseline))
			err = xml.Unmarshal(wire, legacyOut.Interface())
			if array {
				if err == nil {
					t.Fatal("legacy array XML unexpectedly supports decoding")
				}
				t.Logf("legacy [16]byte XML %s decoding error: %v", tag, err)
			} else if err != nil || legacyOut.Elem().FieldByName("V").Interface() != any(legacy) {
				t.Fatalf("legacy XML roundtrip: %v", err)
			}
		})
	}
}

func TestSerializationXMLFieldsLegacyComparison(t *testing.T) {
	for _, tc := range jsonRawPatterns() {
		t.Run(tc.name, func(t *testing.T) {
			checkSerializationXMLFields(t, bid754.Decimal32BIDFromBits(tc.raw.D32), serializationLegacy32(tc.raw.D32), false)
			checkSerializationXMLFields(t, bid754.Decimal64BIDFromBits(tc.raw.D64), serializationLegacy64(tc.raw.D64), false)
		})
	}
	for i, text := range []string{"abcdefghijklmnop", "\t\n<&>\"'abcdefghi", "éééééééé", "😀😀😀😀", "]]>abcdefghijklm"} {
		t.Run(fmt.Sprintf("d128_%d", i), func(t *testing.T) {
			value := serializationText128(text)
			checkSerializationXMLFields(t, value, serializationLegacy128(value.ToBytes()), true)
		})
	}
}

func TestSerializationValueShapeAndNumericZero(t *testing.T) {
	if unsafe.Sizeof(bid754.Decimal32BID{}) != 4 || unsafe.Sizeof(bid754.Decimal64BID{}) != 8 || unsafe.Sizeof(bid754.Decimal128BID{}) != 16 {
		t.Fatal("fixed-width layout changed")
	}
	for _, tc := range jsonRawPatterns() {
		if tc.name != "positive_zero" && tc.name != "negative_zero" {
			continue
		}
		d32 := bid754.Decimal32BIDFromBits(tc.raw.D32)
		d64 := bid754.Decimal64BIDFromBits(tc.raw.D64)
		d128 := bid754.Decimal128BIDFromBytes(tc.raw.D128)
		if !d32.IsZero() || !d64.IsZero() || !d128.IsZero() {
			t.Fatal("numeric zero predicate changed")
		}
		if d32 == (bid754.Decimal32BID{}) || d64 == (bid754.Decimal64BID{}) || d128 == (bid754.Decimal128BID{}) {
			t.Fatal("bit identity replaced by numeric equality")
		}
	}
}
