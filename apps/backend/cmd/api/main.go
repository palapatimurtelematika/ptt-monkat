package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/config"
	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/database"
	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/handler"
	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/repository"
	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/router"
	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/service"
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

	metricSvc := service.NewMetricService(
		repository.NewTargetRepository(db),
		repository.NewMetricRepository(influx, cfg.InfluxOrg, cfg.InfluxBucket),
	)

	app := fiber.New(fiber.Config{AppName: "ptt-monkat api"})
	app.Use(logger.New())
	router.Register(app, handler.NewMetricHandler(metricSvc))

	// Listen off the main goroutine so the deferred Close above actually runs
	// on Ctrl-C — log.Fatal(app.Listen(...)) would skip every deferred flush.
	go func() {
		if err := app.Listen(":" + cfg.APIPort); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()
	log.Printf("api listening on :%s", cfg.APIPort)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("shutting down")
	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
