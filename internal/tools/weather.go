package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go/v2"
)

type weather struct {
	Location  string
	TempC     float64
	Condition string
	WindKph   float64
	Days      []forecastDay
}
type forecastDay struct {
	Date      time.Time
	AvgTempC  float64
	Condition string
	WindKph   float64
}

type response struct {
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

// WeatherTool provides current weather conditions and a 3-day forecast
// for any given location using the WeatherAPI service.
type WeatherTool struct{}

func (WeatherTool) Name() string { return "get_weather" }

func (WeatherTool) Schema() openai.FunctionDefinitionParam {
	return openai.FunctionDefinitionParam{
		Name:        "get_weather",
		Description: openai.String("Get the current weather AND a 3-day forecast for the given location. Always include both in the reply."),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"location": map[string]string{
					"type": "string",
				},
			},
			"required": []string{"location"},
		},
	}
}

func (WeatherTool) Handle(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Location string `json:"location"`
	}
	if err := ParseArgs(args, &p); err != nil {
		return "", errors.New("could not parse location")
	}
	w, err := fetchWeather(ctx, p.Location)
	if err != nil {
		return "", errors.New("weather service unavailable")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %.1f°C, %s, wind %.1f km/h\n", w.Location, w.TempC, w.Condition, w.WindKph)

	for _, d := range w.Days {
		fmt.Fprintf(&b, "%s: %.1f°C avg, %s (wind %.1f km/h)\n",
			d.Date.Format("2006-01-02"), d.AvgTempC, d.Condition, d.WindKph)
	}
	return b.String(), nil
}

// fetchWeather fetches current conditions and a 3-day forecast for a given location.
func fetchWeather(ctx context.Context, location string) (*weather, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	url := fmt.Sprintf("https://api.weatherapi.com/v1/forecast.json?key=%s&q=%s&days=3",
		apiKey, url.QueryEscape(location))

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("forecast API returned %s", res.Status)
	}
	var data response
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

	return &weather{
		Location: data.Location.Name,
		TempC:    data.Current.TempC,
		WindKph:  data.Current.WindKph,
		Days:     days,
	}, nil
}
