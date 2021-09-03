package chart

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/tquellenberg/weatherstation/datastore"
)

type CurrentDataPage struct {
}

type CurrentDataJson struct {
	CurrentTemperature float32 `json:"currentTemperature"`
	CurrentPressure    float32 `json:"currentPressure"`
	CurrentHumidity    float32 `json:"currentHumidity"`
	PressureTrend      string  `json:"pressureTrend"`
}

func CurrentValues(w http.ResponseWriter, req *http.Request) {
	log.Print("Get current values")
	values := datastore.GetLastValues()
	jsonData := CurrentDataJson{
		CurrentTemperature: values[0].Value,
		CurrentPressure:    values[1].Value,
		CurrentHumidity:    values[2].Value,
		PressureTrend:      datastore.GetPressureTrend(),
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonData)
}

func Index(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	err := tmpl.ExecuteTemplate(w, "index.html", CurrentDataPage{})
	if err != nil {
		log.Print(err)
	}
}
