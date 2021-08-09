package proxy

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/zartbot/zadns/geoip"
)

func (p *Proxy) GetResponse(req *dns.Msg) (*dns.Msg, error) {
	resp := new(dns.Msg)
	resp.SetReply(req)
	if len(req.Question) == 0 {
		return resp, nil
	}
	question := req.Question[0]

	serverList := make([]string, 0)
	//RouteTable
	dbRServer := p.DomainRouteLookup(question.Name)
	if len(dbRServer) > 0 {
		serverList = append(serverList, dbRServer...)
	} else {
		serverList = append(serverList, p.server...)
	}
	logrus.Warn("Query: ", question.Name, serverList)

	switch question.Qtype {
	case dns.TypeA, dns.TypeAAAA:
		{
			//DEBUG multilookup
			vaddList := p.MultipleLookup(req, serverList)
			logrus.Warn("LB-Lookup:", vaddList)
			for _, v := range vaddList {
				result := p.geo.Lookup(v)
				distance := geoip.ComputeDistance(31.02, 121.26, result.Latitude, result.Longitude)

				fmt.Printf("LB-Lookup: %20s | %24s |ASN: %-30.30s | City: %-16.16s Region: %-20.20s Country: %-16.16s | Location: %10f,%10f Distance: %10f\n", question.Name, v, result.SPName, result.City, result.Region, result.Country, result.Latitude, result.Longitude, distance)
			}

			answer := p.GetFromCache(question.Name, question.Qtype)
			if len(answer) > 0 {
				resp.Answer = append(resp.Answer, answer...)
			} else {

				//check external Server

				tResp, err := p.RandomLookup(req, serverList)
				if err != nil {
					return resp, err
				}
				//add to cache

				addrList := DecodeTypeAResponse(tResp.Answer)

				//TODO: IP Reputation and GeoIP validation
				/*
					for k, v := range addrList {
						result := p.geo.Lookup(v)
						logrus.Warn("[", k, "|", v, "|", question.Name, "]:", result)
					}*/

				if len(addrList) > 0 {
					if question.Qtype == dns.TypeA {
						p.cacheA.Store(question.Name, addrList, time.Now())
					} else {
						p.cacheAAAA.Store(question.Name, addrList, time.Now())
					}
				}
				return tResp, nil
			}
		}
	default:
		{
			resp, err := p.RandomLookup(req, serverList)
			return resp, err
		}
	}
	return resp, nil
}
