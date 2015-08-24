package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

var id = flag.String("id", "", "ds12b20 id from /sys/bus/w1/devices")
var sample = flag.Float64("sample", float64(1), "sampling frequency in Hz")

func init() {
	flag.Parse()
	if *id == "" {
		log.Fatal("set device id flag")
	}
}

type ds18b20 struct {
	ID string
}

//Return out chan with sampling in Hz
func (this *ds18b20) getOutChan(sampling float64) <-chan float64 {
	file := fmt.Sprintf("/sys/bus/w1/devices/%s/w1_slave", this.ID)
	tempChan := make(chan float64)
	delay := 1 / sampling

	go func() {
		for {
			content, err := ioutil.ReadFile(file)
			if err != nil {
				log.Fatal(err)
			}
			result := string(content)
			tempString := result[len(result)-6 : len(result)-1]
			temp, err := strconv.Atoi(tempString)
			if err != nil {
				log.Fatal(err)
			}
			tempChan <- float64(temp) / float64(1e3)
			time.Sleep(time.Second * time.Duration(delay))
		}
	}()

	return tempChan
}

func main() {
	device := new(ds18b20)
	device.ID = *id
	out := device.getOutChan(*sample)
	for {
		println(fmt.Sprintf("%f", <-out))
	}

}
