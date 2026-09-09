package bid754_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"testing"

	bid754 "github.com/sky1core/bid754/bid754-go"
)

func checkXMLRoundtrip[T comparable](t *testing.T, want T, raw any) {
	t.Helper()
	var actual, legacy bytes.Buffer
	start := xml.StartElement{Name: xml.Name{Local: "raw"}}
	if err := xml.NewEncoder(&actual).EncodeElement(want, start); err != nil {
		t.Fatal(err)
	}
	if err := xml.NewEncoder(&legacy).EncodeElement(raw, start); err != nil {
		t.Fatal(err)
	}
	if actual.String() != legacy.String() {
		t.Fatalf("XML=%q, legacy=%q", actual.String(), legacy.String())
	}
	var got T
	if err := xml.Unmarshal(legacy.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatal("XML changed raw representation")
	}
}

func TestXMLRawRepresentationRoundtrip(t *testing.T) {
	for _, tc := range jsonRawPatterns() {
		t.Run(tc.name, func(t *testing.T) {
			checkXMLRoundtrip(t, bid754.Decimal32BIDFromBits(tc.raw.D32), tc.raw.D32)
			checkXMLRoundtrip(t, bid754.Decimal64BIDFromBits(tc.raw.D64), tc.raw.D64)
		})
	}
	for _, text := range []string{"abcdefghijklmnop", "\t\n\r<&>\"'abcdefgh", "éééééééé", "😀😀😀😀", "�abcdefghijklm"} {
		t.Run(fmt.Sprintf("d128_%x", text), func(t *testing.T) {
			if len(text) != 16 {
				t.Fatalf("fixture has %d bytes", len(text))
			}
			var raw [16]byte
			copy(raw[:], text)
			checkXMLRoundtrip(t, bid754.Decimal128BIDFromBytes(raw), raw)
		})
	}
	type record struct {
		XMLName xml.Name `xml:"raw"`
		D32     bid754.Decimal32BID
		D64     bid754.Decimal64BID
		D128    bid754.Decimal128BID
	}
	want := record{XMLName: xml.Name{Local: "raw"}, D32: bid754.Decimal32BIDFromBits(42), D64: bid754.Decimal64BIDFromBits(43), D128: bid754.Decimal128BIDFromBytes([16]byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p'})}
	encoded, err := xml.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got record
	if err := xml.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatal("nested XML changed raw representation")
	}
}

func checkXMLReject[T comparable](t *testing.T, initial T, inputs []string) {
	t.Helper()
	for i, input := range inputs {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			got := initial
			if err := xml.Unmarshal([]byte(input), &got); err == nil {
				t.Fatalf("accepted %q", input)
			}
			if got != initial {
				t.Fatalf("receiver changed on %q", input)
			}
		})
	}
	var ptr *T
	start := xml.StartElement{Name: xml.Name{Local: "v"}}
	if err := any(ptr).(xml.Unmarshaler).UnmarshalXML(xml.NewDecoder(strings.NewReader("0</v>")), start); err == nil {
		t.Fatal("nil receiver accepted XML")
	}
	got := initial
	if err := any(&got).(xml.Unmarshaler).UnmarshalXML(nil, start); err == nil {
		t.Fatal("nil decoder accepted")
	}
	if got != initial {
		t.Fatal("nil decoder changed receiver")
	}
	if err := any(initial).(xml.Marshaler).MarshalXML(nil, start); err == nil {
		t.Fatal("nil encoder accepted")
	}
}

func TestXMLRawRejectRetainsReceiver(t *testing.T) {
	scalar := []string{"<v></v>", "<v> </v>", "<v>-1</v>", "<v>1.0</v>", "<v>1e0</v>", "<v>null</v>", "<v>42<child/></v>", "<v>42</wrong>", "<v>42", "<v>42&#0;</v>"}
	t.Run("d32", func(t *testing.T) {
		checkXMLReject(t, bid754.Decimal32BIDFromBits(0xfe00002a), append(append([]string{}, scalar...), "<v>4294967296</v>"))
	})
	t.Run("d64", func(t *testing.T) {
		checkXMLReject(t, bid754.Decimal64BIDFromBits(0xfe0000000000002a), append(append([]string{}, scalar...), "<v>18446744073709551616</v>"))
	})
	t.Run("d128", func(t *testing.T) {
		checkXMLReject(t, bid754.Decimal128BIDFromBytes([16]byte{42}), []string{
			"<v/>", "<v>abcdefghijklmno</v>", "<v>abcdefghijklmnopq</v>", "<v>abcdefghijklmnop<child/></v>", "<v>abcdefghijklmnop</wrong>", "<v>abcdefghijklmnop", "<v>abcdefghijklmno&#0;</v>", "<v>abcdefghijklmno\xff</v>",
		})
	})
}

func TestXMLRaw128RejectsLossyBytes(t *testing.T) {
	for _, raw := range [][16]byte{{}, {0xff}, {0x80}, {0xed, 0xa0, 0x80}, {0xef, 0xbf, 0xbe}, {0xef, 0xbf, 0xbf}, {1}, {0xf4, 0x90, 0x80, 0x80}} {
		got := bid754.Decimal128BIDFromBytes(raw)
		if _, err := xml.Marshal(got); err == nil {
			t.Fatalf("accepted lossy XML bytes %x", raw)
		}
	}
}
