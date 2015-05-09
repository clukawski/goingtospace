package main

import (
	"fmt"
	"time"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
	"github.com/kidoman/embd/sensor/bh1750fvi"
)

func main() {
	bus := embd.NewI2CBus(1)

	light := bh1750fvi.New("High2",bus)
	light.Run()

	for {
		current, _ := light.Lighting()
		time.Sleep(1000 * time.Millisecond)
		fmt.Println(current)
	}
}
