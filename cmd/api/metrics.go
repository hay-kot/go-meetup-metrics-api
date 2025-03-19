package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Define Prometheus metrics
	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"endpoint", "method"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_request_duration_seconds",
			Help:    "API request duration in seconds",
			Buckets: prometheus.LinearBuckets(0.01, 0.05, 10), // 10 buckets from 0.01s to 0.5s
		},
		[]string{"endpoint"},
	)

	inFlightRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "api_in_flight_requests",
			Help: "Current number of in-flight requests",
		},
		[]string{"endpoint"},
	)
)
