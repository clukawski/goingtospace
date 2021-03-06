package main

import (
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
	"github.com/kidoman/embd/sensor/bh1750fvi"
	"github.com/kidoman/embd/sensor/bmp180"
	"github.com/kidoman/embd/sensor/l3gd20"
	"github.com/kidoman/embd/sensor/lsm303"
	"log"
	"os"
	"time"
)

var logger *log.Logger

func init() {
	logfile, err := os.Create("logfile.log")
	if err != nil {
		return
	}
	log.Print("Logging to logfile.log")
	logger = log.New(logfile, "", log.LstdFlags|log.Lmicroseconds)
	// defer logfile.Close()
}

func main() {
	bus := embd.NewI2CBus(1)

	go logBH1750FVI(bus) // luminosity
	go logBMP180(bus)    // barometric pressure; temperature; altitude
	go logLSM303(bus)    // magnetometer
	go logL3GD20(bus)    // gyroscope; temperature

	// sleep forever
	select {}
}

func logBH1750FVI(bus embd.I2CBus) {
	sensor := bh1750fvi.New("High2", bus)
	sensor.Run()

	for range time.Tick(time.Second) {
		lighting, _ := sensor.Lighting()
		logger.Print("lighting:", lighting)
	}
}

func logBMP180(bus embd.I2CBus) {
	sensor := bmp180.New(bus)
	sensor.Run()

	for range time.Tick(time.Second) {
		temperature, _ := sensor.Temperature()
		pressure, _ := sensor.Pressure()
		altitude, _ := sensor.Altitude()
		logger.Print("temperature:", temperature, " pressure:", pressure, " altitude:", altitude)
	}
}

func logL3GD20(bus embd.I2CBus) {
	// NOTE: I picked this range at random.
	sensor := l3gd20.New(bus, l3gd20.R500DPS)
	sensor.Start()

	orientations, _ := sensor.Orientations()

	for range time.Tick(time.Second) {
		o := <-orientations
		t, _ := sensor.Temperature()
		logger.Print("orientation:{x:", o.X, " y:", o.Y, " z:", o.Z, "} temperature:", t)
	}
}

func logLSM303(bus embd.I2CBus) {
	sensor := lsm303.New(bus)
	sensor.Run()

	for range time.Tick(time.Second) {
		heading, _ := sensor.Heading()
		logger.Print("heading:", heading)
	}
}
