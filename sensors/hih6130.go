// Package hih6130 allows interfacing with Honewell HIH6130 Humidity/Temperature sensor. This sensor
// can provide humidity and temperature readings
package hih6130

import (
	"math"
	"time"
	"encoding/binary"
	"bytes"
	
	"github.com/kidoman/embd"
)

const (
	address = 0x27
	fakereg = 0x00

	pollDelay = 500
)

// HIH6130 represents a Honewell HIH6130 Humidity/Temperature sensor.
type HIH6130 struct {
	Bus  embd.I2CBus
	Poll int

	temps  chan int
	humids chan float64
	quit   chan struct{}
}

// Init sends the high bit to the sensor and "turns it on"
func (d *HIH6130) Init() error {
	if err := embd.InitGPIO(); err != nil {
		return err
	}
	defer embd.CloseGPIO()
	
	if err := embd.SetDirection(1, embd.Out); err != nil {
		return err
	}

	if err := embd.DigitalWrite(1, embd.High); err != nil {
		return err
	}

	return nil
}

// Stop sends the low bit to the sensor and turns it off.
func (d *HIH6130) Stop() error {
	if err := embd.InitGPIO(); err != nil {
		return err
	}
	defer embd.CloseGPIO()
	
	if err := embd.SetDirection(1, embd.Out); err != nil {
		return err
	}

	if err := embd.DigitalWrite(1, embd.Low); err != nil {
		return err
	}

	return nil
}

// New returns a handle to a HIH6130 sensor.
func New(bus embd.I2CBus) *HIH6130 {
	return &HIH6130{Bus: bus, Poll: pollDelay}
}

func (d *HIH6130) MeasureHumidAndTemp() (float64, int, error) {
	data := make([]byte, 4)
	if err := d.Bus.ReadFromReg(address, fakereg, data); err != nil {
		return 0, 0, err
	}

	// Reading 4 bytes of data. First two are status bits (2) humidity data (6, 8), second two are temperature data (8, 6, with the last two bits DNC)

	// status := uint8(data[0] >> 6) don't need this yet
	hdata := uint16(data[0] & 0x3f) << 8 | uint16(data[1])
	tdata := (uint16(data[2]) << 8 | uint16(data[3])) >> 2

	var humid float64
	var temp float64

	hbytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(hbytes, uint16(hdata))

	tbytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(tbytes, uint16(tdata))

	hbuf := bytes.NewReader(hbytes)
	tbuf := bytes.NewReader(tbytes)

	err := binary.Read(hbuf, binary.LittleEndian, &humid)
	if err != nil {
		return 0, 0, err
	}
	err = binary.Read(tbuf, binary.LittleEndian, &temp)
	if err != nil {
		return 0, 0, err
	}

	var h float64
	var t float64
	
	h = humid / (math.Pow(2, 14) - 1) * 100
	t = temp / (math.Pow(2, 14) - 1) * 165 - 40

	return h, int(t), nil
}

// Run starts the sensor data acquisition loop.
func (d *HIH6130) Run() {
	go func() {
		d.quit = make(chan struct{})
		timer := time.Tick(time.Duration(d.Poll) * time.Millisecond)

		var humid float64
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
					d.humids = make(chan float64)
					d.temps = make(chan int)
				}
			case d.temps <- temp:
			case d.humids <- humid:
			case <-d.quit:
				d.temps = nil
				d.humids = nil
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
