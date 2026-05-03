package geobuf

import (
	"fmt"

	geobuflib "github.com/cairnapp/go-geobuf"
	cairngeojson "github.com/cairnapp/go-geobuf/pkg/geojson"
	"github.com/cairnapp/go-geobuf/pkg/geometry"
	geobufproto "github.com/cairnapp/go-geobuf/proto"
	"github.com/golang/protobuf/proto"
)

// Feature is a narrow view of a decoded GeoJSON feature — just what the adapter
// needs. Keeping this internal type prevents the rest of the codebase from
// coupling to the cairnapp geobuf library's types.
type Feature struct {
	Properties map[string]interface{}
	// Coordinate is [lon, lat] for Point geometries, nil otherwise.
	Coordinate *[2]float64
}

// DecodeFeatures decodes raw geobuf bytes and returns a flat slice of Features.
func DecodeFeatures(data []byte) ([]Feature, error) {
	var msg geobufproto.Data
	if err := proto.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("geobuf: unmarshal: %w", err)
	}
	// The geobuf spec defines Dimensions=0 as "use default of 2", but
	// cairnapp/go-geobuf passes the value straight into a division and panics.
	if msg.Dimensions == 0 {
		msg.Dimensions = 2
	}
	// The geobuf spec defines Precision=0 as "use default of 6 decimal places".
	// cairnapp computes 10^Precision as its internal divisor, so Precision=0 →
	// 10^0=1 → coordinates come out as raw integers. Setting Precision=6 makes
	// cairnapp divide by 10^6=1,000,000, yielding proper WGS84 decimal degrees.
	if msg.Precision == 0 {
		msg.Precision = 6
	}

	decoded := geobuflib.Decode(&msg)
	fc, ok := decoded.(*cairngeojson.FeatureCollection)
	if !ok {
		return nil, fmt.Errorf("geobuf: expected FeatureCollection, got %T", decoded)
	}

	return extractFeatures(fc), nil
}

func extractFeatures(fc *cairngeojson.FeatureCollection) []Feature {
	out := make([]Feature, 0, len(fc.Features))
	for _, f := range fc.Features {
		feat := Feature{
			Properties: map[string]interface{}(f.Properties),
		}
		switch g := f.Geometry.(type) {
		case geometry.Point:
			if len(g) >= 2 {
				coord := [2]float64{g[0], g[1]}
				feat.Coordinate = &coord
			}
		}
		out = append(out, feat)
	}
	return out
}
