package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/explicite/i2c/bh1750"
)

var sample = flag.Float64("sample", float64(1), "sampling frequency in Hz")

func transaction(f func() error) {
	err := f()
	if err != nil {
		panic(err)
	}
}

func init() {
	flag.Parse()
}

func getOutChan(sampling float64, b *bh1750.BH1750) <-chan float64 {
	tmpChan := make(chan float64)
	delay := 1 / sampling
	go func() {
		for {
			lux, _ := b.Lux(bh1750.CON_H_RES_1LX)
			tmpChan <- float64(lux)
			time.Sleep(time.Second * time.Duration(delay))
		}
	}()

	return tmpChan
}

func main() {
	device := &bh1750.BH1750{}
	transaction(func() error { return device.Init(0x23, 1) })
	transaction(device.Active)
	defer device.Deactive()

	out := getOutChan(*sample, device)
	for {
		println(fmt.Sprintf("%f", <-out))
	}

}
