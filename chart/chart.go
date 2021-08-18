package chart

import (
	"log"

	"net/http"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"

	"github.com/tquellenberg/weatherstation/datastore"
)

func generateLineItems(values []float32) []opts.LineData {
	items := make([]opts.LineData, 0)
	for _, v := range values {
		items = append(items, opts.LineData{Value: v})
	}
	return items
}

func values(entries []datastore.Entry) []float32 {
	var s []float32
	for _, v := range entries {
    	s = append(s, v.Value)
	}
	return s
}

func timeValues(entries []datastore.Entry) []string {
	var s []string
	for _, v := range entries {
    	s = append(s, v.Time)
	}
	return s
}

func Httpserver(w http.ResponseWriter, _ *http.Request) {
	data, err := datastore.GetTemperatureSeries()
	if err != nil {
		log.Fatal(err)
		return
	}

	// create a new line instance
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Weather Station",
			Subtitle: "Temperatur",
		}))

	// Put data into instance
	line.SetXAxis(timeValues(data)).
		AddSeries("Temperatur in Grade Celsius", generateLineItems(values(data))).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	line.Render(w)
}