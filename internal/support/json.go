package support

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// Float is a float64 that always encodes with a decimal point, so a distance of
// 1500 metres is rendered as 1500.0 rather than as the integer 1500.
type Float float64

func (f Float) MarshalJSON() ([]byte, error) {
	encoded := strconv.FormatFloat(float64(f), 'f', -1, 64)

	if !strings.Contains(encoded, ".") {
		encoded += ".0"
	}

	return []byte(encoded), nil
}

// NullFloat is a Float that encodes as null when it is absent.
type NullFloat struct {
	Float Float
	Valid bool
}

// FloatValue wraps a present float.
func FloatValue(value float64) NullFloat {
	return NullFloat{Float: Float(value), Valid: true}
}

// FloatPtr wraps a float that may be absent.
func FloatPtr(value *float64) NullFloat {
	if value == nil {
		return NullFloat{}
	}

	return FloatValue(*value)
}

func (f NullFloat) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}

	return f.Float.MarshalJSON()
}

// NullInt is an int64 that encodes as null when it is absent.
type NullInt struct {
	Int   int64
	Valid bool
}

// IntPtr wraps an int that may be absent.
func IntPtr(value *int64) NullInt {
	if value == nil {
		return NullInt{}
	}

	return NullInt{Int: *value, Valid: true}
}

func (i NullInt) MarshalJSON() ([]byte, error) {
	if !i.Valid {
		return []byte("null"), nil
	}

	return []byte(strconv.FormatInt(i.Int, 10)), nil
}

// MarshalJSON encodes a value the way the API renders it: slashes and unicode
// left unescaped, and no trailing newline.
func MarshalJSON(value any) ([]byte, error) {
	var buffer bytes.Buffer

	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(value); err != nil {
		return nil, err
	}

	return bytes.TrimRight(buffer.Bytes(), "\n"), nil
}
