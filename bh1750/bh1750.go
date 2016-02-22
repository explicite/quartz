package main

import (
	"flag"
	"time"

	"github.com/explicite/i2c/bh1750"
	"github.com/influxdata/influxdb/client/v2"
)

var sample = flag.Float64("sample", float64(1), "sampling frequency in Hz")

func transaction(f func() error) {
	err := f()
	if err != nil {
		panic(err)
	}
}

func getOutChan(sampling float64, b *bh1750.BH1750) <-chan client.Point {
	points := make(chan client.Point)
	delay := 1 / sampling
	go func() {
		for {
			illu, _ := b.Illuminance(bh1750.ConHRes1lx)

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
			time.Sleep(time.Second * time.Duration(delay))
		}
	}()

	return points
}

func init() {
	flag.Parse()
}

func main() {
	device := &bh1750.BH1750{}
	transaction(func() error { return device.Init(0x23, 1) })
	transaction(device.Active)
	defer device.Deactive()

	out := getOutChan(*sample, device)
	for {
		point := <-out
		println(point.Fields())
	}

}
