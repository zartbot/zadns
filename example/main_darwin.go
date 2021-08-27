package main

import (
	"github.com/miekg/dns"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/zartbot/zadns/proxy"
)

func main() {
	p := proxy.New("config/route.cfg", "config/hosts.cfg", "config/server.cfg", "model/geoip/geoip.mmdb", "model/geoip/asn.mmdb")

	l := lumberjack.Logger{
		Filename:   "log/request.log",
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
		logrus.Fatal("noRoutedAddress:", err)
	}
	server := &dns.Server{
		Addr: laddr,
		Net:  "udp",
	}
	err = server.ListenAndServe()
	if err != nil {
		logrus.Warn(err)
	}
}
