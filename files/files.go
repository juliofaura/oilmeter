package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/juliofaura/oilmeter/data"
)

var (
	WorkingDir, DataFile, GraphFile, AverageFile string
)

func SaveToFile(msg string) {
	d1 := []byte(msg)
	filename := fmt.Sprintf("%04d-%02d-%02d{%d}-%02d-%02d-%02d.txt", data.Now.Year(), data.Now.Month(), data.Now.Day(), data.Now.Weekday(), data.Now.Hour(), data.Now.Minute(), data.Now.Second())
	err := ioutil.WriteFile(WorkingDir+filename, d1, 0644)
	data.Check(err)
}

func AppendToDataFile(msg string) {
	f, err := os.OpenFile(DataFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(msg + "\n"); err != nil {
		fmt.Println(err)
	}
}

func ReadDataPoint(record []string) (dataPoint data.Datapoint, err error) {
	dataPoint.Timestamp, err = strconv.ParseInt(record[0], 10, 64)
	if err != nil {
		return
	}
	dataPoint.Year, err = strconv.ParseInt(record[1], 10, 64)
	if err != nil {
		return
	}
	dataPoint.Month, err = strconv.ParseInt(record[2], 10, 64)
	if err != nil {
		return
	}
	dataPoint.Day, err = strconv.ParseInt(record[3], 10, 64)
	if err != nil {
		return
	}
	dataPoint.Weekday, err = strconv.ParseInt(record[4], 10, 64)
	if err != nil {
		return
	}
	dataPoint.Hour, err = strconv.ParseInt(record[5], 10, 64)
	if err != nil {
		return
	}
	dataPoint.Minute, err = strconv.ParseInt(record[6], 10, 64)
	if err != nil {
		return
	}
	dataPoint.Second, err = strconv.ParseInt(record[7], 10, 64)
	if err != nil {
		return
	}
	dataPoint.Duration, err = strconv.ParseFloat(record[8], 64)
	if err != nil {
		return
	}
	dataPoint.Distance, err = strconv.ParseFloat(record[9], 64)
	if err != nil {
		return
	}
	dataPoint.Stick, err = strconv.ParseFloat(record[10], 64)
	if err != nil {
		return
	}
	dataPoint.Liters, err = strconv.ParseFloat(record[11], 64)
	return
}
