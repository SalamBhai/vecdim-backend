package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all abstracted application configuration and secrets.
// Direct equivalent of .NET's IOptions<AppSettings>.
type Config struct {
	Port             string
	Env              string // development, production, test
	DatabaseURL      string
	DBHost           string
	DBPort           int
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	CronIntervalMins int
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

// CronInterval returns the interval as a time.Duration.
func (c *Config) CronInterval() time.Duration {
	if c.CronIntervalMins <= 0 {
		return 30 * time.Minute
	}
	return time.Duration(c.CronIntervalMins) * time.Minute
}

// Load reads configuration from .env file (if present) and OS environment variables,
// falling back to safe local development defaults. In production, required secrets
// are strictly enforced using the Fail-Fast principle.
func Load() *Config {
	// 1. Try to load .env file if it exists (in current directory or root)
	loadDotEnv(".env")
	loadDotEnv("../.env")

	// 2. Read with environment variable overrides and sensible defaults
	port := getEnv("PORT", "8080")
	env := getEnv("APP_ENV", "development")

	databaseURL := getEnv("DATABASE_URL", "")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnvInt("DB_PORT", 5432)
	dbUser := getEnv("DB_USER", "vecdim_admin")
	dbName := getEnv("DB_NAME", "vecdim")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	// Secrets hygiene: Enforce explicit secrets in production; fallback to sandbox secrets in dev/test.
	var dbPassword string
	if strings.ToLower(env) == "production" && databaseURL == "" {
		dbPassword = requireEnv("DB_PASSWORD")
		if dbPassword == "vecdim_secret_2026" {
			log.Fatalf("FATAL CONFIG ERROR: In production, DB_PASSWORD cannot use the development default value.")
		}
	} else if databaseURL == "" {
		dbPassword = getEnv("DB_PASSWORD", "vecdim_secret_2026")
	}

	cronInterval := getEnvInt("CRON_INTERVAL_MINS", 30)

	return &Config{
		Port:             port,
		Env:              env,
		DatabaseURL:      databaseURL,
		DBHost:           dbHost,
		DBPort:           dbPort,
		DBUser:           dbUser,
		DBPassword:       dbPassword,
		DBName:           dbName,
		DBSSLMode:        dbSSLMode,
		CronIntervalMins: cronInterval,
	}
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if valStr, exists := os.LookupEnv(key); exists {
		if parsed, err := strconv.Atoi(strings.TrimSpace(valStr)); err == nil {
			return parsed
		}
	}
	return defaultVal
}

// requireEnv enforces that a required environment variable exists and is non-empty.
// If missing in production, it fails fast by terminating the process immediately.
func requireEnv(key string) string {
	val, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(val) == "" {
		log.Fatalf("FATAL CONFIG ERROR: Environment variable '%s' is required in production but not set.", key)
	}
	return strings.TrimSpace(val)
}

// loadDotEnv reads key=value pairs from a .env file and sets them in the process environment.
func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return // File does not exist, ignore
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip comments and empty lines
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Strip optional surrounding quotes
			val = strings.Trim(val, `"'`)
			// Only set if not already set in OS environment
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, val)
			}
		}
	}
}
