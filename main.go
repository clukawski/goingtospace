package main

import (
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
	"github.com/kidoman/embd/sensor/bh1750fvi"
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

	go logLight(bus)

	// sleep forever
	select {}
}

func logLight(bus embd.I2CBus) {
	sensor := bh1750fvi.New("High2", bus)
	sensor.Run()

	ticker := time.Tick(time.Second)

	for range ticker {
		lighting, _ := sensor.Lighting()
		logger.Print("lighting:", lighting)
	}
}
