package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

//how many uint equal 1 USD

type Rates map[string]float64

// api response match the json structure of the response from the exchange rate API
type apiResponse struct {
	Base            string             `json:"base"`
	Date            string             `json:"date"`
	TimeLastUpdated time.Time          `json:"time_last_updated"`
	Rates           map[string]float64 `json:"rates"`
}

// http client so test  it out to swap out
var httpClient = &http.Client{Timeout: 10 * time.Second}

// fetch rates , the base currency is USD and we want to get the rates for the watchlist
// it returns a rates map or an error if something goes wrong
func GetRates(currencies []string) (Rates, error) {
	if len(currencies) == 0 {
		return nil, fmt.Errorf("currency list is empty")
	}
	symbols := strings.Join(currencies, ",")
	url := fmt.Sprintf("https://api.frankfurter.app/latest?base=USD&symbols=%s", symbols)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Rates) == 0 {
		return nil, fmt.Errorf("API returned empty rates")
	}

	return Rates(result.Rates), nil
}
