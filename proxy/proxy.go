package proxy

import (
	"fmt"
	"os"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/zartbot/zadns/geoip"

	"github.com/olekukonko/tablewriter"
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
			p.TableRender(question.Name, vaddList)

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

func (p *Proxy) TableRender(name string, addrList map[string]string) {
	fmt.Printf("[%s] DNS Lookup Result\n\n", name)

	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Addresss ", "ASN", "City", "Region", "Country", "Location", "Distance(KM)", "DNS Server"})
	table.SetAutoFormatHeaders(false)

	for k, v := range addrList {
		result := p.geo.Lookup(k)
		distance := geoip.ComputeDistance(31.02, 121.26, result.Latitude, result.Longitude)
		table.Append([]string{k, fmt.Sprintf("%-30.30s", result.SPName), fmt.Sprintf("%-16.16s", result.City), fmt.Sprintf("%-16.16s", result.Region), fmt.Sprintf("%-16.16s", result.Country), fmt.Sprintf("%6.2f , %6.2f", result.Latitude, result.Longitude), fmt.Sprintf("%8.0f", distance), v})
	}
	table.Render()
}
