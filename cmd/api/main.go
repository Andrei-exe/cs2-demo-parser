package main

import (
	"os"
	"path/filepath"

	parser "cs2-parser/internal/parser"

	fiber "github.com/gofiber/fiber/v3"
	cors "github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit: 500 * 1024 * 1024,
	})
	app.Use(cors.New())

	app.Post("/api/upload", func(c fiber.Ctx) error {
		file, err := c.FormFile("demo")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "No file selected"})
		}

		os.MkdirAll("./uploads", os.ModePerm)
		savePath := filepath.Join("./uploads", file.Filename)

		if err := c.SaveFile(file, savePath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Couldn't save demo locally"})
		}

		matchData, err := parser.ParseDemo(savePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Parsing error: " + err.Error()})
		}

		os.Remove(savePath)

		return c.JSON(matchData)
	})

	app.Listen(":3000")
}
