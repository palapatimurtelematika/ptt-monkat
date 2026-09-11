package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/ujangdoubleday/ptt-monkat/apps/backend/internal/config"
)

// NewMariaDB opens the relational store holding the SNMP master data.
//
// No AutoMigrate on purpose: db/001_mariadb_init.sql owns the schema, so it
// stays reviewable and cannot drift from what actually runs in production.
func NewMariaDB(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.MariaDBDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("mariadb open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("mariadb pool: %w", err)
	}
	// The stage 3 poller reloads the target list on every tick, so the pool
	// needs headroom beyond the API's own traffic.
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("mariadb ping: %w", err)
	}

	return db, nil
}
