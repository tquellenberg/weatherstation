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
		fmt.Sprintf("%3.2f", res.Temperature),
		fmt.Sprintf("%4.2f", res.Pressure),
		fmt.Sprintf("%3.2f", res.Humidity)}

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

func GetTemperatureSeries(start, end time.Time) ([]Entry, error) {
	log.Println("Get temperature series")
	result, err := getDataFromFile(start, end, 1)
	if err != nil {
		return nil, err
	}
	log.Printf("Get temperature series %d", len(result))
	return movingAvg(result, 5), nil
}

func GetPressureSeries(start, end time.Time) ([]Entry, error) {
	log.Println("Get pressure series")
	result, err := getDataFromFile(start, end, 2)
	if err != nil {
		return nil, err
	}
	log.Printf("Get pressure series %d", len(result))
	return movingAvg(result, 5), nil
}

func GetHumiditySeries(start, end time.Time) ([]Entry, error) {
	log.Println("Get humidity series")
	result, err := getDataFromFile(start, end, 3)
	if err != nil {
		return nil, err
	}
	log.Printf("Get humidity series %d", len(result))
	return movingAvg(result, 5), nil
}

func movingAvg(entries []Entry, minutes int) []Entry {
	if len(entries) == 0 || minutes <= 1 {
		return entries
	}
	now := time.Now()
	start, _ := time.ParseInLocation(DateTimeFormat, entries[0].Time, now.Location())
	startMinute := int(start.Minute()/minutes) * minutes
	start = time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), startMinute, 0, 0, now.Location())
	end := start.Add(time.Minute * time.Duration(minutes))
	sumV := 0.0
	counter := 0
	result := make([]Entry, 0)
	for _, e := range entries {
		t, _ := time.ParseInLocation(DateTimeFormat, e.Time, now.Location())
		if t.After(end) {
			if counter > 0 {
				// new avg value
				v := avg(sumV, counter)
				result = append(result, Entry{Time: end.Format(DateTimeFormat), Value: v})
			}
			// step forward
			start = end
			end = start.Add(time.Minute * time.Duration(minutes))
			sumV = 0.0
			counter = 0
		} else {
			sumV = sumV + float64(e.Value)
			counter = counter + 1
		}
	}
	if counter > 0 {
		// last avg value
		v := avg(sumV, counter)
		result = append(result, Entry{Time: entries[len(entries)-1].Time, Value: v})
	}
	return result
}

func avg(sumV float64, counter int) float32 {
	v := sumV / float64(counter)
	return float32(int(v*100.0)) / 100.0
}

// Read time (first value) and value on position 'pos' from csv file
func getDataFromFile(start, end time.Time, pos int) ([]Entry, error) {
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
	for _, line := range lines {
		t, _ := time.ParseInLocation(DateTimeFormat, line[0], now.Location())
		if t.Local().After(start) && t.Local().Before(end) {
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
