package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/explicite/i2c/lps331ap"
)

var sample = flag.Float64("sample", float64(1), "sampling frequency in Hz")

func transaction(f func() error) {
	err := f()
	if err != nil {
		panic(err)
	}
}

func getOutChan(sampling float64, l *lps331ap.LPS331AP) <-chan string {
	tmpChan := make(chan string)
	delay := 1 / sampling
	go func() {
		for {
			pressure, _ := l.Pressure()
			temperature, _ := l.Temperature()
			tmpChan <- fmt.Sprintf("%f %f", pressure, temperature)
			time.Sleep(time.Second * time.Duration(delay))
		}
	}()

	return tmpChan
}

func init() {
	flag.Parse()
}

func main() {
	device := &lps331ap.LPS331AP{}
	transaction(func() error { return device.Init(0x5d, 1) })
	transaction(device.Active)
	defer device.Deactive()

	out := getOutChan(*sample, device)
	for {
		println(<-out)
	}

}
