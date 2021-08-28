package chart

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/tquellenberg/weatherstation/datastore"
	"github.com/tquellenberg/weatherstation/sun"
)

type PageData struct {
	TimeRange string
	Xstart    string
	Xend      string
	Sunrise   string
	Sunset    string
}

// Json: '{value:["2021-08-28 00:10:00", 14.22]}'
type JsonDataEntry struct {
	Value []interface{} `json:"value"`
}

func getTimeRange(req *http.Request) (start, end time.Time) {
	param1 := req.URL.Query().Get("range")
	xRange := 0
	switch param1 {
	case "week":
		xRange = 1
	default:
		xRange = 0
	}

	now := time.Now()
	year, month, day := now.Date()
	xstart := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	if xRange == 1 {
		xstart = xstart.AddDate(0, 0, -6)
	}
	xend := time.Date(year, month, day, 23, 59, 55, 0, now.Location())
	return xstart, xend
}

func TempData(w http.ResponseWriter, req *http.Request) {
	xstart, xend := getTimeRange(req)

	tData, _ := datastore.GetTemperatureSeries(xstart, xend)
	jsonData := make([]JsonDataEntry, 0, len(tData))
	for _, v := range tData {
		jsonData = append(jsonData, JsonDataEntry{Value: []interface{}{v.Time, v.Value}})
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonData)
}

func PressureData(w http.ResponseWriter, req *http.Request) {
	xstart, xend := getTimeRange(req)

	tData, _ := datastore.GetPressureSeries(xstart, xend)
	jsonData := make([]JsonDataEntry, 0, len(tData))
	for _, v := range tData {
		jsonData = append(jsonData, JsonDataEntry{Value: []interface{}{v.Time, v.Value}})
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonData)
}

func HumidityData(w http.ResponseWriter, req *http.Request) {
	xstart, xend := getTimeRange(req)

	tData, _ := datastore.GetHumiditySeries(xstart, xend)
	jsonData := make([]JsonDataEntry, 0, len(tData))
	for _, v := range tData {
		jsonData = append(jsonData, JsonDataEntry{Value: []interface{}{v.Time, v.Value}})
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonData)
}

func Index(w http.ResponseWriter, req *http.Request) {
	xstart, xend := getTimeRange(req)
	sunrise, sunset := sun.GetDayInfo()
	data := PageData{
		TimeRange: req.URL.Query().Get("range"),
		Sunrise:   sunrise.Format(datastore.DateTimeFormat),
		Sunset:    sunset.Format(datastore.DateTimeFormat),
		Xstart:    xstart.Format(datastore.DateTimeFormat),
		Xend:      xend.Format(datastore.DateTimeFormat)}

	tmpl := template.Must(template.ParseFiles("index.html"))
	err := tmpl.Execute(w, data)
	if err != nil {
		log.Print(err)
	}
}
