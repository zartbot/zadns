package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/miekg/dns"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/zartbot/zadns/dga"
	"github.com/zartbot/zadns/proxy"
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

	p := proxy.New("config/route.cfg", "config/hosts.cfg", "config/server.cfg", "model/geoip/geoip.mmdb", "model/geoip/asn.mmdb")
	if cmd.debug {
		p.LogLevel = "debug"
	}

	l := lumberjack.Logger{
		Filename:   "log/dns.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
	logrus.SetOutput(&l)

	dgaModel := dga.New("./model/dga")

	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		if len(r.Question) == 0 {
			return
		}

		question := r.Question[0]
		//DGA Domain security check
		isDGA := dgaModel.Predict(question.Name)
		if isDGA {
			logrus.WithFields(logrus.Fields{
				"Security": "DGA Warning",
				"client":   w.RemoteAddr().String(),
				"Qtype":    question.Qtype,
			}).Warn(question.Name)

			resp := new(dns.Msg)
			resp.SetReply(r)
			w.WriteMsg(resp)
			return
		}

		logrus.WithFields(logrus.Fields{
			"client": w.RemoteAddr().String(),
			"Qtype":  question.Qtype,
		}).Info(question.Name)

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
