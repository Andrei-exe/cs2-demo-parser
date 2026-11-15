package jobs

import (
	"cs2-parser/internal/models"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	mu   sync.RWMutex
	jobs map[string]*Job
	subs map[string]map[chan struct{}]struct{}
	// demo hash -> ID of the queued/running job parsing it, so the same demo
	// uploaded twice at once is parsed only once
	inflight map[string]string
}

func NewStore() *Store {
	return &Store{
		jobs:     make(map[string]*Job),
		subs:     make(map[string]map[chan struct{}]struct{}),
		inflight: make(map[string]string),
	}
}

// Must be called with s.mu held
func (s *Store) add(j *Job) Job {
	j.ID = uuid.NewString()
	j.CreatedAt = time.Now()
	s.jobs[j.ID] = j
	return *j
}

func (s *Store) Create(filePath string) Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.add(&Job{Status: StatusQueued, FilePath: filePath})
}

// CreateOrGet returns the job already parsing the demo with this hash, or
// creates a new queued one. created is false when an existing job was reused.
// The lookup and the insert happen under the same lock, so two concurrent
// uploads of the same demo can't both create a job.
func (s *Store) CreateOrGet(hash, filePath string) (job Job, created bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.inflight[hash]; ok {
		if j := s.jobs[id]; j != nil {
			return *j, false
		}
	}

	j := s.add(&Job{Status: StatusQueued, FilePath: filePath, Hash: hash})
	s.inflight[hash] = j.ID
	return j, true
}

func (s *Store) Get(id string) (Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	j, ok := s.jobs[id]
	if !ok {
		return Job{}, false
	}

	return *j, true
}

func (s *Store) Update(id string, fn func(j *Job)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := s.jobs[id]
	if job == nil {
		return false
	}

	fn(job)
	if job.Status.IsFinal() {
		s.release(job)
	}
	s.notify(id)
	return true
}

func (s *Store) SetRunning(id string) bool {
	return s.Update(id, func(j *Job) {
		j.Status = StatusRunning
	})
}

func (s *Store) SetProgress(id string, p float64) bool {
	return s.Update(id, func(j *Job) {
		j.Progress = p
	})
}

func (s *Store) SetDone(id string, res *models.MatchStats) bool {
	return s.Update(id, func(j *Job) {
		j.Status = StatusDone
		j.Result = res
		j.Progress = 1
		j.FinishedAt = time.Now()
	})
}

// SetCached finishes a job with a result loaded from the cache instead of
// a parse. The job never goes through the pool.
func (s *Store) SetCached(id string, res *models.MatchStats) bool {
	return s.Update(id, func(j *Job) {
		j.Status = StatusDone
		j.Result = res
		j.Progress = 1
		j.Cached = true
		j.FinishedAt = time.Now()
	})
}

func (s *Store) SetFailed(id string, err error) bool {
	return s.Update(id, func(j *Job) {
		j.Status = StatusFailed
		if err != nil {
			j.Error = err.Error()
		}
		j.FinishedAt = time.Now()
	})
}

func (s *Store) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if j := s.jobs[id]; j != nil {
		s.release(j)
	}
	delete(s.jobs, id)
	delete(s.subs, id)
}

// release frees the job's hash so the next upload of that demo starts a new
// job (or hits the result cache). Only removes the entry if it still points
// to this job. Must be called with s.mu held.
func (s *Store) release(j *Job) {
	if j.Hash != "" && s.inflight[j.Hash] == j.ID {
		delete(s.inflight, j.Hash)
	}
}

func (s *Store) Subscribe(id string) (<-chan struct{}, func(), bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.jobs[id] == nil {
		return nil, nil, false
	}

	ch := make(chan struct{}, 1)
	if s.subs[id] == nil {
		s.subs[id] = make(map[chan struct{}]struct{})
	}
	s.subs[id][ch] = struct{}{}

	unsubscribe := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.subs[id], ch)
		if len(s.subs[id]) == 0 {
			delete(s.subs, id)
		}
	}

	return ch, unsubscribe, true
}

func (s *Store) notify(id string) {
	for ch := range s.subs[id] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (s *Store) DeleteExpired(ttl time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var n int
	for id, j := range s.jobs {
		if j.Status.IsFinal() && now.Sub(j.FinishedAt) > ttl {
			delete(s.jobs, id)
			delete(s.subs, id)
			n++
		}
	}

	return n
}
