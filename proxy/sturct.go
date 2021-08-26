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
	cacheA     *tsyncmap.Map
	cacheAAAA  *tsyncmap.Map
	probeCache *cache.Cache
	route      map[string][]string
	routeTree  *radix.Tree
	server     []string
	geo        *geoip.GeoIPDB
	ProbeChan  chan *ProbeMetric
}

func New(routecfg string, hostcfg string, servercfg string, geocfg string, asncfg string) *Proxy {
	cacheTimeout := time.Duration(4 * time.Hour)
	checkFreq := time.Duration(10 * time.Second)
	px := &Proxy{
		cacheA:     tsyncmap.NewMap("dnsCacheIPv4", cacheTimeout, checkFreq, false),
		cacheAAAA:  tsyncmap.NewMap("dnsCacheIPv6", cacheTimeout, checkFreq, false),
		probeCache: cache.New(),
		route:      ReadCfg(routecfg),
		server:     ReadServerListCfg(servercfg),
		geo:        geoip.New(geocfg, asncfg),
		routeTree:  radix.New(),
		ProbeChan:  make(chan *ProbeMetric, 20),
	}

	//TCP Probe db is a cache based db
	px.probeCache.UpperBound = 2000
	px.probeCache.LowerBound = 1800

	//update route table to Radix Tree
	px.updateRadixTree()

	go px.cacheA.Run()
	go px.cacheAAAA.Run()
	//merge local hosts config to local cache
	hosts := ReadCfg(hostcfg)
	for k, v := range hosts {
		px.cacheA.Store(k+".", v, time.Date(2099, 12, 31, 9, 9, 9, 9, time.Local))
	}

	return px
}
