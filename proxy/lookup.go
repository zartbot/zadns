package proxy

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

//SeqLookup is based on serverlist sequence
func (p *Proxy) SeqLookup(msg *dns.Msg, serverList []string) (*dns.Msg, error) {
	for _, s := range serverList {
		resp, err := Lookup(msg, s)
		if err == nil {
			return resp, nil
		}
	}
	return nil, fmt.Errorf("serverNotAvailable")
}

//RandomLookup is based on shuffled serverlist sequence
func (p *Proxy) RandomLookup(msg *dns.Msg, serverList []string) (*dns.Msg, error) {
	tServerList := make([]string, 0)
	tServerList = append(tServerList, serverList...)
	rand.Shuffle(len(tServerList), func(i, j int) { tServerList[i], tServerList[j] = tServerList[j], tServerList[i] })
	for _, s := range tServerList {
		resp, err := Lookup(msg, s)
		if err == nil {
			return resp, nil
		}
	}
	return nil, fmt.Errorf("serverNotAvailable")
}

//Lookup record over multipler server return A/AAAA address list in string array

func (p *Proxy) MultipleLookup(msg *dns.Msg, serverList []string) map[string]string {
	t := time.NewTicker(600 * time.Millisecond)
	tmpMap := make(map[string]string, 0)
	resp := make(chan *DNSReport, 10)
	for _, s := range serverList {
		go lookup2Chan(msg, s, resp)
	}
	for {
		select {
		case <-t.C:
			return tmpMap
		case addr := <-resp:
			for _, v := range addr.Record {
				tmpMap[v] += addr.Server
			}
		}
	}
}

type DNSReport struct {
	Server string
	Record []string
}

func lookup2Chan(msg *dns.Msg, server string, reportChan chan<- *DNSReport) {
	resp, err := Lookup(msg, server)
	if err != nil {
		return
	}

	records := DecodeTypeAResponse(resp.Answer)
	result := &DNSReport{
		Server: server + " ",
		Record: records,
	}

	reportChan <- result
}

//Lookup is used to get the response from external server
func Lookup(msg *dns.Msg, server string) (*dns.Msg, error) {
	c := new(dns.Client)
	c.Net = "udp"

	/*
		o := new(dns.OPT)
		o.Hdr.Name = "."
		o.Hdr.Rrtype = dns.TypeOPT
		e := new(dns.EDNS0_SUBNET)
		e.Code = dns.EDNS0SUBNET
		e.Family = 1         //2 for V6
		e.SourceNetmask = 32 //128 for values
		e.SourceScope = 0
		e.Address = net.ParseIP("101.2.3.4").To4()
		o.Option = append(o.Option, e)
		msg.Extra = append(msg.Extra, o)
	*/
	resp, _, err := c.Exchange(msg, server)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func RoutedAddress() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		logrus.Fatal(err)
	}
	result := conn.LocalAddr().(*net.UDPAddr)
	conn.Close()

	return result.IP.String() + ":53", nil
}
