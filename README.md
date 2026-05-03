# ocm-client

A Go library and CLI tool for querying the [Online Crime Map](https://qps-ocm.s3-ap-southeast-2.amazonaws.com/index.html) — publicly available crime statistics for Queensland, Australia.

> Not affiliated with or endorsed by the Queensland Police Service.

## Installation

```sh
go install github.com/dangrier/ocm-client/cmd/ocm@latest
```

Or build from source:

```sh
git clone https://github.com/dangrier/ocm-client
cd ocm-client
go build ./cmd/ocm
```

## CLI Usage

### List locations

```sh
# All locations
ocm locations

# Filter by type: suburb, postcode, lga, nhw, region, district, patrol, division
ocm locations --type suburb

# Get a single location by code
ocm location 30025
```

### Query offences

```sh
# Last 90 days for a suburb (default)
ocm offences --suburb "Fortitude Valley"

# Multiple locations, custom date range
ocm offences --suburb "Fortitude Valley" --suburb "Newstead" --from 2025-01-01 --to 2025-03-31

# Last N days
ocm offences --postcode "4006" --days 30

# Aggregate by category
ocm offences --lga "Brisbane City" --summary

# Filter to offences within 2 km of a point
ocm offences --suburb "Brisbane City" --near "-27.4698,153.0251" --radius-km 2
```

### Output formats

All commands accept `--output` / `-o` with values `table` (default), `json`, or `csv`.

```sh
ocm offences --suburb "South Brisbane" -o json
ocm offences --suburb "South Brisbane" -o csv > offences.csv
```

When using `--near`, JSON and CSV output include `distance_km` and `heading_deg` columns showing the great-circle distance and compass bearing from the supplied point to each offence.

## Library Usage

```go
import "github.com/dangrier/ocm-client/ocm"

client := ocm.NewClient()

// Look up a location by name
loc, err := client.GetLocationByName(ctx, ocm.LocationTypeSuburb, "Fortitude Valley")

// Fetch offences
locs := []*ocm.Location{loc}
offences, err := client.GetOffences(ctx, dateFrom, dateTo, locs)

for _, o := range offences {
    fmt.Println(o.StartTime, o.Category, o.Coordinate)
}
```

### Location types

| Constant | Aliases accepted by `ParseLocationType` |
|---|---|
| `LocationTypeSuburb` | `suburb` |
| `LocationTypePostcode` | `postcode` |
| `LocationTypeLocalGovernment` | `lga`, `local government area` |
| `LocationTypeNeighbourhoodWatch` | `nhw`, `neighbourhood watch` |
| `LocationTypePoliceRegion` | `region`, `qps region` |
| `LocationTypePoliceDistrict` | `district`, `qps district` |
| `LocationTypePolicePatrolGroup` | `patrol`, `patrol group`, `qps patrol group` |
| `LocationTypePoliceDivision` | `division`, `qps division` |

### Client options

```go
// Override HTTP client (e.g. for custom timeouts or proxies)
client := ocm.NewClient(ocm.WithHTTPClient(myHTTPClient))

// Override API key
client := ocm.NewClient(ocm.WithAPIKey("your-key"))
```

## Architecture

```txt
cmd/ocm/          Entry point (embeds IANA timezone data)
cli/              Cobra commands: locations, location, offences
ocm/              Public library: Client, Location, Offence, models
internal/
  api/            HTTP layer — fetches geobuf responses from AWS API
  auth/           AEST time-based auth token generation
  geobuf/         Decodes Protocol Buffer geobuf into GeoJSON features
```

## Running tests

```sh
go test ./...
```

## Data notes

- Offence coordinates are stored as `[longitude, latitude]` in the `Coordinate` field.
- Offence dates represent the start of the reported period.
- The API provides data at the geobuf (GeoJSON + Protocol Buffers) format.
- Location data is fetched once and cached in-memory for the lifetime of the client.
