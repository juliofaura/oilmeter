package main

//go:generate go run main.go

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/wcharczuk/go-chart/drawing"

	chart "github.com/wcharczuk/go-chart"
)

const (
	dataFile = "./data.txt"
	// dataFile        = "/Users/julio/Dropbox/Gasoleo/data.txt"
	graphFile       = "/Users/julio/Dropbox/Gasoleo/graph.png"
	amountGood      = 1000
	amountDangerous = 600
	timeForAverage  = (3 * 24 * time.Hour) / time.Second
	newGasThreshold = 200
)

type datapoint struct {
	Timestamp                                       int64
	Year, Month, Day, Weekday, Hour, Minute, Second int64
	Duration, Distance, Stick, Liters               float64
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	/*
	   In this example we add a second series, and assign it to the secondary y axis, giving that series it's own range.

	   We also enable all of the axes by setting the `Show` propery of their respective styles to `true`.
	*/

	csvFile, err := os.Open(dataFile)
	if err != nil {
		log.Println(err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(bufio.NewReader(csvFile))
	var data []datapoint
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		var dataPoint datapoint
		dataPoint.Timestamp, err = strconv.ParseInt(line[0], 10, 64)
		check(err)
		dataPoint.Year, err = strconv.ParseInt(line[1], 10, 64)
		check(err)
		dataPoint.Month, err = strconv.ParseInt(line[2], 10, 64)
		check(err)
		dataPoint.Day, err = strconv.ParseInt(line[3], 10, 64)
		check(err)
		dataPoint.Weekday, err = strconv.ParseInt(line[4], 10, 64)
		check(err)
		dataPoint.Hour, err = strconv.ParseInt(line[5], 10, 64)
		check(err)
		dataPoint.Minute, err = strconv.ParseInt(line[6], 10, 64)
		check(err)
		dataPoint.Second, err = strconv.ParseInt(line[7], 10, 64)
		check(err)
		dataPoint.Duration, err = strconv.ParseFloat(line[8], 64)
		check(err)
		dataPoint.Distance, err = strconv.ParseFloat(line[9], 64)
		check(err)
		dataPoint.Stick, err = strconv.ParseFloat(line[10], 64)
		check(err)
		dataPoint.Liters, err = strconv.ParseFloat(line[11], 64)
		check(err)
		data = append(data, dataPoint)
	}
	var XValues []float64
	var YValues []float64
	for _, v := range data {
		XValues = append(XValues, float64(v.Timestamp))
		YValues = append(YValues, v.Liters)
	}

	LastX := XValues[len(XValues)-1]
	LastY := YValues[len(YValues)-1]

	var labelColor drawing.Color

	if LastY > amountGood {
		labelColor = chart.ColorGreen
	} else if LastY > amountDangerous {
		labelColor = chart.ColorYellow
	} else {
		labelColor = chart.ColorRed
	}

	var average = 0.0
	firstPointForAverage, endingPointForAverage := data[len(data)-1], data[len(data)-1]
	var bigChanges = 0.0
	for i := len(data) - 2; i >= 0; i-- {
		if math.Abs(data[i].Liters-data[i+1].Liters) > newGasThreshold {
			bigChanges += data[i+1].Liters - data[i].Liters
		}
		endingPointForAverage = data[i]
		if endingPointForAverage.Timestamp-firstPointForAverage.Timestamp >= int64(timeForAverage) {
			break
		}
	}

	if firstPointForAverage.Timestamp != endingPointForAverage.Timestamp {
		average = -float64(endingPointForAverage.Liters-firstPointForAverage.Liters+bigChanges) / (float64(endingPointForAverage.Timestamp-firstPointForAverage.Timestamp) / (24 * 60 * 60))
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			TickPosition: chart.TickPositionBetweenTicks,
			ValueFormatter: func(v interface{}) string {
				typed := v.(float64) * 1e9
				typedDate := chart.TimeFromFloat64(typed)
				return fmt.Sprintf("%d/%d/%d", typedDate.Month(), typedDate.Day(), typedDate.Year())
			},
			Style: chart.Style{
				TextRotationDegrees: 45,
			},
		},
		YAxis: chart.YAxis{
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%.1f", v.(float64))
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
					FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
					StrokeWidth: 5,
					DotWidth:    4,
				},
				XValues: XValues,
				YValues: YValues,
			},
			chart.AnnotationSeries{
				Annotations: []chart.Value2{
					{
						XValue: LastX,
						YValue: LastY,
						Label:  fmt.Sprintf("%.1f", LastY),
						Style: chart.Style{
							StrokeWidth: 10,
							FontSize:    chart.StyleTextDefaults().FontSize,
							StrokeColor: labelColor,
						},
					},
				},
			},
		},
		Title: "Oil liters vs time (avg is " + fmt.Sprintf("%.2f", average) + " liters / day)",
	}

	f, _ := os.Create(graphFile)
	defer f.Close()
	graph.Render(chart.PNG, f)
}
