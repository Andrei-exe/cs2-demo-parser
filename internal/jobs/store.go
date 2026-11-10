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
}

func NewStore() *Store {
	return &Store{
		jobs: make(map[string]*Job),
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

	if s.jobs[id] == nil {
		return false
	}

	fn(s.jobs[id])
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
