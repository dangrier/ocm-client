package ocm

import (
	"fmt"
	"time"

	"github.com/dangrier/ocm-client/internal/geobuf"
)

func toLocations(features []geobuf.Feature) ([]*Location, error) {
	locs := make([]*Location, 0, len(features))
	for i, f := range features {
		code, err := intProp(f.Properties, "objectid")
		if err != nil {
			return nil, fmt.Errorf("location[%d]: %w", i, err)
		}
		locType, err := strProp(f.Properties, "type")
		if err != nil {
			return nil, fmt.Errorf("location[%d]: %w", i, err)
		}
		name, err := strProp(f.Properties, "label")
		if err != nil {
			return nil, fmt.Errorf("location[%d]: %w", i, err)
		}
		locs = append(locs, &Location{
			Code: code,
			Type: LocationType(locType),
			Name: name,
		})
	}
	return locs, nil
}

func toOffences(features []geobuf.Feature) ([]*Offence, error) {
	offs := make([]*Offence, 0, len(features))
	for i, f := range features {
		dateVal, err := intProp(f.Properties, "date")
		if err != nil {
			return nil, fmt.Errorf("offence[%d]: %w", i, err)
		}
		oCode, err := intProp(f.Properties, "o_code")
		if err != nil {
			return nil, fmt.Errorf("offence[%d]: %w", i, err)
		}
		lCode, err := intProp(f.Properties, "l_code")
		if err != nil {
			return nil, fmt.Errorf("offence[%d]: %w", i, err)
		}

		off := &Offence{
			StartTime: time.Unix(int64(dateVal), 0).UTC(),
			Category:  OffenceCategory(oCode),
			Location:  lCode,
		}
		if f.Coordinate != nil {
			coord := *f.Coordinate
			off.Coordinate = &coord
		}
		offs = append(offs, off)
	}
	return offs, nil
}

func strProp(props map[string]interface{}, key string) (string, error) {
	v, ok := props[key]
	if !ok {
		return "", fmt.Errorf("missing property %q", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("property %q: expected string, got %T", key, v)
	}
	return s, nil
}

func intProp(props map[string]interface{}, key string) (int, error) {
	v, ok := props[key]
	if !ok {
		return 0, fmt.Errorf("missing property %q", key)
	}
	switch n := v.(type) {
	case int:
		return n, nil
	case int64:
		return int(n), nil
	case uint:
		return int(n), nil
	case float64:
		return int(n), nil
	default:
		return 0, fmt.Errorf("property %q: expected number, got %T", key, v)
	}
}
