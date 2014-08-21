// Package bmp180 allows interfacing with Bosch HIH6130 barometric pressure sensor. This sensor
// has the ability to provided compensated temperature and pressure readings.
package hih6130

import (
	"math"
	"sync"
	"time"

	"github.com/golang/glog"
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

	temps  chan uint16
	humids chan float64
	quit   chan struct{}
}

// Init sends the high bit to the sensor and "turns it on"
func (h *HIH6130) Init() error {
	err := h.Bus.WriteByte(address, embd.High)
	return err
}

// Stop sends the low bit to the sensor and turns it off.
func (h *HIH6130) Stop() error {
	err := h.Bus.WriteByte(address, embd.Low)
	return err
}

// New returns a handle to a HIH6130 sensor.
func New(bus embd.I2CBus) *HIH6130 {
	return &HIH6130{Bus: bus, Poll: pollDelay}
}

func (h *HIH6130) MeasureHumidAndTemp() (uint16, uint16) {
	data := make([]byte, 4)
	if err := h.Bus.ReadFromReg(address, fakereg, data); err != nil {
		return 0, 0, err
	}

	// Reading 4 bytes of data. First two are status bits (2) humidity data (6, 8), second two are temperature data (8, 6, with the last two bits DNC)

	status := uint8(data[0] >> 6)
	hdata := uint16(data[0] & 0x3f) << 8 | uint16(data[1])
	tdata := (uint16(data[2]) << 8 | uint16(data[3])) >> 2
}

// Run starts the sensor data acquisition loop.
func (h *HIH6130) Run() {
	go func() {
		d.quit = make(chan struct{})
		timer := time.Tick(time.Duration(d.Poll) * time.Millisecond)

		var temp uint16
		var humid uint16

		for {
			select {
			case <-timer:
				t, err := d.measureTemp()
				if err == nil {
					temp = t
				}
				if err == nil && d.temps == nil {
					d.temps = make(chan uint16)
				}
				p, a, err := d.measurePressureAndAltitude()
				if err == nil {
					pressure = p
					altitude = a
				}
				if err == nil && d.pressures == nil && d.altitudes == nil {
					d.pressures = make(chan uint16)
					d.altitudes = make(chan uint16)
				}
			case h.temps <- temp:
			case h.humid <- humidity:
			case <-h.quit:
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
	if h.quit != nil {
		h.quit <- struct{}{}
	}
}
