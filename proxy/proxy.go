package proxy

import (
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
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
	//logrus.Warn("Query: ", question.Name, serverList)

	switch question.Qtype {
	case dns.TypeA, dns.TypeAAAA:
		{
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
				addrList := p.DecodeTypeAResponse(question.Name, tResp.Answer)
				//TODO: IP Reputation and GeoIP validation
				for k, v := range addrList {
					result := p.geo.Lookup(v)
					logrus.Warn("[", k, question.Name, "]:", result)
				}

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
