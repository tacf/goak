package weatherapp

import (
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"goak/internal/goak"
	"goak/internal/goak/colors"
	"goak/internal/goak/components"
	"goak/internal/goak/layout"
)

const forecastDays = 7

type savedWeather struct {
	place    location
	forecast forecast
}

type weatherView struct {
	app    *goak.App
	client *weatherClient

	searchInput    *components.TextInput
	searchButton   *components.Button
	savedButton    *components.Button
	locationPicker *components.Dropdown
	status         *components.Label

	locationLabel *components.Label
	todayImage    *components.Image
	condition     *components.Label
	temperature   *components.Label
	details       *components.Label
	sun           *components.Label
	forecastGrid  *components.Image

	searchSequence       atomic.Uint64
	forecastSequence     atomic.Uint64
	searchMatches        []location
	saved                []savedWeather
	showingSearchResults bool
}

// Run creates and runs the weather demo application.
func Run() {
	app := goak.NewApp()
	defer app.Destroy()
	app.InitWindow("Goak Weather", 900, 780)
	app.SetAutoDPI(true)

	ui, view := buildWeatherUI(app, newWeatherClient())
	view.loadDefault("Lisbon")
	app.Run(ui)
}

func buildWeatherUI(app *goak.App, client *weatherClient) (*components.UI, *weatherView) {
	ui := components.NewUI()
	ui.Theme().Background = colors.HexOr("#111827", colors.RGB(17, 24, 39))
	root := ui.Root()
	root.SetPadding(14)

	content := root.CreatePanel(layout.PercentOf(100), layout.PercentOf(100))
	content.SetPadding(12)
	content.SetBackground(colors.HexOr("#172033", colors.RGB(23, 32, 51)))
	content.SetAlignment(layout.AlignCenter, layout.AlignStart)

	title := content.CreateLabel(layout.PercentOf(100), layout.StaticPx(38), "GOAK WEATHER")
	title.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	title.SetColor(colors.HexOr("#f8fafc", colors.White))

	searchInput := content.CreateTextInput(layout.PercentOf(100), layout.StaticPx(38), "")
	searchInput.SetPlaceholder("Search city or postal code")
	searchButton := content.CreateButton(layout.StaticPx(220), layout.StaticPx(34), "Search locations")
	savedButton := content.CreateButton(layout.StaticPx(220), layout.StaticPx(32), "Show saved locations")
	locationPicker := content.CreateDropdown(
		layout.PercentOf(100),
		layout.StaticPx(34),
		"Saved locations",
		nil,
	)
	status := content.CreateLabel(layout.PercentOf(100), layout.StaticPx(24), "Loading Lisbon…")
	status.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	status.SetColor(colors.HexOr("#93c5fd", colors.RGB(147, 197, 253)))

	locationLabel := content.CreateLabel(layout.PercentOf(100), layout.StaticPx(34), "Today")
	locationLabel.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	locationLabel.SetColor(colors.HexOr("#f8fafc", colors.White))
	todayImage := content.CreateImage(layout.PercentOf(100), layout.StaticPx(92), nil)
	todayImage.SetFit(components.ImageFitContain)
	condition := content.CreateLabel(layout.PercentOf(100), layout.StaticPx(26), "Waiting for weather…")
	condition.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	temperature := content.CreateLabel(layout.PercentOf(100), layout.StaticPx(28), "")
	temperature.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	temperature.SetColor(colors.HexOr("#fbbf24", colors.RGB(251, 191, 36)))
	details := content.CreateLabel(layout.PercentOf(100), layout.StaticPx(24), "")
	details.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	sun := content.CreateLabel(layout.PercentOf(100), layout.StaticPx(24), "")
	sun.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	sun.SetColor(colors.HexOr("#cbd5e1", colors.RGB(203, 213, 225)))

	previewTitle := content.CreateLabel(layout.PercentOf(100), layout.StaticPx(28), "7-DAY PREVIEW")
	previewTitle.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	previewTitle.SetColor(colors.HexOr("#93c5fd", colors.RGB(147, 197, 253)))
	forecastGrid := content.CreateImage(layout.PercentOf(100), layout.StaticPx(210), nil)
	forecastGrid.SetFit(components.ImageFitContain)

	view := &weatherView{
		app:            app,
		client:         client,
		searchInput:    searchInput,
		searchButton:   searchButton,
		savedButton:    savedButton,
		locationPicker: locationPicker,
		status:         status,
		locationLabel:  locationLabel,
		todayImage:     todayImage,
		condition:      condition,
		temperature:    temperature,
		details:        details,
		sun:            sun,
		forecastGrid:   forecastGrid,
	}
	attribution := content.CreateLabel(
		layout.PercentOf(100),
		layout.StaticPx(20),
		"Weather data by Open-Meteo · Location search by GeoNames",
	)
	attribution.SetAlignment(layout.AlignCenter, layout.AlignCenter)
	attribution.SetColor(colors.HexOr("#94a3b8", colors.RGB(148, 163, 184)))

	search := func() {
		view.search(searchInput.Text())
	}
	searchInput.SetOnSubmitted(func(components.TextInputSubmittedEvent) { search() })
	searchButton.SetOnClick(func(components.ButtonClickEvent) { search() })
	savedButton.SetOnClick(func(components.ButtonClickEvent) {
		view.showSavedLocations()
	})
	locationPicker.SetOnChanged(func(event components.DropdownChangedEvent) {
		if view.showingSearchResults {
			if event.Index >= 0 && event.Index < len(view.searchMatches) {
				view.addLocation(view.searchMatches[event.Index])
			}
			return
		}
		if event.Index >= 0 && event.Index < len(view.saved) {
			view.show(view.saved[event.Index])
		}
	})

	return ui, view
}

func (view *weatherView) loadDefault(query string) {
	go func() {
		places, err := view.client.searchLocations(view.app.Context(), query)
		if err != nil {
			view.dispatchError("Could not load the default location", err)
			return
		}
		if len(places) == 0 || view.searchSequence.Load() != 0 {
			return
		}
		if view.forecastSequence.CompareAndSwap(0, 1) {
			view.fetchForecast(places[0], 1)
		}
	}()
}

func (view *weatherView) search(query string) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 2 {
		view.status.SetText("Enter at least two characters to search.")
		return
	}
	sequence := view.searchSequence.Add(1)
	view.showingSearchResults = true
	view.searchButton.SetLabel("Searching…")
	view.locationPicker.SetLabel("Searching locations…")
	view.locationPicker.SetOptions(nil)
	view.status.SetText("Searching Open-Meteo locations…")
	go func() {
		places, err := view.client.searchLocations(view.app.Context(), query)
		_ = view.app.Dispatch(func() {
			if sequence != view.searchSequence.Load() {
				return
			}
			view.searchButton.SetLabel("Search locations")
			if err != nil {
				view.status.SetText("Search failed: " + err.Error())
				view.searchMatches = nil
				view.locationPicker.SetOptions(nil)
				return
			}
			view.searchMatches = places
			options := make([]components.DropdownOption, len(places))
			for index, place := range places {
				options[index] = components.DropdownOption{
					Label: place.displayName(),
					Value: place.key(),
				}
			}
			view.showingSearchResults = true
			view.locationPicker.SetLabel("Search results — select to add")
			view.locationPicker.SetOptions(options)
			view.status.SetText(fmt.Sprintf("%d matches — select one to add.", len(options)))
		})
	}()
}

func (view *weatherView) addLocation(place location) {
	view.showingSearchResults = false
	view.locationPicker.SetLabel("Loading " + place.Name + "…")
	view.locationPicker.SetOptions(nil)
	view.status.SetText("Loading forecast for " + place.displayName() + "…")
	view.requestForecast(place)
}

func (view *weatherView) requestForecast(place location) {
	sequence := view.forecastSequence.Add(1)
	view.fetchForecast(place, sequence)
}

func (view *weatherView) fetchForecast(place location, sequence uint64) {
	go func() {
		value, err := view.client.forecast(view.app.Context(), place)
		_ = view.app.Dispatch(func() {
			if err != nil {
				if sequence != view.forecastSequence.Load() || view.showingSearchResults {
					return
				}
				view.status.SetText("Forecast failed: " + err.Error())
				view.showSavedLocations()
				return
			}
			saved := savedWeather{place: place, forecast: value}
			index := view.savedIndex(place.key())
			if index < 0 {
				view.saved = append(view.saved, saved)
				index = len(view.saved) - 1
			} else {
				view.saved[index] = saved
			}
			if sequence != view.forecastSequence.Load() || view.showingSearchResults {
				return
			}
			view.syncSavedOptions()
			view.locationPicker.SetSelectedIndex(index)
			view.searchInput.SetText("")
			view.status.SetText("Forecast updated " + formatUpdatedTime(value.Current.Time) + ".")
		})
	}()
}

func (view *weatherView) showSavedLocations() {
	view.showingSearchResults = false
	view.syncSavedOptions()
	if len(view.saved) == 0 {
		view.status.SetText("No saved locations yet — search for one to add.")
	} else {
		view.status.SetText("Select a saved location to change the weather view.")
	}
}

func (view *weatherView) syncSavedOptions() {
	options := make([]components.DropdownOption, len(view.saved))
	for optionIndex, item := range view.saved {
		options[optionIndex] = components.DropdownOption{
			Label: item.place.displayName(),
			Value: item.place.key(),
		}
	}
	view.locationPicker.SetLabel("Saved locations")
	view.locationPicker.SetOptions(options)
}

func (view *weatherView) savedIndex(key string) int {
	for index, item := range view.saved {
		if item.place.key() == key {
			return index
		}
	}
	return -1
}

func (view *weatherView) show(value savedWeather) {
	current := value.forecast.Current
	daily := value.forecast.Daily
	view.locationLabel.SetText("TODAY · " + value.place.displayName())
	view.todayImage.SetSource(weatherArtwork(current.WeatherCode, current.IsDay == 1, 440, 180))
	view.condition.SetText(weatherDescription(current.WeatherCode))
	view.temperature.SetText(fmt.Sprintf(
		"%.0f°C · feels like %.0f°C",
		current.Temperature,
		current.ApparentTemperature,
	))
	view.details.SetText(fmt.Sprintf(
		"Humidity %d%%  ·  Wind %.0f km/h  ·  Precipitation %.1f mm",
		current.RelativeHumidity,
		current.WindSpeed,
		current.Precipitation,
	))
	view.sun.SetText(fmt.Sprintf(
		"Sunrise %s  ·  Sunset %s  ·  %s",
		clockPart(daily.Sunrise[0]),
		clockPart(daily.Sunset[0]),
		value.forecast.Timezone,
	))
	view.forecastGrid.SetSource(forecastGridArtwork(daily, 1700, 420))
}

func (view *weatherView) dispatchError(prefix string, err error) {
	if dispatchErr := view.app.Dispatch(func() {
		view.status.SetText(prefix + ": " + err.Error())
	}); dispatchErr != nil && dispatchErr != goak.ErrAppStopped {
		log.Printf("dispatch weather error: %v", dispatchErr)
	}
}

func formatDay(value string) string {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return parsed.Format("Mon 2 Jan")
}

func clockPart(value string) string {
	if len(value) >= 16 {
		return value[11:16]
	}
	return value
}

func formatUpdatedTime(value string) string {
	if len(value) >= 16 {
		return "at " + value[11:16]
	}
	return "just now"
}
