package main

import (
	"fmt"

	"github.com/explicite/i2c/lps331ap"
)

func transaction(f func() error) {
	err := f()
	if err != nil {
		panic(err)
	}
}

func main() {
	device := &lps331ap.LPS331AP{}
	transaction(func() error { return device.Init(0x5d, 1) })
	transaction(device.Active)
	defer device.Deactive()

	pressure, _ := device.Pressure()
	temperature, _ := device.Temperature()

	fmt.Printf("%.2f \t%.2f\n", pressure, temperature)
}
