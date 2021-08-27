package proxy

import (
	"fmt"
	"net"
	"os"
	"sort"
	"sync/atomic"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/zartbot/zadns/geoip"
)

type ProbeMetric struct {
	DestIP   string
	Duration time.Duration
}

type ProbeCacheEntry struct {
	Latency  time.Duration
	ProbeCnt uint64
	LossCnt  uint64
}

func (p *Proxy) TCPPing(dip string, dport uint16) {
	m := &ProbeMetric{
		DestIP: dip,
	}
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", dip, dport), time.Second*1)
	if err != nil {
		atomic.AddUint64(p.ProbeWorkerCnt, ^uint64(0))
		m.Duration = time.Second * 120
		p.ProbeChan <- m
		return
	}
	conn.Close()
	m.Duration = time.Since(start)
	p.ProbeChan <- m
}

func (p *Proxy) tcpProbe() {
	hosts := p.probeCache.GetKeys()
	for _, host := range hosts {
		for atomic.LoadUint64(p.ProbeWorkerCnt) > p.MaxProberWorker {
			time.Sleep(20 * time.Millisecond)
		}
		atomic.AddUint64(p.ProbeWorkerCnt, 1)
		go p.TCPPing(host, p.ProbeDport)
	}
}

func (p *Proxy) probeStats() {
	for {
		resp := <-p.ProbeChan
		v := p.probeCache.Get(resp.DestIP)
		if v != nil {
			cacheEntry := v.(*ProbeCacheEntry)
			newProbeCache := &ProbeCacheEntry{
				Latency:  resp.Duration,
				ProbeCnt: cacheEntry.ProbeCnt + 1,
			}

			if resp.Duration > time.Second*5 {
				newProbeCache.LossCnt = cacheEntry.LossCnt + 1
			}
			p.probeCache.Set(resp.DestIP, newProbeCache)
		}
	}
}

func (p *Proxy) TCPProbe() {
	logrus.Info("Starting TCP Probe...")
	go p.probeStats()
	for {
		go p.tcpProbe()
		time.Sleep(p.ProbeFrequency)
	}
}

type sortRecord struct {
	Latency time.Duration
	Address string
}
type sortList []sortRecord

func (a sortList) Len() int           { return len(a) }
func (a sortList) Less(i, j int) bool { return a[i].Latency < a[j].Latency }
func (a sortList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (p *Proxy) SortByLatency(addrList []string) []string {
	resultList := make([]sortRecord, 0)
	for _, v := range addrList {
		cacheEntryT := p.probeCache.Get(v)
		if cacheEntryT != nil {
			cacheEntry := cacheEntryT.(*ProbeCacheEntry)
			record := &sortRecord{
				Address: v,
				Latency: cacheEntry.Latency,
			}
			resultList = append(resultList, *record)
		}
	}
	sort.Sort(sortList(resultList))

	result := make([]string, 0)
	for i := 0; i < len(resultList); i++ {
		result = append(result, resultList[i].Address)
	}
	return result
}

func (p *Proxy) DumpCacheTable() {
	keys := p.probeCache.GetKeys()
	p.RenderProbeCacheList(keys)

}

func (p *Proxy) RenderProbeCacheList(s []string) {
	if len(s) == 0 {
		return
	}
	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Addresss ", "SP", "ASN", "City", "Region", "Country", "Location", "Distance(KM)", "TCP Latency", "Loss"})
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	//table.SetRowLine(true)

	for _, v := range s {
		result := p.geo.Lookup(v)
		latency := time.Duration(0)
		loss := float32(0)
		metric := p.probeCache.Get(v)
		if metric != nil {
			latency = metric.(*ProbeCacheEntry).Latency
			loss = float32(metric.(*ProbeCacheEntry).LossCnt) * 100 / float32(metric.(*ProbeCacheEntry).ProbeCnt)

		}

		spStr := fmt.Sprintf("%-40.40s", result.SPName)
		asnStr := fmt.Sprintf("%d", result.ASN)
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
		latencyStr := fmt.Sprintf("%8dms", latency.Milliseconds())

		table.Append([]string{v, spStr, asnStr, cityStr, regionStr, countryStr, geoStr, distanceStr, latencyStr, fmt.Sprintf("%6.2f%%", loss)})
	}
	table.Render()
}
