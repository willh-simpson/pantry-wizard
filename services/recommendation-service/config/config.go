package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	DB_DSN      string
	Port        string
	KafkaBroker string
	Environment string
}

func LoadConfig() *Config {
	dbUser := os.Getenv("RECOMMENDATION_DB_USER")
	dbPass := os.Getenv("RECOMMENDATION_DB_PASSWORD")
	dbHost := os.Getenv("RECOMMENDATION_DB_HOST")
	dbPort := os.Getenv("RECOMMENDATION_DB_PORT")
	dbName := "recommendation_db"

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

	kafkaBroker := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}

	return &Config{
		DB_DSN:      u.String(),
		Port:        ":8084",
		KafkaBroker: kafkaBroker,
	}
}
