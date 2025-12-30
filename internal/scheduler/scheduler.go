package scheduler

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/unitechio/agent/internal/collectors"
	"github.com/unitechio/agent/internal/config"
	"github.com/unitechio/agent/internal/identity"
	"github.com/unitechio/agent/internal/policy"
)

// Scheduler manages periodic execution of collectors with jitter
type Scheduler struct {
	cfg        *config.Config
	policy     *policy.Engine
	identity   *identity.Manager
	logger     *log.Logger
	collectors []collectors.Collector
	stopCh     chan struct{}
	wg         sync.WaitGroup
	mu         sync.Mutex
	running    bool
}

// Job represents a scheduled collection job
type Job struct {
	Name      string
	Interval  time.Duration
	Collector collectors.Collector
	ticker    *time.Ticker
	stopCh    chan struct{}
}

// New creates a new scheduler
func New(cfg *config.Config, policyEngine *policy.Engine, identityMgr *identity.Manager, logger *log.Logger) *Scheduler {
	return &Scheduler{
		cfg:        cfg,
		policy:     policyEngine,
		identity:   identityMgr,
		logger:     logger,
		collectors: collectors.NewDefaultCollectors(),
		stopCh:     make(chan struct{}),
	}
}

// Start begins the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.logger.Println("Starting scheduler...")
	s.running = true

	// Start each collector as a separate job
	for _, collector := range s.collectors {
		s.startCollectorJob(ctx, collector)
	}

	return nil
}

// startCollectorJob starts a periodic job for a collector
func (s *Scheduler) startCollectorJob(ctx context.Context, collector collectors.Collector) {
	interval := s.cfg.CollectionInterval

	// Add jitter: ±10% of interval
	jitter := time.Duration(float64(interval) * 0.1 * (rand.Float64()*2 - 1))
	actualInterval := interval + jitter

	s.logger.Printf("Starting collector '%s' with interval %v (jitter: %v)",
		collector.Name(), actualInterval, jitter)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ticker := time.NewTicker(actualInterval)
		defer ticker.Stop()

		// Run immediately on start
		s.runCollector(ctx, collector)

		for {
			select {
			case <-ticker.C:
				s.runCollector(ctx, collector)
			case <-s.stopCh:
				s.logger.Printf("Stopping collector '%s'", collector.Name())
				return
			case <-ctx.Done():
				s.logger.Printf("Context cancelled, stopping collector '%s'", collector.Name())
				return
			}
		}
	}()
}

// runCollector executes a single collection
func (s *Scheduler) runCollector(ctx context.Context, collector collectors.Collector) {
	s.logger.Printf("Running collector: %s", collector.Name())

	// Create a timeout context for the collection
	collectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	start := time.Now()
	data, err := collector.Collect(collectCtx)
	duration := time.Since(start)

	if err != nil {
		s.logger.Printf("Collector '%s' failed: %v", collector.Name(), err)
		return
	}

	s.logger.Printf("Collector '%s' completed in %v", collector.Name(), duration)

	// TODO: Send data to backend via sender
	// For now, just log the data
	s.logger.Printf("Collected data from '%s': %+v", collector.Name(), data)
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.logger.Println("Stopping scheduler...")
	close(s.stopCh)

	// Wait for all jobs to complete (with timeout)
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Println("All collector jobs stopped")
	case <-time.After(30 * time.Second):
		s.logger.Println("Warning: timeout waiting for collector jobs to stop")
	}

	s.running = false
}

// AddCollector adds a new collector to the scheduler
func (s *Scheduler) AddCollector(collector collectors.Collector) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.collectors = append(s.collectors, collector)

	// If scheduler is already running, start this collector immediately
	if s.running {
		s.startCollectorJob(context.Background(), collector)
	}
}

// RemoveCollector removes a collector by name
func (s *Scheduler) RemoveCollector(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, collector := range s.collectors {
		if collector.Name() == name {
			s.collectors = append(s.collectors[:i], s.collectors[i+1:]...)
			s.logger.Printf("Removed collector: %s", name)
			return
		}
	}
}
