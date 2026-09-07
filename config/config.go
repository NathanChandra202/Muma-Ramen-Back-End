package config

import "os"

type Config struct {
	Port      string
	JWTSecret string
	DBPath    string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "muma-ramen-secret-key-2024"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "muma_ramen.db"
	}

	return &Config{
		Port:      port,
		JWTSecret: jwtSecret,
		DBPath:    dbPath,
	}
}
