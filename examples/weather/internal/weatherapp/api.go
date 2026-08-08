package weatherapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultGeocodingURL = "https://geocoding-api.open-meteo.com/v1/search"
	defaultForecastURL  = "https://api.open-meteo.com/v1/forecast"
	maxResponseBytes    = 2 << 20
)

type weatherClient struct {
	httpClient   *http.Client
	geocodingURL string
	forecastURL  string
}

func newWeatherClient() *weatherClient {
	return &weatherClient{
		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},
		geocodingURL: defaultGeocodingURL,
		forecastURL:  defaultForecastURL,
	}
}

type location struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Country   string  `json:"country"`
	Admin1    string  `json:"admin1"`
}

func (l location) key() string {
	if l.ID != 0 {
		return strconv.FormatInt(l.ID, 10)
	}
	return fmt.Sprintf("%.4f,%.4f", l.Latitude, l.Longitude)
}

func (l location) displayName() string {
	parts := []string{l.Name}
	if l.Admin1 != "" && !strings.EqualFold(l.Admin1, l.Name) {
		parts = append(parts, l.Admin1)
	}
	if l.Country != "" {
		parts = append(parts, l.Country)
	}
	return strings.Join(parts, ", ")
}

type geocodingResponse struct {
	Results []location `json:"results"`
	Error   bool       `json:"error"`
	Reason  string     `json:"reason"`
}

func (client *weatherClient) searchLocations(ctx context.Context, query string) ([]location, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 2 {
		return nil, errors.New("enter at least two characters")
	}
	values := url.Values{
		"name":     {query},
		"count":    {"6"},
		"language": {"en"},
		"format":   {"json"},
	}
	var response geocodingResponse
	if err := client.getJSON(ctx, client.geocodingURL+"?"+values.Encode(), &response); err != nil {
		return nil, err
	}
	if response.Error {
		return nil, apiError(response.Reason)
	}
	if len(response.Results) == 0 {
		return nil, fmt.Errorf("no locations found for %q", query)
	}
	return response.Results, nil
}

type currentWeather struct {
	Time                string  `json:"time"`
	Temperature         float64 `json:"temperature_2m"`
	ApparentTemperature float64 `json:"apparent_temperature"`
	RelativeHumidity    int     `json:"relative_humidity_2m"`
	Precipitation       float64 `json:"precipitation"`
	WeatherCode         int     `json:"weather_code"`
	WindSpeed           float64 `json:"wind_speed_10m"`
	IsDay               int     `json:"is_day"`
}

type dailyWeather struct {
	Time                     []string  `json:"time"`
	WeatherCode              []int     `json:"weather_code"`
	TemperatureMax           []float64 `json:"temperature_2m_max"`
	TemperatureMin           []float64 `json:"temperature_2m_min"`
	PrecipitationProbability []int     `json:"precipitation_probability_max"`
	Sunrise                  []string  `json:"sunrise"`
	Sunset                   []string  `json:"sunset"`
	WindSpeedMax             []float64 `json:"wind_speed_10m_max"`
}

type forecast struct {
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	Timezone  string         `json:"timezone"`
	Current   currentWeather `json:"current"`
	Daily     dailyWeather   `json:"daily"`
	Error     bool           `json:"error"`
	Reason    string         `json:"reason"`
}

func (client *weatherClient) forecast(ctx context.Context, place location) (forecast, error) {
	values := url.Values{
		"latitude": {
			strconv.FormatFloat(place.Latitude, 'f', -1, 64),
		},
		"longitude": {
			strconv.FormatFloat(place.Longitude, 'f', -1, 64),
		},
		"current": {
			"temperature_2m,apparent_temperature,relative_humidity_2m," +
				"precipitation,weather_code,wind_speed_10m,is_day",
		},
		"daily": {
			"weather_code,temperature_2m_max,temperature_2m_min," +
				"precipitation_probability_max,sunrise,sunset,wind_speed_10m_max",
		},
		"timezone":      {"auto"},
		"forecast_days": {"7"},
	}
	var result forecast
	if err := client.getJSON(ctx, client.forecastURL+"?"+values.Encode(), &result); err != nil {
		return forecast{}, err
	}
	if result.Error {
		return forecast{}, apiError(result.Reason)
	}
	if err := validateForecast(result); err != nil {
		return forecast{}, err
	}
	return result, nil
}

func validateForecast(value forecast) error {
	const expectedDays = 7
	lengths := []int{
		len(value.Daily.Time),
		len(value.Daily.WeatherCode),
		len(value.Daily.TemperatureMax),
		len(value.Daily.TemperatureMin),
		len(value.Daily.PrecipitationProbability),
		len(value.Daily.Sunrise),
		len(value.Daily.Sunset),
		len(value.Daily.WindSpeedMax),
	}
	for _, length := range lengths {
		if length < expectedDays {
			return fmt.Errorf("Open-Meteo returned an incomplete %d-day forecast", expectedDays)
		}
	}
	return nil
}

func (client *weatherClient) getJSON(ctx context.Context, endpoint string, destination any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "goak-weather-demo/1.0")
	response, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("weather service: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 512))
		return fmt.Errorf("weather service returned %s: %s", response.Status, strings.TrimSpace(string(message)))
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes))
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode weather response: %w", err)
	}
	return nil
}

func apiError(reason string) error {
	if strings.TrimSpace(reason) == "" {
		reason = "unknown API error"
	}
	return fmt.Errorf("Open-Meteo: %s", reason)
}
