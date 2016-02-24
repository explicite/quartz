package ds12b20

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

// New channel with measure points from ds12b20.
func New(id string, sampling float64) <-chan client.Point {
	file := fmt.Sprintf("/sys/bus/w1/devices/%s/w1_slave", id)
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
