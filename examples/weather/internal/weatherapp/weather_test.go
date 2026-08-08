package weatherapp

import (
	"context"
	"image/color"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/tacf/goak"
	"github.com/tacf/goak/layout"
)

func TestSearchLocationsBuildsQueryAndDecodesResults(t *testing.T) {
	var received url.Values
	client := newWeatherClient()
	client.geocodingURL = "https://weather.test/search"
	client.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		received = request.URL.Query()
		return jsonResponse(`{
			"results":[{
				"id":2267057,
				"name":"Lisbon",
				"latitude":38.72509,
				"longitude":-9.1498,
				"timezone":"Europe/Lisbon",
				"country":"Portugal",
				"admin1":"Lisbon District"
			}]
		}`), nil
	})}

	places, err := client.searchLocations(context.Background(), " Lisbon ")
	if err != nil {
		t.Fatalf("search locations: %v", err)
	}
	if received.Get("name") != "Lisbon" || received.Get("count") != "6" ||
		received.Get("language") != "en" {
		t.Fatalf("unexpected query: %v", received)
	}
	if len(places) != 1 || places[0].Name != "Lisbon" ||
		places[0].displayName() != "Lisbon, Lisbon District, Portugal" {
		t.Fatalf("unexpected places: %#v", places)
	}
}

func TestSearchLocationsRejectsShortAndEmptyResults(t *testing.T) {
	client := newWeatherClient()
	if _, err := client.searchLocations(context.Background(), "x"); err == nil {
		t.Fatal("short query did not return an error")
	}

	client.geocodingURL = "https://weather.test/search"
	client.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(`{"results":[]}`), nil
	})}
	if _, err := client.searchLocations(context.Background(), "missing"); err == nil ||
		!strings.Contains(err.Error(), "no locations") {
		t.Fatalf("empty search error = %v", err)
	}
}

func TestForecastBuildsSevenDayQueryAndValidatesResponse(t *testing.T) {
	var received url.Values
	client := newWeatherClient()
	client.forecastURL = "https://weather.test/forecast"
	client.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		received = request.URL.Query()
		return jsonResponse(validForecastJSON), nil
	})}

	value, err := client.forecast(context.Background(), location{
		Name:      "Lisbon",
		Latitude:  38.72509,
		Longitude: -9.1498,
	})
	if err != nil {
		t.Fatalf("forecast: %v", err)
	}
	if received.Get("forecast_days") != "7" || received.Get("timezone") != "auto" {
		t.Fatalf("unexpected forecast settings: %v", received)
	}
	if !strings.Contains(received.Get("current"), "relative_humidity_2m") ||
		!strings.Contains(received.Get("daily"), "precipitation_probability_max") {
		t.Fatalf("missing forecast variables: %v", received)
	}
	if value.Current.WeatherCode != 1 || len(value.Daily.Time) != forecastDays {
		t.Fatalf("unexpected forecast: %#v", value)
	}
}

func TestForecastRejectsIncompleteDailyData(t *testing.T) {
	value := forecast{
		Daily: dailyWeather{
			Time:                     make([]string, forecastDays),
			WeatherCode:              make([]int, forecastDays),
			TemperatureMax:           make([]float64, forecastDays),
			TemperatureMin:           make([]float64, forecastDays),
			PrecipitationProbability: make([]int, forecastDays-1),
			Sunrise:                  make([]string, forecastDays),
			Sunset:                   make([]string, forecastDays),
			WindSpeedMax:             make([]float64, forecastDays),
		},
	}
	if err := validateForecast(value); err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("validation error = %v", err)
	}
}

func TestWeatherDescriptionsAndArtwork(t *testing.T) {
	cases := map[int]string{
		0:  "Clear sky",
		45: "Fog",
		63: "Rain",
		75: "Heavy snow",
		99: "Thunderstorm with hail",
	}
	for code, want := range cases {
		if got := weatherDescription(code); got != want {
			t.Errorf("weatherDescription(%d) = %q, want %q", code, got, want)
		}
	}

	clearDay := weatherArtwork(0, true, 220, 110)
	stormNight := weatherArtwork(95, false, 220, 110)
	if clearDay.Bounds().Dx() != 220 || clearDay.Bounds().Dy() != 110 {
		t.Fatalf("unexpected artwork bounds: %v", clearDay.Bounds())
	}
	dayPixel := color.RGBAModel.Convert(clearDay.At(0, 0)).(color.RGBA)
	nightPixel := color.RGBAModel.Convert(stormNight.At(0, 0)).(color.RGBA)
	if dayPixel == nightPixel {
		t.Fatalf("day and night artwork backgrounds are identical: %#v", dayPixel)
	}
}

func TestWeatherFormatting(t *testing.T) {
	if got, want := formatDay("2026-07-31"), "Fri 31 Jul"; got != want {
		t.Fatalf("formatDay = %q, want %q", got, want)
	}
	if got, want := clockPart("2026-07-31T06:42"), "06:42"; got != want {
		t.Fatalf("clockPart = %q, want %q", got, want)
	}
	if got, want := formatTemperatureRange(18.2, 25.4), "18 / 25°C"; got != want {
		t.Fatalf("formatTemperatureRange = %q, want %q", got, want)
	}
}

func TestWeatherUIFitsWindowAndRegistersInteractiveComponents(t *testing.T) {
	app := goak.NewApp()
	defer app.Destroy()
	ui, view := buildWeatherUI(app, newWeatherClient())
	layout.Layout(ui.Root().Container(), 900, 780)

	if len(ui.Images()) != 2 || len(ui.TextInputs()) != 1 || len(ui.Dropdowns()) != 1 {
		t.Fatalf(
			"unexpected component counts: images=%d inputs=%d dropdowns=%d",
			len(ui.Images()),
			len(ui.TextInputs()),
			len(ui.Dropdowns()),
		)
	}
	if view.locationPicker != ui.Dropdowns()[0] {
		t.Fatal("location picker was not registered")
	}
	if view.forecastGrid != ui.Images()[1] {
		t.Fatal("horizontal forecast grid was not registered")
	}
	content := ui.Root().Container().Children[0]
	last := content.Children[len(content.Children)-1].Bounds
	contentBottom := content.Bounds.Y + content.Bounds.H - content.Padding
	if last.Y+last.H > contentBottom {
		t.Fatalf("weather UI overflows its window: last=%#v content=%#v", last, content.Bounds)
	}
}

func TestForecastGridRendersSevenHorizontalCards(t *testing.T) {
	var value forecast
	client := newWeatherClient()
	client.forecastURL = "https://weather.test/forecast"
	client.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(validForecastJSON), nil
	})}
	var err error
	value, err = client.forecast(context.Background(), location{Latitude: 1, Longitude: 1})
	if err != nil {
		t.Fatalf("forecast fixture: %v", err)
	}

	grid := forecastGridArtwork(value.Daily, 850, 210)
	if grid.Bounds().Dx() != 850 || grid.Bounds().Dy() != 210 {
		t.Fatalf("unexpected grid bounds: %v", grid.Bounds())
	}
	active := color.RGBAModel.Convert(grid.At(20, 20)).(color.RGBA)
	inactive := color.RGBAModel.Convert(grid.At(140, 20)).(color.RGBA)
	if active == inactive {
		t.Fatalf("today card is not visually highlighted: active=%#v inactive=%#v", active, inactive)
	}
	for index := range forecastDays {
		cardCenterX := gridOuterPadding + index*((850-gridOuterPadding*2-gridCardGap*(forecastDays-1))/forecastDays+gridCardGap) + 10
		pixel := color.RGBAModel.Convert(grid.At(cardCenterX, 20)).(color.RGBA)
		if pixel == (color.RGBA{R: 17, G: 24, B: 39, A: 255}) {
			t.Fatalf("card %d was not rendered", index)
		}
	}
}

const validForecastJSON = `{
	"latitude":38.75,
	"longitude":-9.125,
	"timezone":"Europe/Lisbon",
	"current":{
		"time":"2026-07-31T07:30",
		"temperature_2m":19.5,
		"apparent_temperature":20.1,
		"relative_humidity_2m":84,
		"precipitation":0,
		"weather_code":1,
		"wind_speed_10m":8.2,
		"is_day":1
	},
	"daily":{
		"time":["2026-07-31","2026-08-01","2026-08-02","2026-08-03","2026-08-04","2026-08-05","2026-08-06"],
		"weather_code":[1,2,3,61,0,45,80],
		"temperature_2m_max":[28,27,26,24,29,25,23],
		"temperature_2m_min":[18,18,17,16,19,17,16],
		"precipitation_probability_max":[0,10,20,70,0,15,80],
		"sunrise":["2026-07-31T06:38","2026-08-01T06:39","2026-08-02T06:40","2026-08-03T06:41","2026-08-04T06:42","2026-08-05T06:43","2026-08-06T06:44"],
		"sunset":["2026-07-31T20:48","2026-08-01T20:47","2026-08-02T20:46","2026-08-03T20:45","2026-08-04T20:44","2026-08-05T20:43","2026-08-06T20:42"],
		"wind_speed_10m_max":[18,17,16,22,13,12,24]
	}
}`

type roundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}
}
