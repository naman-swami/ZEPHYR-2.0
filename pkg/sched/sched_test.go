package sched

import (
	"context"
	"container/heap"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	tb := NewTokenBucket(100, 5) // 100/sec, burst 5
	ctx := context.Background()

	// Should consume 5 immediately without delay
	for i := 0; i < 5; i++ {
		if err := tb.Wait(ctx); err != nil {
			t.Fatalf("unexpected token bucket wait error: %v", err)
		}
	}
}

func TestMemoryGovernor(t *testing.T) {
	gov := NewMemoryGovernor(64 * 1024 * 1024) // 64 MiB = 4 granules of 16 MiB
	ctx := context.Background()

	// Acquire 32 MiB (2 granules)
	release1, err := gov.Acquire(ctx, 32*1024*1024)
	if err != nil {
		t.Fatalf("failed acquiring memory granules: %v", err)
	}
	defer release1()

	// Acquire another 32 MiB (2 granules)
	release2, err := gov.Acquire(ctx, 32*1024*1024)
	if err != nil {
		t.Fatalf("failed acquiring second memory granules: %v", err)
	}
	defer release2()
}

func TestAgingPriorityHeap(t *testing.T) {
	h := &readyHeap{}
	heap.Init(h)

	// Task A has low initial CP score (10) but has waited for 100 seconds (boost = 5.0) -> effective 15.0
	taskA := ScheduledTask{
		Name:    "taskA",
		CPScore: 10,
		ReadyAt: time.Now().Add(-100 * time.Second),
	}

	// Task B has high initial CP score (12) but was ready right now -> effective 12.0
	taskB := ScheduledTask{
		Name:    "taskB",
		CPScore: 12,
		ReadyAt: time.Now(),
	}

	heap.Push(h, taskB)
	heap.Push(h, taskA)

	// With aging boost, Task A must pop first, eliminating starvation!
	popped := heap.Pop(h).(ScheduledTask)
	if popped.Name != "taskA" {
		t.Errorf("expected taskA to pop first due to aging boost, got: %s", popped.Name)
	}
}

func TestMetricsPrometheusServeHTTP(t *testing.T) {
	m := NewMetrics()
	m.TasksSubmitted.Add(10)
	m.TasksCompleted.Add(8)
	m.CacheHits.Add(5)
	m.ExecWaitHist.Observe(12.5)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	m.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got: %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "zephyr_tasks_submitted_total 10") {
		t.Errorf("missing zephyr_tasks_submitted_total in metrics output")
	}
	if !strings.Contains(body, "zephyr_cache_hits_total 5") {
		t.Errorf("missing zephyr_cache_hits_total in metrics output")
	}
	if !strings.Contains(body, "zephyr_exec_wait_ms_bucket") {
		t.Errorf("missing histogram buckets in metrics output")
	}
}

func TestSchedulerExecution(t *testing.T) {
	m := NewMetrics()
	s := NewScheduler(4, m)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)

	var executedCount atomic.Int32
	for i := 0; i < 5; i++ {
		s.Submit(ScheduledTask{
			Name:        "unit_task",
			CPScore:     float64(i),
			MemEstimate: 16 * 1024 * 1024,
			RunFunc: func(ctx context.Context) error {
				executedCount.Add(1)
				return nil
			},
		})
	}

	// Give workers a moment to execute
	time.Sleep(100 * time.Millisecond)
	s.Stop()
	s.Wait()

	if executedCount.Load() != 5 {
		t.Errorf("expected 5 executed tasks, got %d", executedCount.Load())
	}
	if m.TasksCompleted.Load() != 5 {
		t.Errorf("expected 5 completed tasks in metrics, got %d", m.TasksCompleted.Load())
	}
}
