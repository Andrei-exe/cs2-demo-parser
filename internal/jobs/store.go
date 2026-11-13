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
}

func NewStore() *Store {
	return &Store{
		jobs: make(map[string]*Job),
		subs: make(map[string]map[chan struct{}]struct{}),
	}
}

func (s *Store) Create(filePath string) Job {
	newID := uuid.New().String()

	newJob := &Job{
		ID:        newID,
		Status:    StatusQueued,
		CreatedAt: time.Now(),
		FilePath:  filePath,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[newID] = newJob

	return *newJob
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

func (s *Store) SetFailed(id string, err error) bool {
	return s.Update(id, func(j *Job) {
		s.jobs[id].Status = StatusFailed
		if err != nil {
			j.Error = err.Error()
		}
		j.FinishedAt = time.Now()
	})
}

func (s *Store) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.jobs, id)
	delete(s.subs, id)
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
