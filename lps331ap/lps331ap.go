package lps331ap

import (
	"time"

	"github.com/explicite/i2c/lps331ap"
	"github.com/influxdata/influxdb/client/v2"
)

// New channel with measure points from lps331ap.
func New(addr byte, bus byte, sampling float64) <-chan client.Point {
	device := &lps331ap.LPS331AP{}
	device.Init(addr, bus)
	device.Active()

	points := make(chan client.Point)
	go func() {
		defer device.Deactive()
		for {
			press, pressErr := device.Pressure()
			tmp, tmpErr := device.Temperature()

			if pressErr != nil && tmpErr != nil {
				tags := map[string]string{
					"sensor": "lps331ap",
					"type":   "weather",
					"press":  "kPa",
					"tmp":    "°C",
				}

				fields := map[string]interface{}{
					"press": press,
					"tmp":   tmp,
				}

				point, _ := client.NewPoint("lps331ap", tags, fields, time.Now())
				points <- *point
			}

			time.Sleep(time.Second * time.Duration(sampling))
		}
	}()

	return points
}
