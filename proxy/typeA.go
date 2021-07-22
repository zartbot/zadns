package proxy

import (
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

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

func (p *Proxy) GetFromCache(question string, qtype uint16) []dns.RR {
	var cacheFound bool = false
	var cacheResult interface{}

	if qtype == dns.TypeA {
		cacheResult, cacheFound = p.cacheA.Load(question)
	} else {
		cacheResult, cacheFound = p.cacheAAAA.Load(question)
	}

	result := make([]dns.RR, 0)
	if cacheFound {
		rrList := cacheResult.([]string)
		for _, v := range rrList {
			if qtype == dns.TypeA {
				answer, err := dns.NewRR(fmt.Sprintf("%s A %s", question, v))
				if err == nil {
					result = append(result, answer)
				}
			} else {
				answer, err := dns.NewRR(fmt.Sprintf("%s AAAA %s", question, v))
				if err == nil {
					result = append(result, answer)
				}
			}
		}
	}
	return result
}
