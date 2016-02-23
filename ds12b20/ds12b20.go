package main

import (
	"flag"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
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
func (this *ds18b20) getOutChan(sampling float64) <-chan client.Point {
	file := fmt.Sprintf("/sys/bus/w1/devices/%s/w1_slave", this.ID)
	points := make(chan client.Point)
	delay := 1 / sampling

	go func() {
		for {
			content, err := ioutil.ReadFile(file)
			if err != nil {
				log.Fatal(err)
			}
			result := string(content)
			tmpStr := result[len(result)-6 : len(result)-1]
			tmp, err := strconv.Atoi(tmpStr)
			if err != nil {
				log.Fatal(err)
			}

			tags := map[string]string{
				"sensor": "ds12b20",
				"type":   "weather",
				"tmp":    "°C",
			}

			fields := map[string]interface{}{
				"tmp": float64(tmp) / float64(1e3),
			}

			point, _ := client.NewPoint("bs12b20", tags, fields, time.Now())
			points <- *point
			time.Sleep(time.Second * time.Duration(delay))
		}
	}()

	return points
}

func main() {
	device := ds18b20{*id}

	out := device.getOutChan(*sample)
	for {
		point := <-out
		println(point.Fields())
	}

}
