package bid754

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"unicode/utf8"
)

func (d Decimal32BID) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.raw)
}

func (d *Decimal32BID) UnmarshalJSON(data []byte) error {
	if d == nil {
		return fmt.Errorf("bid754: Decimal32BID.UnmarshalJSON: nil receiver")
	}
	raw, isNull, err := decodeRawJSONUint(data, 32, true)
	if err != nil {
		return err
	}
	if !isNull {
		d.raw = uint32(raw)
	}
	return nil
}

func (d Decimal64BID) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.raw)
}

func (d *Decimal64BID) UnmarshalJSON(data []byte) error {
	if d == nil {
		return fmt.Errorf("bid754: Decimal64BID.UnmarshalJSON: nil receiver")
	}
	raw, isNull, err := decodeRawJSONUint(data, 64, true)
	if err != nil {
		return err
	}
	if !isNull {
		d.raw = raw
	}
	return nil
}

func (d Decimal128BID) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.raw)
}

func (d *Decimal128BID) UnmarshalJSON(data []byte) error {
	if d == nil {
		return fmt.Errorf("bid754: Decimal128BID.UnmarshalJSON: nil receiver")
	}
	if json.Valid(data) {
		trimmed := bytes.TrimSpace(data)
		if bytes.Equal(trimmed, []byte("null")) {
			return nil
		}
		if trimmed[0] == '"' {
			text, err := decodeRawJSONString(trimmed)
			if err != nil {
				return err
			}
			return d.UnmarshalText([]byte(text))
		}
		if trimmed[0] != '[' {
			return fmt.Errorf("bid754: Decimal128BID JSON requires a 16-byte array")
		}
	}
	var elems []json.RawMessage
	if err := json.Unmarshal(data, &elems); err != nil {
		return err
	}
	if elems == nil {
		return nil
	}
	if len(elems) != 16 {
		return fmt.Errorf("bid754: Decimal128BID JSON requires exactly 16 bytes, got %d", len(elems))
	}
	var raw [16]byte
	for i, elem := range elems {
		if bytes.Equal(bytes.TrimSpace(elem), []byte("null")) {
			return fmt.Errorf("bid754: Decimal128BID JSON byte %d cannot be null", i)
		}
		value, _, err := decodeRawJSONUint(elem, 8, false)
		if err != nil {
			return fmt.Errorf("bid754: Decimal128BID JSON byte %d: %w", i, err)
		}
		raw[i] = byte(value)
	}
	d.raw = raw
	return nil
}

func decodeRawJSONUint(data []byte, bits int, allowText bool) (uint64, bool, error) {
	if !json.Valid(data) {
		var raw json.RawMessage
		return 0, false, json.Unmarshal(data, &raw)
	}
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		return 0, true, nil
	}
	if allowText && data[0] == '"' {
		text, err := decodeRawJSONString(data)
		if err != nil {
			return 0, false, err
		}
		data = []byte(text)
	}
	raw, err := parseRawUint(data, bits)
	return raw, false, err
}

func decodeRawJSONString(data []byte) (string, error) {
	if !utf8.Valid(data) {
		return "", fmt.Errorf("bid754: raw JSON text contains invalid UTF-8")
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return "", err
	}
	for i := 1; i < len(data)-1; i++ {
		if data[i] != '\\' {
			continue
		}
		i++
		if data[i] != 'u' {
			continue
		}
		code, _ := strconv.ParseUint(string(data[i+1:i+5]), 16, 16)
		i += 4
		if code >= 0xd800 && code <= 0xdbff {
			if i+6 >= len(data)-1 || data[i+1] != '\\' || data[i+2] != 'u' {
				return "", fmt.Errorf("bid754: raw JSON text contains an unpaired surrogate")
			}
			low, _ := strconv.ParseUint(string(data[i+3:i+7]), 16, 16)
			if low < 0xdc00 || low > 0xdfff {
				return "", fmt.Errorf("bid754: raw JSON text contains an unpaired surrogate")
			}
			i += 6
		} else if code >= 0xdc00 && code <= 0xdfff {
			return "", fmt.Errorf("bid754: raw JSON text contains an unpaired surrogate")
		}
	}
	return text, nil
}
