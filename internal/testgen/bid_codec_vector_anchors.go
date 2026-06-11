package testgen

import (
	"fmt"
	"strconv"
	"strings"
)

type bidCodecVectorAnchor struct {
	Type          string
	Hex           string
	HexHi         string
	Sign          bool
	Coefficient   string
	Exponent      int32
	Kind          string
	Payload       string
	DecimalString string
	Canonical     bool
	EncodedHex    string
	EncodedHi     string
}

var bidCodecVectorAnchors = []bidCodecVectorAnchor{
	{
		Type: "bid32", Hex: "32800001", Sign: false, Coefficient: "1", Exponent: 0, Kind: "normal",
		DecimalString: "+1E+0", Canonical: true, EncodedHex: "32800001",
	},
	{
		Type: "bid32", Hex: "b2800000", Sign: true, Coefficient: "", Exponent: 0, Kind: "zero",
		DecimalString: "-0", Canonical: true, EncodedHex: "b2800000",
	},
	{
		Type: "bid32", Hex: "7c000001", Sign: false, Coefficient: "", Exponent: 0, Kind: "qnan", Payload: "1",
		DecimalString: "+NaN1", Canonical: true, EncodedHex: "7c000001",
	},
	{
		Type: "bid64", Hex: "31c0000000000001", Sign: false, Coefficient: "1", Exponent: 0, Kind: "normal",
		DecimalString: "+1E+0", Canonical: true, EncodedHex: "31c0000000000001",
	},
	{
		Type: "bid64", Hex: "b1c0000000000000", Sign: true, Coefficient: "", Exponent: 0, Kind: "zero",
		DecimalString: "-0", Canonical: true, EncodedHex: "b1c0000000000000",
	},
	{
		Type: "bid64", Hex: "7c00000000000001", Sign: false, Coefficient: "", Exponent: 0, Kind: "qnan", Payload: "1",
		DecimalString: "+NaN1", Canonical: true, EncodedHex: "7c00000000000001",
	},
	{
		Type: "bid128", Hex: "0000000000000001", HexHi: "3040000000000000", Sign: false, Coefficient: "1", Exponent: 0, Kind: "normal",
		DecimalString: "+1E+0", Canonical: true, EncodedHex: "0000000000000001", EncodedHi: "3040000000000000",
	},
	{
		Type: "bid128", Hex: "0000000000000000", HexHi: "8000000000000000", Sign: true, Coefficient: "", Exponent: -6176, Kind: "zero",
		DecimalString: "-0E-6176", Canonical: true, EncodedHex: "0000000000000000", EncodedHi: "8000000000000000",
	},
	{
		Type: "bid128", Hex: "0000000000000001", HexHi: "7c00000000000000", Sign: false, Coefficient: "", Exponent: 0, Kind: "qnan", Payload: "1",
		DecimalString: "+NaN1", Canonical: true, EncodedHex: "0000000000000001", EncodedHi: "7c00000000000000",
	},
}

func bidCodecAnchorJSON() string {
	var b strings.Builder
	b.WriteString("[")
	for i, a := range bidCodecVectorAnchors {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("{")
		writeJSONField(&b, "type", a.Type, false)
		writeJSONField(&b, "hex", a.Hex, true)
		if a.HexHi != "" {
			writeJSONField(&b, "hex_hi", a.HexHi, true)
		}
		writeJSONBoolField(&b, "sign", a.Sign, true)
		writeJSONField(&b, "coefficient", a.Coefficient, true)
		writeJSONIntField(&b, "exponent", int(a.Exponent), true)
		writeJSONField(&b, "kind", a.Kind, true)
		if a.Payload != "" {
			writeJSONField(&b, "payload", a.Payload, true)
		}
		writeJSONField(&b, "decimal_string", a.DecimalString, true)
		writeJSONBoolField(&b, "canonical", a.Canonical, true)
		writeJSONField(&b, "encoded_hex", a.EncodedHex, true)
		if a.EncodedHi != "" {
			writeJSONField(&b, "encoded_hi", a.EncodedHi, true)
		}
		b.WriteString("}")
	}
	b.WriteString("]")
	return b.String()
}

func writeJSONField(b *strings.Builder, key, value string, comma bool) {
	if comma {
		b.WriteString(",")
	}
	fmt.Fprintf(b, "%q:%q", key, value)
}

func writeJSONBoolField(b *strings.Builder, key string, value bool, comma bool) {
	if comma {
		b.WriteString(",")
	}
	fmt.Fprintf(b, "%q:%t", key, value)
}

func writeJSONIntField(b *strings.Builder, key string, value int, comma bool) {
	if comma {
		b.WriteString(",")
	}
	fmt.Fprintf(b, "%q:%d", key, value)
}

func bidCodecGoAnchorSnippet() string {
	var b strings.Builder
	b.WriteString("var anchorVectors = []vectorEntry{\n")
	for _, a := range bidCodecVectorAnchors {
		fmt.Fprintf(&b, "\t{Type: %q, Hex: %q", a.Type, a.Hex)
		if a.HexHi != "" {
			fmt.Fprintf(&b, ", HexHi: %q", a.HexHi)
		}
		fmt.Fprintf(&b, ", Sign: %t, Coefficient: %q, Exponent: %d, Kind: %q", a.Sign, a.Coefficient, a.Exponent, a.Kind)
		if a.Payload != "" {
			fmt.Fprintf(&b, ", Payload: %q", a.Payload)
		}
		fmt.Fprintf(&b, ", DecimalString: %q, Canonical: %t, EncodedHex: %q", a.DecimalString, a.Canonical, a.EncodedHex)
		if a.EncodedHi != "" {
			fmt.Fprintf(&b, ", EncodedHi: %q", a.EncodedHi)
		}
		b.WriteString("},\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func bidCodecRustAnchorSnippet() string {
	var b strings.Builder
	b.WriteString("fn anchor_vectors() -> Vec<Vector> {\n    vec![\n")
	for _, a := range bidCodecVectorAnchors {
		fmt.Fprintf(&b, "        Vector { typ: %q.to_string(), hex: %q.to_string(), hex_hi: %s, sign: %t, coefficient: %q.to_string(), exponent: %d, kind: %q.to_string(), decimal_string: %q.to_string(), canonical: %t, payload: %s, encoded_hex: Some(%q.to_string()), encoded_hi: %s },\n",
			a.Type, a.Hex, rustOptionString(a.HexHi), a.Sign, a.Coefficient, a.Exponent, a.Kind, a.DecimalString, a.Canonical, rustOptionString(a.Payload), a.EncodedHex, rustOptionString(a.EncodedHi))
	}
	b.WriteString("    ]\n}\n")
	return b.String()
}

func rustOptionString(value string) string {
	if value == "" {
		return "None"
	}
	return fmt.Sprintf("Some(%q.to_string())", value)
}

func bidCodecPythonAnchorSnippet() string {
	var b strings.Builder
	b.WriteString("_ANCHOR_VECTORS = [\n")
	for _, a := range bidCodecVectorAnchors {
		b.WriteString("    {\n")
		writePythonStringField(&b, "type", a.Type)
		writePythonStringField(&b, "hex", a.Hex)
		if a.HexHi != "" {
			writePythonStringField(&b, "hex_hi", a.HexHi)
		}
		writePythonBoolField(&b, "sign", a.Sign)
		writePythonStringField(&b, "coefficient", a.Coefficient)
		writePythonIntField(&b, "exponent", int(a.Exponent))
		writePythonStringField(&b, "kind", a.Kind)
		if a.Payload != "" {
			writePythonStringField(&b, "payload", a.Payload)
		}
		writePythonStringField(&b, "decimal_string", a.DecimalString)
		writePythonBoolField(&b, "canonical", a.Canonical)
		writePythonStringField(&b, "encoded_hex", a.EncodedHex)
		if a.EncodedHi != "" {
			writePythonStringField(&b, "encoded_hi", a.EncodedHi)
		}
		b.WriteString("    },\n")
	}
	b.WriteString("]\n")
	return b.String()
}

func writePythonStringField(b *strings.Builder, key, value string) {
	fmt.Fprintf(b, "        %q: %q,\n", key, value)
}

func writePythonBoolField(b *strings.Builder, key string, value bool) {
	pyValue := "False"
	if value {
		pyValue = "True"
	}
	fmt.Fprintf(b, "        %q: %s,\n", key, pyValue)
}

func writePythonIntField(b *strings.Builder, key string, value int) {
	fmt.Fprintf(b, "        %q: %d,\n", key, value)
}

func bidCodecJSAnchorSnippet() string {
	return "const anchorVectors = " + bidCodecPrettyJSON("  ") + ";\n"
}

func bidCodecJSAnchorArray() string {
	return bidCodecPrettyJSON("  ")
}

func bidCodecJavaAnchorJSONLiteral() string {
	return strconv.Quote(bidCodecAnchorJSON())
}

func bidCodecSwiftAnchorSnippet() string {
	var b strings.Builder
	b.WriteString("private let anchorVectors: [Vector] = [\n")
	for _, a := range bidCodecVectorAnchors {
		fmt.Fprintf(&b, "    Vector(type: %q, hex: %q, hex_hi: %s, sign: %t, coefficient: %q, exponent: %d, kind: %q, payload: %s, decimal_string: %q, canonical: %t, encoded_hex: %q, encoded_hi: %s),\n",
			a.Type, a.Hex, swiftOptionalString(a.HexHi), a.Sign, a.Coefficient, a.Exponent, a.Kind, swiftOptionalString(a.Payload), a.DecimalString, a.Canonical, a.EncodedHex, swiftOptionalString(a.EncodedHi))
	}
	b.WriteString("]\n")
	return b.String()
}

func swiftOptionalString(value string) string {
	if value == "" {
		return "nil"
	}
	return fmt.Sprintf("%q", value)
}

func bidCodecPrettyJSON(indent string) string {
	raw := bidCodecAnchorJSON()
	var out strings.Builder
	level := 0
	inString := false
	escape := false
	for _, r := range raw {
		switch {
		case escape:
			out.WriteRune(r)
			escape = false
		case r == '\\':
			out.WriteRune(r)
			escape = inString
		case r == '"':
			out.WriteRune(r)
			inString = !inString
		case inString:
			out.WriteRune(r)
		case r == '[' || r == '{':
			out.WriteRune(r)
			level++
			out.WriteRune('\n')
			out.WriteString(strings.Repeat(indent, level))
		case r == ']' || r == '}':
			level--
			out.WriteRune('\n')
			out.WriteString(strings.Repeat(indent, level))
			out.WriteRune(r)
		case r == ',':
			out.WriteRune(r)
			out.WriteRune('\n')
			out.WriteString(strings.Repeat(indent, level))
		case r == ':':
			out.WriteString(": ")
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}
