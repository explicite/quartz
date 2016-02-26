package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/explicite/quartz/bh1750"
	"github.com/explicite/quartz/lps331ap"
	"github.com/explicite/quartz/si7021"
	"github.com/influxdata/influxdb/client/v2"
)

var (
	logTrace   *log.Logger
	logInfo    *log.Logger
	logWarning *log.Logger
	logError   *log.Logger
)

func logging(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	logTrace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	logInfo = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	logWarning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	logError = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func merge(cs ...<-chan client.Point) <-chan client.Point {
	var wg sync.WaitGroup
	out := make(chan client.Point)

	output := func(c <-chan client.Point) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	logging(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	sampling := 2.0

	logInfo.Println("lps331AP initialization")
	l := lps331ap.New(0x5d, 0x01, sampling)

	logInfo.Println("bh1750 initialization")
	b := bh1750.New(0x23, 0x01, sampling)

	logInfo.Println("si7021 initialization")
	s := si7021.New(0x40, 0x01, sampling)

	for point := range merge(l, b, s) {
		//TODO send to influxdb
		println(point.Fields())
	}
}
