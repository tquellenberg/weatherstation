package chart

import (
	"log"

	"net/http"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"

	"github.com/tquellenberg/weatherstation/datastore"
	"github.com/tquellenberg/weatherstation/sun"
)

func generateLineItems(entries []datastore.Entry) []opts.LineData {
	items := make([]opts.LineData, 0, len(entries))
	for _, v := range entries {
		items = append(items, opts.LineData{Value: []interface{}{v.Time, v.Value}, Symbol: "none"})
	}
	return items
}

func lineChart(name string, data []datastore.Entry) *charts.Line {
	// create a new line instance
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{Title: name}),
		charts.WithXAxisOpts(opts.XAxis{Type: "time"}),
		charts.WithYAxisOpts(opts.YAxis{Min: "dataMin", Max: "dataMax"}, 0))
	// Put data into instance
	sunrise, sunset := sun.GetDayInfo()
	line.
		AddSeries(name, generateLineItems(data)).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}),
			charts.WithMarkLineNameXAxisItemOpts(opts.MarkLineNameXAxisItem{
				Name:  "sunrise",
				XAxis: sunrise.Format(datastore.DateTimeFormat)}),
			charts.WithMarkLineNameXAxisItemOpts(opts.MarkLineNameXAxisItem{
				Name:  "sunset",
				XAxis: sunset.Format(datastore.DateTimeFormat)}))
	return line
}

func Httpserver(w http.ResponseWriter, _ *http.Request) {
	log.Println("Httpserver Request")

	tData, err := datastore.GetTemperatureSeries()
	if err != nil {
		log.Println(err)
		return
	}

	pData, err := datastore.GetPressureSeries()
	if err != nil {
		log.Println(err)
		return
	}

	hData, err := datastore.GetHumiditySeries()
	if err != nil {
		log.Println(err)
		return
	}

	page := components.NewPage()
	page.AddCharts(
		lineChart("Temperatur", tData),
		lineChart("Pressure", pData),
		lineChart("Humidity", hData))
	page.Render(w)

	log.Println("Httpserver Request OKAY")
}
