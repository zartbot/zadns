package proxy

import (
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

func DecodeTypeAResponse(answer []dns.RR) []string {
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

func (p *Proxy) GetFromCache(question string, qtype uint16) []string {
	var cacheFound bool = false
	var cacheResult interface{}

	if qtype == dns.TypeA {
		cacheResult, cacheFound = p.typeACache.Load(question)
	} else {
		cacheResult, cacheFound = p.typeAAAACache.Load(question)
	}

	if cacheFound {
		return cacheResult.([]string)
	}
	return make([]string, 0)
}

func BuildRR(rrList []string, question string, qtype uint16) []dns.RR {
	result := make([]dns.RR, 0)

	for _, v := range rrList {
		if qtype == dns.TypeA {
			answer, err := dns.NewRR(fmt.Sprintf("%s A %s", question, v))
			answer.Header().Ttl = 0
			if err == nil {
				result = append(result, answer)
			}
		} else {
			answer, err := dns.NewRR(fmt.Sprintf("%s AAAA %s", question, v))
			answer.Header().Ttl = 0
			if err == nil {
				result = append(result, answer)
			}
		}
	}
	return result
}
