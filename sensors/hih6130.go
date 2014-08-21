// Package bmp180 allows interfacing with Bosch HIH6130 barometric pressure sensor. This sensor
// has the ability to provided compensated temperature and pressure readings.
package hih6130

import (
	"math"
	"sync"
	"time"
	"encoding/binary"
	
	"github.com/kidoman/embd"
)

const (
	address = 0x27
	fakereg = 0x00

	pollDelay = 500
)

// HIH6130 represents a Bosch HIH6130 barometric sensor.
type HIH6130 struct {
	Bus  embd.I2CBus
	Poll int

	cmu sync.RWMutex

	temps  chan int
	humids chan float64
	quit   chan struct{}
}

// Init sends the high bit to the sensor and "turns it on"
func (d *HIH6130) Init() error {
	err := d.Bus.WriteByte(address, embd.High)
	return err
}

// Stop sends the low bit to the sensor and turns it off.
func (d *HIH6130) Stop() error {
	err := d.Bus.WriteByte(address, embd.Low)
	return err
}

// New returns a handle to a HIH6130 sensor.
func New(bus embd.I2CBus) *HIH6130 {
	return &HIH6130{Bus: bus, Poll: pollDelay}
}

func (d *HIH6130) MeasureHumidAndTemp() (int, int) {
	data := make([]byte, 4)
	if err := d.Bus.ReadFromReg(address, fakereg, data); err != nil {
		return 0, 0, err
	}

	// Reading 4 bytes of data. First two are status bits (2) humidity data (6, 8), second two are temperature data (8, 6, with the last two bits DNC)

	status := uint8(data[0] >> 6)
	hdata := uint16(data[0] & 0x3f) << 8 | uint16(data[1])
	tdata := (uint16(data[2]) << 8 | uint16(data[3])) >> 2

	var humid int
	var temp int

	hbuf := bytes.NewReader([]byte(hdata))
	tbuf := bytes.NewReader([]byte(tdata))

	err := binary.Read(hbuf, binary.LittleEndian, &humid)
	if err != nil {
		return 0, 0, err
	}
	err := binary.Read(tbuf, binary.LittleEndian, &temp)
	if err != nil {
		return 0, 0, err
	}

	var h float64
	var t int
	
	h = humid / (math.Pow(2, 14) - 1) * 100
	t = temp / (math.Pow(2, 14) - 1) * 165 - 40

	return h, t, nil
}

// Run starts the sensor data acquisition loop.
func (d *HIH6130) Run() {
	go func() {
		d.quit = make(chan struct{})
		timer := time.Tick(time.Duration(d.Poll) * time.Millisecond)

		var humid int
		var temp int

		for {
			select {
			case <-timer:
				h, t, err := d.MeasureHumidAndTemp()
				if err == nil {
					humid = h
					temp = t
				}
				if err == nil && d.humids == nil && d.temps == nil {
					d.humid = make(chan float64)
					d.temp = make(chan int)
				}
			case d.temps <- temp:
			case d.humid <- humidity:
			case <-d.quit:
				d.temps = nil
				d.humid = nil
				return
			}
		}
	}()

	return
}

// Close.
func (d *HIH6130) Close() {
	if d.quit != nil {
		d.quit <- struct{}{}
	}
}
