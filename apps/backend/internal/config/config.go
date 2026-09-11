package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	APIPort      string
	MariaDBDSN   string
	InfluxURL    string
	InfluxToken  string
	InfluxOrg    string
	InfluxBucket string
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
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
