package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/dangrier/ocm-client/internal/auth"
)

const baseURL = "https://4w0qhtalkj.execute-api.ap-southeast-2.amazonaws.com"

// Provider handles raw HTTP communication with the QPS OCM API.
// It returns raw response bytes; decoding is the caller's responsibility.
type Provider struct {
	hc     *http.Client
	apiKey string
}

// NewProvider creates a Provider. Pass nil for hc to use http.DefaultClient.
func NewProvider(hc *http.Client, apiKeyOverride string) *Provider {
	if hc == nil {
		hc = http.DefaultClient
	}
	key := auth.APIKey()
	if apiKeyOverride != "" {
		key = apiKeyOverride
	}
	return &Provider{hc: hc, apiKey: key}
}

// FetchLocations fetches the location lookup table (no auth required).
func (p *Provider) FetchLocations(ctx context.Context) ([]byte, error) {
	return p.get(ctx, baseURL+"/dev/lut", false)
}

// FetchOffences fetches offence data for the given date range and location codes.
// dateFrom and dateTo must be formatted as MM-DD-YYYY.
// codes must not be empty.
func (p *Provider) FetchOffences(ctx context.Context, dateFrom, dateTo string, codes []int) ([]byte, error) {
	if len(codes) == 0 {
		return nil, fmt.Errorf("api: FetchOffences: codes must not be empty")
	}
	parts := make([]string, len(codes))
	for i, c := range codes {
		parts[i] = strconv.Itoa(c)
	}
	url := fmt.Sprintf("%s/dev/offences/%s/%s/%s", baseURL, dateFrom, dateTo, strings.Join(parts, ","))
	return p.get(ctx, url, true)
}

func (p *Provider) get(ctx context.Context, url string, requiresAuth bool) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("api: build request: %w", err)
	}
	// Mimic the browser headers sent by the official QPS OCM web app.
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://qps-ocm.s3-ap-southeast-2.amazonaws.com")
	req.Header.Set("Referer", "https://qps-ocm.s3-ap-southeast-2.amazonaws.com/")
	if requiresAuth {
		req.Header.Set("X-API-Key", p.apiKey)
		req.Header.Set("Authorization", "Basic "+basicToken())
	}

	resp, err := p.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api: %s: %w", url, err)
	}
	defer resp.Body.Close() //nolint:all // who cares really if the body fails to close?

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("api: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api: %s: status %d: %s", url, resp.StatusCode, body)
	}
	return body, nil
}

// basicToken is a variable so tests can swap it out.
var basicToken = func() string { return auth.BasicToken() }
