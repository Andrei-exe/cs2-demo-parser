package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cs2-parser/internal/jobs"
	parser "cs2-parser/internal/parser"

	fiber "github.com/gofiber/fiber/v3"
	cors "github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/google/uuid"
)

func main() {
	workers := flag.Int("workers", 2, "number of parallel parse workers")
	queueSize := flag.Int("queue", 10, "max jobs waiting in queue")
	timeout := flag.Duration("timeout", 2*time.Minute, "max time per parse")
	flag.Parse()

	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		log.Fatalf("Couldn't create uploads folder: %v", err)
	}

	store := jobs.NewStore()
	pool := jobs.NewPool(store, parser.ParseDemo, *workers, *queueSize, *timeout)
	pool.Start(context.Background())

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

		savePath := filepath.Join("./uploads", uuid.NewString()+".dem")

		if err := c.SaveFile(file, savePath); err != nil {
			_ = os.Remove(savePath)
			return c.Status(500).JSON(fiber.Map{"error": "Couldn't save demo locally"})
		}

		job := store.Create(savePath)
		if err := pool.Submit(job.ID); err != nil {
			store.Delete(job.ID)
			_ = os.Remove(savePath)

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

	log.Fatal(app.Listen(":3000"))
}
