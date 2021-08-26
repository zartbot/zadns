package main

import (
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/zartbot/zadns/dga"
	"github.com/zartbot/zadns/proxy"
)

func main() {

	p := proxy.New("config/route.cfg", "config/hosts.cfg", "config/server.cfg", "model/geoip/geoip.mmdb", "model/geoip/asn.mmdb")

	dgaModel := dga.New("./model/dga")
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		if len(r.Question) == 0 {
			return
		}
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
