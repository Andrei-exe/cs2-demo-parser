package jobs

import (
	"context"
	"cs2-parser/internal/models"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Fake parser that reports some progress and returns empty stats
func okParse(ctx context.Context, path string, onProgress func(float64)) (*models.MatchStats, error) {
	onProgress(0.5)
	return &models.MatchStats{}, nil
}

// Fake parser that crashes, like a corrupted demo would
func panicParse(ctx context.Context, path string, onProgress func(float64)) (*models.MatchStats, error) {
	panic("corrupted demo")
}

// Fake parser that never finishes on its own, it only stops when ctx is cancelled
func slowParse(ctx context.Context, path string, onProgress func(float64)) (*models.MatchStats, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

// Creates an empty file in a temp dir that the test deletes automatically
func tempDemo(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "demo.dem")
	if err := os.WriteFile(path, []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// Polls the store until the job reaches done/failed or the deadline passes
func waitFinished(t *testing.T, s *Store, id string) Job {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		j, ok := s.Get(id)
		if ok && (j.Status == StatusDone || j.Status == StatusFailed) {
			return j
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s did not finish in time", id)
	return Job{}
}

func TestPoolProcessesJobs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewStore()
	p := NewPool(s, okParse, 2, 10, time.Second)
	p.Start(ctx)

	var ids []string
	var paths []string
	for range 5 {
		path := tempDemo(t)
		job := s.Create(path)
		if err := p.Submit(job.ID); err != nil {
			t.Fatalf("Submit: %v", err)
		}
		ids = append(ids, job.ID)
		paths = append(paths, path)
	}

	for i, id := range ids {
		j := waitFinished(t, s, id)
		if j.Status != StatusDone {
			t.Errorf("job %d: status = %q, want %q (error: %s)", i, j.Status, StatusDone, j.Error)
		}
		if j.Progress != 1 {
			t.Errorf("job %d: progress = %v, want 1", i, j.Progress)
		}
		if j.Result == nil {
			t.Errorf("job %d: result is nil", i)
		}
		if _, err := os.Stat(paths[i]); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("job %d: demo file was not deleted", i)
		}
	}

	cancel()
	p.Wait()
}

func TestPoolRecoversFromPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewStore()
	p := NewPool(s, panicParse, 1, 10, time.Second)
	p.Start(ctx)

	// Two jobs on a single worker: if the first panic killed the worker,
	// the second job would never finish
	first := s.Create(tempDemo(t))
	second := s.Create(tempDemo(t))
	p.Submit(first.ID)
	p.Submit(second.ID)

	for _, id := range []string{first.ID, second.ID} {
		j := waitFinished(t, s, id)
		if j.Status != StatusFailed {
			t.Errorf("status = %q, want %q", j.Status, StatusFailed)
		}
		if !strings.Contains(j.Error, "parser panic") {
			t.Errorf("error = %q, want it to mention the panic", j.Error)
		}
	}

	cancel()
	p.Wait()
}

func TestPoolTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewStore()
	p := NewPool(s, slowParse, 1, 10, 50*time.Millisecond)
	p.Start(ctx)

	job := s.Create(tempDemo(t))
	p.Submit(job.ID)

	j := waitFinished(t, s, job.ID)
	if j.Status != StatusFailed {
		t.Errorf("status = %q, want %q", j.Status, StatusFailed)
	}
	if !strings.Contains(j.Error, context.DeadlineExceeded.Error()) {
		t.Errorf("error = %q, want deadline exceeded", j.Error)
	}

	cancel()
	p.Wait()
}

func TestPoolQueueFull(t *testing.T) {
	s := NewStore()
	// Pool is never started, so nobody drains the queue
	p := NewPool(s, okParse, 1, 2, time.Second)

	for i := range 2 {
		if err := p.Submit(s.Create("x.dem").ID); err != nil {
			t.Fatalf("Submit %d: unexpected error %v", i, err)
		}
	}

	err := p.Submit(s.Create("x.dem").ID)
	if !errors.Is(err, ErrQueueFull) {
		t.Errorf("third Submit: err = %v, want ErrQueueFull", err)
	}
}

func TestPoolStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	s := NewStore()
	p := NewPool(s, okParse, 3, 10, time.Second)
	p.Start(ctx)

	cancel()

	done := make(chan struct{})
	go func() {
		p.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("workers did not stop after context was cancelled")
	}
}
