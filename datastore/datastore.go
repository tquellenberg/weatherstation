package datastore

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"encoding/csv"

	"github.com/tquellenberg/weatherstation/bme280"
)

const DateTimeFormat = "2006-01-02 15:04:05"

var dataDir string = "."
var filename = "results.csv"

type Range int

const (
	TODAY Range = iota
	ALL
)

type Entry struct {
	Time  string
	Value float32
}

func SetDataDir(newDataDir string) {
	dataDir = newDataDir
	dataDir = strings.TrimSuffix(dataDir, "/")
	if dataDir != "" && dataDir != "." {
		err := os.MkdirAll(dataDir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	log.Printf("Data directory set to %s", dataDir)
}

func getFilename() string {
	return dataDir + "/" + filename
}
func AppendToStore(res bme280.Result) {
	column := []string{time.Now().Format(DateTimeFormat),
		fmt.Sprintf("%3.2f", float32(res.Temperature)/100.0),
		fmt.Sprintf("%4.2f", float32(res.Pressure)/100.0),
		fmt.Sprintf("%3.2f", float32(res.Humidity)/1024.0)}

	f, err := os.OpenFile(getFilename(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println("Error: ", err)
		return
	}
	w := csv.NewWriter(f)
	w.Write(column)
	w.Flush()
	f.Close()
}

func GetTemperatureSeries() ([]Entry, error) {
	log.Println("Get temperature series")
	result, err := getDataFromFile(TODAY, 1)
	if err != nil {
		return nil, err
	}
	log.Printf("Get temperature series %d", len(result))
	return result, nil
}

func GetPressureSeries() ([]Entry, error) {
	log.Println("Get pressure series")
	result, err := getDataFromFile(TODAY, 2)
	if err != nil {
		return nil, err
	}
	log.Printf("Get pressure series %d", len(result))
	return result, nil
}

func GetHumiditySeries() ([]Entry, error) {
	log.Println("Get humidity series")
	result, err := getDataFromFile(TODAY, 3)
	if err != nil {
		return nil, err
	}
	log.Printf("Get humidity series %d", len(result))
	return result, nil
}

// Read time (first value) and value on position 'pos' from csv file
func getDataFromFile(r Range, pos int) ([]Entry, error) {
	f, err := os.OpenFile(getFilename(), os.O_RDONLY, 0644)
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	defer f.Close()

	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}

	result := make([]Entry, 0, len(lines))
	now := time.Now()
	var v float64
	var start time.Time
	if r == TODAY {
		year, month, day := now.Date()
		start = time.Date(year, month, day, 0, 0, 0, 0, now.Location())
		log.Printf("Start %v", start)
	}
	for _, line := range lines {
		t, _ := time.ParseInLocation(DateTimeFormat, line[0], now.Location())
		if r == ALL || t.Local().After(start) {
			e := Entry{}
			e.Time = line[0]
			if v, err = strconv.ParseFloat(line[pos], 32); err != nil {
				return nil, err
			}
			e.Value = float32(v)
			result = append(result, e)
		}
	}
	return result, nil
}
