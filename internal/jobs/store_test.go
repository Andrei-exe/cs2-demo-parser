package jobs

import (
	"errors"
	"sync"
	"testing"
)

// Hammers the store from many goroutines at once
// Run with: go test -race ./internal/jobs/...
func TestStoreConcurrent(t *testing.T) {
	s := NewStore()
	job := s.Create("demo.dem")

	const workers = 50
	var wg sync.WaitGroup

	// Writers: simulate parse workers updating the same job
	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.SetRunning(job.ID)
			s.SetProgress(job.ID, float64(i)/workers)
		}()
	}

	// Readers: simulate HTTP handlers polling the job
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, ok := s.Get(job.ID); !ok {
				t.Errorf("job %s disappeared while reading", job.ID)
			}
		}()
	}

	// Creators: simulate new uploads arriving at the same time
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			j := s.Create("other.dem")
			s.SetFailed(j.ID, errors.New("boom"))
		}()
	}

	wg.Wait()

	got, ok := s.Get(job.ID)
	if !ok {
		t.Fatalf("job %s not found after concurrent updates", job.ID)
	}
	if got.Status != StatusRunning {
		t.Errorf("status = %q, want %q", got.Status, StatusRunning)
	}

	// Updating a job that doesn't exist must not panic
	if s.SetProgress("does-not-exist", 0.5) {
		t.Error("SetProgress on missing job returned true, want false")
	}
}
