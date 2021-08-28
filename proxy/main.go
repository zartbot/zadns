package proxy

import (
	"math/rand"
	"time"

	"github.com/armon/go-radix"
	"github.com/zartbot/zadns/cache"
	"github.com/zartbot/zadns/geoip"
	"github.com/zartbot/zadns/tsyncmap"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Proxy struct {
	typeAAAACache     *tsyncmap.Map
	typeACache        *tsyncmap.Map
	PTRCache          *tsyncmap.Map
	probeCache        *cache.Cache
	route             map[string][]string
	routeTree         *radix.Tree
	server            []string
	geo               *geoip.GeoIPDB
	ProbeChan         chan *ProbeMetric
	LogLevel          string
	ProbeLFUCacheSize int
	CacheTimeOut      time.Duration
	LookupTimeOut     time.Duration
	ProbeFrequency    time.Duration
	Latitude          float64
	Longitude         float64
	ProbeWorkerCnt    *uint64
	MaxProberWorker   uint64
	ProbeDport        uint16
	BestRecordNum     int
}

func New(routecfg string, hostcfg string, servercfg string, geocfg string, asncfg string) *Proxy {
	px := &Proxy{
		CacheTimeOut:      time.Duration(4 * time.Hour),
		LookupTimeOut:     time.Duration(500 * time.Millisecond),
		probeCache:        cache.New(),
		route:             ReadCfg(routecfg),
		server:            ReadServerListCfg(servercfg),
		geo:               geoip.New(geocfg, asncfg),
		routeTree:         radix.New(),
		ProbeChan:         make(chan *ProbeMetric, 20),
		ProbeLFUCacheSize: 2000,
		ProbeFrequency:    time.Second * 30,
		Latitude:          31.24,
		Longitude:         129.1,
		MaxProberWorker:   20,
		ProbeDport:        443,
		BestRecordNum:     3,
	}

	var cnt uint64 = 0
	px.ProbeWorkerCnt = &cnt

	px.typeACache = tsyncmap.NewMap("dnsCacheIPv4", px.CacheTimeOut, time.Duration(20*time.Second), false)
	go px.typeACache.Run()

	px.typeAAAACache = tsyncmap.NewMap("dnsCacheIPv6", px.CacheTimeOut, time.Duration(20*time.Second), false)
	go px.typeAAAACache.Run()

	px.PTRCache = tsyncmap.NewMap("PTRRecord", px.CacheTimeOut, time.Duration(20*time.Second), false)
	go px.PTRCache.Run()

	//TCP Probe db is a cache based db
	px.probeCache.UpperBound = px.ProbeLFUCacheSize
	px.probeCache.LowerBound = px.ProbeLFUCacheSize - 200

	//update route table to Radix Tree
	px.updateRadixTree()

	//merge local hosts config to local cache
	hosts := ReadCfg(hostcfg)
	for k, v := range hosts {
		px.typeACache.Store(k+".", v, time.Date(2099, 12, 31, 9, 9, 9, 9, time.Local))
		for i := 0; i < len(v); i++ {
			px.PTRCache.Store(v[i], k+".", time.Date(2099, 12, 31, 9, 9, 9, 9, time.Local))
		}
	}

	return px
}
