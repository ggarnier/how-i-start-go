package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type openWeatherMap struct{}

func (w openWeatherMap) temperature(city string) (float64, error) {
	begin := time.Now()
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=" + city)
	log.Printf("openWeatherMap took %s", time.Since(begin).String())
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var d struct {
		Main struct {
			Kelvin float64 `json:"temp"`
		} `json:"main"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}

	log.Printf("openWeatherMap: %s: %.2f", city, d.Main.Kelvin)
	return d.Main.Kelvin, nil
}
