package main


import (
    "fmt"
    "time"

    "net/http"

    "github.com/tquellenberg/weatherstation/bme280"
    "github.com/tquellenberg/weatherstation/chart"
    "github.com/tquellenberg/weatherstation/datastore"
)


// The I2C address which this device listens to.
const Address = 0x76


func initHttp() {
    http.HandleFunc("/", chart.Httpserver)
    http.ListenAndServe(":8081", nil)
}

func main() {
    initHttp()

	d := bme280.InitBme280(Address)
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