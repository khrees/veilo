package services

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
)

// Worker represents a generic periodic background job runner
type Worker struct {
	name     string
	interval time.Duration
	runFn    func(ctx context.Context) error
}

// NewWorker creates a new generic background worker
func NewWorker(name string, interval time.Duration, runFn func(ctx context.Context) error) *Worker {
	return &Worker{
		name:     name,
		interval: interval,
		runFn:    runFn,
	}
}

// Start spawns the background ticker goroutine
func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	log.Infof("Worker %s started (runs every %v)", w.name, w.interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Infof("Worker %s stopped", w.name)
				return
			case <-ticker.C:
				log.Infof("Worker %s running job...", w.name)
				if err := w.runFn(ctx); err != nil {
					log.Errorf("Worker %s job failed: %v", w.name, err)
				} else {
					log.Infof("Worker %s job completed successfully", w.name)
				}
			}
		}
	}()
}
