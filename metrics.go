package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tquellenberg/weatherstation/bme280"
)

var (
	tempGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "tomsweather",
			Name:      "temperature",
			Help:      "Temperature in degrees Celsius"})
	pressureGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "tomsweather",
			Name:      "pressure",
			Help:      "Air pressure in hectopascal"})
	humidityGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "tomsweather",
			Name:      "humidity",
			Help:      "Humidity in percent"})
)

func InitMetrics() {
	prometheus.MustRegister(tempGauge)
	prometheus.MustRegister(pressureGauge)
	prometheus.MustRegister(humidityGauge)
}

func UpdateMetrics(v bme280.Result) {
	tempGauge.Set(float64(v.Temperature))
	pressureGauge.Set(float64(v.Pressure))
	humidityGauge.Set(float64(v.Humidity))
}
