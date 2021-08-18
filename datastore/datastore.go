package datastore

import (
	"os"
	"time"
	"fmt"
	"strconv"

	"encoding/csv"

	"github.com/tquellenberg/weatherstation/bme280"
)

const Filename = "results.csv"

const DateTimeFormat = "2006-01-02 15:04:05"

type Entry struct {
	Time  string
	Value float32
}

func AppendToStore(res bme280.Result) {
    column := []string{time.Now().Format(DateTimeFormat), 
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

func GetTemperatureSeries() ([]Entry, error) {
	f, err := os.OpenFile(Filename, os.O_RDONLY, 0644)
    if err != nil {
        fmt.Println("Error: ", err)
        return nil, err
    }
    defer f.Close()

    lines, err := csv.NewReader(f).ReadAll()
    if err != nil {
        return nil, err
    }

    result := make([]Entry, 0, len(lines))
    var v float64
    for _, line := range lines {
    	e := Entry{}
    	e.Time = line[0]
        if v, err = strconv.ParseFloat(line[1], 32); err != nil {
            return nil, err
        }
        e.Value = float32(v)
        result = append(result, e)
    }
    return result, nil
}

