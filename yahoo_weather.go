package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

type yahooWeather struct{}

func (w yahooWeather) temperature(city string) (float64, error) {
	begin := time.Now()
	resp, err := http.Get("https://query.yahooapis.com/v1/public/yql?q=select%20*%20from%20weather.forecast%20where%20woeid%20in%20(select%20woeid%20from%20geo.places(1)%20where%20text%3D%22" + city + "%22)&format=json&env=store%3A%2F%2Fdatatables.org%2Falltableswithkeys")
	log.Printf("yahooWeather took %s", time.Since(begin).String())
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var d struct {
		Query struct {
			Results struct {
				Channel struct {
					Item struct {
						Condition struct {
							Temp string `json:"temp"`
						} `json:"condition"`
					} `json:"item"`
				} `json:"channel"`
			} `json:"results"`
		} `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}

	tempFarenheit, err := strconv.ParseFloat(d.Query.Results.Channel.Item.Condition.Temp, 64)
	if err != nil {
		return 0, err
	}

	tempKelvin := (tempFarenheit + 459.67) * 5 / 9

	log.Printf("yahooWeather: %s: %.2f", city, tempKelvin)
	return tempKelvin, nil
}
