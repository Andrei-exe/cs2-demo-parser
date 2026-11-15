package jobs

import (
	"cs2-parser/internal/models"
	"errors"
	"strings"
	"sync"
	"testing"
)

var testHash = strings.Repeat("a", 64)

func TestCreateOrGetReusesInflightJob(t *testing.T) {
	s := NewStore()

	first, created := s.CreateOrGet(testHash, "demo.dem")
	if !created {
		t.Fatal("first CreateOrGet: created = false, want true")
	}

	second, created := s.CreateOrGet(testHash, "demo.dem")
	if created {
		t.Error("second CreateOrGet: created = true, want false")
	}
	if second.ID != first.ID {
		t.Errorf("second job ID = %s, want the in-flight %s", second.ID, first.ID)
	}

	other, created := s.CreateOrGet(strings.Repeat("b", 64), "other.dem")
	if !created || other.ID == first.ID {
		t.Error("a different hash must get its own job")
	}
}

// Once a job is finished (or deleted) its hash is free again
func TestCreateOrGetAfterFinish(t *testing.T) {
	tests := []struct {
		name   string
		finish func(s *Store, id string)
	}{
		{"done", func(s *Store, id string) { s.SetDone(id, &models.MatchStats{}) }},
		{"failed", func(s *Store, id string) { s.SetFailed(id, errors.New("boom")) }},
		{"deleted", func(s *Store, id string) { s.Delete(id) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStore()
			first, _ := s.CreateOrGet(testHash, "demo.dem")
			tc.finish(s, first.ID)

			second, created := s.CreateOrGet(testHash, "demo.dem")
			if !created || second.ID == first.ID {
				t.Errorf("got old job %s back after it finished", first.ID)
			}
		})
	}
}

// Many uploads of the same demo at once must all land on a single job
func TestCreateOrGetConcurrent(t *testing.T) {
	s := NewStore()

	const uploads = 50
	ids := make(chan string, uploads)
	var wg sync.WaitGroup
	for range uploads {
		wg.Go(func() {
			j, _ := s.CreateOrGet(testHash, "demo.dem")
			ids <- j.ID
		})
	}
	wg.Wait()
	close(ids)

	unique := make(map[string]struct{})
	for id := range ids {
		unique[id] = struct{}{}
	}
	if len(unique) != 1 {
		t.Errorf("%d jobs created for the same demo, want 1", len(unique))
	}
}

func TestSetCached(t *testing.T) {
	s := NewStore()
	res := &models.MatchStats{MapName: "de_inferno"}

	job, _ := s.CreateOrGet(testHash, "demo.dem")
	s.SetCached(job.ID, res)
	got, ok := s.Get(job.ID)
	if !ok {
		t.Fatal("cached job not found")
	}
	if got.Status != StatusDone || !got.Cached || got.Progress != 1 || got.Result != res {
		t.Errorf("cached job = %+v, want done, cached, progress 1 and the given result", got)
	}
	if got.FinishedAt.IsZero() {
		t.Error("FinishedAt not set, the janitor would never remove this job")
	}

	// A cached job is final, so the hash is free again
	if _, created := s.CreateOrGet(testHash, "demo.dem"); !created {
		t.Error("SetCached did not release the hash")
	}
}

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
