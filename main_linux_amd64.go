package main

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/zartbot/zadns/dga"
	"github.com/zartbot/zadns/proxy"
)

func main() {

	p := proxy.New("config/route.cfg", "config/hosts.cfg", "config/server.cfg", "model/geoip/geoip.mmdb", "model/geoip/asn.mmdb")
	p.LogLevel = "debug"

	dgaModel := dga.New("./model/dga")

	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		if len(r.Question) == 0 {
			return
		}
		//TODO : Dump Request.Name/Type and Request IP for security check
		fmt.Printf("%20s|%d|%8d|%-60s\n", w.RemoteAddr().String(), len(r.Question), r.Question[0].Qtype, r.Question[0].Name)

		question := r.Question[0]
		//DGA Domain security check
		isDGA := dgaModel.Predict(question.Name)
		if isDGA {
			logrus.Warn("DGA: ", question.Name)
			resp := new(dns.Msg)
			resp.SetReply(r)
			w.WriteMsg(resp)
			return
		}
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
