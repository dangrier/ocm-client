package ocm

import (
	"testing"
	"time"

	"github.com/dangrier/ocm-client/internal/geobuf"
)

func TestToLocations(t *testing.T) {
	features := []geobuf.Feature{
		{Properties: map[string]interface{}{
			"objectid": float64(1234),
			"type":     "Suburb",
			"label":    "Brisbane City",
		}},
	}
	locs, err := toLocations(features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 1 {
		t.Fatalf("want 1 location, got %d", len(locs))
	}
	l := locs[0]
	if l.Code != 1234 {
		t.Errorf("Code: want 1234, got %d", l.Code)
	}
	if l.Type != LocationTypeSuburb {
		t.Errorf("Type: want Suburb, got %q", l.Type)
	}
	if l.Name != "Brisbane City" {
		t.Errorf("Name: want Brisbane City, got %q", l.Name)
	}
}

func TestToOffences(t *testing.T) {
	coord := [2]float64{153.0, -27.5}
	features := []geobuf.Feature{
		{
			Properties: map[string]interface{}{
				"date":   float64(1700000000),
				"o_code": float64(90),
				"l_code": float64(1234),
			},
			Coordinate: &coord,
		},
	}
	offs, err := toOffences(features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(offs) != 1 {
		t.Fatalf("want 1 offence, got %d", len(offs))
	}
	o := offs[0]
	if o.Category != OffenceCategoryAssault {
		t.Errorf("Category: want Assault, got %v", o.Category)
	}
	if o.Location != 1234 {
		t.Errorf("Location: want 1234, got %d", o.Location)
	}
	if want := time.Unix(1700000000, 0).UTC(); !o.StartTime.Equal(want) {
		t.Errorf("StartTime: want %v, got %v", want, o.StartTime)
	}
	if o.Coordinate == nil {
		t.Fatal("Coordinate: want non-nil")
	}
	if *o.Coordinate != coord {
		t.Errorf("Coordinate: want %v, got %v", coord, *o.Coordinate)
	}
}

func TestToOffences_NoCoordinate(t *testing.T) {
	features := []geobuf.Feature{
		{Properties: map[string]interface{}{
			"date":   float64(1700000000),
			"o_code": float64(90),
			"l_code": float64(1234),
		}},
	}
	offs, err := toOffences(features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if offs[0].Coordinate != nil {
		t.Error("Coordinate: want nil for feature with no geometry")
	}
}
