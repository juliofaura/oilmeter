package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"time"

	rpio "github.com/stianeikeland/go-rpio"
	chart "github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

type datapoint struct {
	Timestamp                                       int64
	Year, Month, Day, Weekday, Hour, Minute, Second int64
	Duration, Distance, Stick, Liters               float64
}

const (
	// sensor = "HC-SR04"
	sensor          = "VL53L1X"
	maxsamples      = 80
	delay           = time.Millisecond * 500
	timeout         = time.Second
	ceiling         = 134.7
	amountGood      = 1000
	amountDangerous = 600
	timeForAverage  = (6 * 24 * 60 * 60)
	newGasThreshold = 200
)

var (
	trigPin = rpio.Pin(27)
	echoPin = rpio.Pin(17)

	workingDir, dataFile, graphFile string

	litersTable = []float64{
		0,
		3.5527328462448,
		10.0247326637267,
		18.3725529464608,
		28.2183707163832,
		39.3410291761309,
		51.5894296250806,
		64.8512518469467,
		79.0383716964191,
		94.0789905505373,
		109.91295336688,
		126.488761712572,
		143.761564449007,
		161.69174861514,
		180.243917421517,
		199.386128134566,
		219.089310343002,
		239.32681298639,
		260.07404553232,
		281.308189439568,
		303.007963055353,
		325.153427791659,
		347.725826648087,
		370.707448406554,
		394.081512435341,
		417.83207021062,
		441.943920526747,
		466.40253601196,
		491.193999054705,
		516.304945620071,
		541.722515725553,
		567.434309571892,
		593.428348503432,
		619.693040114645,
		646.217146933477,
		672.989758204295,
		700.00026436809,
		727.238333898943,
		754.693892206211,
		782.357102353728,
		810.218347382084,
		838.268214049181,
		866.497477828712,
		894.897089026865,
		923.458159895046,
		952.171952631289,
		981.029868175764,
		1010.0234357166,
		1039.14430283168,
		1068.38422619995,
		1097.73506282297,
		1127.18876170313,
		1156.73735593039,
		1186.37295513379,
		1216.08773825783,
		1245.87394662738,
		1275.72387726737,
		1305.6298764465,
		1335.58433341577,
		1365.57967431509,
		1395.60835622234,
		1425.66286132079,
		1455.73569116171,
		1485.81936099999,
		1515.90639418122,
		1545.98931655889,
		1576.06065092107,
		1606.11291140557,
		1636.13859788279,
		1666.13019028501,
		1696.08014286057,
		1725.98087833044,
		1755.82478192416,
		1785.60419527079,
		1815.31141011929,
		1844.93866186104,
		1874.47812282557,
		1903.92189531813,
		1933.2620043653,
		1962.49039013188,
		1991.59889996877,
		2020.57928004784,
		2049.42316653488,
		2078.12207624681,
		2106.66739673293,
		2135.05037571322,
		2163.26210979838,
		2191.29353240684,
		2219.13540078288,
		2246.77828200736,
		2274.21253787693,
		2301.42830851038,
		2328.41549451939,
		2355.16373755613,
		2381.66239902075,
		2407.900536676,
		2433.86687887403,
		2459.54979604855,
		2484.93726906335,
		2510.01685393155,
		2534.77564232584,
		2559.20021718358,
		2583.27660256515,
		2606.99020674099,
		2630.32575725073,
		2653.26722638063,
		2675.79774512153,
		2697.89950316718,
		2719.5536318491,
		2740.74006601442,
		2761.4373796459,
		2781.62258835655,
		2801.2709095543,
		2820.35546772896,
		2838.84692743029,
		2856.71302919873,
		2873.91799247094,
		2890.42173164463,
		2906.17880211966,
		2921.13694265572,
		2935.23498903259,
		2948.39975791085,
		2960.54113248609,
		2971.54373270813,
		2981.25129727342,
		2989.43254072289,
		2995.68301629862,
		3000,
	}
	now = time.Now()
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func saveToFile(msg string) {
	d1 := []byte(msg)
	filename := fmt.Sprintf("%04d-%02d-%02d{%d}-%02d-%02d-%02d.txt", now.Year(), now.Month(), now.Day(), now.Weekday(), now.Hour(), now.Minute(), now.Second())
	err := ioutil.WriteFile(workingDir+filename, d1, 0644)
	check(err)
}

func appendToDataFile(msg string) {
	f, err := os.OpenFile(dataFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(msg + "\n"); err != nil {
		fmt.Println(err)
	}
}

// Returns the distance in mm
func takeMeasurement() (measurement float64, err error) {
	switch sensor {
	case "HC-SR04":
		// Send the pulse
		trigPin.High()
		time.Sleep(time.Microsecond * 10)
		// echoPin.Detect(rpio.FallEdge)
		trigPin.Low()
		startingTime := time.Now()

		// Detect the echo

		// for time.Since(startingTime) < timeout && !echoPin.EdgeDetected() {
		// }

		// First wait echo pin to settle
		for time.Since(startingTime) < timeout && echoPin.Read() != rpio.High {
		}
		// Then wait for the echo
		for time.Since(startingTime) < timeout && echoPin.Read() != rpio.Low {
		}

		// Measure the distance
		duration := time.Since(startingTime)
		if duration >= timeout {
			err = errors.New("Timeout in measurement")
		} else {
			durationUs := float64(duration) / float64(time.Microsecond)
			measurement = (durationUs * .343) / 2
		}
	case "VL53L1X":
		cmd := exec.Command("python", workingDir+"distance.py")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err == nil {
			errStr := string(stderr.Bytes())
			measurement, err = strconv.ParseFloat(errStr, 64)
		}
	default:
		log.Fatal("Unknow sensor selected: ", sensor)
	}
	time.Sleep(delay)
	return
}

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Usage: " + os.Args[0] + " <working dir>, where <working dir> is the directory where the .txt and .png files will be placed")
		os.Exit(1)
	}

	workingDir = os.Args[1] + "/"
	dataFile = workingDir + "data.txt"
	graphFile = workingDir + "graph.png"

	fmt.Println("Starting")
	time.Sleep(2 * time.Second)

	// Now configuring GPIO

	// Open and map memory to access gpio, check for errors
	fmt.Println("Opening rpio ...")
	check(rpio.Open())

	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pin to output mode
	fmt.Println("Configuring pins ...")
	trigPin.Output()
	echoPin.Input()
	trigPin.Low()

	// Some time to settle
	fmt.Println("Waiting to settle ...")
	time.Sleep(2 * time.Second)

	fmt.Println("Taking a measurement, date/time is", time.Now())

	// Measuring samples
	fmt.Printf("Starting samples (ceiling is %.1f)...\n", ceiling)

	atLeastOneValidMeasurement := false

	distances := []float64{}
	for i := 0; i < maxsamples; i++ {
		thisDistance, err := takeMeasurement()
		if err != nil {
			fmt.Println(err)
		} else if thisDistance < 0 {
			fmt.Println("Negative distance measured (", thisDistance, "), discarding")
		} else {
			// fmt.Println("Distance is", thisDistance)
			distances = append(distances, thisDistance/10.0)
			atLeastOneValidMeasurement = true
		}
	}

	if !atLeastOneValidMeasurement {
		log.Fatal("No valid measurements")
	}

	// Calculate the distance and the liters
	fmt.Printf("Calculating everything (good samples are %d)...\n", len(distances))
	sort.Float64s(distances)

	var distance float64
	// // Using the average of the middle tranch
	// for i := len(distances) / 3; i < len(distances)-len(distances)/3; i++ {
	// 	distance += distances[i]
	// }
	// distance /= float64(len(distances) - 2*(len(distances)/3))

	// Using the median
	distance = distances[len(distances)/2]

	stick := ceiling - distance
	var liters float64

	if stick < 0 {
		liters = 0
	} else if stick >= float64(len(litersTable)) {
		liters = 3000
	} else {
		tranch := int(stick)
		liters = litersTable[tranch] + (litersTable[tranch+1]-litersTable[tranch])*(stick-float64(tranch))
	}

	message := fmt.Sprintf("Duration (us) = %.1f\nDistance (cm) = %.1f\nStick (cm) = %.1f\nLiters (l) = %.1f\n",
		0.0, distance, stick, liters,
	)
	fmt.Println(message)
	saveToFile(message)

	dataline := fmt.Sprintf("%0d,%d,%d,%d,%d,%d,%d,%d,%f,%f,%f,%f",
		now.Unix(),
		now.Year(),
		now.Month(),
		now.Day(),
		now.Weekday(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		0.0,
		distance,
		stick,
		liters,
	)
	appendToDataFile(dataline)

	/*
	 * Now rendering report
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
		firstPointForAverage = data[i]
		if endingPointForAverage.Timestamp-firstPointForAverage.Timestamp >= int64(timeForAverage) {
			break
		}
	}

	if firstPointForAverage.Timestamp != endingPointForAverage.Timestamp {
		average = -float64(endingPointForAverage.Liters-firstPointForAverage.Liters-bigChanges) / (float64(endingPointForAverage.Timestamp-firstPointForAverage.Timestamp) / (24 * 60 * 60))
	}

	fmt.Printf("Average consumption is %.2f liters/day\n", average)
	fmt.Println("Starting / ending liters: ", firstPointForAverage.Liters, " / ", endingPointForAverage.Liters, " (big changes are", bigChanges, ")")
	fmt.Println("Starting / ending timestamps: ", firstPointForAverage.Timestamp, " / ", endingPointForAverage.Timestamp)

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
