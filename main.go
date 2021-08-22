package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/tquellenberg/weatherstation/bme280"
	"github.com/tquellenberg/weatherstation/chart"
	"github.com/tquellenberg/weatherstation/datastore"
	"github.com/tquellenberg/weatherstation/opensensemap"
	"github.com/tquellenberg/weatherstation/sun"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Position struct {
		Latitude  float64
		Longitude float64
	}
	OpensenseMap struct {
		BoxId      string `yaml:"boxId"`
		TempSensor string `yaml:"tempSensor"`
		PresSensor string `yaml:"presSensor"`
		HumiSensor string `yaml:"humiSensor"`
	} `yaml:"opensenseMap"`
	Bme280 struct {
		I2cAddress int
	}
	Http struct {
		Port int
	}
}

// The I2C address which this device listens to.
const DEFAULT_I2C_ADDRESS = 0x76

// Http port for web server
const DEFAULT_HTTP_PORT = 8082

func initHttp(port int) {
	log.Print("Http: Init")
	http.HandleFunc("/", chart.Httpserver)
	go func() {
		addr := fmt.Sprintf(":%d", port)
		log.Printf("Http: Start listening on %s", addr)
		log.Println(http.ListenAndServe(addr, nil))
	}()
}

func setDefault(config *Config) {
	if config.Http.Port == 0 {
		config.Http.Port = DEFAULT_HTTP_PORT
	}
	if config.Bme280.I2cAddress == 0 {
		config.Bme280.I2cAddress = DEFAULT_I2C_ADDRESS
	}
}

func readConfig() Config {
	config := Config{}

	b, err := ioutil.ReadFile("weatherstation.yml")
	if err != nil {
		log.Println(err)
		return config
	}

	err = yaml.Unmarshal(b, &config)
	if err != nil {
		log.Printf("error: %v", err)
	}
	return config
}

func sendOpensensemapData(opensensemapToken *string, v bme280.Result, config *Config) {
	opensensemap.PostFloatValue(*opensensemapToken, float32(v.Temperature)/100.0, 2,
		config.OpensenseMap.BoxId, config.OpensenseMap.TempSensor)
	opensensemap.PostFloatValue(*opensensemapToken, float32(v.Pressure)/100.0, 1,
		config.OpensenseMap.BoxId, config.OpensenseMap.PresSensor)
	opensensemap.PostFloatValue(*opensensemapToken, float32(v.Humidity)/1024.0, 1,
		config.OpensenseMap.BoxId, config.OpensenseMap.HumiSensor)
}

func main() {
	noDataReading := flag.Bool("noDataReading", false, "do not read new values")
	dataDir := flag.String("dataDir", "./data", "directory for storing data files")
	opensensemapToken := flag.String("opensensemapToken", "", "API token for opensensemap")
	flag.Parse()

	config := readConfig()
	setDefault(&config)

	datastore.SetDataDir(*dataDir)

	initHttp(config.Http.Port)

	sun.InitLocation(config.Position.Latitude, config.Position.Longitude)

	if *noDataReading {
		log.Print("No new data will be read.")
		// wait forever
		select {}
	} else {
		d, err := bme280.InitBme280(config.Bme280.I2cAddress)
		if err != nil {
			log.Println(err)
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

			if *opensensemapToken != "" {
				sendOpensensemapData(opensensemapToken, v, &config)
			}

			time.Sleep(time.Minute)
		}
	}
}
