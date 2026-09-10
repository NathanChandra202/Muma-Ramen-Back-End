package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port      string
	JWTSecret string
	// Database
	DBDriver   string // "sqlite" atau "postgres"
	DBPath     string // untuk SQLite
	DBHost     string // untuk PostgreSQL
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// Load membaca file .env (jika ada) lalu mengambil nilai dari environment variables
func Load() *Config {
	loadDotEnv(".env")

	port := getEnv("PORT", "8081")
	jwtSecret := getEnv("JWT_SECRET", "muma-ramen-secret-key-2024")
	dbDriver := getEnv("DB_DRIVER", "sqlite")

	return &Config{
		Port:       port,
		JWTSecret:  jwtSecret,
		DBDriver:   dbDriver,
		DBPath:     getEnv("DB_PATH", "muma_ramen.db"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "muma_ramen"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// PostgresDSN mengembalikan connection string untuk PostgreSQL
func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// getEnv mengambil nilai env var, dengan fallback jika kosong
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// loadDotEnv membaca file .env dan set environment variables
func loadDotEnv(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		return // file .env tidak wajib ada
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip komentar dan baris kosong
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Hanya set jika belum ada di environment (env var sistem lebih prioritas)
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
