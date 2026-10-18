// Package sched provides a hardened worker pool, process admission governor,
// aging-aware critical path scheduler, and zero-dependency telemetry engine.
//
// Key architectural guarantees:
//   1. Subprocess Token Bucket: Enforces maximum spawn rates and bursts to prevent
//      kernel process table exhaustion / fork storms during wide fan-outs.
//   2. Memory Governor: Implements weighted semaphore admission control based on
//      cgroup v2, cgroup v1, or /proc/meminfo to prevent Out-Of-Memory (OOM) kills.
//   3. Aging-Aware Priority Heap: Dynamically boosts critical-path priority over time,
//      guaranteeing zero starvation for low-priority tasks under heavy load.
//   4. Prometheus & OTLP Telemetry: Exposes live Prometheus exposition format
//      metrics and spec-compliant OTLP/JSON using only Go's standard library.
package sched

import (
	"container/heap"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// TOKEN BUCKET SPAWN-RATE GOVERNOR
// ============================================================================

type TokenBucket struct {
	mu     sync.Mutex
	tokens float64
	rate   float64 // tokens per second
	burst  float64
	last   time.Time
}

func NewTokenBucket(rate, burst float64) *TokenBucket {
	if rate <= 0 {
		rate = 50.0
	}
	if burst <= 0 {
		burst = 200.0
	}
	return &TokenBucket{
		tokens: burst,
		rate:   rate,
		burst:  burst,
		last:   time.Now(),
	}
}

func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		tb.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(tb.last).Seconds()
		tb.tokens = tb.tokens + elapsed*tb.rate
		if tb.tokens > tb.burst {
			tb.tokens = tb.burst
		}
		tb.last = now

		if tb.tokens >= 1.0 {
			tb.tokens -= 1.0
			tb.mu.Unlock()
			return nil
		}

		deficit := (1.0 - tb.tokens) / tb.rate
		tb.mu.Unlock()

		delay := time.Duration(deficit * float64(time.Second))
		if delay < time.Millisecond {
			delay = time.Millisecond
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}

// ============================================================================
// MEMORY GOVERNOR (CGROUP & MEMINFO DETECTION)
// ============================================================================

func DetectMemoryBudget() int64 {
	// 1. Linux cgroup v2
	if b, err := os.ReadFile("/sys/fs/cgroup/memory.max"); err == nil {
		s := strings.TrimSpace(string(b))
		if s != "max" {
			if v, err := strconv.ParseInt(s, 10, 64); err == nil && v > 0 {
				return v
			}
		}
	}

	// 2. Linux cgroup v1
	if b, err := os.ReadFile("/sys/fs/cgroup/memory/memory.limit_in_bytes"); err == nil {
		s := strings.TrimSpace(string(b))
		if v, err := strconv.ParseInt(s, 10, 64); err == nil && v > 0 && v < (1<<50) {
			return v
		}
	}

	// 3. /proc/meminfo fallback
	if b, err := os.ReadFile("/proc/meminfo"); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if v, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
						return v * 1024 // kB to bytes
					}
				}
			}
		}
	}

	// Default fallback: 4 GiB heuristic
	return 4 * 1024 * 1024 * 1024
}

type MemoryGovernor struct {
	granule int64
	tokens  chan struct{}
}

func NewMemoryGovernor(budgetBytes int64) *MemoryGovernor {
	const granule = 16 * 1024 * 1024 // 16 MiB granules
	if budgetBytes <= 0 {
		budgetBytes = 2 * 1024 * 1024 * 1024 // 2 GiB default
	}
	n := budgetBytes / granule
	if n < 1 {
		n = 1
	}
	return &MemoryGovernor{
		granule: granule,
		tokens:  make(chan struct{}, n),
	}
}

func (g *MemoryGovernor) Acquire(ctx context.Context, estBytes int64) (release func(), err error) {
	if estBytes <= 0 {
		estBytes = 256 * 1024 * 1024 // 256 MiB default task estimate
	}
	n := (estBytes + g.granule - 1) / g.granule
	if n < 1 {
		n = 1
	}
	if n > int64(cap(g.tokens)) {
		n = int64(cap(g.tokens))
	}

	for i := int64(0); i < n; i++ {
		select {
		case g.tokens <- struct{}{}:
		case <-ctx.Done():
			for j := int64(0); j < i; j++ {
				<-g.tokens
			}
			return nil, ctx.Err()
		}
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			for i := int64(0); i < n; i++ {
				<-g.tokens
			}
		})
	}, nil
}

// ============================================================================
// HISTOGRAM & TELEMETRY REGISTRY
// ============================================================================

type Histogram struct {
	name   string
	bounds []float64
	counts []atomic.Int64
	count  atomic.Int64
	sumMs  atomic.Int64
}

func NewHistogram(name string, bounds []float64) *Histogram {
	return &Histogram{
		name:   name,
		bounds: bounds,
		counts: make([]atomic.Int64, len(bounds)+1),
	}
}

func (h *Histogram) Observe(v float64) {
	h.count.Add(1)
	h.sumMs.Add(int64(v))
	for i, b := range h.bounds {
		if v <= b {
			h.counts[i].Add(1)
			return
		}
	}
	h.counts[len(h.bounds)].Add(1)
}

type Metrics struct {
	TasksSubmitted  atomic.Int64
	TasksStarted    atomic.Int64
	TasksCompleted  atomic.Int64
	TasksFailed     atomic.Int64
	CacheHits       atomic.Int64
	CacheMisses     atomic.Int64
	ExecThrottled   atomic.Int64
	MemThrottled    atomic.Int64
	QueueDepthGauge func() int64

	ExecWaitHist *Histogram
}

func NewMetrics() *Metrics {
	return &Metrics{
		QueueDepthGauge: func() int64 { return 0 },
		ExecWaitHist: NewHistogram("zephyr_exec_wait_ms",
			[]float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}),
	}
}

func (m *Metrics) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	var b strings.Builder
	counter := func(name, help string, v int64) {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s counter\n%s %d\n", name, help, name, name, v)
	}

	counter("zephyr_tasks_submitted_total", "Total tasks submitted to the orchestrator.", m.TasksSubmitted.Load())
	counter("zephyr_tasks_started_total", "Total tasks started.", m.TasksStarted.Load())
	counter("zephyr_tasks_completed_total", "Total tasks completed successfully.", m.TasksCompleted.Load())
	counter("zephyr_tasks_failed_total", "Total tasks failed.", m.TasksFailed.Load())
	counter("zephyr_cache_hits_total", "Total action cache hits.", m.CacheHits.Load())
	counter("zephyr_cache_misses_total", "Total action cache misses.", m.CacheMisses.Load())
	counter("zephyr_exec_throttled_total", "Tasks delayed by spawn rate token bucket.", m.ExecThrottled.Load())
	counter("zephyr_mem_throttled_total", "Tasks delayed by memory governor.", m.MemThrottled.Load())

	fmt.Fprintf(&b, "# HELP zephyr_queue_depth Current ready queue depth.\n# TYPE zephyr_queue_depth gauge\nzephyr_queue_depth %d\n",
		m.QueueDepthGauge())

	h := m.ExecWaitHist
	fmt.Fprintf(&b, "# HELP %s Subprocess launch latency in milliseconds.\n# TYPE %s histogram\n", h.name, h.name)
	cum := int64(0)
	for i, bound := range h.bounds {
		cum += h.counts[i].Load()
		fmt.Fprintf(&b, "%s_bucket{le=\"%g\"} %d\n", h.name, bound, cum)
	}
	cum += h.counts[len(h.bounds)].Load()
	fmt.Fprintf(&b, "%s_bucket{le=\"+Inf\"} %d\n%s_count %d\n%s_sum %d\n",
		h.name, cum, h.name, h.count.Load(), h.name, h.sumMs.Load())

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Write([]byte(b.String()))
}

func (m *Metrics) ExportOTLPJSON(endpoint string) error {
	metrics := []any{
		otlpSum("zephyr_tasks_submitted_total", m.TasksSubmitted.Load()),
		otlpSum("zephyr_tasks_completed_total", m.TasksCompleted.Load()),
		otlpSum("zephyr_tasks_failed_total", m.TasksFailed.Load()),
		otlpSum("zephyr_cache_hits_total", m.CacheHits.Load()),
		otlpSum("zephyr_cache_misses_total", m.CacheMisses.Load()),
		otlpGauge("zephyr_queue_depth", m.QueueDepthGauge()),
	}
	body, err := json.Marshal(map[string]any{
		"resourceMetrics": []any{map[string]any{
			"resource": map[string]any{"attributes": []any{
				map[string]any{"key": "service.name", "value": map[string]any{"stringValue": "zephyr"}},
			}},
			"scopeMetrics": []any{map[string]any{
				"scope":   map[string]any{"name": "zephyr.taskrunner"},
				"metrics": metrics,
			}},
		}},
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(endpoint, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("otlp export error: HTTP %d", resp.StatusCode)
	}
	return nil
}

func otlpSum(name string, v int64) any {
	return map[string]any{
		"name": name,
		"sum": map[string]any{
			"aggregationTemporality": "AGGREGATION_TEMPORALITY_CUMULATIVE",
			"isMonotonic":             true,
			"dataPoints":              []any{map[string]any{"asInt": v}},
		},
	}
}

func otlpGauge(name string, v int64) any {
	return map[string]any{
		"name": name,
		"gauge": map[string]any{
			"dataPoints": []any{map[string]any{"asInt": v}},
		},
	}
}

// ============================================================================
// AGING-AWARE CRITICAL PATH PRIORITY QUEUE
// ============================================================================

type ScheduledTask struct {
	Name        string
	CPScore     float64
	MemEstimate int64
	ReadyAt     time.Time
	RunFunc     func(context.Context) error
}

type readyHeap []ScheduledTask

func (h readyHeap) Len() int { return len(h) }
func (h readyHeap) Less(i, j int) bool {
	// Effective priority = CPScore + (ageBoostPerSec * secondsWaiting)
	// Eliminates low-critical-path task starvation under heavy workloads.
	const ageBoostPerSec = 0.05
	ei := h[i].CPScore + ageBoostPerSec*time.Since(h[i].ReadyAt).Seconds()
	ej := h[j].CPScore + ageBoostPerSec*time.Since(h[j].ReadyAt).Seconds()
	return ei > ej
}
func (h readyHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *readyHeap) Push(x any)       { *h = append(*h, x.(ScheduledTask)) }
func (h *readyHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// ============================================================================
// SCHEDULER
// ============================================================================

type Scheduler struct {
	workers   int
	spawnRate *TokenBucket
	memGov    *MemoryGovernor
	metrics   *Metrics
	mu        sync.Mutex
	ready     readyHeap
	wake      chan struct{}
	wg        sync.WaitGroup
	stopped   atomic.Bool
}

func NewScheduler(maxWorkers int, m *Metrics) *Scheduler {
	if maxWorkers <= 0 {
		maxWorkers = runtime.GOMAXPROCS(0)
	}
	if m == nil {
		m = NewMetrics()
	}

	budget := DetectMemoryBudget() * 7 / 10 // 70% commit threshold
	s := &Scheduler{
		workers:   maxWorkers,
		spawnRate: NewTokenBucket(50, 200),
		memGov:    NewMemoryGovernor(budget),
		metrics:   m,
		wake:      make(chan struct{}, 1),
	}
	m.QueueDepthGauge = func() int64 {
		s.mu.Lock()
		defer s.mu.Unlock()
		return int64(len(s.ready))
	}
	return s
}

func (s *Scheduler) Submit(task ScheduledTask) {
	s.metrics.TasksSubmitted.Add(1)
	if task.ReadyAt.IsZero() {
		task.ReadyAt = time.Now()
	}
	s.mu.Lock()
	heap.Push(&s.ready, task)
	s.mu.Unlock()

	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(ctx)
	}
}

func (s *Scheduler) Wait() {
	s.wg.Wait()
}

func (s *Scheduler) Stop() {
	s.stopped.Store(true)
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *Scheduler) worker(ctx context.Context) {
	defer s.wg.Done()

	for {
		s.mu.Lock()
		if s.ready.Len() == 0 {
			s.mu.Unlock()
			if s.stopped.Load() {
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-s.wake:
			}
			continue
		}

		task := heap.Pop(&s.ready).(ScheduledTask)
		s.mu.Unlock()

		waitDuration := time.Since(task.ReadyAt)
		s.metrics.ExecWaitHist.Observe(float64(waitDuration.Milliseconds()))

		// Token bucket admission
		if err := s.spawnRate.Wait(ctx); err != nil {
			s.metrics.TasksFailed.Add(1)
			continue
		}

		// Memory governor admission
		release, err := s.memGov.Acquire(ctx, task.MemEstimate)
		if err != nil {
			s.metrics.MemThrottled.Add(1)
			s.metrics.TasksFailed.Add(1)
			continue
		}

		s.metrics.TasksStarted.Add(1)
		execErr := task.RunFunc(ctx)
		release()

		if execErr != nil {
			s.metrics.TasksFailed.Add(1)
		} else {
			s.metrics.TasksCompleted.Add(1)
		}
	}
}
