package main

import (
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/zartbot/zadns/proxy"
	"os"
)

var cmd = struct {
	debug bool
	lat   float64
	long  float64
}{

	false,
	31.02,
	121.1,
}

func init() {
	flag.BoolVar(&cmd.debug, "debug", cmd.debug, "Debug mode")
	flag.Float64Var(&cmd.lat, "lat", cmd.lat, "Latitude")
	flag.Float64Var(&cmd.lat, "long", cmd.long, "Longitude")
	flag.Parse()
}

func main() {
	p := proxy.New("config\\route.cfg", "config\\hosts.cfg", "config\\server.cfg", "model\\geoip\\geoip.mmdb", "model\\geoip\\asn.mmdb")

	if cmd.debug {
		p.LogLevel = "debug"
	}

	l := lumberjack.Logger{
		Filename:   "log\\request.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
	logrus.SetOutput(&l)

	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		logrus.WithFields(logrus.Fields{
			"client": w.RemoteAddr().String(),
			"Qtype":  r.Question[0].Qtype,
		}).Info(r.Question[0].Name)

		//proxy logic
		result, err := p.GetResponse(r)
		if err == nil {
			w.WriteMsg(result)
		}
	})

	go p.TCPProbe()

	laddr, err := proxy.RoutedAddress()
	if err != nil {
		fmt.Printf("noRoutedAddress: %v\n", err)
		os.Exit(1)
	}
	server := &dns.Server{
		Addr: laddr,
		Net:  "udp",
	}
	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("listen failed: %v\n", err)
		os.Exit(1)
	}
}
