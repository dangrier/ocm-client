package ocm

import (
	"testing"

	"github.com/dangrier/ocm-client/internal/geobuf"
)

func TestToLocations_MissingProperty(t *testing.T) {
	cases := []struct {
		name  string
		props map[string]interface{}
	}{
		{"missing objectid", map[string]interface{}{"type": "Suburb", "label": "X"}},
		{"missing type", map[string]interface{}{"objectid": float64(1), "label": "X"}},
		{"missing label", map[string]interface{}{"objectid": float64(1), "type": "Suburb"}},
	}
	for _, c := range cases {
		_, err := toLocations([]geobuf.Feature{{Properties: c.props}})
		if err == nil {
			t.Errorf("toLocations(%s): want error, got nil", c.name)
		}
	}
}

func TestToLocations_WrongType(t *testing.T) {
	// objectid must be numeric, not a string
	_, err := toLocations([]geobuf.Feature{{Properties: map[string]interface{}{
		"objectid": "not-a-number",
		"type":     "Suburb",
		"label":    "X",
	}}})
	if err == nil {
		t.Error("toLocations(objectid=string): want error, got nil")
	}
}

func TestIntProp_AllNumericTypes(t *testing.T) {
	cases := []struct {
		name string
		val  interface{}
		want int
	}{
		{"int", int(42), 42},
		{"int64", int64(42), 42},
		{"uint", uint(42), 42},
		{"float64", float64(42), 42},
	}
	for _, c := range cases {
		props := map[string]interface{}{"k": c.val}
		got, err := intProp(props, "k")
		if err != nil {
			t.Errorf("intProp(%s): unexpected error: %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("intProp(%s) = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestIntProp_Missing(t *testing.T) {
	_, err := intProp(map[string]interface{}{}, "missing")
	if err == nil {
		t.Error("intProp(missing key): want error, got nil")
	}
}

func TestIntProp_WrongType(t *testing.T) {
	_, err := intProp(map[string]interface{}{"k": "string"}, "k")
	if err == nil {
		t.Error("intProp(string value): want error, got nil")
	}
}

func TestToOffences_MissingProperty(t *testing.T) {
	base := map[string]interface{}{
		"date":   float64(1700000000),
		"o_code": float64(90),
		"l_code": float64(1234),
	}
	for _, drop := range []string{"date", "o_code", "l_code"} {
		props := make(map[string]interface{})
		for k, v := range base {
			props[k] = v
		}
		delete(props, drop)
		_, err := toOffences([]geobuf.Feature{{Properties: props}})
		if err == nil {
			t.Errorf("toOffences(missing %q): want error, got nil", drop)
		}
	}
}

func TestToLocations_Multiple(t *testing.T) {
	features := []geobuf.Feature{
		{Properties: map[string]interface{}{"objectid": float64(1), "type": "Suburb", "label": "Alpha"}},
		{Properties: map[string]interface{}{"objectid": float64(2), "type": "Postcode", "label": "4000"}},
	}
	locs, err := toLocations(features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 2 {
		t.Fatalf("want 2 locations, got %d", len(locs))
	}
}
