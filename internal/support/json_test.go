package support_test

import (
	"testing"

	"velostats/internal/support"
)

func TestFloatKeepsItsZeroFraction(t *testing.T) {
	cases := map[float64]string{
		1500:    "1500.0",
		10.63:   "10.63",
		51.2189: "51.2189",
		0:       "0.0",
		-100:    "-100.0",
	}

	for value, want := range cases {
		encoded, err := support.MarshalJSON(support.Float(value))
		if err != nil {
			t.Fatalf("MarshalJSON(%v): %v", value, err)
		}

		if string(encoded) != want {
			t.Errorf("MarshalJSON(%v) = %s, want %s", value, encoded, want)
		}
	}
}

func TestNullFloatEncodesAbsentValuesAsNull(t *testing.T) {
	encoded, err := support.MarshalJSON(support.NullFloat{})
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	if string(encoded) != "null" {
		t.Errorf("MarshalJSON = %s, want null", encoded)
	}
}

func TestMarshalJSONLeavesSlashesUnescaped(t *testing.T) {
	encoded, err := support.MarshalJSON(map[string]string{"url": "http://localhost/a&b"})
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	if string(encoded) != `{"url":"http://localhost/a&b"}` {
		t.Errorf("MarshalJSON = %s", encoded)
	}
}
