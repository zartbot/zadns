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

func (p *Proxy) MultipleLookup(msg *dns.Msg, serverList []string) []string {
	t := time.NewTicker(600 * time.Millisecond)
	tmpMap := make(map[string]interface{}, 0)
	resp := make(chan []string, 10)
	for _, s := range serverList {
		go lookup2Chan(msg, s, resp)
	}
	for {
		select {
		case <-t.C:
			result := make([]string, 0)
			for k, _ := range tmpMap {
				result = append(result, k)
			}
			return result
		case addr := <-resp:
			for _, v := range addr {
				tmpMap[v] = 1
			}
		}
	}
}

/*
func (p *Proxy) DecodeTypeAResponse(question string, answer []dns.RR) []string {
	addressList := make([]string, 0)
	for _, v := range answer {
		if v.Header().Class == dns.TypeA {
			record := strings.Split(v.String(), "\t")
			ipAddr := strings.TrimSpace(record[4])
			if net.ParseIP(ipAddr) != nil {
				addressList = append(addressList, ipAddr)
			}
		}
	}
	return addressList
}
*/

func lookup2Chan(msg *dns.Msg, server string, resp chan<- []string) {
	result, err := Lookup(msg, server)
	if err != nil {
		return
	}
	resultlist := DecodeTypeAResponse(result.Answer)
	logrus.Info("server:", server, "|REsult:", resultlist)
	resp <- resultlist
}

//Lookup is used to get the response from external server
func Lookup(msg *dns.Msg, server string) (*dns.Msg, error) {
	c := new(dns.Client)
	c.Net = "udp"
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
