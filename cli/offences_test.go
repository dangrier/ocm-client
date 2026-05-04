package cli

import (
	"math"
	"testing"

	"github.com/dangrier/ocm-client/ocm"
)

// ---- haversineKm ----

func TestHaversineKm_SamePoint(t *testing.T) {
	if d := haversineKm(0, 0, 0, 0); d != 0 {
		t.Errorf("same point: want 0, got %f", d)
	}
}

func TestHaversineKm_KnownDistance(t *testing.T) {
	// Brisbane CBD to Sydney CBD, expected ~733 km great-circle distance.
	// Accept ±5 km for floating-point and Earth-radius approximation.
	brisLat, brisLon := -27.4698, 153.0251
	sydLat, sydLon := -33.8688, 151.2093
	d := haversineKm(brisLat, brisLon, sydLat, sydLon)
	if math.Abs(d-733) > 5 {
		t.Errorf("Brisbane→Sydney: want ~733 km, got %.1f km", d)
	}
}

func TestHaversineKm_Symmetric(t *testing.T) {
	a := haversineKm(-27.47, 153.02, -33.87, 151.21)
	b := haversineKm(-33.87, 151.21, -27.47, 153.02)
	if math.Abs(a-b) > 1e-9 {
		t.Errorf("haversine not symmetric: %f vs %f", a, b)
	}
}

// ---- bearingDeg ----

func TestBearingDeg_Cardinals(t *testing.T) {
	p := &proximityContext{lat: 0, lon: 0}
	cases := []struct {
		name     string
		lat, lon float64
		want     float64
	}{
		{"north", 1, 0, 0},
		{"east", 0, 1, 90},
		{"south", -1, 0, 180},
		{"west", 0, -1, 270},
	}
	for _, c := range cases {
		got := p.bearingDeg(c.lat, c.lon)
		if math.Abs(got-c.want) > 0.5 {
			t.Errorf("bearingDeg(%s): want %.0f°, got %.2f°", c.name, c.want, got)
		}
	}
}

func TestBearingDeg_NortheastQuadrant(t *testing.T) {
	p := &proximityContext{lat: 0, lon: 0}
	got := p.bearingDeg(1, 1) // northeast
	if got <= 0 || got >= 90 {
		t.Errorf("northeast bearing should be in (0,90), got %.2f", got)
	}
}

func TestBearingDeg_AlwaysPositive(t *testing.T) {
	p := &proximityContext{lat: -27.47, lon: 153.02}
	targets := [][2]float64{{-27.5, 153.0}, {-27.45, 153.05}, {-27.47, 152.9}}
	for _, pt := range targets {
		b := p.bearingDeg(pt[0], pt[1])
		if b < 0 || b >= 360 {
			t.Errorf("bearing out of [0,360): %.4f", b)
		}
	}
}

// ---- newProximityContext ----

func TestNewProximityContext_Valid(t *testing.T) {
	p, err := newProximityContext("-27.4698,153.0251", 1.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(p.lat-(-27.4698)) > 1e-9 || math.Abs(p.lon-153.0251) > 1e-9 {
		t.Errorf("parsed lat/lon wrong: got %v,%v", p.lat, p.lon)
	}
	if p.radiusKm != 1.5 {
		t.Errorf("radiusKm = %v, want 1.5", p.radiusKm)
	}
}

func TestNewProximityContext_WithSpaces(t *testing.T) {
	_, err := newProximityContext(" -27.47 , 153.02 ", 1)
	if err != nil {
		t.Errorf("whitespace around values should be tolerated: %v", err)
	}
}

func TestNewProximityContext_BadFormat(t *testing.T) {
	cases := []string{
		"",
		"notanumber",
		"-27.47",         // missing lon
		"abc,153.02",     // bad lat
		"-27.47,notanum", // bad lon
	}
	for _, s := range cases {
		_, err := newProximityContext(s, 1)
		if err == nil {
			t.Errorf("newProximityContext(%q): want error, got nil", s)
		}
	}
}

// ---- proximityContext.filter ----

func coord(lon, lat float64) *[2]float64 { c := [2]float64{lon, lat}; return &c }

func TestFilter_WithinRadius(t *testing.T) {
	p := &proximityContext{lat: 0, lon: 0, radiusKm: 200}
	offs := []*ocm.Offence{
		{Coordinate: coord(0.5, 0.5)}, // ~78 km — inside
		{Coordinate: coord(5, 5)},     // ~785 km — outside
		{Coordinate: nil},             // no coordinate — always dropped
	}
	got := p.filter(offs)
	if len(got) != 1 {
		t.Errorf("want 1 result within 200 km, got %d", len(got))
	}
}

func TestFilter_ExactBoundary(t *testing.T) {
	p := &proximityContext{lat: 0, lon: 0, radiusKm: 100}
	// Point at ~111 km north — just outside 100 km radius.
	offs := []*ocm.Offence{{Coordinate: coord(0, 1.0)}}
	got := p.filter(offs)
	if len(got) != 0 {
		t.Errorf("point outside radius should be filtered: got %d results", len(got))
	}

	// Point at ~55 km northeast — inside 100 km.
	offs2 := []*ocm.Offence{{Coordinate: coord(0.35, 0.35)}}
	got2 := p.filter(offs2)
	if len(got2) != 1 {
		t.Errorf("point inside radius should pass: got %d results", len(got2))
	}
}

func TestFilter_NilCoordinateAlwaysDropped(t *testing.T) {
	p := &proximityContext{lat: 0, lon: 0, radiusKm: 1e9} // huge radius
	offs := []*ocm.Offence{{Coordinate: nil}}
	got := p.filter(offs)
	if len(got) != 0 {
		t.Errorf("nil-coordinate offence should be dropped even with huge radius")
	}
}

func TestFilter_Empty(t *testing.T) {
	p := &proximityContext{lat: 0, lon: 0, radiusKm: 100}
	got := p.filter(nil)
	if len(got) != 0 {
		t.Errorf("nil input: want empty, got %d", len(got))
	}
}
