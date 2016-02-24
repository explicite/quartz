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
	"github.com/explicite/i2c/si7021"
)

var tick = flag.Float64("tick", float64(3600), "tick in sec")
var mes = flag.String("mes", "", "measurement")

var (
	logTrace   *log.Logger
	logInfo    *log.Logger
	logWarning *log.Logger
	logError   *log.Logger
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

func out(l *lps331ap.LPS331AP, b *bh1750.BH1750, s *si7021.SI7021) <-chan string {
	tmp := make(chan string)
	go func() {
		for {
			temp, _ := l.Temperature()
			press, _ := l.Pressure()
			illu, _ := b.Illuminance(bh1750.ConHRes1lx)
			rh, _ := s.RelativeHumidity(false)
			str := fmt.Sprintf(
				"%s quantity=tmp value=%f\n%s quantity=press value=%f\n%s quantity=lux value=%f\n%s quantity=rh value=%f",
				*mes, temp, *mes, press, *mes, illu, *mes, rh)
			tmp <- str
			logTrace.Println(str)
			time.Sleep(time.Duration(*tick) * time.Second)
		}
	}()

	return tmp
}

func main() {
	logging(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	logTrace.Println("lps331AP initialization")
	l := &lps331ap.LPS331AP{}
	transaction(func() error { return l.Init(0x5d, 0x01) })
	transaction(l.Active)
	defer l.Deactive()

	logTrace.Println("bh1750 initialization")
	b := &bh1750.BH1750{}
	transaction(func() error { return b.Init(0x23, 0x01) })
	transaction(b.Active)
	defer b.Deactive()

	logTrace.Println("si7021 initialization")
	s := &si7021.SI7021{}
	transaction(func() error { return s.Init(0x40, 0x01) })
	transaction(s.Active)
	defer s.Deactive()
	out := out(l, b, s)
	for {
		msg := []byte(<-out)
		resp, err := http.Post("http://localhost:8086/write?db=quartz", "text/plain", bytes.NewBuffer(msg))
		if err != nil {
			logWarning.Println(fmt.Sprintf("cannot send measurement:%v", err))
		}
		logInfo.Println(fmt.Sprintf("response:%v", resp))

	}
}
