package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"time"

	i2c "github.com/d2r2/go-i2c"
	logger "github.com/d2r2/go-logger"
	shell "github.com/d2r2/go-shell"
	vl53l0x "github.com/d2r2/go-vl53l0x"
)

const (
	maxsamples = 3
	delay      = time.Millisecond * 500
	timeout    = time.Second
	ceiling    = 133
	dataDir    = "/home/pi/Gasoleo/data/"
)

var (
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
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func saveToFile(msg string) {
	d1 := []byte(msg)
	now := time.Now()
	filename := fmt.Sprintf("%d-%d-%d{%d}-%d-%d-%d.dat", now.Year(), now.Month(), now.Day(), now.Weekday(), now.Hour(), now.Minute(), now.Second())
	err := ioutil.WriteFile(dataDir+filename, d1, 0644)
	check(err)
}

var lg = logger.NewPackageLogger("main",
	logger.DebugLevel,
	// logger.InfoLevel,
)

func main() {
	defer logger.FinalizeLogger()
	// Create new connection to i2c-bus on 1 line with address 0x40.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err := i2c.NewI2C(0x29, 1)
	if err != nil {
		lg.Fatal(err)
	}
	defer i2c.Close()

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** !!! READ THIS !!!")
	lg.Notify("*** You can change verbosity of output, by modifying logging level of modules \"i2c\", \"vl53l0x\".")
	lg.Notify("*** Uncomment/comment corresponding lines with call to ChangePackageLogLevel(...)")
	lg.Notify("*** !!! READ THIS !!!")
	lg.Notify("**********************************************************************************************")
	// Uncomment/comment next line to suppress/increase verbosity of output
	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
	logger.ChangePackageLogLevel("vl53l0x", logger.InfoLevel)

	sensor := vl53l0x.NewVl53l0x()
	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Reset sensor")
	lg.Notify("**********************************************************************************************")
	err = sensor.Reset(i2c)
	if err != nil {
		lg.Fatalf("Error reseting sensor: %s", err)
	}
	// It's highly recommended to reset sensor before repeated initialization.
	// By default, sensor initialized with "RegularRange" and "RegularAccuracy" parameters.

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Initialize sensor")
	lg.Notify("**********************************************************************************************")
	err = sensor.Init(i2c)
	if err != nil {
		lg.Fatalf("Failed to initialize sensor: %s", err)
	}
	rev, err := sensor.GetProductMinorRevision(i2c)
	if err != nil {
		lg.Fatalf("Error getting sensor minor revision: %s", err)
	}
	lg.Infof("Sensor minor revision = %d", rev)

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Сonfigure sensor")
	lg.Notify("**********************************************************************************************")
	rngConfig := vl53l0x.RegularRange
	speedConfig := vl53l0x.GoodAccuracy
	lg.Infof("Configure sensor with  %q and %q",
		rngConfig, speedConfig)
	err = sensor.Config(i2c, rngConfig, speedConfig)
	if err != nil {
		lg.Fatalf("Failed to initialize sensor: %s", err)
	}

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Single shot range measurement mode")
	lg.Notify("**********************************************************************************************")
	rng, err := sensor.ReadRangeSingleMillimeters(i2c)
	if err != nil {
		lg.Fatalf("Failed to measure range: %s", err)
	}
	lg.Infof("Measured range = %v mm", rng)

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Continuous shot range measurement mode")
	lg.Notify("**********************************************************************************************")
	var freq uint32 = 100
	times := 2000
	lg.Infof("Made measurement each %d milliseconds, %d times", freq, times)
	err = sensor.StartContinuous(i2c, freq)
	if err != nil {
		lg.Fatalf("Can't start continuous measures: %s", err)
	}
	// create context with cancellation possibility
	ctx, cancel := context.WithCancel(context.Background())
	// use done channel as a trigger to exit from signal waiting goroutine
	done := make(chan struct{})
	defer close(done)
	// build actual signals list to control
	signals := []os.Signal{os.Kill, os.Interrupt}
	if shell.IsLinuxMacOSFreeBSD() {
		signals = append(signals, syscall.SIGTERM)
	}
	// run goroutine waiting for OS termination events, including keyboard Ctrl+C
	shell.CloseContextOnSignals(cancel, done, signals...)

	for i := 0; i < times; i++ {
		rng, err := sensor.ReadRangeContinuousMillimeters(i2c)
		if err != nil {
			lg.Fatalf("Failed to measure range: %s", err)
		}
		lg.Infof("Measured range = %v mm", rng)
		select {
		// Check for termination request.
		case <-ctx.Done():
			err = sensor.StopContinuous(i2c)
			if err != nil {
				lg.Fatal(err)
			}
			lg.Fatal(ctx.Err())
		default:
		}
	}
	err = sensor.StopContinuous(i2c)
	if err != nil {
		lg.Fatalf("Error stopping continuous measures: %s", err)
	}

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Reconfigure sensor")
	lg.Notify("**********************************************************************************************")
	rngConfig = vl53l0x.RegularRange
	speedConfig = vl53l0x.RegularAccuracy
	lg.Infof("Reconfigure sensor with %q and %q",
		rngConfig, speedConfig)
	err = sensor.Config(i2c, rngConfig, speedConfig)
	if err != nil {
		lg.Fatalf("Failed to initialize sensor: %s", err)
	}

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Single shot range measurement mode")
	lg.Notify("**********************************************************************************************")
	rng, err = sensor.ReadRangeSingleMillimeters(i2c)
	if err != nil {
		lg.Fatalf("Failed to measure range: %s", err)
	}
	lg.Infof("Measured range = %v mm", rng)

}
