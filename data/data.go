package data

import "time"

const (
	// sensor = "HC-SR04"
	Sensor                = "VL53L1X"
	Maxsamples            = 80
	Delay                 = time.Millisecond * 500
	Timeout               = time.Second
	AmountGood            = 1000
	AmountDangerous       = 600
	TimeForAverage        = (6 * 24 * 60 * 60)
	NewGasThreshold       = 200
	GasFilteringThreshold = 100
)

type Datapoint struct {
	Timestamp                                       int64
	Year, Month, Day, Weekday, Hour, Minute, Second int64
	Duration, Distance, Stick, Liters               float64
}

var (
	Ceiling = 136.5 // in cm
	Now     = time.Now()
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}
