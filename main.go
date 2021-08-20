package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tquellenberg/weatherstation/bme280"
	"github.com/tquellenberg/weatherstation/chart"
	"github.com/tquellenberg/weatherstation/datastore"
)

// The I2C address which this device listens to.
const I2cAddress = 0x76

// Http port for web server
const Port = 8082

func initHttp() {
	log.Print("Http: Init")
	http.HandleFunc("/", chart.Httpserver)
	go func() {
		addr := fmt.Sprintf(":%d", Port)
		log.Printf("Http: Start listening on %s", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}()
}

func main() {
	noDataReading := flag.Bool("noDataReading", false, "do not read new values")
	dataDir := flag.String("dataDir", "./data", "directory for storing data files")
	flag.Parse()

	datastore.SetDataDir(*dataDir)

	initHttp()

	if *noDataReading {
		log.Print("No new data will be read.")
		// wait forever
		select {}
	} else {
		d, err := bme280.InitBme280(I2cAddress)
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
}
