package main

import (
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/zartbot/zadns/proxy"
	"time"
)

func main() {
	p := proxy.New("config/route.cfg", "config/hosts.cfg", "config/server.cfg", "model/geoip/geoip.mmdb", "model/geoip/asn.mmdb")

	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		//proxy logic
		result, err := p.GetResponse(r)
		if err == nil {
			w.WriteMsg(result)
		}
	})

	laddr, err := proxy.RoutedAddress()
	if err != nil {
		logrus.Fatal("noRoutedAddress:", err)
	}
	server := &dns.Server{
		Addr: laddr,
		Net:  "udp",
	}
	go p.TCPProbe(time.Second * 5)
	err = server.ListenAndServe()
	if err != nil {
		logrus.Warn(err)
	}
}
