package main


import (
    "fmt"
    "time"
    "os"

    "encoding/csv"

    "github.com/tquellenberg/weatherstation/bme280"
)

const Filename = "results.csv"

// The I2C address which this device listens to.
const Address = 0x76

func appendToFile(res bme280.Result) {
    column := []string{time.Now().Format("2006-01-02 15:04:05"), 
                       fmt.Sprintf("%3.2f", float32(res.Temperature)/100.0), 
                       fmt.Sprintf("%4.2f", float32(res.Pressure)/100.0), 
                       fmt.Sprintf("%3.2f", float32(res.Humidity)/1024.0)}

    f, err := os.OpenFile(Filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
        fmt.Println("Error: ", err)
        return
    }
    w := csv.NewWriter(f)
    w.Write(column)
    w.Flush()
    f.Close()
}

func main() {
	d := bme280.InitBme280(Address)
    time.Sleep(time.Second)

    for {
        d.SetConfiguration()
        v := d.ReadValues()

        fmt.Printf("Temp: %3.2f Grad C\n", float32(v.Temperature)/100.0)
        fmt.Printf("Pres: %4.2f hPa\n", float32(v.Pressure)/100.0)
        fmt.Printf("Humi: %3.2f %%\n", float32(v.Humidity)/1024.0)

        appendToFile(v)

        time.Sleep(time.Minute)
    }
}