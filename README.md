# Documentación del medidor del tanque de gasóleo

## Descripción general

Se trata de poner un sensor de distancia dentro del tanque de gasóleo y conectarlo a una raspberry para medir periódicamente el gasóleo que queda, comunicándoselo a un servidor que lo publique en un interfaz web o móvil (y/o que envíe alarmas en situaciones relevantes)

Usamos un medidor de ultrasonidos HC-SR04 para medir el espacio entre el techo del tanque y el nivel de gasóleo en el tanque. Aproximamos la forma del tanque a la de un cilindro tumbado de radio 634.5mm, siendo la fórmula que relaciona los litros de gasóleo con el nivel desde el fondo del tanque la siguiente:

```litros de gasóleo = K*(ASIN((nivel-R)/R)+0.5*SIN(2*ASIN((nivel-R)/R))+PI()/2)```

, donde ```R = 634.5 mm``` y ```K = 954.5 litros``` (esta fórmula responde a la integral de una circunferencia, que es la base del tanque como se ha dicho)

(Cálculos detallados en https://www.dropbox.com/s/0gnpqd8h2pqaif8/C%C3%A1lculos%20tanque%20de%20gasoleo.ods?dl=0 )


---
## Medidor

- Actualizar a VL53L1X

Se pone el sensor HC-SR04 en una barquilla de PVC en forma de flecha. Se atan dos sedales:
- Uno en el centro de la barquilla, con una marca de cinta aislante en el extremo. Este sedal se usa para tensar la barquilla en el techo del tanque, de forma que quede perpendicular al gasóleo del interior del tanque
- Otro en la punta de la flecha, sin ninguna marca en el extremo. Este sedal se usa para tirar de la barquilla cuando se quiere sacar del tubo

El medidor se introduce por un tubo de PVC de 1 m de largo. Una vez tensada la barquilla, el sensor se queda aproximadamente a 133 cm del fondo del tanque, Una medida de 133 cm de distancia corresponde por tanto a cero litros, y una medida de 6 cm corresponde a la capacidad máxima del tanque, que es de 3000 litros. Por tanto la medida del nivel de gasóleo en función de la distancia medida por el HC-SR04 es de ```nivel = 133 - distancia```, con un valor de la distancia que debe estar entre 0 y 127

Para la implementación práctica hemos definido un array con las cantidades de gasóleo que corresponden a cada nivel medido, en intervalos de 1 cm. Luego se interpola linealmente entre dichos intervalos.

El sensor se ha conectado a un cable con una serie de conectores:
- VCC -> cable marrón (1)
- Trig -> cable rojo (2)
- Echo -> cable naranja (3)
- GND -> cable amarillo (4)

----
## Setup Arduino

Se ha usado un Arduino uno, instalando el IDE desde https://www.arduino.cc/ , versión 1.8.10

Para comunicarse con el Arduino desde el Mac, se ha instalado un driver de USB -> puerto serie:
https://github.com/adrianmihalko/ch340g-ch34g-ch34x-mac-os-x-driver

Concretamente:

```
brew tap adrianmihalko/ch340g-ch34g-ch34x-mac-os-x-driver https://github.com/adrianmihalko/ch340g-ch34g-ch34x-mac-os-x-driver
brew cask install wch-ch34x-usb-serial-driver
```

Y luego seleccionar ```/dev/cu.Bluetooth-Incoming-Port``` como puerto

- Programa con HC-SR04
- Programa con VL53L1X

El HC-SR04 se ha conectado al Arduino de la siguiente forma:
- VCC (cable marríon) -> VCC
- Trig (cable rojo) -> pin 9
- Echo (cable naranja) -> pin 10
- GND (cable amarillo) -> GND

El Arduino sólo se ha usado para calibrar las medidas y comprobar el sensor. Este es el código que se ha usado:

```
const int trigPin = 9; 
const int echoPin = 10;

const int samples = 20;
const int time_delay = 500;
const int ceiling = 133;

float duration, distance, stick, liters;
int tranch;

const float liters_table[] = {
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
  };
const int liters_table_size = 127;

void setup() {
 pinMode(trigPin, OUTPUT); 
 pinMode(echoPin, INPUT); 
 Serial.begin(9600);
}

void loop() {

  duration = 0;
  for(int i=0; i<samples; i++) {
    digitalWrite(trigPin, LOW);
    delayMicroseconds(2);
    digitalWrite(trigPin, HIGH);
    delayMicroseconds(10);
    digitalWrite(trigPin, LOW);
    duration += pulseIn(echoPin, HIGH);
    delay(time_delay);
  }
  duration /= samples;

  distance = (duration*.0343)/2;
  stick = ceiling-distance;
  if(stick < 0) {
    liters = 0;
  } else if(stick >= liters_table_size) {
    liters = 3000;
  } else {
    tranch = stick;
    liters = liters_table[tranch] + (liters_table[tranch+1] - liters_table[tranch])*(stick - tranch);
  }
  
  
  Serial.print("Distance: ");
  Serial.print(distance);
  Serial.print(", Stick: ");
  Serial.print(stick);
  Serial.print("; Liters: ");
  Serial.println(liters);
  
}
```

---
## Setup Raspberry

Se ha instalado Raspbian, user es "pi", passwd es "capullete". Se le ha puesto una IP estática: 192.168.86.24 , y se puede conectar a "Mongolillo's Territory" y a "Wifrijoles"

- Medidor
- Antena Wifi USB
- Samba
- Crontab
- Programa con HC-SR04
- Programa con VL53L1X

## Análisis y reports

