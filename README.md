# Documentación del medidor del tanque de gasóleo

## Descripción general

Se trata de poner un sensor de distancia dentro del tanque de gasóleo y conectarlo a una raspberry para medir periódicamente el gasóleo que queda, comunicándoselo a un servidor que lo publique en un interfaz web o móvil (y/o que envíe alarmas en situaciones relevantes)

Usamos un medidor de ultrasonidos HC-SR04 para medir el espacio entre el techo del tanque y el nivel de gasóleo en el tanque. Aproximamos la forma del tanque a la de un cilindro tumbado de radio 634.5mm, siendo la fórmula que relaciona los litros de gasóleo con el nivel desde el fondo del tanque la siguiente:

```litros de gasóleo = K*(ASIN((nivel-R)/R)+0.5*SIN(2*ASIN((nivel-R)/R))+PI()/2)```

, donde ```R = 634.5 mm``` y ```K = 954.5 litros``` (esta fórmula responde a la integral de una circunferencia, que es la base del tanque como se ha dicho)

(Cálculos detallados en https://www.dropbox.com/s/0gnpqd8h2pqaif8/C%C3%A1lculos%20tanque%20de%20gasoleo.ods?dl=0 )

La Raspberry utiliza un crontab para realizar las mediciones, una por hora, y añade la medida a un fichero guardado en un diosco compartido (con un servidor Samba). El mismo programa de la Raspberry hace un pequeño análisis y actualiza un informe que puede consultarse en el disco compartido de forma remota

---
## Medidor

Se desueldan los altavoces del HC-SR04, y se reconectan a través de cables no muy largos (unos 10 cm) de nuevo a la placa. Los altavoces se ponen en la boca de un tubo de PVC de 1 m de largo, y se fijan con bluetac o similar. Entonces se introduce el tubo de PVC por el tubo de medición del depósito, hasta que la boca del tubo emerge en el depósito - esto se sabe porque las medidas de distance empiezan a tener sentido (mientras que los altavoces están en el tubo del depósito las medidas no tienen sentido). Una vez situada la boca del tubo de PVC con los altavoces fijados, el sensor se queda aproximadamente a 144 cm del fondo del tanque, con lo que una medida de 144 cm de distancia corresponde a cero litros, y una medida de 17 cm corresponde a la capacidad máxima del tanque, que es de 3000 litros. Por tanto la medida del nivel de gasóleo en función de la distancia medida por el HC-SR04 es de ```nivel = 144 - distancia```, con un valor de la distancia que debe estar entre 0 y 127 (las cantidades exactas deben calibrarse en el momento de la colocación del tubo, según la medida que se haga con la varilla de medir)

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

El HC-SR04 se ha conectado al Arduino de la siguiente forma:
- VCC (cable marríon) -> VCC
- Trig (cable rojo) -> pin 9
- Echo (cable naranja) -> pin 10
- GND (cable amarillo) -> GND

El Arduino sólo se ha usado para calibrar las medidas y comprobar el sensor. El código usado está en https://github.com/juliofaura/oilmeter/Sensors/HC-SR04/Arduino/Distance-HC-SR04.ino

---
## Setup Raspberry

Se ha instalado Raspbian y se ha conectado a la red wifi a través de una antena auxiliar USB (red "Mongolillo's Territory"). Con Raspi-config se ha habilitado ssh, y se han instalado las claves públicas para acceder sin passsword

El medidor se alimenta con 5V, pero para el pin de lectura del eco es necesario utilizar un puente de resistencias R+2R. La conexión es de la siguiente forma:
- VCC del sensor (cable marríon) -> VCC (pin 1)
- GND del sensor (cable amarillo) -> GND (pin 9)
- Trig del sensor (cable rojo) -> GPIO 04 (pin 7)
- Echo del sensor  (cable naranja) -> A un extremo del puente de resistencias (a la resistencia R)
- El otro extremo del puente de resistencias (a la resistencia 2R) a GND (pin 6)
- El medio del puente (en el que se unen ambas resistencias) -> GPIO 17 (pin 11)

POR AQUIIIII
De cara a poder comunicar los datos de forma que sean accesibles desde fuera, se ha montado un disco remoto por Samba. 
- Samba => entry in fstab, incl .smbcredentials and _netdev, mount -a in /etc/rc.local (adding sleep 20 just in case), run raspi-config then "Boot Options" then enable "Wait for network at boot" -> so disks are mounted by rc.local
- Crontab (incl)

El código utilizado está en https://github.com/juliofaura/oilmeter/oilmeter.go

## Análisis y reports

