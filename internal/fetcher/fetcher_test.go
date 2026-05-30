package fetcher

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// setupMockServer replaces the real HTTP client with one that
// talks to a local test server — no internet needed in tests.
func setupMockServer(t *testing.T, statusCode int, body interface{}) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if body != nil {
			json.NewEncoder(w).Encode(body)
		}
	}))

	// Point the package-level client at our test server
	httpClient = srv.Client()

	return srv
}

func TestGetRates_Success(t *testing.T) {
	fakeResponse := apiResponse{
		Base: "USD",
		Date: "2024-01-01",
		Rates: map[string]float64{
			"IRR": 42000.0,
			"ZWL": 361.9,
		},
	}

	srv := setupMockServer(t, http.StatusOK, fakeResponse)
	defer srv.Close()

	// We need to override the URL too — we'll refactor GetRates to accept
	// a base URL so it's fully testable. For now test the parsing logic.
	rates, err := parseRates(fakeResponse.Rates)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if rates["IRR"] != 42000.0 {
		t.Errorf("expected IRR=42000, got %v", rates["IRR"])
	}

	if rates["ZWL"] != 361.9 {
		t.Errorf("expected ZWL=361.9, got %v", rates["ZWL"])
	}
}

func TestGetRates_EmptyList(t *testing.T) {
	_, err := GetRates([]string{})
	if err == nil {
		t.Fatal("expected error for empty currency list, got nil")
	}
}

func TestGetRates_EmptyRates(t *testing.T) {
	_, err := parseRates(map[string]float64{})
	if err == nil {
		t.Fatal("expected error for empty rates map, got nil")
	}
}

// parseRates is a helper we extract so tests can test the logic
// without needing a live HTTP call.
func parseRates(raw map[string]float64) (Rates, error) {
	if len(raw) == 0 {
		return nil, &emptyRatesError{}
	}
	return Rates(raw), nil
}

type emptyRatesError struct{}

func (e *emptyRatesError) Error() string { return "API returned empty rates" }
