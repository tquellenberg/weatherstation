package datastore

import (
	"container/list"
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

// Three entries with the last values for temp(0), pressure(1) and humidity(2)
var lastValues []Entry

// Last 30 pressure entries; used to determine the air pressure trend
var pressureQueue = list.New()

const pressureQueueMaxLength = 30

// Position in CSV file
type CsvPos int

const (
	DatePos CsvPos = iota
	TemperaturePos
	PressurePos
	HumidityPos
)

func (pos CsvPos) String() string {
	return []string{"date", "temperature", "pressure", "humidity"}[pos]
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

func updateLastValue(res bme280.Result, t string) {
	l := make([]Entry, 0, 3)
	l = append(l, Entry{Time: t, Value: res.Temperature})
	l = append(l, Entry{Time: t, Value: res.Pressure})
	l = append(l, Entry{Time: t, Value: res.Humidity})
	lastValues = l
}

func updatePressureQueue(res bme280.Result, t string) {
	pressureQueue.PushBack(Entry{Time: t, Value: res.Pressure})
	for pressureQueue.Len() > pressureQueueMaxLength {
		pressureQueue.Remove(pressureQueue.Front())
	}
}

func GetLastValues() []Entry {
	if len(lastValues) < 3 {
		l := make([]Entry, 0, 3)
		l = append(l, Entry{Time: "", Value: 0.0})
		l = append(l, Entry{Time: "", Value: 0.0})
		l = append(l, Entry{Time: "", Value: 0.0})
		return l
	}
	return lastValues
}

// Return "up", "down" or ""
func GetPressureTrend() string {
	if pressureQueue.Len() > 0 {
		currentPressure := GetLastValues()[1].Value
		pastPressure := pressureQueue.Front().Value.(Entry).Value
		if currentPressure > pastPressure {
			return "up"
		} else if currentPressure < pastPressure {
			return "down"
		}
	}
	return ""
}

func AppendToStore(res bme280.Result) {
	t := time.Now().Format(DateTimeFormat)
	column := []string{t,
		fmt.Sprintf("%3.2f", res.Temperature),
		fmt.Sprintf("%4.2f", res.Pressure),
		fmt.Sprintf("%3.2f", res.Humidity)}

	updateLastValue(res, t)
	updatePressureQueue(res, t)
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
	return getDataSeries(start, end, TemperaturePos)
}

func GetPressureSeries(start, end time.Time) ([]Entry, error) {
	return getDataSeries(start, end, PressurePos)
}

func GetHumiditySeries(start, end time.Time) ([]Entry, error) {
	return getDataSeries(start, end, HumidityPos)
}

func getDataSeries(start, end time.Time, csvPos CsvPos) ([]Entry, error) {
	log.Printf("Get %s series", csvPos)
	result, err := getDataFromFile(start, end, csvPos)
	if err != nil {
		return nil, err
	}
	log.Printf("Get %s series %d", csvPos, len(result))
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
func getDataFromFile(start, end time.Time, pos CsvPos) ([]Entry, error) {
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
		t, _ := time.ParseInLocation(DateTimeFormat, line[DatePos], now.Location())
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
