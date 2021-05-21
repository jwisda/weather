package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

//basic weather struct defined in requirements
type weather struct {
	Temperature string `json:"temp"`
	Conditions  string `json:"conditions"`
	Alerts      string `json:"alerts"`
}

//only using 1 api id for now but could use more in the future for concurrent access
var APIIDS = [1]string{"24572ec02742e943b625df83a904b2ca"}
var getWeatherUrl = "https://api.openweathermap.org/data/2.5/onecall?lat=%v&lon=%v&exclude=minutely,hourly,daily&units=imperial&appid=%v"
var hot = 85.0
var cold = 40.0
var temps = [3]string{"Cold", "Moderate", "Hot"}

func main() {
	http.HandleFunc("/forecast", ForecastHandler)
	http.HandleFunc("/ping", Pong)

	log.Print("server starting at localhost:8080 ... ")
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Printf("the HTTP server failed to start: %s", err)
		os.Exit(1)
	}
}

func Pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Pong!!"))
}

func ForecastHandler(w http.ResponseWriter, r *http.Request) {
	//using default values for now, could return an error in the future
	lattitude := "33.44"
	longitude := "-94.04"

	if coords := r.URL.Query().Get("coords"); len(coords) > 0 {
		if c := strings.Split(coords, ","); len(c) > 1 {
			lattitude = strings.TrimSpace(c[0])
			longitude = strings.TrimSpace(c[1])
		}
	}

	if lat := r.URL.Query().Get("lat"); len(lat) > 0 {
		lattitude = lat
	}
	if lon := r.URL.Query().Get("lon"); len(lon) > 0 {
		longitude = lon
	}

	forecast := getForecastFromOpenWeather(lattitude, longitude)
	temperature := temps[1]
	conditions := ""
	alerts := ""

	//temperature
	//temperature
	if current, ok := forecast["current"].(map[string]interface{}); ok {
		currentTemp := current["temp"].(float64)
		if currentTemp > hot {
			temperature = temps[2]
		}
		if currentTemp < cold {
			temperature = temps[0]
		}

	}
	//conditions
	if currentWeather, ok := forecast["current"].(map[string]interface{}); ok {
		if weatherList, aok := currentWeather["weather"].([]interface{}); aok {
			if len(weatherList) > 0 {
				weather := weatherList[0].(map[string]interface{})
				conditions = fmt.Sprintf("%v: %v", weather["main"], weather["description"])
			}
		}
	}

	//alerts
	if alertList, ok := forecast["alerts"].([]interface{}); ok {
		if len(alertList) > 0 {
			a := alertList[0].(map[string]interface{})
			alerts = fmt.Sprintf("%v", a["event"])
		}
	}

	if json, err := json.Marshal(weather{temperature, conditions, alerts}); err == nil {
		w.Write(json)
		return
	}

	w.Write([]byte("Unable to return forecast at this time"))
}

//could add a bunch of checks for error conditions but for now should just fail nicely
func getForecastFromOpenWeather(lat, long string) map[string]interface{} {
	url := fmt.Sprintf(getWeatherUrl, lat, long, APIIDS[0])
	req, _ := http.NewRequest("GET", url, bytes.NewReader([]byte("")))
	h := req.Header
	h.Add("Accept", "application/json")
	h.Add("Connection", "close")

	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	var forecast map[string]interface{}
	json.Unmarshal(body, &forecast)
	return forecast
}
