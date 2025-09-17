package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type weather struct {
	Location  string
	TempC     float64
	Condition string
	WindKph   float64
}

type currentWeatherResponse struct {
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
type forecastDay struct {
	Date      time.Time
	AvgTempC  float64
	Condition string
	WindKph   float64
}
type Forecast struct {
	Days []forecastDay
}
type forecastResponse struct {
	Forecast struct {
		Forecastday []struct {
			Date string `json:"date"`
			Day  struct {
				AvgTempC  float64 `json:"avgtemp_c"`
				Condition struct {
					Text string `json:"text"`
				} `json:"condition"`
				MaxWindKph float64 `json:"maxwind_kph"`
			} `json:"day"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

// FetchWeather fetches current conditions for a given location.
func FetchWeather(ctx context.Context, location string) (*weather, error) {
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
	var data currentWeatherResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &weather{
		Location:  data.Location.Name,
		TempC:     data.Current.TempC,
		Condition: data.Current.Condition.Text,
		WindKph:   data.Current.WindKph,
	}, nil
}

// FetchForecast fetches a 3-day forecast for a given location.
func FetchForecast(ctx context.Context, location string) (*Forecast, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	url := fmt.Sprintf("https://api.weatherapi.com/v1/forecast.json?key=%s&q=%s&days=3",
		apiKey, url.QueryEscape(location))

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("forecast API returned %s", res.Status)
	}
	var data forecastResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	days := make([]forecastDay, 0, len(data.Forecast.Forecastday))
	for _, d := range data.Forecast.Forecastday {
		t, _ := time.Parse("2006-01-02", d.Date)
		days = append(days, forecastDay{
			Date:      t,
			AvgTempC:  d.Day.AvgTempC,
			Condition: d.Day.Condition.Text,
			WindKph:   d.Day.MaxWindKph,
		})
	}

	return &Forecast{
		Days: days,
	}, nil
}
