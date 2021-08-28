package chart

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/tquellenberg/weatherstation/datastore"
	"github.com/tquellenberg/weatherstation/sun"
)

type PageData struct {
	Temperature ChartData
	Pressure    ChartData
	Humidity    ChartData
	Xstart      string
	Xend        string
	Sunrise     string
	Sunset      string
}

type ChartData struct {
	Values    []Value
	LastValue float32
}

type Value struct {
	Time  string
	Value float32
}

func Index(w http.ResponseWriter, req *http.Request) {
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
	sunrise, sunset := sun.GetDayInfo()
	data := PageData{
		Sunrise: sunrise.Format(datastore.DateTimeFormat),
		Sunset:  sunset.Format(datastore.DateTimeFormat),
		Xstart:  xstart.Format(datastore.DateTimeFormat), Xend: xend.Format(datastore.DateTimeFormat)}

	tData, err := datastore.GetTemperatureSeries(xstart, xend)
	if err != nil {
		log.Println(err)
		return
	}
	values := []Value{}
	for _, v := range tData {
		values = append(values, Value{Time: v.Time, Value: v.Value})
	}
	data.Temperature.Values = values
	data.Temperature.LastValue = values[len(values)-1].Value

	hData, err := datastore.GetHumiditySeries(xstart, xend)
	if err != nil {
		log.Println(err)
		return
	}
	values = []Value{}
	for _, v := range hData {
		values = append(values, Value{Time: v.Time, Value: v.Value})
	}
	data.Humidity.Values = values
	data.Humidity.LastValue = values[len(values)-1].Value

	pData, err := datastore.GetPressureSeries(xstart, xend)
	if err != nil {
		log.Println(err)
		return
	}
	values = []Value{}
	for _, v := range pData {
		values = append(values, Value{Time: v.Time, Value: v.Value})
	}
	data.Pressure.Values = values
	data.Pressure.LastValue = values[len(values)-1].Value

	tmpl := template.Must(template.ParseFiles("index.html"))
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Print(err)
	}
}
