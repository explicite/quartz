package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/explicite/i2c/bh1750"
	"github.com/explicite/i2c/lps331ap"
)

var tick = flag.Float64("tick", float64(3600), "tick in sec")
var mes = flag.String("mes", "", "measurement")

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func transaction(f func() error) {
	err := f()
	if err != nil {
		panic(err)
	}
}

func init() {
	flag.Parse()
	if *mes == "" {
		panic("set measurement name")
	}
}

func Logging(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func out(l *lps331ap.LPS331AP, b *bh1750.BH1750) <-chan string {
	tmp := make(chan string)
	go func() {
		for {
			temp, _ := l.Temperature()
			press, _ := l.Pressure()
			lux, _ := b.Lux(bh1750.CON_H_RES_1LX)
			str := fmt.Sprintf(
				"%s quantity=tmp value=%f\n%s quantity=press value=%f\n%s quantity=lux value=%f",
				*mes, temp, *mes, press, *mes, lux)
			tmp <- str
			time.Sleep(time.Duration(*tick) * time.Second)
		}
	}()

	return tmp
}

func main() {
	Logging(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	Trace.Println("lps331AP initialization")
	l := &lps331ap.LPS331AP{}
	transaction(func() error { return l.Init(0x5d, 1) })
	transaction(l.Active)
	defer l.Deactive()

	Trace.Println("bh1750 initialization")
	b := &bh1750.BH1750{}
	transaction(func() error { return b.Init(0x23, 1) })
	transaction(b.Active)
	defer b.Deactive()

	out := out(l, b)
	for {
		println(<-out)
		msg := []byte(<-out)
		resp, err := http.Post("http://localhost:8086/write?db=quartz", "text/plain", bytes.NewBuffer(msg))
		if err != nil {
			Warning.Println(fmt.Sprintf("cannot send measurement:%v", err))
		}
		Info.Println(fmt.Sprintf("response:%v", resp))

	}
}
