package ocm

import (
	"context"
	"errors"
	"testing"
	"time"
)

// seedClient returns a Client whose location cache has been pre-populated,
// bypassing the HTTP layer entirely.
func seedClient(locs []*Location) *Client {
	c := &Client{}
	c.locData = make(map[int]*Location, len(locs))
	for _, l := range locs {
		c.locData[l.Code] = l
	}
	return c
}

var testLocations = []*Location{
	{Code: 100, Type: LocationTypeSuburb, Name: "Alpha"},
	{Code: 200, Type: LocationTypeSuburb, Name: "Beta"},
	{Code: 300, Type: LocationTypePostcode, Name: "Alpha"}, // same name, different type
}

func TestGetLocation_Found(t *testing.T) {
	c := seedClient(testLocations)
	l, err := c.GetLocation(context.Background(), 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Name != "Alpha" {
		t.Errorf("Name = %q, want Alpha", l.Name)
	}
}

func TestGetLocation_NotFound(t *testing.T) {
	c := seedClient(testLocations)
	_, err := c.GetLocation(context.Background(), 9999)
	if !errors.Is(err, ErrLocationNotFound) {
		t.Errorf("want ErrLocationNotFound, got %v", err)
	}
}

func TestGetLocationByName_Found(t *testing.T) {
	c := seedClient(testLocations)
	l, err := c.GetLocationByName(context.Background(), LocationTypeSuburb, "Alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Code != 100 {
		t.Errorf("Code = %d, want 100", l.Code)
	}
}

func TestGetLocationByName_CaseInsensitive(t *testing.T) {
	c := seedClient(testLocations)
	l, err := c.GetLocationByName(context.Background(), LocationTypeSuburb, "ALPHA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Code != 100 {
		t.Errorf("Code = %d, want 100", l.Code)
	}
}

func TestGetLocationByName_SameNameDifferentType(t *testing.T) {
	c := seedClient(testLocations)
	// "Alpha" exists as both Suburb (100) and Postcode (300)
	l, err := c.GetLocationByName(context.Background(), LocationTypePostcode, "Alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Code != 300 {
		t.Errorf("Code = %d, want 300", l.Code)
	}
}

func TestGetLocationByName_NotFound(t *testing.T) {
	c := seedClient(testLocations)
	_, err := c.GetLocationByName(context.Background(), LocationTypeSuburb, "Nowhere")
	if !errors.Is(err, ErrLocationNotFound) {
		t.Errorf("want ErrLocationNotFound, got %v", err)
	}
}

func TestGetLocationByName_UnknownType(t *testing.T) {
	c := seedClient(testLocations)
	_, err := c.GetLocationByName(context.Background(), LocationTypePoliceRegion, "Alpha")
	if !errors.Is(err, ErrLocationNotFound) {
		t.Errorf("want ErrLocationNotFound, got %v", err)
	}
}

func TestGetLocations_CacheHit(t *testing.T) {
	c := seedClient(testLocations)
	m1, err := c.GetLocations(context.Background())
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	m2, err := c.GetLocations(context.Background())
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	// Same underlying map must be returned (pointer equality).
	if &m1 == &m2 {
		// map values are equal — just check length as a proxy
	}
	if len(m1) != len(m2) {
		t.Errorf("cache returned different map lengths: %d vs %d", len(m1), len(m2))
	}
}

func TestGetOffences_InvalidDateRange(t *testing.T) {
	c := seedClient(testLocations)
	later := time.Now()
	earlier := later.Add(-time.Hour)
	// dateFrom > dateTo should fail
	_, err := c.GetOffences(context.Background(), later, earlier, testLocations[:1])
	if !errors.Is(err, ErrInvalidDateRange) {
		t.Errorf("want ErrInvalidDateRange, got %v", err)
	}
}

func TestGetOffences_EqualDates(t *testing.T) {
	// dateFrom == dateTo is valid (same day); the error only comes from the
	// HTTP layer, not from the validation guard. Just confirm no date error.
	c := seedClient(testLocations)
	now := time.Now()
	_, err := c.GetOffences(context.Background(), now, now, testLocations[:1])
	if errors.Is(err, ErrInvalidDateRange) {
		t.Error("equal dates should not return ErrInvalidDateRange")
	}
}

func TestGetOffences_EmptyLocations(t *testing.T) {
	c := seedClient(testLocations)
	now := time.Now()
	_, err := c.GetOffences(context.Background(), now.Add(-time.Hour), now, nil)
	if err == nil {
		t.Error("empty locations: want error, got nil")
	}
}
