# go-meetup-metrics-api

Example of a metrics insturmented API

## Prometheus Metric Types Explained

### 1. Counter

A **Counter** is a cumulative metric that only increases or resets to zero on restart.

**Real-World Use Cases:**

- Request count per API endpoint
- Error count by type
- Total bytes processed
- Completed job count

**Example in Go:**

```go
requestCounter = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "api_requests_total",
        Help: "Total number of API requests",
    },
    []string{"endpoint", "method"},
)

// Usage
requestCounter.WithLabelValues("endpoint-a", "GET").Inc()
```

**Key Query:** `rate(api_requests_total[5m])` - Request rate over 5 minutes

### 2. Gauge

A **Gauge** represents a value that can arbitrarily increase or decrease.

**Real-World Use Cases:**

- Current memory usage
- Active connections
- Queue size
- CPU utilization
- Temperature readings

**Example in Go:**

```go
inFlightRequests = promauto.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "api_in_flight_requests",
        Help: "Current number of in-flight requests",
    },
    []string{"endpoint"},
)

// Usage
inFlightRequests.WithLabelValues("endpoint-a").Inc()
inFlightRequests.WithLabelValues("endpoint-a").Dec()
```

**Key Query:** `max_over_time(api_in_flight_requests[1h])` - Maximum value over an hour

### 3. Histogram

A **Histogram** tracks the distribution of values across configurable buckets.

**Real-World Use Cases:**

- API response time distribution
- Request payload size distribution
- Database query time
- File processing duration

**Example in Go:**

```go
requestDuration = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "api_request_duration_seconds",
        Help:    "API request duration in seconds",
        Buckets: prometheus.LinearBuckets(0.01, 0.05, 10), // 10 buckets from 0.01s to 0.5s
    },
    []string{"endpoint"},
)

// Usage
duration := time.Since(startTime).Seconds()
requestDuration.WithLabelValues("endpoint-a").Observe(duration)
```

**Key Query:** `histogram_quantile(0.95, sum(rate(api_request_duration_seconds_bucket[5m])) by (le, endpoint))` - 95th percentile response time

### 4. Summary

A **Summary** is similar to a histogram but pre-calculates quantiles in the application.

**Real-World Use Cases:**

- Response time monitoring with precise quantiles
- Resource usage distribution
- Any case where pre-calculated quantiles are preferred over bucket-based approximations

**Example in Go:**

```go
requestDurationSummary = promauto.NewSummaryVec(
    prometheus.SummaryOpts{
        Name:       "api_request_duration_summary_seconds",
        Help:       "API request duration in seconds (summary)",
        Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
    },
    []string{"endpoint"},
)
```

**Key Query:** `api_request_duration_summary_seconds{quantile="0.99"}` - Direct 99th percentile reading

## Histogram vs. Summary: Quick Comparison

| Feature     | Histogram                           | Summary                            |
| ----------- | ----------------------------------- | ---------------------------------- |
| Calculation | Query time                          | Application                        |
| Aggregation | Possible across instances           | Not possible for quantiles         |
| Flexibility | Any quantile                        | Only pre-configured                |
| Best for    | General monitoring with aggregation | Precise quantiles, single instance |
