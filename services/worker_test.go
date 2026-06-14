package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/khrees/veilo/services"
)

func TestWorker(t *testing.T) {
	t.Parallel()
	called := make(chan bool, 1)
	w := services.NewWorker("test-worker", 10*time.Millisecond, func(ctx context.Context) error {
		select {
		case called <- true:
		default:
		}
		return nil
	})

	ctx := t.Context()

	w.Start(ctx)

	// Wait for callback to be called
	select {
	case <-called:
		// success
	case <-time.After(200 * time.Millisecond):
		t.Fatal("worker callback was not called")
	}
}
