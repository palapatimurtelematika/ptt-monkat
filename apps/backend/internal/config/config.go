package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	APIPort      string
	MariaDBDSN   string
	InfluxURL    string
	InfluxToken  string
	InfluxOrg    string
	InfluxBucket string

	// Poller. SNMPTimeout*(1+SNMPRetries) must stay well under PollInterval
	// so a slow tick cannot overrun the next one.
	PollInterval    time.Duration
	PollConcurrency int
	SNMPTimeout     time.Duration
	SNMPRetries     int
}

// Load reads .env (if present) then the environment. Missing vars fall back to
// the dev defaults below so `npm run dev` works on a fresh clone.
func Load() Config {
	_ = godotenv.Load()

	return Config{
		APIPort:      env("API_PORT", "3000"),
		MariaDBDSN:   env("MARIADB_DSN", "monkat:changeme@tcp(127.0.0.1:3307)/ptt_monkat?parseTime=true&charset=utf8mb4"),
		InfluxURL:    env("INFLUX_URL", "http://127.0.0.1:8087"),
		InfluxToken:  env("INFLUX_TOKEN", ""),
		InfluxOrg:    env("INFLUX_ORG", "ptt"),
		InfluxBucket: env("INFLUX_BUCKET", "snmp_metrics"),

		PollInterval:    envDuration("POLL_INTERVAL", time.Minute),
		PollConcurrency: envInt("POLL_CONCURRENCY", 50),
		SNMPTimeout:     envDuration("SNMP_TIMEOUT", 2*time.Second),
		SNMPRetries:     envInt("SNMP_RETRIES", 1),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(os.Getenv(key)); err == nil && d > 0 {
		return d
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if n, err := strconv.Atoi(os.Getenv(key)); err == nil && n > 0 {
		return n
	}
	return fallback
}
