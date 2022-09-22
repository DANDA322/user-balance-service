package converter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

const base = "RUB"

type RemoteConverter struct {
	URL    string
	apiKey string
	client *http.Client
}

type Response struct {
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
	Success   bool               `json:"success"`
	Timestamp int                `json:"timestamp"`
}

func NewConverter(url, apiKey string) *RemoteConverter {
	return &RemoteConverter{
		URL:    url,
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (c *RemoteConverter) GetRate(ctx context.Context, currency string) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return 0, err
	}
	values := url.Values{}
	values.Add("symbols", currency)
	values.Add("base", base)
	req.URL.RawQuery = values.Encode()
	req.Header.Set("apikey", c.apiKey)
	res, err := c.client.Do(req)
	if res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return 0, err
	}
	response := Response{}
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return 0, err
	}
	return response.Rates[currency], nil
}
