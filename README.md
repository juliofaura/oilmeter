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

Se ha usado una Raspbery Pi 3B, aunque también se podría hacer con una A+. Para el setup:
- Se instala Raspbian (https://www.raspberrypi.org/downloads/raspbian/)
- Se cambia el nombre del host, editando los ficheros ```/etc/hostname``` y ```/etc/hosts``` (este último sólo para el alias de localhost)
- Se registran las redes wifi a las que conectarse, editando el fichero ```/etc/wpa_supplicant/wpa_supplicant.conf``` y añadiendo las entradas relevantes:
```
network={
    ssid="xxx"
    psk="xxx"
    key_mgmt=WPA-PSK
}
```
- Se hacen la(s) IP(s) estática(s), editando el fichero ```/etc/dhcpcp.conf``` para añadir los datos correctos:
```
# Example static IP configuration:
interface wlan0
static ip_address=192.168.1.XX/24
static routers=192.168.1.1
static domain_name_servers=80.58.61.250
```
- Con ```Raspi-config``` se habilita SSH, dentro de "Interfacing options"
- Se instalan claves públicas para poder acceder vía ssh sin password, haciendo ```ssh-copy-id pi@<raspberry pi IP>``` desde donde se quiera acceder a futuro
- Se inhabilita el acceso por SSH con password, editando el fichero ```/etc/ssh/sshd_config``` e inhabilitando todo el acceso con password:
```
ChallengeResponseAuthentication no
PasswordAuthentication no
UsePAM no
PermitRootLogin no
```
, y después reiniciando el servicio de SSH con ```/etc/init.d/ssh reload``` o con ```sudo systemctl reload ssh```

y se ha conectado a la red wifi a través de una antena auxiliar USB (red "Mongolillo's Territory"). Con Raspi-config se ha habilitado ssh, y se han instalado las claves públicas para acceder sin passsword

El medidor se alimenta con 5V, pero para el pin de lectura del eco es necesario utilizar un puente de resistencias R+2R. La conexión es a través de una placa auxiliar, de la siguiente forma:

1  2  3  4
o  o  o  o
          
o-[ R1 ]-o
o-[ R2 ]-o
o-[ R3 ]-o
          
[] [] [] []
5  6  7  8

Conexiones con el sensor:
- VCC del sensor (cable marrón) -> pin 1
- Trig del sensor (cable rojo) -> pin 2
- Echo del sensor (cable naranja) -> pin 3
- GND del sensor (cable amarillo) -> pin 4

Conexiones con la raspberry:
- Pin 5 (cable mnarrón) -> VCC de la Raspberry (pin 2)
- Pin 6 (cable rojo)-> GPIO 27 (pin 13)
- Pin 7 (cable naranja)-> GPIO 17 (pin 11)
- Pin 8 (cable amarillo)-> GBD de la Raspberry (pin 9)

Internamente, la placa auxiliar hace las siguientes conexiones:
- Pin 1 a pin 5, y pin 4 a pin 8 (conexiones de VCC y de GND)
- Pin 2 (trig del sensor) a pin 6 (GPIO 27)
- Pin 3 (echo del sensor) a una pata de R1
- Pin 7 (GPIO17) a la otra pata de R1
- Esta segunda pata de R1 a una pata de R2
- La otra pata de R2 a una pata de R3 (conexión en serie)
- La otra pata de R3 al pin 8 (GND)

<!-- - VCC del sensor (cable marríon) -> VCC (pin 2)
- GND del sensor (cable amarillo) -> GND (pin 9)
- Trig del sensor (cable rojo) -> GPIO 27 (pin 13)
- Echo del sensor  (cable naranja) -> A un extremo del puente de resistencias (a la resistencia R)
- El otro extremo del puente de resistencias (a la resistencia 2R) a GND (pin 6)
- El medio del puente (en el que se unen ambas resistencias) -> GPIO 17 (pin 11) -->

Para usar el VL53L1X:
- VCC a VCC 3.3V (pin 1, cable rojo)
- SDA a SDA I2C1 (Pin 3, cable naranja)
- SCL a SCL I2C1 (Pin5, cable amarillo)
- GND a GND (pin 7, cable marrón)


De cara a poder comunicar los datos de forma que sean accesibles desde fuera, se ha montado un disco remoto por Samba. Para ello hay que:
- Añadir la siguiente línea a ```/etc/fstab```:
```
//192.168.1.24/Gas /home/pi/Gasoleo cifs credentials=/home/pi/.smbcredentials,uid=pi,gid=pi,_netdev,auto 0 0
```
- Añadir un fichero ```.smbcredentials```:
```
username=pi
password=<passwd>
```
- Montar el disco en startup, añadiendo  ```mount -a``` en ```/etc/rc.local``` (y añadir ```sleep 20``` just in case)
- Correr ```raspi-config```, seleccionar ```Boot Options```, y habilitar ```Wait for network at boot```


Para usar el VL53L1X y en general I2C:
- Correr ```raspi-config```, seleccionar ```Interfacing Options```, y habilitar ```I2C```
- Instalar librerías y herramientas:
```
sudo apt-get install -y i2c-tools
sudo pip install smbus2
sudo pip install vl53l1x
```

Para contectar medidores de temperatura y en general comunicaciones W1 (en GPIO 04, que es el pin por defecto):
```
sudo echo dtoverlay=w1-gpio-pullup,gpiopin=4 >> /boot/config.txt
sudo modprobe w1_gpio && sudo modprobe w1_therm
sudo modprobe wire
sudo modprobe w1-gpio
sudo modprobe w1-therm
```

Finalmente, poner en el crontab una medida diaria, o una por hora etc.

El código utilizado está en https://github.com/juliofaura/oilmeter/oilmeter.go


## Setup caldera

Read heat: GPIO 17 (white cable) => up (1) is on, down (0) is off
Read power: GPIO 27 (gray cable) => down (0) is on, up (1) is off

Relés:
  Caldera off / on: GPIO 14 (gray cable) and GIO 15 (white cable) => down is off, up is on
  Heat on / off: GPIO 23 => down is off, up is on

Placa auxiliar:

1  2  3  4  5  6
o  o  o  o  o  o

    -xxxx-
    -xxxx-
    -xxxx-
    -xxxx-
    -xxxx-
    -xxxx-
    -xxxx-
    -xxxx-

1 -> polo + del switch de heat (cable granate)
2 -> polo - del switch de heat (cable azul)
4 -> GPIO 27 (cable gris)
5 -> GPIO 17 (cable blanco)
6 -> GND (cable negro)

Set power: GPIO 14 & GPIO 15 [output, 0 is off, 1 is on]
Set heat: GPIO 23 [output, 0 is off, 1 is on]
Read power: GPIO 27 [input, pullup, 0 is on, 1 is off]
Read heat: GPIO 17 [input, pullup, 1 is on, 0 is off]

Config:
raspi-gpio set 14-15 op dh
raspi-gpio set 23 op dl
raspi-gpio set 17 ip pu
raspi-gpio set 27 ip pu

set power:
On -> raspi-gpio set 14-15 dh
Off -> raspi-gpio set 14-15 dl

set heat:
On -> raspi-gpio set 23 dh
Off -> raspi-gpio set 23 dl

Read power:
if [ -n "$(raspi-gpio get 27 | grep level=0)" ] ; then echo On; else echo Off; fi

Read heat:
if [ -n "$(raspi-gpio get 17 | grep level=1)" ] ; then echo On; else echo Off; fi

Last measure: 47



/home/pi/Local/oilmeter /home/pi/Local; cp /home/pi/Local/2020* /home/pi/Gasoleo/data/; cp /home/pi/Local/data.txt /home/pi/Gasoleo; cp /home/pi/Local/graph.png /home/pi/Gasoleo