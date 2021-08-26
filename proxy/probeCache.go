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

type CacheMetirc struct {
	Latency  time.Duration
	ProbeCnt uint64
	LossCnt  uint64
}

func (p *Proxy) TCPPing(dip string, dport uint16, workerCnt *uint64) {
	m := &ProbeMetric{
		DestIP: dip,
	}
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", dip, dport), time.Second*2)
	if err != nil {
		atomic.AddUint64(workerCnt, ^uint64(0))
		m.Duration = time.Second * 100
		p.ProbeChan <- m
		return
	}
	conn.Close()
	m.Duration = time.Since(start)
	p.ProbeChan <- m
}

func (p *Proxy) tcpProbe() {
	var workerCnt uint64 = 0
	hosts := p.probeCache.GetKeys()
	for _, host := range hosts {
		for atomic.LoadUint64(&workerCnt) > 20 {
			time.Sleep(200 * time.Millisecond)
		}
		atomic.AddUint64(&workerCnt, 1)
		go p.TCPPing(host, 443, &workerCnt)
	}
}

func (p *Proxy) probeStats() {
	for {
		resp := <-p.ProbeChan
		v := p.probeCache.Get(resp.DestIP)
		if v != nil {
			cachedMetric := v.(*CacheMetirc)
			newMetric := &CacheMetirc{
				Latency:  resp.Duration,
				ProbeCnt: cachedMetric.ProbeCnt + 1,
			}

			if resp.Duration > time.Second*5 {
				newMetric.LossCnt = cachedMetric.LossCnt + 1
			}
			p.probeCache.Set(resp.DestIP, newMetric)
		}
	}
}

func (p *Proxy) TCPProbe(freq time.Duration) {
	logrus.Info("Starting TCP Probe...")
	go p.probeStats()
	for {
		go p.tcpProbe()
		time.Sleep(freq)
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

func (p *Proxy) GetFastResult(addrList []string) []string {
	metricList := make([]sortRecord, 0)
	for _, v := range addrList {
		metric := p.probeCache.Get(v)
		if metric != nil {

			m := metric.(*CacheMetirc)
			record := &sortRecord{
				Address: v,
				Latency: m.Latency,
			}
			metricList = append(metricList, *record)
		}
	}
	sort.Sort(sortList(metricList))
	//DEBUG
	go p.PrintCacheByList(metricList)

	result := make([]string, 0)
	for i := 0; i < len(metricList); i++ {
		result = append(result, metricList[i].Address)
		// only return 3 fastest server to client
		if i == 2 {
			break
		}
	}
	return result
}

func (p *Proxy) PrintProbeCacheTable() {
	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Addresss ", "ASN", "City", "Region", "Country", "Location", "Distance(KM)", "TCP Latency", "Loss"})
	table.SetAutoFormatHeaders(false)

	keys := p.probeCache.GetKeys()
	for _, v := range keys {
		result := p.geo.Lookup(v)
		latency := time.Duration(0)
		loss := float32(0)
		metric := p.probeCache.Get(v)
		if metric != nil {
			latency = metric.(*CacheMetirc).Latency
			loss = float32(metric.(*CacheMetirc).LossCnt) * 100 / float32(metric.(*CacheMetirc).ProbeCnt)

		}

		distance := geoip.ComputeDistance(31.02, 121.26, result.Latitude, result.Longitude)
		table.Append([]string{v, fmt.Sprintf("%-30.30s", result.SPName), fmt.Sprintf("%-16.16s", result.City), fmt.Sprintf("%-16.16s", result.Region), fmt.Sprintf("%-16.16s", result.Country), fmt.Sprintf("%6.2f , %6.2f", result.Latitude, result.Longitude), fmt.Sprintf("%8.0f", distance), fmt.Sprintf("%12s", latency.String()), fmt.Sprintf("%6.2f%%", loss)})
	}
	table.Render()
}

func (p *Proxy) PrintCacheByList(s sortList) {
	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Addresss ", "ASN", "City", "Region", "Country", "Location", "Distance(KM)", "TCP Latency"})
	table.SetAutoFormatHeaders(false)

	for _, v := range s {
		result := p.geo.Lookup(v.Address)
		latency := time.Duration(0)
		loss := float32(0)
		metric := p.probeCache.Get(v.Address)
		if metric != nil {
			latency = metric.(*CacheMetirc).Latency
			loss = float32(metric.(*CacheMetirc).LossCnt) * 100 / float32(metric.(*CacheMetirc).ProbeCnt)

		}

		distance := geoip.ComputeDistance(31.02, 121.26, result.Latitude, result.Longitude)
		table.Append([]string{v.Address, fmt.Sprintf("%-30.30s", result.SPName), fmt.Sprintf("%-16.16s", result.City), fmt.Sprintf("%-16.16s", result.Region), fmt.Sprintf("%-16.16s", result.Country), fmt.Sprintf("%6.2f , %6.2f", result.Latitude, result.Longitude), fmt.Sprintf("%8.0f", distance), fmt.Sprintf("%12s", latency.String()), fmt.Sprintf("%6.2f%%", loss)})
	}
	table.Render()
}
