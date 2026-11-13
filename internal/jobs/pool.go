package jobs

import (
	"context"
	"cs2-parser/internal/models"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sync"
	"time"
)

type ParseFunc func(ctx context.Context, path string, onProgress func(float64)) (*models.MatchStats, error)

var ErrQueueFull = errors.New("job queue is full")
var ErrShuttingDown = errors.New("server is shutting down")

type Pool struct {
	store    *Store
	parse    ParseFunc
	queue    chan string
	workers  int
	timeout  time.Duration
	wg       sync.WaitGroup
	stop     chan struct{}
	stopOnce sync.Once
}

func NewPool(store *Store, parse ParseFunc, workers, queueSize int, timeout time.Duration) *Pool {
	return &Pool{
		store:   store,
		parse:   parse,
		queue:   make(chan string, queueSize),
		workers: workers,
		timeout: timeout,
		stop:    make(chan struct{}),
	}
}

func (p *Pool) Start(ctx context.Context) {
	for range p.workers {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

func (p *Pool) Submit(id string) error {
	select {
	case <-p.stop:
		return ErrShuttingDown
	default:
	}

	select {
	case p.queue <- id:
		return nil
	default:
		return ErrQueueFull
	}
}

func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case <-p.stop:
			return
		default:
		}

		select {
		case <-ctx.Done():
			return
		case <-p.stop:
			return
		case id := <-p.queue:
			p.process(ctx, id)
		}
	}
}

func (p *Pool) process(ctx context.Context, id string) {
	job, ok := p.store.Get(id)
	if !ok {
		return
	}

	defer func() {
		err := os.Remove(job.FilePath)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			log.Printf("Couldn't delete %s: %v", job.FilePath, err)
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			p.store.SetFailed(id, fmt.Errorf("parser panic: %v", r))
		}
	}()

	p.store.SetRunning(id)

	start := time.Now()
	defer func() {
		j, _ := p.store.Get(id)
		log.Printf("job %s %s in %s", id, j.Status, time.Since(start).Round(time.Millisecond))
	}()

	jobCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	res, err := p.parse(jobCtx, job.FilePath, func(f float64) {
		p.store.SetProgress(id, f)
	})

	if err != nil {
		p.store.SetFailed(id, err)
	} else {
		p.store.SetDone(id, res)
	}
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) Shutdown(ctx context.Context) error {
	p.stopOnce.Do(func() { close(p.stop) })
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
