/*
A blinker example using go-rpio library.
Requires administrator rights to run
Toggles a LED on physical pin 19 (mcu pin 10)
Connect a LED with resistor from pin 19 to ground.
*/

package main

import (
	"fmt"
	"os"
	"time"

	rpio "github.com/stianeikeland/go-rpio"
)

const (
	samples = 5
	delay   = time.Millisecond * 500
	ceiling = 133
)

var (
	trigPin = rpio.Pin(4)
	echoPin = rpio.Pin(17)

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

func main() {
	// Open and map memory to access gpio, check for errors
	fmt.Println("Opening rpio ...")
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

	var duration time.Duration

	// Measuring samples
	fmt.Println("Starting samples ...")

	for i := 0; i < samples; i++ {
		fmt.Printf("Measuring sample #%v ... \n", i)
		// Detect any remaining pulses
		for echoPin.EdgeDetected() {
			fmt.Println("Whoa, spurious edge detected")
		}
		// Send the pulse
		trigPin.High()
		time.Sleep(time.Microsecond * 10)
		trigPin.Low()
		startingTime := time.Now()

		// Detect the echo
		// for time.Since(startingTime) < time.Second && echoPin.Read() == rpio.Low {
		// 	fmt.Println("Pin is ", echoPin.Read(), ", so far is ", time.Since(startingTime), ", time is ", time.Now())
		// }
		// fmt.Println("... and Pin is ", echoPin.Read(), ", so far is ", time.Since(startingTime), ", time is ", time.Now())
		for i := 0; i < 20; i++ {
			fmt.Println("Pin is ", echoPin.Read(), ", so far is ", time.Since(startingTime), ", time is ", time.Now())
		}
		fmt.Println("... and Pin is ", echoPin.Read(), ", so far is ", time.Since(startingTime), ", time is ", time.Now())

		// Measure the distance
		thisDuration := time.Since(startingTime)
		thisDurationUs := float64(thisDuration) / float64(time.Microsecond)
		fmt.Printf("Done! this duration is %.1f us\n", thisDurationUs)
		duration += thisDuration

		// Wait until echo fades
		time.Sleep(delay)

	}

	// Calculate the distance and the liters
	fmt.Println("Calculating everything ... ")
	duration /= samples // Calculate the average
	durationUs := float64(duration) / float64(time.Microsecond)
	fmt.Printf("Average duration is %.1f us\n", durationUs)
	distance := (durationUs * .0343) / 2
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

	fmt.Printf("Duration (us) is %.1f, distance (cm) is %.1f, stick (cm) is %.1f, liters (l) is %.1f\n",
		durationUs, distance, stick, liters,
	)

}
