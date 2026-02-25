package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	DB_DSN      string
	Port        string
	Environment string
}

func LoadConfig() *Config {
	dbUser := os.Getenv("IDENTITY_DB_USER")
	dbPass := os.Getenv("IDENTITY_DB_PASSWORD")
	dbHost := os.Getenv("IDENTITY_DB_HOST")
	dbPort := os.Getenv("IDENTITY_DB_PORT")
	dbName := "identity_db"

	if dbHost == "" {
		dbHost = "localhost"
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(dbUser, dbPass),
		Host:   fmt.Sprintf("%s:%s", dbHost, dbPort),
		Path:   dbName,
	}

	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()

	return &Config{
		DB_DSN: u.String(),
		Port:   ":8081",
	}
}
