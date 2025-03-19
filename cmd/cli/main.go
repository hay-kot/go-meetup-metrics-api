package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Stats to track requests
type Stats struct {
	totalRequests      uint64
	successfulRequests uint64
	failedRequests     uint64
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	err := run()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run load generator")
	}
}

func run() error {
	// Define command line flags
	host := flag.String("host", "localhost", "Target server host")
	port := flag.Int("port", 8080, "Target server port")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent workers")
	verbose := flag.Bool("verbose", false, "Enable verbose logging of each request")
	flag.Parse()

	// Initialize stats
	stats := &Stats{}

	// Setup base URL and available endpoints
	baseURL := fmt.Sprintf("http://%s:%d", *host, *port)
	endpoints := []string{
		"/api/endpoint-a",
		"/api/endpoint-b",
		"/api/endpoint-c",
	}

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS interrupts (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		log.Info().Msg("interrupt received, stopping load generator...")
		cancel()
	}()

	log.Info().
		Str("host", *host).
		Int("port", *port).
		Int("concurrency", *concurrency).
		Msg("starting load generator (press Ctrl+C to stop)")

	// Use a WaitGroup to manage concurrent workers
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := range *concurrency {
		wg.Add(1)
		go worker(ctx, &wg, i, baseURL, endpoints, stats, *verbose)
	}

	// Start a goroutine to print stats periodically
	go statsReporter(ctx, stats)

	// Wait for all workers to complete
	wg.Wait()

	// Print final stats
	printStats(stats)

	return nil
}

func worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	id int,
	baseURL string,
	endpoints []string,
	stats *Stats,
	verbose bool,
) {
	defer wg.Done()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	log.Info().Int("worker_id", id).Msg("worker started")

	// Run until context is canceled
	for {
		select {
		case <-ctx.Done():
			log.Info().Int("worker_id", id).Msg("worker stopped")
			return
		default:
			// Choose a random endpoint
			endpoint := endpoints[rand.Intn(len(endpoints))]
			url := baseURL + endpoint

			// Make the request
			start := time.Now()
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				atomic.AddUint64(&stats.failedRequests, 1)
				log.Error().Err(err).Str("url", url).Msg("failed to create request")
				continue
			}

			resp, err := client.Do(req)
			duration := time.Since(start)

			// Track statistics
			atomic.AddUint64(&stats.totalRequests, 1)

			if err != nil {
				atomic.AddUint64(&stats.failedRequests, 1)
				if verbose {
					log.Error().Err(err).Str("url", url).Msg("request failed")
				}
				continue
			}

			// Ensure we close the response body
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()

			// Check status code
			if resp.StatusCode != http.StatusOK {
				atomic.AddUint64(&stats.failedRequests, 1)
				if verbose {
					log.Error().
						Int("status", resp.StatusCode).
						Str("url", url).
						Msg("received non-OK response")
				}
			} else {
				atomic.AddUint64(&stats.successfulRequests, 1)
				if verbose {
					log.Info().
						Str("url", url).
						Dur("duration", duration).
						Msg("request completed")
				}
			}

			// Small sleep to avoid absolute hammering
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		}
	}
}

func statsReporter(ctx context.Context, stats *Stats) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastTotal uint64

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			current := atomic.LoadUint64(&stats.totalRequests)
			rps := float64(current-lastTotal) / 5.0 // 5-second window
			lastTotal = current

			log.Info().
				Uint64("total", current).
				Uint64("success", atomic.LoadUint64(&stats.successfulRequests)).
				Uint64("failures", atomic.LoadUint64(&stats.failedRequests)).
				Float64("requests_per_second", rps).
				Msg("load statistics")
		}
	}
}

func printStats(stats *Stats) {
	total := atomic.LoadUint64(&stats.totalRequests)
	success := atomic.LoadUint64(&stats.successfulRequests)
	failed := atomic.LoadUint64(&stats.failedRequests)

	successRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}

	log.Info().
		Uint64("total_requests", total).
		Uint64("successful_requests", success).
		Uint64("failed_requests", failed).
		Float64("success_rate_percent", successRate).
		Msg("load test completed")
}
