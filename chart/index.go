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
	Values []Value
}

type Value struct {
	Time  string
	Value float32
}

func Index(w http.ResponseWriter, _ *http.Request) {
	now := time.Now()
	year, month, day := now.Date()
	xstart := time.Date(year, month, day, 0, 0, 0, 0, now.Location()).Format(datastore.DateTimeFormat)
	xend := time.Date(year, month, day, 23, 59, 55, 0, now.Location()).Format(datastore.DateTimeFormat)
	sunrise, sunset := sun.GetDayInfo()
	data := PageData{
		Sunrise: sunrise.Format(datastore.DateTimeFormat),
		Sunset:  sunset.Format(datastore.DateTimeFormat),
		Xstart:  xstart, Xend: xend}

	tData, err := datastore.GetTemperatureSeries()
	if err != nil {
		log.Println(err)
		return
	}
	for _, v := range tData {
		data.Temperature.Values = append(data.Temperature.Values, Value{Time: v.Time, Value: v.Value})
	}

	hData, err := datastore.GetHumiditySeries()
	if err != nil {
		log.Println(err)
		return
	}
	for _, v := range hData {
		data.Humidity.Values = append(data.Humidity.Values, Value{Time: v.Time, Value: v.Value})
	}

	pData, err := datastore.GetPressureSeries()
	if err != nil {
		log.Println(err)
		return
	}
	for _, v := range pData {
		data.Pressure.Values = append(data.Pressure.Values, Value{Time: v.Time, Value: v.Value})
	}

	tmpl := template.Must(template.ParseFiles("index.html"))
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Print(err)
	}
}
