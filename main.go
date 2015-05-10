package main

import (
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
	"github.com/kidoman/embd/sensor/bh1750fvi"
	"github.com/kidoman/embd/sensor/bmp180"
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

	// sleep forever
	select {}
}

func logBH1750FVI(bus embd.I2CBus) {
	sensor := bh1750fvi.New("High2", bus)
	sensor.Run()

	ticker := time.Tick(time.Second)

	for range ticker {
		lighting, _ := sensor.Lighting()
		logger.Print("lighting:", lighting)
	}
}

func logBMP180(bus embd.I2CBus) {
	sensor := bmp180.New(bus)
	sensor.Run()

	ticker := time.Tick(time.Second)

	for range ticker {
		temperature := sensor.Temperature()
		pressure := sensor.Pressure()
		altitute := sensor.Altitude()
		logger.Print("temperature:", temperature, " pressure:", pressure, " altitude:", altitude)
	}
}
