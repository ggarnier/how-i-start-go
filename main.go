package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	wundergroundAPIKey := flag.String("wunderground.api.key", "0123456789abcdef", "wunderground.com API key")
	flag.Parse()

	mw := multiWeatherProvider{
		openWeatherMap{},
		weatherUnderground{apiKey: *wundergroundAPIKey},
		yahooWeather{},
	}

	http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		city := strings.SplitN(r.URL.Path, "/", 3)[2]

		temp, err := mw.temperature(city)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"city": city,
			"temp": temp,
			"took": time.Since(begin).String(),
		})
	})

	http.ListenAndServe(":8080", nil)
}

type weatherProvider interface {
	temperature(city string) (float64, error) // in Kelvin, naturally
}

type multiWeatherProvider []weatherProvider

func (w multiWeatherProvider) temperature(city string) (float64, error) {
	timeoutMilliseconds := 500

	// Make a channel for temperatures, and a channel for errors.
	// Each provider will push a value into only one.
	temps := make(chan float64, len(w))
	errs := make(chan error, len(w))

	// For each provider, spawn a goroutine with an anonymous function.
	// That function will invoke the temperature method, and forward the response.
	for _, provider := range w {
		go func(p weatherProvider) {
			k, err := p.temperature(city)
			if err != nil {
				errs <- err
				return
			}
			temps <- k
		}(provider)
	}

	sum := 0.0
	size := len(w)

	// Collect a temperature or an error from each provider.
	for i := 0; i < len(w); i++ {
		select {
		case temp := <-temps:
			sum += temp
		case err := <-errs:
			return 0, err
		case <-time.After(time.Duration(timeoutMilliseconds) * time.Millisecond):
			log.Println("timeout")
			size = size - 1
		}
	}

	if size == 0 {
		return 0, fmt.Errorf("no API responded in less than %d ms", timeoutMilliseconds)
	}

	// Return the average, same as before.
	return sum / float64(size), nil
}
