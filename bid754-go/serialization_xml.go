package bid754

import (
	"encoding/xml"
	"fmt"
	"strings"
)

func (d Decimal32BID) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if e == nil {
		return fmt.Errorf("bid754: Decimal32BID.MarshalXML: nil encoder")
	}
	return e.EncodeElement(d.raw, start)
}

func (d *Decimal32BID) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return fmt.Errorf("bid754: Decimal32BID.UnmarshalXML: nil receiver")
	}
	text, err := decodeRawXMLText(dec, start)
	if err != nil {
		return err
	}
	return d.UnmarshalText([]byte(text))
}

func (d Decimal64BID) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if e == nil {
		return fmt.Errorf("bid754: Decimal64BID.MarshalXML: nil encoder")
	}
	return e.EncodeElement(d.raw, start)
}

func (d *Decimal64BID) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return fmt.Errorf("bid754: Decimal64BID.UnmarshalXML: nil receiver")
	}
	text, err := decodeRawXMLText(dec, start)
	if err != nil {
		return err
	}
	return d.UnmarshalText([]byte(text))
}

func (d Decimal128BID) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if e == nil {
		return fmt.Errorf("bid754: Decimal128BID.MarshalXML: nil encoder")
	}
	if err := validateRaw128Text(d.raw[:]); err != nil {
		return err
	}
	return e.EncodeElement(d.raw, start)
}

func (d *Decimal128BID) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return fmt.Errorf("bid754: Decimal128BID.UnmarshalXML: nil receiver")
	}
	text, err := decodeRawXMLText(dec, start)
	if err != nil {
		return err
	}
	return d.UnmarshalText([]byte(text))
}

func decodeRawXMLText(dec *xml.Decoder, start xml.StartElement) (string, error) {
	if dec == nil {
		return "", fmt.Errorf("bid754: UnmarshalXML: nil decoder")
	}
	var text strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", err
		}
		switch tok := tok.(type) {
		case xml.CharData:
			text.Write(tok)
		case xml.StartElement:
			return "", fmt.Errorf("bid754: raw XML value cannot contain child elements")
		case xml.EndElement:
			if tok.Name != start.Name {
				return "", fmt.Errorf("bid754: mismatched raw XML end element")
			}
			return text.String(), nil
		}
	}
}
