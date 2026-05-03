package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/dangrier/ocm-client/ocm"
	"github.com/spf13/cobra"
)

var offencesCmd = &cobra.Command{
	Use:   "offences",
	Short: "Query crime offences for one or more locations",
	Example: `  ocm offences --suburb "Brisbane City" --days 90
  ocm offences --suburb "Brisbane City" --from 2026-01-01 --to 2026-03-31 --summary
  ocm offences --suburb "Brisbane City" --near "-27.47,153.02" --radius-km 1 --output csv`,
	RunE: runOffences,
}

var (
	flagSuburb   []string
	flagPostcode []string
	flagLGA      []string
	flagNHW      []string
	flagRegion   []string
	flagDistrict []string
	flagPatrol   []string
	flagDivision []string
	flagCodes    []int
	flagDays     int
	flagFrom     string
	flagTo       string
	flagSummary  bool
	flagNear     string
	flagRadiusKm float64
)

func init() {
	offencesCmd.Flags().StringArrayVar(&flagSuburb, "suburb", nil, "Suburb name (repeatable)")
	offencesCmd.Flags().StringArrayVar(&flagPostcode, "postcode", nil, "Postcode name (repeatable)")
	offencesCmd.Flags().StringArrayVar(&flagLGA, "lga", nil, "Local Government Area name (repeatable)")
	offencesCmd.Flags().StringArrayVar(&flagNHW, "nhw", nil, "Neighbourhood Watch name (repeatable)")
	offencesCmd.Flags().StringArrayVar(&flagRegion, "region", nil, "QPS Region name (repeatable)")
	offencesCmd.Flags().StringArrayVar(&flagDistrict, "district", nil, "QPS District name (repeatable)")
	offencesCmd.Flags().StringArrayVar(&flagPatrol, "patrol", nil, "QPS Patrol Group name (repeatable)")
	offencesCmd.Flags().StringArrayVar(&flagDivision, "division", nil, "QPS Division name (repeatable)")
	offencesCmd.Flags().IntSliceVar(&flagCodes, "code", nil, "Raw location code (repeatable)")
	offencesCmd.Flags().IntVar(&flagDays, "days", 90, "Number of days back from today (mutually exclusive with --from/--to)")
	offencesCmd.Flags().StringVar(&flagFrom, "from", "", "Start date YYYY-MM-DD (mutually exclusive with --days)")
	offencesCmd.Flags().StringVar(&flagTo, "to", "", "End date YYYY-MM-DD (mutually exclusive with --days)")
	offencesCmd.Flags().BoolVar(&flagSummary, "summary", false, "Aggregate results by offence category")
	offencesCmd.Flags().StringVar(&flagNear, "near", "", "Filter to offences within --radius-km of this point, as \"lat,lon\"")
	offencesCmd.Flags().Float64Var(&flagRadiusKm, "radius-km", 0, "Radius in km for --near filter")
	offencesCmd.MarkFlagsMutuallyExclusive("days", "from")
	offencesCmd.MarkFlagsMutuallyExclusive("days", "to")
	offencesCmd.MarkFlagsRequiredTogether("near", "radius-km")
}

func runOffences(cmd *cobra.Command, args []string) error {
	// Resolve date range.
	var dateFrom, dateTo time.Time
	if flagFrom != "" || flagTo != "" {
		var err error
		if flagFrom == "" || flagTo == "" {
			return fmt.Errorf("--from and --to must both be provided")
		}
		dateFrom, err = time.Parse("2006-01-02", flagFrom)
		if err != nil {
			return fmt.Errorf("invalid --from date: %w", err)
		}
		dateTo, err = time.Parse("2006-01-02", flagTo)
		if err != nil {
			return fmt.Errorf("invalid --to date: %w", err)
		}
	} else {
		dateTo = time.Now()
		dateFrom = dateTo.AddDate(0, 0, -flagDays)
	}

	// Collect location specs.
	type locSpec struct {
		locType ocm.LocationType
		name    string
	}
	var specs []locSpec
	for _, n := range flagSuburb {
		specs = append(specs, locSpec{ocm.LocationTypeSuburb, n})
	}
	for _, n := range flagPostcode {
		specs = append(specs, locSpec{ocm.LocationTypePostcode, n})
	}
	for _, n := range flagLGA {
		specs = append(specs, locSpec{ocm.LocationTypeLocalGovernment, n})
	}
	for _, n := range flagNHW {
		specs = append(specs, locSpec{ocm.LocationTypeNeighbourhoodWatch, n})
	}
	for _, n := range flagRegion {
		specs = append(specs, locSpec{ocm.LocationTypePoliceRegion, n})
	}
	for _, n := range flagDistrict {
		specs = append(specs, locSpec{ocm.LocationTypePoliceDistrict, n})
	}
	for _, n := range flagPatrol {
		specs = append(specs, locSpec{ocm.LocationTypePolicePatrolGroup, n})
	}
	for _, n := range flagDivision {
		specs = append(specs, locSpec{ocm.LocationTypePoliceDivision, n})
	}

	if len(specs) == 0 && len(flagCodes) == 0 {
		return fmt.Errorf("at least one location flag is required (--suburb, --postcode, --code, ...)")
	}

	ctx, cancel := clientWithTimeout()
	defer cancel()

	// Resolve named locations.
	var locations []*ocm.Location
	for _, spec := range specs {
		l, err := s.client.GetLocationByName(ctx, spec.locType, spec.name)
		if err != nil {
			return err
		}
		locations = append(locations, l)
	}
	// Resolve code-based locations.
	for _, code := range flagCodes {
		l, err := s.client.GetLocation(ctx, code)
		if err != nil {
			return err
		}
		locations = append(locations, l)
	}

	offences, err := s.client.GetOffences(ctx, dateFrom, dateTo, locations)
	if err != nil {
		return err
	}

	var proxCtx *proximityContext
	if flagNear != "" {
		proxCtx, err = newProximityContext(flagNear, flagRadiusKm)
		if err != nil {
			return err
		}
		offences = proxCtx.filter(offences)
	}

	locNames := make(map[int]string, len(locations))
	for _, l := range locations {
		locNames[l.Code] = l.Name
	}

	if flagSummary {
		return writeSummary(s.output, offences)
	}
	return writeOffences(s.output, offences, locNames, proxCtx)
}

// proximityContext holds the centre point for --near filtering and provides
// distance and bearing calculations relative to that point.
type proximityContext struct {
	lat, lon float64
	radiusKm float64
}

func newProximityContext(nearFlag string, radiusKm float64) (*proximityContext, error) {
	parts := strings.SplitN(nearFlag, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("--near must be \"lat,lon\", got %q", nearFlag)
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return nil, fmt.Errorf("--near: invalid lat %q: %w", parts[0], err)
	}
	lon, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return nil, fmt.Errorf("--near: invalid lon %q: %w", parts[1], err)
	}
	return &proximityContext{lat: lat, lon: lon, radiusKm: radiusKm}, nil
}

// filter returns only offences that have a coordinate within the radius.
// Offences without a coordinate are always dropped.
func (p *proximityContext) filter(offs []*ocm.Offence) []*ocm.Offence {
	out := offs[:0]
	for _, o := range offs {
		if o.Coordinate != nil && p.distKm(o.Coordinate[1], o.Coordinate[0]) <= p.radiusKm {
			out = append(out, o)
		}
	}
	return out
}

// distKm returns the great-circle distance in km from the centre to (lat, lon).
func (p *proximityContext) distKm(lat, lon float64) float64 {
	return haversineKm(p.lat, p.lon, lat, lon)
}

// bearingDeg returns the initial bearing in degrees (0–360, clockwise from north)
// from the centre to (lat, lon).
func (p *proximityContext) bearingDeg(lat, lon float64) float64 {
	lat1 := p.lat * math.Pi / 180
	lat2 := lat * math.Pi / 180
	dLon := (lon - p.lon) * math.Pi / 180
	x := math.Sin(dLon) * math.Cos(lat2)
	y := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)
	return math.Mod(math.Atan2(x, y)*180/math.Pi+360, 360)
}

func writeOffences(format string, offs []*ocm.Offence, locNames map[int]string, prox *proximityContext) error {
	resolveLoc := func(code int) string {
		if name, ok := locNames[code]; ok {
			return name
		}
		return fmt.Sprintf("%d", code)
	}

	switch strings.ToLower(format) {
	case "json":
		type offenceJSON struct {
			StartTime  time.Time   `json:"start_time"`
			Category   string      `json:"category"`
			Location   string      `json:"location"`
			Coordinate *[2]float64 `json:"coordinate,omitempty"`
			DistanceKm *float64    `json:"distance_km,omitempty"`
			HeadingDeg *float64    `json:"heading_deg,omitempty"`
		}
		out := make([]offenceJSON, len(offs))
		for i, o := range offs {
			rec := offenceJSON{
				StartTime:  o.StartTime,
				Category:   o.Category.String(),
				Location:   resolveLoc(o.Location),
				Coordinate: o.Coordinate,
			}
			if prox != nil && o.Coordinate != nil {
				// Coordinate is [lon, lat]
				d := prox.distKm(o.Coordinate[1], o.Coordinate[0])
				h := prox.bearingDeg(o.Coordinate[1], o.Coordinate[0])
				rec.DistanceKm = &d
				rec.HeadingDeg = &h
			}
			out[i] = rec
		}
		return json.NewEncoder(os.Stdout).Encode(out)

	case "csv":
		w := csv.NewWriter(os.Stdout)
		header := []string{"start_time", "category", "location", "lat", "lon"}
		if prox != nil {
			header = append(header, "distance_km", "heading_deg")
		}
		_ = w.Write(header)
		for _, o := range offs {
			lon, lat := "", ""
			if o.Coordinate != nil {
				lon = strconv.FormatFloat(o.Coordinate[0], 'f', 6, 64)
				lat = strconv.FormatFloat(o.Coordinate[1], 'f', 6, 64)
			}
			row := []string{
				o.StartTime.Format(time.RFC3339),
				o.Category.String(),
				resolveLoc(o.Location),
				lat,
				lon,
			}
			if prox != nil {
				dist, heading := "", ""
				if o.Coordinate != nil {
					dist = strconv.FormatFloat(prox.distKm(o.Coordinate[1], o.Coordinate[0]), 'f', 3, 64)
					heading = strconv.FormatFloat(prox.bearingDeg(o.Coordinate[1], o.Coordinate[0]), 'f', 1, 64)
				}
				row = append(row, dist, heading)
			}
			_ = w.Write(row)
		}
		w.Flush()
		return w.Error()

	default:
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "DATE\tCATEGORY\tLOCATION")
		for _, o := range offs {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", o.StartTime.Format("2006-01-02"), o.Category, resolveLoc(o.Location))
		}
		return tw.Flush()
	}
}

func writeSummary(format string, offs []*ocm.Offence) error {
	counts := make(map[ocm.OffenceCategory]int)
	for _, o := range offs {
		counts[o.Category]++
	}

	type row struct {
		Category string `json:"category"`
		Count    int    `json:"count"`
	}
	rows := make([]row, 0, len(counts))
	for cat, n := range counts {
		rows = append(rows, row{cat.String(), n})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Count > rows[j].Count
	})

	switch strings.ToLower(format) {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(rows)
	case "csv":
		w := csv.NewWriter(os.Stdout)
		_ = w.Write([]string{"category", "count"})
		for _, r := range rows {
			_ = w.Write([]string{r.Category, fmt.Sprintf("%d", r.Count)})
		}
		w.Flush()
		return w.Error()
	default:
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "CATEGORY\tCOUNT")
		for _, r := range rows {
			fmt.Fprintf(tw, "%s\t%d\n", r.Category, r.Count)
		}
		return tw.Flush()
	}
}

// haversineKm returns the great-circle distance in km between two WGS84 points.
func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	lat1R := lat1 * math.Pi / 180
	lat2R := lat2 * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1R)*math.Cos(lat2R)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return earthRadiusKm * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
