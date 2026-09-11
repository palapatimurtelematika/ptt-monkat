// Command poller runs the SNMP engine on its own, for deployments that keep
// the collector separate from the API.
//
// cmd/api already starts the same engine in-process. Do not run both against
// one database — every target would be polled twice.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/config"
	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/database"
	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/poller"
	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/repository"
)

func main() {
	cfg := config.Load()

	db, err := database.NewMariaDB(cfg)
	if err != nil {
		log.Fatalf("mariadb: %v", err)
	}

	influx, err := database.NewInflux(cfg)
	if err != nil {
		log.Fatalf("influxdb: %v", err)
	}
	defer influx.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	poller.New(
		repository.NewTargetRepository(db),
		repository.NewMetricRepository(influx, cfg.InfluxOrg, cfg.InfluxBucket),
		cfg,
	).Run(ctx)
}
