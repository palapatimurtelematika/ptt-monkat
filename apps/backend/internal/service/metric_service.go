package service

import (
	"context"

	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/models"
	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/repository"
)

type MetricService struct {
	targets *repository.TargetRepository
	metrics *repository.MetricRepository
}

func NewMetricService(targets *repository.TargetRepository, metrics *repository.MetricRepository) *MetricService {
	return &MetricService{targets: targets, metrics: metrics}
}

// Latest pairs every active target with its newest reading.
//
// An Influx failure fails the call rather than returning every device as
// unpolled — a dashboard full of nulls is indistinguishable from a dead network.
func (s *MetricService) Latest(ctx context.Context) ([]models.LatestMetric, error) {
	targets, err := s.targets.FindActive(ctx)
	if err != nil {
		return nil, err
	}

	points, err := s.metrics.LatestAll(ctx)
	if err != nil {
		return nil, err
	}

	return joinLatest(targets, points), nil
}

// joinLatest is kept free of DB handles so it is testable without mocks.
// MariaDB drives the result: a point with no matching active target is dropped
// (decommissioned device), a target with no point keeps nil value/timestamp.
func joinLatest(targets []models.SnmpTarget, points map[string]models.MetricPoint) []models.LatestMetric {
	out := make([]models.LatestMetric, 0, len(targets))

	for _, t := range targets {
		row := models.LatestMetric{
			TargetID:   t.ID,
			DeviceName: t.DeviceName,
			IPAddress:  t.IPAddress,
			OID:        t.OID,
			MetricName: t.MetricName,
			Unit:       t.Unit,
		}

		if p, ok := points[models.MetricKey(t.IPAddress, t.MetricName)]; ok {
			value, ts := p.Value, p.Time
			row.Value = &value
			row.Timestamp = &ts
		}

		out = append(out, row)
	}

	return out
}
