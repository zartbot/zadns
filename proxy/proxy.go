package proxy

import (
	"fmt"
	"os"
	"time"

	"github.com/miekg/dns"
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

	switch question.Qtype {
	case dns.TypeA, dns.TypeAAAA:
		{
			cacheList := p.GetFromCache(question.Name, question.Qtype)
			bestList := p.SortByLatency(cacheList)

			if len(bestList) > 0 {
				if p.LogLevel == "debug" {
					p.RenderProbeCacheList(bestList)
				}
				if len(bestList) > p.BestRecordNum {
					answer := BuildRR(bestList[:p.BestRecordNum], question.Name, question.Qtype)
					resp.Answer = append(resp.Answer, answer...)
				} else {
					answer := BuildRR(bestList, question.Name, question.Qtype)
					resp.Answer = append(resp.Answer, answer...)
				}
			} else {
				//check external Server
				records := p.MultipleLookup(req, serverList)
				if p.LogLevel == "debug" {
					p.TableRender(question.Name, records)
				} //TODO: IP Reputation and GeoIP validation

				addrList := make([]string, 0)
				for k, _ := range records {
					cacheEntry := &ProbeCacheEntry{
						Latency: time.Second * 120,
					}
					isNew := p.probeCache.Update(k, cacheEntry)
					if isNew {
						go p.TCPPing(k, p.ProbeDport)
					}
					addrList = append(addrList, k)
				}
				resp.Answer = BuildRR(addrList, question.Name, question.Qtype)
				if len(addrList) > 0 {
					if question.Qtype == dns.TypeA {
						p.typeACache.Store(question.Name, addrList, time.Now())
					} else {
						p.typeAAAACache.Store(question.Name, addrList, time.Now())
					}
				}
				return resp, nil
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

	if len(addrList) == 0 {
		return
	}
	fmt.Printf("[%s] DNS Lookup Result\n\n", name)

	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Addresss ", "ASN", "City", "Region", "Country", "Location", "Distance(KM)", "DNS Server"})
	table.SetAutoFormatHeaders(false)

	for k, v := range addrList {
		result := p.geo.Lookup(k)

		spStr := fmt.Sprintf("%-30.30s", result.SPName)
		cityStr := fmt.Sprintf("%-16.16s", result.City)
		regionStr := fmt.Sprintf("%-16.16s", result.Region)
		countryStr := fmt.Sprintf("%-16.16s", result.Country)
		geoStr := fmt.Sprintf("%6.2f , %6.2f", result.Latitude, result.Longitude)

		distance := geoip.ComputeDistance(p.Latitude, p.Longitude, result.Latitude, result.Longitude)

		latencyByDistance := distance / 75

		/*
		  LightSpeed over Fiber is nearly 150,000km/s
		  RTT(ms) = distance *2 / Fiber_LightSpeed *1000 = 2 * distance /150,000 * 1000 = distance /100
		  Each hop contribute 3ms latency,based on average QoS and forwarding latency estimation
		*/
		distanceStr := fmt.Sprintf("%6.0fkm[%3.0fms]", distance, latencyByDistance)
		if result.Latitude == 0 && result.Longitude == 0 {
			distanceStr = fmt.Sprintf("%12s", "")
		}

		table.Append([]string{k, spStr, cityStr, regionStr, countryStr, geoStr, distanceStr, v})
	}
	table.Render()
}
