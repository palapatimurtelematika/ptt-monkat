package database

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/config"
)

// NewInflux opens the time-series store holding the polling results.
// The caller owns the client and must Close() it so buffered writes flush.
func NewInflux(cfg config.Config) (influxdb2.Client, error) {
	client := influxdb2.NewClient(cfg.InfluxURL, cfg.InfluxToken)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ready, err := client.Ready(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("influxdb ready: %w", err)
	}
	if ready == nil || ready.Status == nil || *ready.Status != "ready" {
		client.Close()
		return nil, fmt.Errorf("influxdb not ready at %s", cfg.InfluxURL)
	}

	return client, nil
}
