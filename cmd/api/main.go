package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	parser "cs2-parser/internal/parser"

	fiber "github.com/gofiber/fiber/v3"
	cors "github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {
	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		log.Fatalf("Couldn't create uploads folder: %v", err)
	}

	app := fiber.New(fiber.Config{
		BodyLimit: 500 * 1024 * 1024,
	})
	app.Use(cors.New())

	app.Post("/api/upload", func(c fiber.Ctx) error {
		file, err := c.FormFile("demo")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "No file selected"})
		}

		savePath := filepath.Join("./uploads", file.Filename)
		defer func() {
			err := os.Remove(savePath)
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				log.Printf("Couldn't delete %s: %v", savePath, err)
			}
		}()

		if err := c.SaveFile(file, savePath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Couldn't save demo locally"})
		}

		matchData, err := parser.ParseDemo(context.Background(), savePath, nil)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Parsing error: " + err.Error()})
		}

		return c.JSON(matchData)
	})

	log.Fatal(app.Listen(":3000"))
}
