package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type Weather struct {
	Location  string
	TempC     float64
	Condition string
	WindKph   float64
}

type Response struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
		WindKph float64 `json:"wind_kph"`
	} `json:"current"`
}

func FetchWeather(ctx context.Context, location string) (*Weather, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s",
		apiKey, url.QueryEscape(location))

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned %s", res.Status)
	}
	var data Response
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &Weather{
		Location:  data.Location.Name,
		TempC:     data.Current.TempC,
		Condition: data.Current.Condition.Text,
		WindKph:   data.Current.WindKph,
	}, nil
}
