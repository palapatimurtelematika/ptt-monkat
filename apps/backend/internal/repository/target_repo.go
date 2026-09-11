package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/models"
)

type TargetRepository struct {
	db *gorm.DB
}

func NewTargetRepository(db *gorm.DB) *TargetRepository {
	return &TargetRepository{db: db}
}

// FindActive returns every target the poller and the dashboard care about.
// Rides the idx_poller (is_active, ip_address) index.
func (r *TargetRepository) FindActive(ctx context.Context) ([]models.SnmpTarget, error) {
	var targets []models.SnmpTarget
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("device_name, metric_name").
		Find(&targets).Error
	if err != nil {
		return nil, fmt.Errorf("find active targets: %w", err)
	}
	return targets, nil
}
