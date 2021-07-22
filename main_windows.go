package main

import (
	"regexp"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/zartbot/zadns/proxy"
)

func main() {
	p := proxy.New("config\\route.cfg", "config\\hosts.cfg", "config\\server.cfg", "model\\geoip\\geoip.mmdb", "model\\geoip\\asn.mmdb")

	exp := "aaa.*\\.cisco.com"
	match1, _ := regexp.MatchString(exp, "aaa.www.cisco.com")
	logrus.Warn(match1)
	match, _ := regexp.MatchString(exp, "www.cisaco.com.cisco.com.aaaa.bbb")
	logrus.Warn(match)

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
	err = server.ListenAndServe()
	if err != nil {
		logrus.Warn(err)
	}
}
