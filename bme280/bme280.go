package bme280

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
)

/**
 * BME280 / Bosch Sensortec
 * Combined temperature, humidity and pressure sensor
 * https://cdn-shop.adafruit.com/datasheets/BST-BME280_DS001-10.pdf
 *
 * Access via i2c protocoll.
 * https://en.wikipedia.org/wiki/I%C2%B2C
**/

type BME280 struct {
	dev *i2c.Dev
	cv  CompensationValues
}

const (
	CTRL_MEAS_ADDR     = 0xF4
	CTRL_HUMIDITY_ADDR = 0xF2
	CTRL_CONFIG        = 0xF5
	REG_PRESSURE       = 0xF7
	REG_CALIBRATION    = 0x88
	REG_CALIBRATION_H1 = 0xA1
	REG_CALIBRATION_H2 = 0xE1
	CMD_RESET          = 0xE0

	WHO_AM_I = 0xD0
	CHIP_ID  = 0x60
)

type TemperatureCompensation struct {
	t1 int32
	t2 int32
	t3 int32
}
type PressureCompensation struct {
	p1 int32
	p2 int32
	p3 int32
	p4 int32
	p5 int32
	p6 int32
	p7 int32
	p8 int32
	p9 int32
}
type HumidityCompensation struct {
	h1 int32
	h2 int32
	h3 int32
	h4 int32
	h5 int32
	h6 int32
}
type CompensationValues struct {
	temperatureCompensation TemperatureCompensation
	pressureCompensation    PressureCompensation
	humidityCompensation    HumidityCompensation
}

type Result struct {
	Temperature int32
	Pressure    uint32
	Humidity    uint32
}

func writeReadTx(d *i2c.Dev, b byte, size int) []byte {
	write := []byte{b}
	read := make([]byte, size)
	if err := d.Tx(write, read); err != nil {
		log.Fatal(err)
	}
	return read
}

func devCheck(d *i2c.Dev) {
	read := writeReadTx(d, WHO_AM_I, 1)
	log.Printf("%#x\n", read[0])
	log.Printf("%v\n", read)
}

// unsigned int from two bytes (little-endian)
func uint16LE(b0 byte, b1 byte) uint16 {
	return uint16(b1)<<8 | uint16(b0)
}

// signed int from two bytes (little-endian)
func int16LE(b0 byte, b1 byte) int16 {
	return int16(b1)<<8 | int16(b0)
}

func readCompensationValues(dev *i2c.Dev) CompensationValues {
	var cv CompensationValues

	read := writeReadTx(dev, REG_CALIBRATION, 24)
	cv.temperatureCompensation.t1 = int32(uint16LE(read[0], read[1]))
	cv.temperatureCompensation.t2 = int32(int16LE(read[2], read[3]))
	cv.temperatureCompensation.t3 = int32(int16LE(read[4], read[5]))

	cv.pressureCompensation.p1 = int32(uint16LE(read[6], read[7]))
	cv.pressureCompensation.p2 = int32(int16LE(read[8], read[9]))
	cv.pressureCompensation.p3 = int32(int16LE(read[10], read[11]))
	cv.pressureCompensation.p4 = int32(int16LE(read[12], read[13]))
	cv.pressureCompensation.p5 = int32(int16LE(read[14], read[15]))
	cv.pressureCompensation.p6 = int32(int16LE(read[16], read[17]))
	cv.pressureCompensation.p7 = int32(int16LE(read[18], read[19]))
	cv.pressureCompensation.p8 = int32(int16LE(read[20], read[21]))
	cv.pressureCompensation.p9 = int32(int16LE(read[22], read[23]))

	read2 := writeReadTx(dev, REG_CALIBRATION_H1, 1)
	cv.humidityCompensation.h1 = int32(uint8(read2[0]))

	read3 := writeReadTx(dev, REG_CALIBRATION_H2, 7)
	cv.humidityCompensation.h2 = int32(int16LE(read3[0], read3[1]))
	cv.humidityCompensation.h3 = int32(uint8(read3[2]))
	cv.humidityCompensation.h4 = int32((int16(read3[3]) << 4) | (int16(read3[4] & 0x0F)))
	cv.humidityCompensation.h5 = int32((int16(read3[5]) << 4) | (int16(read3[4]) >> 4))
	cv.humidityCompensation.h6 = int32(read3[6])

	return cv
}

func (d *BME280) SetConfiguration() {
	// oversampling ×1
	d.dev.Write([]byte{CTRL_HUMIDITY_ADDR, 0x01})

	// Pressure oversampling x1   Temperature oversampling x1, Forced mode
	read := writeReadTx(d.dev, CTRL_MEAS_ADDR, 1)
	fmt.Printf("CTRL_MEAS_ADDR: %#x\n", read[0])
	OVERSAMPLE_TEMP := 1
	OVERSAMPLE_PRES := 1
	MODE := 1
	control := OVERSAMPLE_TEMP<<5 | OVERSAMPLE_PRES<<2 | MODE
	d.dev.Write([]byte{CTRL_MEAS_ADDR, byte(control)})
	read2 := writeReadTx(d.dev, CTRL_MEAS_ADDR, 1)
	fmt.Printf("CTRL_MEAS_ADDR: %#x\n", read2[0])

	// Filter OFF
	d.dev.Write([]byte{CTRL_CONFIG, 0x00})
}

func (d *BME280) ReadValues() Result {
	read4 := writeReadTx(d.dev, REG_PRESSURE, 8)

	rawPressure := int32((uint32(read4[0]) << 12) | (uint32(read4[1]) << 4) | (uint32(read4[2]) >> 4))
	fmt.Printf("raw Pressure: %d\n", rawPressure)

	rawTemp := int32((uint32(read4[3]) << 12) | (uint32(read4[4]) << 4) | (uint32(read4[5]) >> 4))
	fmt.Printf("raw Temp: %d\n", rawTemp)

	rawHumidity := int32((uint32(read4[6]) << 8) | uint32(read4[7]))
	fmt.Printf("raw Humidity: %d\n", rawHumidity)

	return compensation(d.cv, rawTemp, rawPressure, rawHumidity)
}

func compensation(cv CompensationValues, rawTemp int32, rawPressure int32, rawHumidity int32) Result {
	var r Result

	// Temperature compensation (int32)
	tvar1 := ((rawTemp >> 3) - (cv.temperatureCompensation.t1 << 1)) * cv.temperatureCompensation.t2
	tvar2 := (((rawTemp >> 4) - cv.temperatureCompensation.t1) *
		((rawTemp >> 4) - cv.temperatureCompensation.t1) >> 12) * cv.temperatureCompensation.t3
	tFine := (tvar1 >> 11) + (tvar2 >> 14)
	r.Temperature = ((tFine*5 + 128) >> 8)

	// Pressure compensation (int32)
	var1 := (tFine >> 1) - 64000
	var2 := (((var1 >> 2) * (var1 >> 2)) >> 11) * cv.pressureCompensation.p6
	var2 = var2 + ((var1 * cv.pressureCompensation.p5) << 1)
	var2 = (var2 >> 2) + (cv.pressureCompensation.p4 << 16)
	var1 = (((cv.pressureCompensation.p3 * (((var1 >> 2) * (var1 >> 2)) >> 13)) >> 3) +
		((cv.pressureCompensation.p2 * var1) >> 1)) >> 18
	var1 = ((32768 + var1) * cv.pressureCompensation.p1) >> 15
	if var1 == 0 {
		r.Pressure = 0 // avoid exception caused by division by zero
	} else {
		p := uint32((1048576-rawPressure)-(var2>>12)) * 3125
		if p < 0x80000000 {
			p = (p << 1) / uint32(var1)
		} else {
			p = (p / uint32(var1)) * 2
		}
		var1 = (cv.pressureCompensation.p9 * (int32(((p >> 3) * (p >> 3)) >> 13))) >> 12
		var2 = (int32(p>>2) * cv.pressureCompensation.p8) >> 13
		r.Pressure = uint32(int32(p) + ((var1 + var2 + cv.pressureCompensation.p7) >> 4))
	}

	// Humidity compensation (int32)
	v_x1_u32r := tFine - 76800
	v_x1_u32r = (((((rawHumidity << 14) - (cv.humidityCompensation.h4 << 20) -
		(cv.humidityCompensation.h5 * v_x1_u32r)) + 16384) >> 15) *
		(((((((v_x1_u32r*cv.humidityCompensation.h6)>>10)*
			(((v_x1_u32r*cv.humidityCompensation.h3)>>11)+32768))>>10)+2097152)*
			cv.humidityCompensation.h2 + 8192) >> 14))

	v_x1_u32r = v_x1_u32r - (((((v_x1_u32r >> 15) * (v_x1_u32r >> 15)) >> 7) * cv.humidityCompensation.h1) >> 4)
	if v_x1_u32r < 0 {
		fmt.Printf("Humidity value too small: %d\n", v_x1_u32r)
		v_x1_u32r = 0
	}
	if v_x1_u32r > 419430400 {
		fmt.Printf("Humidity value too big: %d\n", v_x1_u32r)
		v_x1_u32r = 419430400
	}
	r.Humidity = uint32(v_x1_u32r >> 12)

	return r
}

func InitBme280(address int) (*BME280, error) {
	log.Print("init Bme280")

	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
		return nil, err
	}
	log.Print("host okay")

	// Use i2creg I²C bus registry to find the first available I²C bus.
	b, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	log.Print("i2c okay")
	// defer b.Close()

	// Dev is a valid conn.Conn.
	d := &i2c.Dev{Addr: uint16(address), Bus: b}

	devCheck(d)

	compensationValues := readCompensationValues(d)

	return &BME280{d, compensationValues}, nil
}
