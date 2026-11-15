package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"cs2-parser/internal/jobs"
	"cs2-parser/internal/models"
	parser "cs2-parser/internal/parser"
	"cs2-parser/internal/storage"

	fiber "github.com/gofiber/fiber/v3"
	cors "github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/sse"
)

func main() {
	workers := flag.Int("workers", 2, "number of parallel parse workers")
	queueSize := flag.Int("queue", 10, "max jobs waiting in queue")
	timeout := flag.Duration("timeout", 2*time.Minute, "max time per parse")
	shutdownTimeout := flag.Duration("shutdown-timeout", 30*time.Second, "max time to wait for running jobs on shutdown")
	jobTTL := flag.Duration("job-ttl", 30*time.Minute, "how long finished jobs are kept in memory")
	storageDir := flag.String("storage", "./storage", "where demos and cached results are kept")
	flag.Parse()

	st, err := storage.New(*storageDir, parser.Version)
	if err != nil {
		log.Fatalf("Couldn't init storage: %v", err)
	}
	if err := st.ClearTmp(); err != nil {
		log.Fatalf("Couldn't clear storage tmp: %v", err)
	}

	store := jobs.NewStore()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	store.StartJanitor(ctx, *jobTTL, min(*jobTTL, time.Minute))
	context.AfterFunc(ctx, stop)

	workCtx, abortWork := context.WithCancel(context.Background())
	defer abortWork()

	pool := jobs.NewPool(store, parser.ParseDemo, *workers, *queueSize, *timeout)
	// Runs before the job is marked done: once the hash is released from
	// in-flight, the result is already on disk for the next upload
	pool.OnSuccess = func(job jobs.Job, res *models.MatchStats) {
		if job.Hash == "" {
			return
		}
		if err := st.SaveResult(job.Hash, res); err != nil {
			log.Printf("Couldn't cache result for %s: %v", job.Hash, err)
		}
	}
	pool.Start(workCtx)

	app := fiber.New(fiber.Config{
		BodyLimit: 500 * 1024 * 1024,
	})
	app.Use(cors.New())

	app.Post("/api/upload", func(c fiber.Ctx) error {
		file, err := c.FormFile("demo")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "No file selected"})
		}

		if !strings.EqualFold(filepath.Ext(file.Filename), ".dem") {
			return c.Status(400).JSON(fiber.Map{"error": "Only .dem files are accepted"})
		}

		src, err := file.Open()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Couldn't read uploaded demo"})
		}
		defer func() { _ = src.Close() }()

		// One pass over the data: written to disk and hashed at the same time
		hash, err := st.SaveDemo(src)
		if err != nil {
			log.Printf("save demo: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Couldn't save demo locally"})
		}

		// Claim the hash first and only then look in the cache. The pool writes
		// the result before releasing the hash, so there is no moment where a
		// finished demo is neither in-flight nor cached
		job, created := store.CreateOrGet(hash, st.DemoPath(hash))
		if !created {
			// Same demo is already queued or being parsed: follow that job
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"job_id": job.ID})
		}

		res, err := st.LoadResult(hash)
		if err == nil {
			store.SetCached(job.ID, res)
			log.Printf("job %s served from cache (%s)", job.ID, hash)
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"job_id": job.ID, "cached": true})
		}
		if !errors.Is(err, storage.ErrNotFound) {
			// Corrupt or unreadable cache entry: parse again, the new result overwrites it
			log.Printf("cached result for %s unusable, parsing again: %v", hash, err)
		}

		if err := pool.Submit(job.ID); err != nil {
			// The demo stays in storage: it's content-addressed, a retry reuses it
			store.Delete(job.ID)

			if errors.Is(err, jobs.ErrShuttingDown) {
				return c.Status(503).JSON(fiber.Map{"error": "Server is shutting down"})
			}

			c.Set("Retry-After", "10")
			return c.Status(503).JSON(fiber.Map{"error": "Server is busy, try again in a few seconds"})
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"job_id": job.ID})
	})

	app.Get("/api/jobs/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		job, ok := store.Get(id)
		if !ok {
			return c.Status(404).JSON(fiber.Map{"error": "Job not found"})
		}

		return c.JSON(job)
	})

	app.Get("/api/jobs/:id/events", func(c fiber.Ctx) error {
		if _, ok := store.Get(c.Params("id")); !ok {
			return c.Status(404).JSON(fiber.Map{"error": "Job not found"})
		}
		return c.Next()
	}, sse.New(sse.Config{
		Handler: func(c fiber.Ctx, stream *sse.Stream) error {
			id := strings.Clone(c.Params("id"))
			updates, unsubscribe, ok := store.Subscribe(id)
			if !ok {
				return nil
			}
			defer unsubscribe()

			send := func() (bool, error) {
				job, _ := store.Get(id)

				var err error
				switch job.Status {
				case jobs.StatusDone:
					err = stream.Event(sse.Event{Name: "done", Data: job})
				case jobs.StatusFailed:
					err = stream.Event(sse.Event{Name: "failed", Data: fiber.Map{"error": job.Error}})
				default:
					err = stream.Event(sse.Event{Name: "progress", Data: fiber.Map{"status": job.Status, "progress": job.Progress}})
				}

				return job.Status.IsFinal(), err
			}

			if done, err := send(); done || err != nil {
				return err
			}

			for {
				select {
				case <-updates:
					if done, err := send(); done || err != nil {
						return err
					}
				case <-stream.Done():
					return stream.Err()
				case <-ctx.Done():
					return nil
				}
			}
		},
	}))

	if err := app.Listen(":3000", fiber.ListenConfig{
		GracefulContext: ctx,
		ShutdownTimeout: 10 * time.Second,
	}); err != nil {
		log.Printf("server error: %v", err)
	}

	log.Println("HTTP stopped, waiting for running jobs...")
	shCtx, cancel := context.WithTimeout(context.Background(), *shutdownTimeout)
	defer cancel()

	if err := pool.Shutdown(shCtx); err != nil {
		log.Println("jobs didn't finish in time, aborting them")
		abortWork()
		pool.Wait()
	}
	log.Println("bye")
}
