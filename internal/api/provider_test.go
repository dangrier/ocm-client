package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestProvider returns a Provider pointed at srv with the basicToken
// function swapped out so tests don't depend on real timestamps.
func newTestProvider(srv *httptest.Server) *Provider {
	p := NewProvider(srv.Client(), "test-key")
	basicToken = func() string { return "test-token" }
	// Redirect the hardcoded baseURL by wrapping the transport to rewrite hosts.
	p.hc = &http.Client{
		Transport: &rewriteTransport{base: srv.URL, inner: srv.Client().Transport},
	}
	return p
}

// rewriteTransport rewrites every request to target the test server.
type rewriteTransport struct {
	base  string
	inner http.RoundTripper
}

func (r *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	// Replace the host portion so requests reach the test server.
	parsed, _ := http.NewRequest(req.Method, r.base+req.URL.RequestURI(), req.Body)
	parsed.Header = req.Header
	return r.inner.RoundTrip(parsed)
}

func TestFetchLocations_Headers(t *testing.T) {
	var gotHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := newTestProvider(srv)
	p.get(context.Background(), srv.URL+"/dev/lut", false) //nolint

	for _, h := range []string{"Accept", "Origin", "Referer"} {
		if gotHeaders.Get(h) == "" {
			t.Errorf("expected header %q to be set on unauthenticated request", h)
		}
	}
	if gotHeaders.Get("X-API-Key") != "" {
		t.Error("X-API-Key should NOT be set on unauthenticated request")
	}
	if gotHeaders.Get("Authorization") != "" {
		t.Error("Authorization should NOT be set on unauthenticated request")
	}
}

func TestFetchOffences_AuthHeaders(t *testing.T) {
	var gotHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := newTestProvider(srv)
	p.get(context.Background(), srv.URL+"/dev/offences/01-01-2026/01-31-2026/1", true) //nolint

	if gotHeaders.Get("X-API-Key") != "test-key" {
		t.Errorf("X-API-Key = %q, want %q", gotHeaders.Get("X-API-Key"), "test-key")
	}
	if gotHeaders.Get("Authorization") != "Basic test-token" {
		t.Errorf("Authorization = %q, want %q", gotHeaders.Get("Authorization"), "Basic test-token")
	}
}

func TestGet_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	p := newTestProvider(srv)
	_, err := p.get(context.Background(), srv.URL+"/any", false)
	if err == nil {
		t.Fatal("want error for 403 response, got nil")
	}
}

func TestFetchOffences_EmptyCodes(t *testing.T) {
	p := NewProvider(nil, "")
	_, err := p.FetchOffences(context.Background(), "01-01-2026", "01-31-2026", nil)
	if err == nil {
		t.Fatal("want error for empty codes, got nil")
	}
}
