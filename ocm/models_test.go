package ocm

import (
	"encoding/json"
	"testing"
)

func TestParseLocationType(t *testing.T) {
	cases := []struct {
		in   string
		want LocationType
	}{
		{"suburb", LocationTypeSuburb},
		{"Suburb", LocationTypeSuburb},
		{"  suburb  ", LocationTypeSuburb},
		{"postcode", LocationTypePostcode},
		{"lga", LocationTypeLocalGovernment},
		{"local government area", LocationTypeLocalGovernment},
		{"nhw", LocationTypeNeighbourhoodWatch},
		{"neighbourhood watch", LocationTypeNeighbourhoodWatch},
		{"region", LocationTypePoliceRegion},
		{"qps region", LocationTypePoliceRegion},
		{"district", LocationTypePoliceDistrict},
		{"qps district", LocationTypePoliceDistrict},
		{"patrol", LocationTypePolicePatrolGroup},
		{"patrol group", LocationTypePolicePatrolGroup},
		{"qps patrol group", LocationTypePolicePatrolGroup},
		{"division", LocationTypePoliceDivision},
		{"qps division", LocationTypePoliceDivision},
	}
	for _, c := range cases {
		got, err := ParseLocationType(c.in)
		if err != nil {
			t.Errorf("ParseLocationType(%q): unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseLocationType(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseLocationType_Unknown(t *testing.T) {
	_, err := ParseLocationType("galaxy")
	if err == nil {
		t.Error("ParseLocationType(unknown): want error, got nil")
	}
}

func TestOffenceCategoryString(t *testing.T) {
	cases := []struct {
		cat  OffenceCategory
		want string
	}{
		{OffenceCategoryAssault, "Assault"},
		{OffenceCategoryDrugOffences, "Drug Offences"},
		{OffenceCategoryHomicide, "Homicide"},
		{OffenceCategoryMiscellaneousOffences, "Miscellaneous Offences"},
	}
	for _, c := range cases {
		if got := c.cat.String(); got != c.want {
			t.Errorf("%d.String() = %q, want %q", int(c.cat), got, c.want)
		}
	}
}

func TestOffenceCategoryString_Unknown(t *testing.T) {
	unknown := OffenceCategory(9999)
	got := unknown.String()
	if got != "Unknown(9999)" {
		t.Errorf("unknown category String() = %q, want %q", got, "Unknown(9999)")
	}
}

func TestOffenceCategoryMarshalJSON(t *testing.T) {
	b, err := json.Marshal(OffenceCategoryAssault)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	got := string(b)
	if got != `"Assault"` {
		t.Errorf("MarshalJSON = %s, want %q", got, "Assault")
	}
}

func TestOffenceCategoryMarshalJSON_InStruct(t *testing.T) {
	o := Offence{Category: OffenceCategoryDrugOffences}
	b, err := json.Marshal(o)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]interface{}
	json.Unmarshal(b, &m)
	if got, ok := m["Category"]; !ok || got != "Drug Offences" {
		t.Errorf("Category in JSON = %v, want %q", got, "Drug Offences")
	}
}

func TestParseOffenceCategory(t *testing.T) {
	cases := []struct {
		in   string
		want OffenceCategory
	}{
		{"Assault", OffenceCategoryAssault},
		{"assault", OffenceCategoryAssault},
		{"ASSAULT", OffenceCategoryAssault},
		{"Drug Offences", OffenceCategoryDrugOffences},
		{"drug offences", OffenceCategoryDrugOffences},
		{"Miscellaneous Offences", OffenceCategoryMiscellaneousOffences},
	}
	for _, c := range cases {
		got, err := ParseOffenceCategory(c.in)
		if err != nil {
			t.Errorf("ParseOffenceCategory(%q): unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseOffenceCategory(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseOffenceCategory_Unknown(t *testing.T) {
	_, err := ParseOffenceCategory("jaywalking")
	if err == nil {
		t.Error("ParseOffenceCategory(unknown): want error, got nil")
	}
}
