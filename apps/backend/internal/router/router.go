package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/handler"
)

// Register mounts the route table.
//
// No CORS middleware: Vite proxies /api to this server, so the browser stays
// same-origin. Add it only if the frontend is ever served from another origin.
func Register(app *fiber.App, metrics *handler.MetricHandler) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api")
	api.Get("/metrics/latest", metrics.GetLatest)
}
