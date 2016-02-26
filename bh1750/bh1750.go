package bh1750

import (
	"time"

	"github.com/explicite/i2c/bh1750"
	"github.com/influxdata/influxdb/client/v2"
)

// New channel with measure points from bh1750.
func New(addr byte, bus byte, sampling float64) <-chan client.Point {
	device := &bh1750.BH1750{}
	device.Init(addr, bus)
	device.Active()

	points := make(chan client.Point)
	go func() {
		defer device.Deactive()
		for {
			illu, _ := device.Illuminance(bh1750.ConHRes1lx)

			tags := map[string]string{
				"sensor":      "bh1750",
				"type":        "weather",
				"illuminance": "lx",
			}

			fields := map[string]interface{}{
				"illu": illu,
			}

			point, _ := client.NewPoint("bh1750", tags, fields, time.Now())
			points <- *point
			time.Sleep(time.Second * time.Duration(sampling))
		}
	}()

	return points
}
