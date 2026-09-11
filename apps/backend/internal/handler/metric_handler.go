package handler

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/service"
)

type MetricHandler struct {
	svc *service.MetricService
}

func NewMetricHandler(svc *service.MetricService) *MetricHandler {
	return &MetricHandler{svc: svc}
}

// GetLatest handles GET /api/metrics/latest.
func (h *MetricHandler) GetLatest(c *fiber.Ctx) error {
	data, err := h.svc.Latest(c.UserContext())
	if err != nil {
		// Detail goes to the log; the client gets a generic message so DSNs
		// and bucket names never leak out of the trust boundary.
		log.Printf("metrics/latest: %v", err)
		return fiber.NewError(fiber.StatusBadGateway, "datastore unavailable")
	}
	return c.JSON(data)
}
