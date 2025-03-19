package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	err := run()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run program")
	}
}

func run() error {
	// Define command line flags for host and port
	host := flag.String("host", "localhost", "Server host")
	port := flag.Int("port", 8080, "Server port")
	flag.Parse()

	// Setup HTTP routes
	http.HandleFunc("/api/endpoint-a", handleEndpoint("endpoint-a"))
	http.HandleFunc("/api/endpoint-b", handleEndpoint("endpoint-b"))
	http.HandleFunc("/api/endpoint-c", handleEndpoint("endpoint-c"))

	// Expose Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())
	log.Info().Msg("prometheus metrics enabled at /metrics")

	// Start the server
	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Info().Str("address", addr).Msg("starting server")
	return http.ListenAndServe(addr, nil)
}

type Response struct {
	Endpoint string `json:"endpoint"`
}

func handleEndpoint(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Track in-flight requests
		inFlightRequests.WithLabelValues(name).Inc()
		defer inFlightRequests.WithLabelValues(name).Dec()

		// Increment request counter metric
		requestCounter.WithLabelValues(name, r.Method).Inc()

		// Log incoming request
		log.Info().Str("endpoint", name).Str("method", r.Method).Str("path", r.URL.Path).Msg("request received")

		// Wait for a random time between 0-500ms
		waitTime := time.Duration(rand.Intn(500)) * time.Millisecond
		log.Debug().Dur("wait_time", waitTime).Str("endpoint", name).Msg("waiting before response")
		time.Sleep(waitTime)

		// Prepare response
		resp := Response{Endpoint: name}

		// Set content type and encode response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error().Err(err).Str("endpoint", name).Msg("failed to encode response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Observe request duration metric
		duration := time.Since(startTime).Seconds()
		requestDuration.WithLabelValues(name).Observe(duration)

		log.Info().Str("endpoint", name).Dur("response_time", waitTime).Float64("duration_seconds", duration).Msg("response sent")
	}
}
