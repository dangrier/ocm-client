package ocm

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dangrier/ocm-client/internal/api"
	"github.com/dangrier/ocm-client/internal/geobuf"
)

// Client provides access to the QPS Online Crime Map API.
type Client struct {
	provider *api.Provider

	mu      sync.Mutex
	locData map[int]*Location
	nameIdx map[LocationType]map[string]*Location // built lazily
}

// Option configures a Client.
type Option func(*Client, *clientConfig)

type clientConfig struct {
	hc     *http.Client
	apiKey string
}

// WithHTTPClient overrides the HTTP client used for requests.
func WithHTTPClient(hc *http.Client) Option {
	return func(_ *Client, cfg *clientConfig) { cfg.hc = hc }
}

// WithAPIKey overrides the default API key.
func WithAPIKey(key string) Option {
	return func(_ *Client, cfg *clientConfig) { cfg.apiKey = key }
}

// NewClient creates a new Client with optional configuration.
func NewClient(opts ...Option) *Client {
	cfg := &clientConfig{}
	c := &Client{}
	for _, o := range opts {
		o(c, cfg)
	}
	c.provider = api.NewProvider(cfg.hc, cfg.apiKey)
	return c
}

// GetLocations returns all known locations, keyed by location code.
// Results are cached after the first successful call.
func (c *Client) GetLocations(ctx context.Context) (map[int]*Location, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.locData != nil {
		return c.locData, nil
	}
	locs, err := c.fetchLocations(ctx)
	if err != nil {
		return nil, err
	}
	c.locData = locs
	c.nameIdx = nil // invalidate name index on cache refresh
	return c.locData, nil
}

// GetLocation returns a single location by code, or ErrLocationNotFound.
func (c *Client) GetLocation(ctx context.Context, code int) (*Location, error) {
	locs, err := c.GetLocations(ctx)
	if err != nil {
		return nil, err
	}
	l, ok := locs[code]
	if !ok {
		return nil, fmt.Errorf("ocm: code %d: %w", code, ErrLocationNotFound)
	}
	return l, nil
}

// GetLocationByName returns the first location matching the given type and name
// (case-insensitive). Returns ErrLocationNotFound if no match.
func (c *Client) GetLocationByName(ctx context.Context, locType LocationType, name string) (*Location, error) {
	if _, err := c.GetLocations(ctx); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.nameIdx == nil {
		c.buildNameIndex()
	}
	byName, ok := c.nameIdx[locType]
	if !ok {
		return nil, fmt.Errorf("ocm: type %q: %w", locType, ErrLocationNotFound)
	}
	l, ok := byName[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("ocm: %q %q: %w", locType, name, ErrLocationNotFound)
	}
	return l, nil
}

// buildNameIndex builds the name lookup index. Must be called with c.mu held.
func (c *Client) buildNameIndex() {
	c.nameIdx = make(map[LocationType]map[string]*Location)
	for _, l := range c.locData {
		if _, ok := c.nameIdx[l.Type]; !ok {
			c.nameIdx[l.Type] = make(map[string]*Location)
		}
		c.nameIdx[l.Type][strings.ToLower(l.Name)] = l
	}
}

// GetOffences returns offences for the given date range and set of locations.
// dateFrom must be before or equal to dateTo.
func (c *Client) GetOffences(ctx context.Context, dateFrom, dateTo time.Time, locations []*Location) ([]*Offence, error) {
	if dateFrom.After(dateTo) {
		return nil, ErrInvalidDateRange
	}
	if len(locations) == 0 {
		return nil, fmt.Errorf("ocm: GetOffences: at least one location required")
	}
	if c.provider == nil {
		return nil, fmt.Errorf("ocm: GetOffences: no provider configured")
	}

	codes := make([]int, len(locations))
	for i, l := range locations {
		codes[i] = l.Code
	}

	from := dateFrom.Format("01-02-2006")
	to := dateTo.Format("01-02-2006")

	raw, err := c.provider.FetchOffences(ctx, from, to, codes)
	if err != nil {
		return nil, fmt.Errorf("ocm: %w", err)
	}

	features, err := geobuf.DecodeFeatures(raw)
	if err != nil {
		return nil, fmt.Errorf("ocm: decode offences: %w", err)
	}

	offs, err := toOffences(features)
	if err != nil {
		return nil, fmt.Errorf("ocm: %w", err)
	}

	result := make([]*Offence, len(offs))
	copy(result, offs)
	return result, nil
}

func (c *Client) fetchLocations(ctx context.Context) (map[int]*Location, error) {
	raw, err := c.provider.FetchLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("ocm: %w", err)
	}

	features, err := geobuf.DecodeFeatures(raw)
	if err != nil {
		return nil, fmt.Errorf("ocm: decode locations: %w", err)
	}

	locs, err := toLocations(features)
	if err != nil {
		return nil, fmt.Errorf("ocm: %w", err)
	}

	m := make(map[int]*Location, len(locs))
	for _, l := range locs {
		m[l.Code] = l
	}
	return m, nil
}
