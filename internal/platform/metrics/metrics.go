package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Registry struct {
	service       string
	startedAtUnix int64
	requestsTotal atomic.Int64
	inflight      atomic.Int64

	mu              sync.RWMutex
	statusCounters  map[int]*atomic.Int64
	latencyBuckets  []time.Duration
	latencyCounters []*atomic.Int64
}

func NewRegistry(service string) *Registry {
	buckets := []time.Duration{
		10 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		250 * time.Millisecond,
		500 * time.Millisecond,
		time.Second,
		2 * time.Second,
		5 * time.Second,
	}
	counters := make([]*atomic.Int64, len(buckets)+1)
	for i := range counters {
		counters[i] = &atomic.Int64{}
	}
	return &Registry{
		service:         service,
		startedAtUnix:   time.Now().Unix(),
		statusCounters:  map[int]*atomic.Int64{},
		latencyBuckets:  buckets,
		latencyCounters: counters,
	}
}

func (r *Registry) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/metrics" {
			next.ServeHTTP(w, req)
			return
		}

		start := time.Now()
		r.inflight.Add(1)
		defer r.inflight.Add(-1)

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, req)

		r.requestsTotal.Add(1)
		r.recordStatus(recorder.status)
		r.recordLatency(time.Since(start))
	})
}

func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(r.Render()))
	})
}

func (r *Registry) Render() string {
	metricPrefix := sanitizeMetricName(r.service)
	var builder strings.Builder

	fmt.Fprintf(&builder, "# HELP %s_http_requests_total Total HTTP requests handled by the service.\n", metricPrefix)
	fmt.Fprintf(&builder, "# TYPE %s_http_requests_total counter\n", metricPrefix)
	fmt.Fprintf(&builder, "%s_http_requests_total %d\n", metricPrefix, r.requestsTotal.Load())

	fmt.Fprintf(&builder, "# HELP %s_http_inflight_requests Current number of inflight HTTP requests.\n", metricPrefix)
	fmt.Fprintf(&builder, "# TYPE %s_http_inflight_requests gauge\n", metricPrefix)
	fmt.Fprintf(&builder, "%s_http_inflight_requests %d\n", metricPrefix, r.inflight.Load())

	fmt.Fprintf(&builder, "# HELP %s_process_start_time_seconds Process start time in unix seconds.\n", metricPrefix)
	fmt.Fprintf(&builder, "# TYPE %s_process_start_time_seconds gauge\n", metricPrefix)
	fmt.Fprintf(&builder, "%s_process_start_time_seconds %d\n", metricPrefix, r.startedAtUnix)

	fmt.Fprintf(&builder, "# HELP %s_http_response_status_total HTTP responses by status code.\n", metricPrefix)
	fmt.Fprintf(&builder, "# TYPE %s_http_response_status_total counter\n", metricPrefix)
	for _, status := range r.statuses() {
		counter := r.statusCounter(status)
		fmt.Fprintf(&builder, "%s_http_response_status_total{code=\"%d\"} %d\n", metricPrefix, status, counter.Load())
	}

	fmt.Fprintf(&builder, "# HELP %s_http_request_duration_bucket HTTP request latency buckets in milliseconds.\n", metricPrefix)
	fmt.Fprintf(&builder, "# TYPE %s_http_request_duration_bucket histogram\n", metricPrefix)
	var cumulative int64
	for idx, bucket := range r.latencyBuckets {
		count := r.latencyCounters[idx].Load()
		cumulative += count
		fmt.Fprintf(&builder, "%s_http_request_duration_bucket{le=\"%d\"} %d\n", metricPrefix, bucket.Milliseconds(), cumulative)
	}
	cumulative += r.latencyCounters[len(r.latencyCounters)-1].Load()
	fmt.Fprintf(&builder, "%s_http_request_duration_bucket{le=\"+Inf\"} %d\n", metricPrefix, cumulative)
	fmt.Fprintf(&builder, "%s_http_request_duration_count %d\n", metricPrefix, cumulative)

	return builder.String()
}

func (r *Registry) recordStatus(status int) {
	r.statusCounter(status).Add(1)
}

func (r *Registry) statusCounter(status int) *atomic.Int64 {
	r.mu.RLock()
	counter := r.statusCounters[status]
	r.mu.RUnlock()
	if counter != nil {
		return counter
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	counter = r.statusCounters[status]
	if counter == nil {
		counter = &atomic.Int64{}
		r.statusCounters[status] = counter
	}
	return counter
}

func (r *Registry) statuses() []int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	statuses := make([]int, 0, len(r.statusCounters))
	for status := range r.statusCounters {
		statuses = append(statuses, status)
	}
	sort.Ints(statuses)
	return statuses
}

func (r *Registry) recordLatency(duration time.Duration) {
	for idx, bucket := range r.latencyBuckets {
		if duration <= bucket {
			r.latencyCounters[idx].Add(1)
			return
		}
	}
	r.latencyCounters[len(r.latencyCounters)-1].Add(1)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func sanitizeMetricName(value string) string {
	value = strings.ReplaceAll(value, "-", "_")
	var builder strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			builder.WriteRune(r)
		}
	}
	sanitized := builder.String()
	if sanitized == "" {
		return "service"
	}
	if sanitized[0] >= '0' && sanitized[0] <= '9' {
		return "service_" + sanitized
	}
	return sanitized
}

func MustParseInt64(value string) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		panic(err)
	}
	return parsed
}
