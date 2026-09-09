package bid754_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"reflect"
	"strings"
	"testing"

	bid754 "github.com/sky1core/bid754/bid754-go"
)

func serializationJSONField[T any](value T, tag string) any {
	switch tag {
	case "value":
		return struct {
			V T `json:"v"`
		}{value}
	case "string":
		return struct {
			V T `json:"v,string"`
		}{value}
	case "omitempty":
		return struct {
			V T `json:"v,omitempty"`
		}{value}
	case "omitzero":
		return struct {
			V T `json:"v,omitzero"`
		}{value}
	case "ignored":
		return struct {
			V T `json:"-"`
		}{value}
	default:
		panic("unknown fixture tag")
	}
}

func checkSerializationJSONTags[T comparable, L comparable](t *testing.T, value T, legacy L, zeroBits, numericZero, array bool) {
	t.Helper()
	for _, tag := range []string{"value", "string", "omitempty", "omitzero", "ignored"} {
		t.Run(tag, func(t *testing.T) {
			current := serializationJSONField(value, tag)
			baseline := serializationJSONField(legacy, tag)
			wire, err := json.Marshal(current)
			if err != nil {
				t.Fatal(err)
			}
			old, err := json.Marshal(baseline)
			if err != nil {
				t.Fatal(err)
			}
			different := tag == "string" && !array || tag == "omitempty" && zeroBits && !array || tag == "omitzero" && numericZero && !zeroBits
			if bytes.Equal(wire, old) == different {
				t.Fatalf("tag difference=%v, new=%s legacy=%s", different, wire, old)
			}
			if different {
				t.Logf("legacy difference: new=%s, legacy=%s", wire, old)
			}
			got := reflect.New(reflect.TypeOf(current))
			if err := json.Unmarshal(wire, got.Interface()); err != nil {
				t.Fatal(err)
			}
			if string(wire) != "{}" && got.Elem().FieldByName("V").Interface() != any(value) {
				t.Fatal("field tag changed raw bits")
			}
			oldOut := reflect.New(reflect.TypeOf(baseline))
			if err := json.Unmarshal(old, oldOut.Interface()); err != nil {
				t.Fatal(err)
			}
			if tag == "string" && !array {
				got.Elem().FieldByName("V").Set(reflect.ValueOf(value))
				if err := json.Unmarshal(old, got.Interface()); err != nil || got.Elem().FieldByName("V").Interface() != any(value) {
					t.Fatal("quoted legacy scalar must preserve raw bits")
				}
			}
		})
	}
}

func TestSerializationJSONLegacyTagDifferences(t *testing.T) {
	for _, tc := range jsonRawPatterns() {
		t.Run(tc.name, func(t *testing.T) {
			zeroBits := tc.name == "all_zero_bits"
			numericZero := tc.name == "positive_zero" || tc.name == "negative_zero" || zeroBits
			checkSerializationJSONTags(t, bid754.Decimal32BIDFromBits(tc.raw.D32), serializationLegacy32(tc.raw.D32), zeroBits, numericZero, false)
			checkSerializationJSONTags(t, bid754.Decimal64BIDFromBits(tc.raw.D64), serializationLegacy64(tc.raw.D64), zeroBits, numericZero, false)
			checkSerializationJSONTags(t, bid754.Decimal128BIDFromBytes(tc.raw.D128), serializationLegacy128(tc.raw.D128), zeroBits, numericZero, true)
		})
	}
}

func checkSerializationLongJSON[T comparable](t *testing.T, initial T, inputs []string) {
	t.Helper()
	for i, input := range inputs {
		for _, direct := range []bool{false, true} {
			got := initial
			var err error
			if direct {
				err = any(&got).(json.Unmarshaler).UnmarshalJSON([]byte(input))
			} else {
				err = json.Unmarshal([]byte(input), &got)
			}
			if err == nil || got != initial {
				t.Fatalf("input %d direct=%v: atomic rejection failed", i, direct)
			}
			if len(err.Error()) > 256 {
				t.Fatalf("input %d: error length=%d", i, len(err.Error()))
			}
		}
	}
}

func TestSerializationJSONBoundedErrors(t *testing.T) {
	huge := strings.Repeat("9", 1<<20)
	inputs := []string{huge, huge + ".0", "\"" + huge + "\"", "0" + huge, "[" + huge + "]"}
	checkSerializationLongJSON(t, bid754.Decimal32BIDFromBits(42), inputs)
	checkSerializationLongJSON(t, bid754.Decimal64BIDFromBits(42), inputs)
	for _, value := range []string{huge, huge + ".0", "\"" + huge + "\""} {
		inputs = append(inputs, "["+strings.Repeat("0,", 15)+value+"]")
	}
	checkSerializationLongJSON(t, bid754.Decimal128BIDFromBytes([16]byte{42}), inputs)
}

func checkSerializationXMLFieldReject[T comparable](t *testing.T, initial T, inputs []string) {
	t.Helper()
	for _, tag := range []string{"value", "attr", "chardata", "cdata"} {
		for i, text := range inputs {
			wire, err := xml.Marshal(serializationXMLField(text, tag))
			if err != nil {
				t.Fatal(err)
			}
			current := serializationXMLField(initial, tag)
			got := reflect.New(reflect.TypeOf(current))
			got.Elem().Set(reflect.ValueOf(current))
			err = xml.Unmarshal(wire, got.Interface())
			if err == nil || got.Elem().FieldByName("V").Interface() != any(initial) {
				t.Fatalf("tag %s input %d: rejection changed receiver or returned no error", tag, i)
			}
			if len(err.Error()) > 256 {
				t.Fatalf("tag %s input %d: error length=%d", tag, i, len(err.Error()))
			}
		}
	}
}

func TestSerializationXMLFieldRejectAtomicity(t *testing.T) {
	invalid := []string{"", " ", "-1", "-0", "+1", "1.0", "1e0", "0x1", "1_0", "null", strings.Repeat("9", 1<<20)}
	checkSerializationXMLFieldReject(t, bid754.Decimal32BIDFromBits(42), append(append([]string{}, invalid...), "4294967296"))
	checkSerializationXMLFieldReject(t, bid754.Decimal64BIDFromBits(42), append(append([]string{}, invalid...), "18446744073709551616"))
	checkSerializationXMLFieldReject(t, serializationText128("abcdefghijklmnop"), []string{"", strings.Repeat("a", 15), strings.Repeat("a", 17), strings.Repeat("a", 1<<20), "YWJjZGVmZ2hpamtsbW5vcA=="})
	for _, raw := range [][16]byte{{}, {0xff}, {1}, {0xef, 0xbf, 0xbe}} {
		for _, tag := range []string{"value", "attr", "chardata", "cdata"} {
			if _, err := xml.Marshal(serializationXMLField(bid754.Decimal128BIDFromBytes(raw), tag)); err == nil {
				t.Fatalf("XML %s accepted lossy raw bytes %x", tag, raw)
			}
		}
		if _, err := json.Marshal(map[bid754.Decimal128BID]string{bid754.Decimal128BIDFromBytes(raw): "v"}); err == nil {
			t.Fatalf("JSON map key accepted lossy raw bytes %x", raw)
		}
	}
}

func TestSerializationXMLCDATARejectsCarriageReturnLoss(t *testing.T) {
	for _, text := range []string{"\rabcdefghijklmno", "\r\nabcdefghijklmn"} {
		value := serializationText128(text)
		wire, err := xml.Marshal(serializationXMLField(text, "cdata"))
		if err != nil {
			t.Fatal(err)
		}
		var normalized struct {
			V string `xml:",cdata"`
		}
		if err := xml.Unmarshal(wire, &normalized); err != nil {
			t.Fatal(err)
		}
		if normalized.V == text {
			t.Fatal("native XML CDATA unexpectedly preserved carriage return")
		}
		if _, err := value.MarshalText(); err == nil {
			t.Fatal("MarshalText accepted CDATA-lossy carriage return")
		}
		for _, tag := range []string{"attr", "chardata", "cdata"} {
			if _, err := xml.Marshal(serializationXMLField(value, tag)); err == nil {
				t.Fatalf("XML %s accepted carriage return", tag)
			}
		}
		checkXMLRoundtrip(t, value, serializationLegacy128(value.ToBytes()))
		var decoded bid754.Decimal128BID
		if err := decoded.UnmarshalText([]byte(text)); err != nil || decoded != value {
			t.Fatal("raw text decoder changed carriage return")
		}
	}
	t.Log("d128 MarshalText rejects CR because native XML CDATA normalizes it; XML element values retain legacy escaped CR; UnmarshalText preserves already decoded CR")
}

func checkSerializationXMLOmitTags[T comparable, L comparable](t *testing.T, value T, legacy L, zeroBits bool) {
	t.Helper()
	for _, tag := range []string{"omitempty", "attr,omitempty", "ignored"} {
		wire, err := xml.Marshal(serializationXMLField(value, tag))
		if err != nil {
			t.Fatal(err)
		}
		old, err := xml.Marshal(serializationXMLField(legacy, tag))
		if err != nil {
			t.Fatal(err)
		}
		different := zeroBits && tag != "ignored"
		if bytes.Equal(wire, old) == different {
			t.Fatalf("XML %s difference=%v, new=%q legacy=%q", tag, different, wire, old)
		}
		if different {
			t.Logf("XML %s legacy difference: new=%q legacy=%q", tag, wire, old)
		}
	}
}

func TestSerializationXMLLegacyOmitTags(t *testing.T) {
	for _, raw := range []uint64{0, 42} {
		checkSerializationXMLOmitTags(t, bid754.Decimal32BIDFromBits(uint32(raw)), serializationLegacy32(raw), raw == 0)
		checkSerializationXMLOmitTags(t, bid754.Decimal64BIDFromBits(raw), serializationLegacy64(raw), raw == 0)
	}
	value := serializationText128("abcdefghijklmnop")
	checkSerializationXMLOmitTags(t, value, serializationLegacy128(value.ToBytes()), false)
}
