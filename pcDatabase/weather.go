package pcDatabase

import (
	"encoding/json"
	"fmt"
	forecast "github.com/mlbright/forecast/v2"
)

const (
	forecastPrefix     = "@forecast "
	FORECASTIO_API_KEY = "FILL_ME"
)

type WeatherCmdStruct struct {
	Cmd       string `json:"cmd"`
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
	Units     string `json:"units"`
	Language  string `json:"language"`
}

// EXAMPLE RETURNED JSON
// https://api.forecast.io/forecast/FILL_ME/37.8267,-122.423

func GetWeather(message string) string {
	// unmarshal message into latitude string, longitude string, unit string, language string
	weatherdata := WeatherCmdStruct{}
	err := json.Unmarshal([]byte(message), &weatherdata)
	if err != nil {
		ERROR.Println("error in GetWeather Unmarshalling into WeatherCmdStruct:", err)
		return message // return original message back
	}
	// now use API to get forecast
	f, err := forecast.Get(FORECASTIO_API_KEY, weatherdata.Latitude, weatherdata.Longitude, "now", forecast.AUTO)
	if err != nil {
		ERROR.Println(err)
	}
	// forecastBytes, err := json.Marshal(f)
	// if err != nil {
	// 	ERROR.Println("error in json.Marshal in GetWeather:")
	// 	ERROR.Println(err)
	// } else {
	//     TRACE.Println(string(forecastBytes))
	// }
	// create string to send
	// first, determine which weather-icon to include
	wIcon := "<i class=\"wi "
	switch f.Hourly.Icon {
	case "clear-day":
		wIcon += "wi-day-sunny"
	case "clear-night":
		wIcon += "wi-night-clear"
	case "rain":
		wIcon += "wi-day-rain"
	case "snow":
		wIcon += "wi-day-snow"
	case "sleet":
		wIcon += "wi-day-sleet"
	case "wind":
		wIcon += "wi-day-windy"
	case "fog":
		wIcon += "wi-day-fog"
	case "cloudy":
		wIcon += "wi-day-cloudy"
	case "partly-cloudy-day":
		wIcon += "wi-day-cloudy"
	case "partly-cloudy-night":
		wIcon += "wi-night-partly-cloudy"
	case "hail":
		wIcon += "wi-day-hail"
	case "thunderstorm":
		wIcon += "wi-day-thunderstorm"
	case "tornado":
		wIcon += "wi-tornado"
	default:
		wIcon += "na"
	}
	wIcon += "\"></i>"
	// now create HTML string to return
	retstr := `<div id="forecast-embedded">` +
		`<div id="forecast-header">` +
		`<i>Forecast provided by <a target="_blank" href="http://forecast.io/#/f/33.7748323,-84.3804785">forecast.io</a></i>` +
		`</div>` +
		`<div id="forecast-main">` +
		`<span id="forecast-icon">` +
		wIcon +
		`</span>` +
		`<span id="forecast-temp">` +
		fmt.Sprintf("%2.0f", f.Currently.Temperature) +
		`<i class="wi wi-fahrenheit"></i>` +
		`</span>` +
		`</div>` +
		`<div id="forecast-summary">` +
		f.Hourly.Summary +
		`</div>` +
		`</div>`
	TRACE.Println("weather string: " + retstr)
	return retstr

	// cheap solution
	// return `<iframe id="forecast_embed" type="text/html" frameborder="0" height="245" width="100%" src="http://forecast.io/embed/#lat=` + weatherdata.Latitude + `&lon=` + weatherdata.Longitude + `"> </iframe>`;
}
