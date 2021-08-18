package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tquellenberg/weatherstation/bme280"
	"github.com/tquellenberg/weatherstation/datastore"
)

// The I2C address which this device listens to.
const Address = 0x76

func initHttp() {
	log.Print("init Http")
	//	http.HandleFunc("/", chart.Httpserver)
	//	http.ListenAndServe(":8082", nil)
	log.Print("Http okay")
}

func main() {
	initHttp()

	d, err := bme280.InitBme280(Address)
	if err != nil {
		log.Fatal(err)
		return
	}
	time.Sleep(time.Second)

	for {
		d.SetConfiguration()
		v := d.ReadValues()

		fmt.Printf("Temp: %3.2f Grad C\n", float32(v.Temperature)/100.0)
		fmt.Printf("Pres: %4.2f hPa\n", float32(v.Pressure)/100.0)
		fmt.Printf("Humi: %3.2f %%\n", float32(v.Humidity)/1024.0)

		datastore.AppendToStore(v)

		time.Sleep(time.Minute)
	}
}
