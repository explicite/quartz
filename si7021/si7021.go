package si7021

import (
	"flag"
	"time"

	"github.com/explicite/i2c/si7021"
	"github.com/influxdata/influxdb/client/v2"
)

var sample = flag.Float64("sample", float64(1), "sampling frequency in Hz")

func transaction(f func() error) {
	err := f()
	if err != nil {
		panic(err)
	}
}

func getOutChan(sampling float64, s *si7021.SI7021) <-chan client.Point {
	points := make(chan client.Point)
	delay := 1 / sampling
	go func() {
		for {
			rh, _ := s.RelativeHumidity(false)
			tmp, _ := s.Temperature(false)

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
			time.Sleep(time.Second * time.Duration(delay))
		}
	}()

	return points
}

func init() {
	flag.Parse()
}

func main() {
	device := &si7021.SI7021{}
	transaction(func() error { return device.Init(0x40, 1) })
	transaction(device.Active)
	defer device.Deactive()

	out := getOutChan(*sample, device)
	for {
		point := <-out
		println(point.Fields())
	}

}
