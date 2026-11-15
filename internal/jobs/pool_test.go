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

// Submits a job and fails the test right away if the queue rejects it
func mustSubmit(t *testing.T, p *Pool, id string) {
	t.Helper()
	if err := p.Submit(id); err != nil {
		t.Fatalf("Submit(%s): %v", id, err)
	}
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
	for range 5 {
		job := s.Create(tempDemo(t))
		if err := p.Submit(job.ID); err != nil {
			t.Fatalf("Submit: %v", err)
		}
		ids = append(ids, job.ID)
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
	}

	cancel()
	p.Wait()
}

// The demo is content-addressed storage now, the pool must leave it alone
func TestPoolKeepsDemoFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewStore()
	p := NewPool(s, okParse, 1, 10, time.Second)
	p.Start(ctx)

	path := tempDemo(t)
	job := s.Create(path)
	mustSubmit(t, p, job.ID)
	waitFinished(t, s, job.ID)

	if _, err := os.Stat(path); err != nil {
		t.Errorf("demo file is gone after parse: %v", err)
	}

	cancel()
	p.Wait()
}

func TestPoolOnSuccessRunsBeforeDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewStore()
	p := NewPool(s, okParse, 1, 10, time.Second)

	hash := strings.Repeat("a", 64)
	var calls int
	var statusDuringHook Status
	p.OnSuccess = func(job Job, res *models.MatchStats) {
		calls++
		if job.Hash != hash {
			t.Errorf("hook got hash %q, want %q", job.Hash, hash)
		}
		if res == nil {
			t.Error("hook got nil result")
		}
		j, _ := s.Get(job.ID)
		statusDuringHook = j.Status
	}
	p.Start(ctx)

	job, _ := s.CreateOrGet(hash, tempDemo(t))
	mustSubmit(t, p, job.ID)
	waitFinished(t, s, job.ID)

	// waitFinished saw "done", which is set after the hook returns, so
	// reading calls here is ordered after the hook's writes
	if calls != 1 {
		t.Errorf("OnSuccess called %d times, want 1", calls)
	}
	if statusDuringHook != StatusRunning {
		t.Errorf("status during hook = %q, want %q", statusDuringHook, StatusRunning)
	}

	cancel()
	p.Wait()
}

func TestPoolOnSuccessSkippedOnFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewStore()
	p := NewPool(s, slowParse, 1, 10, 20*time.Millisecond)
	called := make(chan struct{}, 1)
	p.OnSuccess = func(Job, *models.MatchStats) { called <- struct{}{} }
	p.Start(ctx)

	job := s.Create(tempDemo(t))
	mustSubmit(t, p, job.ID)
	if j := waitFinished(t, s, job.ID); j.Status != StatusFailed {
		t.Fatalf("status = %q, want %q", j.Status, StatusFailed)
	}

	select {
	case <-called:
		t.Error("OnSuccess ran for a failed job, failures must not be cached")
	default:
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
	mustSubmit(t, p, first.ID)
	mustSubmit(t, p, second.ID)

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
	mustSubmit(t, p, job.ID)

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
