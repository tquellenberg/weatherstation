package chart

import (
	"log"

	"net/http"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"

	"github.com/tquellenberg/weatherstation/datastore"
)

func generateLineItems(entries []datastore.Entry) []opts.LineData {
	items := make([]opts.LineData, 0, len(entries))
	for _, v := range entries {
		items = append(items, opts.LineData{Value: []interface{}{v.Time, v.Value}, Symbol: "none"})
	}
	return items
}

func Httpserver(w http.ResponseWriter, _ *http.Request) {
	log.Println("Httpserver Request")

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
		}),
		charts.WithXAxisOpts(opts.XAxis{Type: "time"}),
		charts.WithYAxisOpts(opts.YAxis{Min: "dataMin", Max: "dataMax"}, 0))

	// Put data into instance
	line.
		AddSeries("Temperatur in Grade Celsius", generateLineItems(data)).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	line.Render(w)

	log.Println("Httpserver Request OKAY")
}
