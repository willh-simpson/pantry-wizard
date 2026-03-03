package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	DB_DSN         string
	UserServiceURL string
	Port           string
	CognitoAppID   string
	CognitoPoolID  string
	AWSRegion      string
}

func LoadConfig() *Config {
	dbUser := os.Getenv("RECIPE_DB_USER")
	dbPass := os.Getenv("RECIPE_DB_PASSWORD")
	dbHost := os.Getenv("RECIPE_DB_HOST")
	dbPort := os.Getenv("RECIPE_DB_PORT")
	dbName := "recipe_db"

	userServiceHost := os.Getenv("USER_SERVICE_HOST")
	if userServiceHost == "" {
		userServiceHost = "localhost"
	}

	userServicePort := os.Getenv("USER_SERVICE_PORT")
	userServiceURL := fmt.Sprintf("http://%s:%s", userServiceHost, userServicePort)

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

	appID := os.Getenv("COGNITO_CLIENT_ID")
	poolID := os.Getenv("COGNITO_USER_POOL_ID")
	awsRegion := os.Getenv("AWS_REGION")

	return &Config{
		DB_DSN:         u.String(),
		UserServiceURL: userServiceURL,
		Port:           ":8083",
		CognitoAppID:   appID,
		CognitoPoolID:  poolID,
		AWSRegion:      awsRegion,
	}
}
