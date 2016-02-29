package main

import (
	"flag"
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

var addr = flag.String("addr", "", "database addr")
var db = flag.String("db", "", "database name")
var user = flag.String("user", "", "database user")
var pass = flag.String("pass", "", "database password")

func init() {
	flag.Parse()
	if *addr == "" {
		panic("set database addr")
	}

	if *db == "" {
		panic("set database name")
	}

	if *user == "" {
		panic("set database user")
	}

	if *pass == "" {
		panic("set database pass")
	}
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

	sampling := 30.0

	// Setup influxdb udp connection
	// Make client
	c, cErr := client.NewHTTPClient(client.HTTPConfig{
		Addr:     *addr,
		Username: *user,
		Password: *pass,
	})

	if cErr != nil {
		panic(cErr)
	}

	logInfo.Printf("influxdb connection established (%s, %s)\n", *addr, *user)

	logInfo.Println("lps331AP initialization")
	l := lps331ap.New(0x5d, 0x01, sampling)

	logInfo.Println("bh1750 initialization")
	b := bh1750.New(0x23, 0x01, sampling)

	logInfo.Println("si7021 initialization")
	s := si7021.New(0x40, 0x01, sampling)

	points := merge(l, b, s)
	batchSize := 30

	for {
		// Create a new point batch
		bp, bpErr := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  *db,
			Precision: "s",
		})

		if bpErr != nil {
			panic(bpErr)
		}

		for i := 0; i < batchSize; i++ {
			p := <-points
			bp.AddPoint(&p)
		}

		writeErr := c.Write(bp)
		if writeErr != nil {
			logInfo.Printf("Write batch points with error: %v", writeErr)
		}
	}

}
