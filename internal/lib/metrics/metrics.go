package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

const (
	UploadMethod = "upload"
	GetMethod    = "get"
	ListMethod   = "list"
)

var (
	requestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "tage",
		Subsystem: "test",
		Name:      "requests",
	}, []string{"type"})
)

var (
	h *Prom
)

func init() {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		requestCounter,
		collectors.NewGoCollector(),
	)

	h = &Prom{
		registry: reg,
	}
}

type Prom struct {
	registry *prometheus.Registry
}

func GetProm() *Prom {
	return h
}

func (h *Prom) GetHandler() http.Handler {
	return promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{})
}

func (h *Prom) IncRequest(name string) {
	requestCounter.WithLabelValues(name).Inc()
}
