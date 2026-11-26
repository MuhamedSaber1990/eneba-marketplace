package httpapi

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func intPtr(n int) *int { return &n }

// samePrice compares two optional prices: both missing, or both the same value.
func samePrice(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func TestParseListParams(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantErr     string // part of the error message, "" when no error is expected
		wantSort    string
		wantPage    int
		wantMin     *int
		wantMax     *int
		wantQuery   string
		wantPlatLen int
	}{
		{name: "no parameters uses the defaults", query: "", wantSort: "newest", wantPage: 1},
		{name: "several platforms", query: "platform=Steam&platform=Xbox", wantSort: "newest", wantPage: 1, wantPlatLen: 2},
		{name: "sort and page", query: "sort=price_asc&page=3", wantSort: "price_asc", wantPage: 3},
		{name: "prices in euros become cents", query: "min_price=10&max_price=39.99", wantSort: "newest", wantPage: 1, wantMin: intPtr(1000), wantMax: intPtr(3999)},
		{name: "a price of zero is a real limit", query: "min_price=0", wantSort: "newest", wantPage: 1, wantMin: intPtr(0)},
		{name: "search text is trimmed, and a search sorts by relevance", query: "q=++crown++", wantSort: "relevance", wantPage: 1, wantQuery: "crown"},
		{name: "a search can still be sorted another way", query: "q=crown&sort=price_asc", wantSort: "price_asc", wantPage: 1, wantQuery: "crown"},
		{name: "relevance is accepted without a search", query: "sort=relevance", wantSort: "relevance", wantPage: 1},

		{name: "unknown sort", query: "sort=price", wantErr: "unknown sort"},
		{name: "sort cannot carry SQL", query: "sort=name%3BDROP+TABLE+products", wantErr: "unknown sort"},
		{name: "page zero", query: "page=0", wantErr: "page must be"},
		{name: "page is not a number", query: "page=two", wantErr: "page must be"},
		{name: "negative price", query: "min_price=-5", wantErr: "min_price must be"},
		{name: "price is not a number", query: "max_price=cheap", wantErr: "max_price must be"},
		{name: "price is NaN", query: "max_price=NaN", wantErr: "max_price must be"},
		{name: "min above max", query: "min_price=50&max_price=5", wantErr: "higher than max_price"},
		{name: "search text too long", query: "q=" + strings.Repeat("a", 101), wantErr: "q is too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("bad test query: %v", err)
			}

			got, err := parseListParams(values)

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want one containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Sort != tt.wantSort {
				t.Errorf("Sort = %q, want %q", got.Sort, tt.wantSort)
			}
			if got.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", got.Page, tt.wantPage)
			}
			if !samePrice(got.MinPriceCents, tt.wantMin) {
				t.Errorf("MinPriceCents = %v, want %v", got.MinPriceCents, tt.wantMin)
			}
			if !samePrice(got.MaxPriceCents, tt.wantMax) {
				t.Errorf("MaxPriceCents = %v, want %v", got.MaxPriceCents, tt.wantMax)
			}
			if got.Query != tt.wantQuery {
				t.Errorf("Query = %q, want %q", got.Query, tt.wantQuery)
			}
			if len(got.Platforms) != tt.wantPlatLen {
				t.Errorf("len(Platforms) = %d, want %d", len(got.Platforms), tt.wantPlatLen)
			}
		})
	}
}

// Without a database, every route that needs one must answer 503 in JSON
// instead of crashing on the missing pool.
func TestRoutesWithoutDatabase(t *testing.T) {
	router := NewRouter(nil, Config{})

	for _, path := range []string{
		"/api/products", "/api/products/some-slug", "/api/products/some-slug/cover.svg", "/api/filters", "/api/search/suggest?q=kart",
		"/api/auth/me", "/api/orders", "/api/orders/x", "/api/home", "/api/home/banners", "/api/home/deals", "/api/home/best-sellers", "/api/home/new-releases", "/api/home/platforms",
	} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))

			if response.Code != http.StatusServiceUnavailable {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
			}
			if !strings.Contains(response.Body.String(), `"error"`) {
				t.Fatalf("body is not a JSON error: %s", response.Body.String())
			}
		})
	}
}
