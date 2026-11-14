package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	tunerReady = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_ready",
		Help: "tuner server accepting connections",
	})
)
