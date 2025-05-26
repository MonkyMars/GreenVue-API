package location

import (
	"encoding/json"
	"fmt"
	"greenvue/lib"
	"net/url"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
)

var client = resty.New().SetTimeout(10 * time.Second)

type OpenCageResponse struct {
	Results []struct {
		Geometry struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geometry"`
		Components struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"components"`
	} `json:"results"`
}

func GetFullLocation(country, city string) (lib.Location, error) {
	if country == "" || city == "" {
		return lib.Location{}, fmt.Errorf("country and city must not be empty")
	}

	query := url.QueryEscape(fmt.Sprintf("%s, %s", city, country))
	apiKey := os.Getenv("OPENCAGE_API_KEY")

	url := fmt.Sprintf("https://api.opencagedata.com/geocode/v1/json?q=%s&key=%s", query, apiKey)

	resp, err := client.R().Get(url)
	if err != nil {
		return lib.Location{}, fmt.Errorf("failed to fetch location: %w", err)
	}

	body := resp.Body()

	if resp.StatusCode() != 200 {
		return lib.Location{}, fmt.Errorf("failed to fetch location: %s", string(body))
	}

	var data OpenCageResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return lib.Location{}, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(data.Results) == 0 {
		return lib.Location{}, fmt.Errorf("no results found for %s, %s", city, country)
	}

	// Validate that the first result matches the expected city and country
	if data.Results[0].Components.City != city {
		return lib.Location{}, fmt.Errorf("Please give a valid city name")
	}

	lat := data.Results[0].Geometry.Lat
	lng := data.Results[0].Geometry.Lng

	location := lib.Location{
		Country:   country,
		City:      city,
		Latitude:  lat,
		Longitude: lng,
	}

	return location, nil
}
