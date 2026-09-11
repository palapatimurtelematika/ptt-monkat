package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/config"
)

func main() {
	cfg := config.Load()

	app := fiber.New(fiber.Config{AppName: "ptt-monkat api"})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	log.Fatal(app.Listen(":" + cfg.APIPort))
}
