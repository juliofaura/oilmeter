package main

//go:generate go run main.go

import (
	"fmt"
	"log"
	"os"

	chart "github.com/wcharczuk/go-chart"
)

const (
	// dataFile   = "/home/pi/Gasoleo/data.txt"
	dataFile = "/Users/julio/Dropbox/Gasoleo/data.txt"
)

func main() {

	/*
	   In this example we add a second series, and assign it to the secondary y axis, giving that series it's own range.

	   We also enable all of the axes by setting the `Show` propery of their respective styles to `true`.
	*/

	log.Println("Opening file")
	d, err := os.OpenFile(dataFile,
		os.O_RDONLY, 0)
	if err != nil {
		log.Println(err)
	}
	defer d.Close()

	var arr []byte
	arr = make([]byte, 20)
	str := string("")

	for r, err := d.Read(arr); r != 0; {
		fmt.Println("r is", r)
		if err != nil {
			log.Fatalln(err)
		}
		str += string(arr)
	}
	fmt.Println("Str is", str)

	graph := chart.Chart{
		XAxis: chart.XAxis{
			TickPosition: chart.TickPositionBetweenTicks,
			ValueFormatter: func(v interface{}) string {
				typed := v.(float64)
				typedDate := chart.TimeFromFloat64(typed)
				return fmt.Sprintf("%d-%d\n%d", typedDate.Month(), typedDate.Day(), typedDate.Year())
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
				YValues: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			},
		},
	}

	f, _ := os.Create("output.png")
	defer f.Close()
	graph.Render(chart.PNG, f)
}
