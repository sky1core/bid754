package bid754

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode/utf8"
)

func (d Decimal32BID) MarshalText() ([]byte, error) {
	return strconv.AppendUint(nil, uint64(d.raw), 10), nil
}

func (d *Decimal32BID) UnmarshalText(text []byte) error {
	if d == nil {
		return fmt.Errorf("bid754: Decimal32BID.UnmarshalText: nil receiver")
	}
	raw, err := parseRawUint(bytes.TrimSpace(text), 32)
	if err != nil {
		return err
	}
	d.raw = uint32(raw)
	return nil
}

func (d Decimal64BID) MarshalText() ([]byte, error) {
	return strconv.AppendUint(nil, d.raw, 10), nil
}

func (d *Decimal64BID) UnmarshalText(text []byte) error {
	if d == nil {
		return fmt.Errorf("bid754: Decimal64BID.UnmarshalText: nil receiver")
	}
	raw, err := parseRawUint(bytes.TrimSpace(text), 64)
	if err != nil {
		return err
	}
	d.raw = raw
	return nil
}

func (d Decimal128BID) MarshalText() ([]byte, error) {
	if err := validateRaw128Text(d.raw[:]); err != nil {
		return nil, err
	}
	if bytes.ContainsRune(d.raw[:], '\r') {
		return nil, fmt.Errorf("bid754: Decimal128BID text cannot preserve carriage returns in XML CDATA")
	}
	return append([]byte(nil), d.raw[:]...), nil
}

func (d *Decimal128BID) UnmarshalText(text []byte) error {
	if d == nil {
		return fmt.Errorf("bid754: Decimal128BID.UnmarshalText: nil receiver")
	}
	if err := validateRaw128Text(text); err != nil {
		return err
	}
	copy(d.raw[:], text)
	return nil
}

func parseRawUint(text []byte, bits int) (uint64, error) {
	if len(text) == 0 {
		return 0, fmt.Errorf("bid754: raw uint%d requires decimal digits", bits)
	}
	first := len(text)
	for i, c := range text {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("bid754: raw uint%d requires decimal digits", bits)
		}
		if first == len(text) && c != '0' {
			first = i
		}
	}
	if first == len(text) {
		return 0, nil
	}
	text = text[first:]
	if len(text) > 20 {
		return 0, fmt.Errorf("bid754: raw uint%d out of range", bits)
	}
	return strconv.ParseUint(string(text), 10, bits)
}

func validateRaw128Text(text []byte) error {
	if len(text) != 16 {
		return fmt.Errorf("bid754: Decimal128BID text requires exactly 16 bytes, got %d", len(text))
	}
	if !utf8.Valid(text) {
		return fmt.Errorf("bid754: Decimal128BID text bytes must be valid UTF-8")
	}
	for _, r := range string(text) {
		if !(r == '\t' || r == '\n' || r == '\r' || r >= 0x20 && r <= 0xd7ff || r >= 0xe000 && r <= 0xfffd || r >= 0x10000 && r <= 0x10ffff) {
			return fmt.Errorf("bid754: Decimal128BID text contains an unrepresentable character U+%04X", r)
		}
	}
	return nil
}
