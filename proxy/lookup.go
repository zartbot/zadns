package proxy

import (
	"fmt"
	"math/rand"
	"net"

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
