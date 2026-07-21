package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Address          string
	DatabaseDriver   string
	DatabaseHost     string
	DatabasePort     string
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseSSLMode  string
	DatabaseSchema   string
	DatabaseTimezone string
	JWTSecret        string
	MigrationDir     string
	RequestTimeoutS  int
}

func Load() Config {
	_ = loadDotEnv(".env")

	address := getEnv("APP_ADDR", ":8080")
	databaseDriver := getEnv("DATABASE_DRIVER", "pgx")
	databaseHost := getEnv("DATABASE_HOST", "localhost")
	databasePort := getEnv("DATABASE_PORT", "5432")
	databaseName := os.Getenv("DATABASE_NAME")
	databaseUser := os.Getenv("DATABASE_USERNAME")
	databasePassword := os.Getenv("DATABASE_PASSWORD")
	databaseSSLMode := getEnv("DATABASE_SSLMODE", "disable")
	databaseSchema := getEnv("DATABASE_SCHEMA", "public")
	databaseTimezone := getEnv("DATABASE_TIMEZONE", "Asia/Jakarta")
	jwtSecret := getEnv("JWT_SECRET", "development-jwt-secret")
	migrationDir := getEnv("MIGRATION_DIR", "migrations")

	return Config{
		Address:          address,
		DatabaseDriver:   databaseDriver,
		DatabaseHost:     databaseHost,
		DatabasePort:     databasePort,
		DatabaseName:     databaseName,
		DatabaseUser:     databaseUser,
		DatabasePassword: databasePassword,
		DatabaseSSLMode:  databaseSSLMode,
		DatabaseSchema:   databaseSchema,
		DatabaseTimezone: databaseTimezone,
		JWTSecret:        jwtSecret,
		MigrationDir:     migrationDir,
		RequestTimeoutS:  30,
	}
}

func (config Config) Validate() error {
	if config.DatabaseHost == "" {
		return fmt.Errorf("DATABASE_HOST is required")
	}
	if config.DatabasePort == "" {
		return fmt.Errorf("DATABASE_PORT is required")
	}
	if config.DatabaseName == "" {
		return fmt.Errorf("DATABASE_NAME is required")
	}
	if config.DatabaseUser == "" {
		return fmt.Errorf("DATABASE_USERNAME is required")
	}
	if config.DatabasePassword == "" {
		return fmt.Errorf("DATABASE_PASSWORD is required")
	}
	if config.DatabaseTimezone == "" {
		return fmt.Errorf("DATABASE_TIMEZONE is required")
	}
	return nil
}

func (config Config) DatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s search_path=%s TimeZone=%s",
		config.DatabaseHost,
		config.DatabasePort,
		config.DatabaseName,
		config.DatabaseUser,
		config.DatabasePassword,
		config.DatabaseSSLMode,
		config.DatabaseSchema,
		config.DatabaseTimezone,
	)
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		_ = os.Setenv(key, value)
	}

	return scanner.Err()
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
