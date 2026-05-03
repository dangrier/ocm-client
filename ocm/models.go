package ocm

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type LocationType string

const (
	LocationTypeSuburb             LocationType = "Suburb"
	LocationTypePostcode           LocationType = "Postcode"
	LocationTypeLocalGovernment    LocationType = "Local Government Area"
	LocationTypeNeighbourhoodWatch LocationType = "Neighbourhood Watch"
	LocationTypePoliceRegion       LocationType = "QPS Region"
	LocationTypePoliceDistrict     LocationType = "QPS District"
	LocationTypePolicePatrolGroup  LocationType = "QPS Patrol Group"
	LocationTypePoliceDivision     LocationType = "QPS Division"
)

func ParseLocationType(s string) (LocationType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "suburb":
		return LocationTypeSuburb, nil
	case "postcode":
		return LocationTypePostcode, nil
	case "lga", "local government", "local government area":
		return LocationTypeLocalGovernment, nil
	case "nhw", "neighbourhood watch":
		return LocationTypeNeighbourhoodWatch, nil
	case "region", "qps region":
		return LocationTypePoliceRegion, nil
	case "district", "qps district":
		return LocationTypePoliceDistrict, nil
	case "patrol", "patrol group", "qps patrol group":
		return LocationTypePolicePatrolGroup, nil
	case "division", "qps division":
		return LocationTypePoliceDivision, nil
	default:
		return "", fmt.Errorf("unknown location type %q", s)
	}
}

type OffenceCategory int

const (
	OffenceCategoryHomicide                   OffenceCategory = 10
	OffenceCategoryOtherHomicide              OffenceCategory = 20
	OffenceCategoryAssault                    OffenceCategory = 90
	OffenceCategoryRobbery                    OffenceCategory = 150
	OffenceCategoryOtherOffencesAgainstPerson OffenceCategory = 180
	OffenceCategoryUnlawfulEntry              OffenceCategory = 220
	OffenceCategoryArson                      OffenceCategory = 280
	OffenceCategoryOtherPropertyDamage        OffenceCategory = 290
	OffenceCategoryUnlawfulUseOfMotorVehicle  OffenceCategory = 300
	OffenceCategoryOtherTheft                 OffenceCategory = 310
	OffenceCategoryFraud                      OffenceCategory = 360
	OffenceCategoryHandlingStolenGoods        OffenceCategory = 400
	OffenceCategoryDrugOffences               OffenceCategory = 460
	OffenceCategoryProstitution               OffenceCategory = 520
	OffenceCategoryLiquor                     OffenceCategory = 610
	OffenceCategoryGamingRacingBetting        OffenceCategory = 620
	OffenceCategoryTrespassingVagrancy        OffenceCategory = 640
	OffenceCategoryWeaponsActOffences         OffenceCategory = 645
	OffenceCategoryGoodOrderOffences          OffenceCategory = 650
	OffenceCategoryStockRelatedOffences       OffenceCategory = 710
	OffenceCategoryTrafficRelatedOffences     OffenceCategory = 740
	OffenceCategoryMiscellaneousOffences      OffenceCategory = 790
)

var offenceCategoryNames = map[OffenceCategory]string{
	OffenceCategoryHomicide:                   "Homicide",
	OffenceCategoryOtherHomicide:              "Other Homicide",
	OffenceCategoryAssault:                    "Assault",
	OffenceCategoryRobbery:                    "Robbery",
	OffenceCategoryOtherOffencesAgainstPerson: "Other Offences Against the Person",
	OffenceCategoryUnlawfulEntry:              "Unlawful Entry",
	OffenceCategoryArson:                      "Arson",
	OffenceCategoryOtherPropertyDamage:        "Other Property Damage",
	OffenceCategoryUnlawfulUseOfMotorVehicle:  "Unlawful Use of Motor Vehicle",
	OffenceCategoryOtherTheft:                 "Other Theft",
	OffenceCategoryFraud:                      "Fraud",
	OffenceCategoryHandlingStolenGoods:        "Handling Stolen Goods",
	OffenceCategoryDrugOffences:               "Drug Offences",
	OffenceCategoryProstitution:               "Prostitution",
	OffenceCategoryLiquor:                     "Liquor",
	OffenceCategoryGamingRacingBetting:        "Gaming, Racing and Betting",
	OffenceCategoryTrespassingVagrancy:        "Trespassing and Vagrancy",
	OffenceCategoryWeaponsActOffences:         "Weapons Act Offences",
	OffenceCategoryGoodOrderOffences:          "Good Order Offences",
	OffenceCategoryStockRelatedOffences:       "Stock Related Offences",
	OffenceCategoryTrafficRelatedOffences:     "Traffic Related Offences",
	OffenceCategoryMiscellaneousOffences:      "Miscellaneous Offences",
}

func (c OffenceCategory) String() string {
	if name, ok := offenceCategoryNames[c]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", int(c))
}

func (c OffenceCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func ParseOffenceCategory(s string) (OffenceCategory, error) {
	lower := strings.ToLower(strings.TrimSpace(s))
	for cat, name := range offenceCategoryNames {
		if strings.ToLower(name) == lower {
			return cat, nil
		}
	}
	return 0, fmt.Errorf("unknown offence category %q", s)
}

type Location struct {
	Code int
	Type LocationType
	Name string
}

type Offence struct {
	StartTime  time.Time
	Category   OffenceCategory
	Location   int
	Coordinate *[2]float64 // [lon, lat], nil if absent
}
