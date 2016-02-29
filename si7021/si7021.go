package si7021

import (
	"time"

	"github.com/explicite/i2c/si7021"
	"github.com/influxdata/influxdb/client/v2"
)

// New channel with measure points from si7021.
func New(addr byte, bus byte, sampling float64) <-chan client.Point {
	device := &si7021.SI7021{}
	device.Init(addr, bus)
	device.Active()

	points := make(chan client.Point)
	go func() {
		defer device.Deactive()
		for {
			rh, rhErr := device.RelativeHumidity(false)
			tmp, tmpErr := device.Temperature(false)

			if rhErr == nil && tmpErr == nil {
				tags := map[string]string{
					"sensor": "si7021",
					"type":   "weather",
					"rh":     "%",
					"tmp":    "°C",
				}

				fields := map[string]interface{}{
					"rh":  rh,
					"tmp": tmp,
				}

				point, _ := client.NewPoint("si7021", tags, fields, time.Now())
				points <- *point
			}

			time.Sleep(time.Second * time.Duration(sampling))
		}
	}()

	return points
}
