package bond_coredns

import (
	"sync"

	"github.com/coredns/coredns/plugin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var requestCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: plugin.Namespace,
	Subsystem: "bond-coredns",
	Name:      "request_count_total",
	Help:      "Counter of requests made.",
}, []string{"server"})

var once sync.Once
