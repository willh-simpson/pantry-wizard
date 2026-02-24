package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"github.com/willh-simpson/pantry-wizard/services/interaction-service/domain/database"
)

func main() {
	forceVersion := flag.Int("force", -1, "force a specific migration version to clear dirty state")
	flag.Parse()

	dbUser := os.Getenv("INTERACTION_DB_USER")
	dbPass := os.Getenv("INTERACTION_DB_PASSWORD")
	dbHost := os.Getenv("INTERACTION_DB_HOST")
	dbPort := os.Getenv("INTERACTION_DB_PORT")
	dbName := "interaction_db"

	if dbHost == "" {
		dbHost = "localhost" // fallback if system is running locally instead of via docker
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

	dsn := u.String()

	kafkaBroker := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}

	// if db is marked dirty and in maintenance mode then this will force migration without manually entering db
	if *forceVersion != -1 {
		log.Printf("maintenance: forcing database version to %d...", *forceVersion)

		if err := database.ForceMigration(dsn, *forceVersion); err != nil {
			log.Fatalf("force migration failed: %v", err)
		}

		log.Println("force migration successful. exiting")

		return // don't start server normally after forcing migration
	}

	migrationPath := "migrations"
	if err := database.RunMigrations(dsn, migrationPath); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	ginRouter := gin.Default()
	ginRouter.GET("/health", func(c *gin.Context) {
		dialer := &kafka.Dialer{
			Timeout:   5 * time.Second,
			DualStack: true,
		}
		conn, err := dialer.DialContext(c, "tcp", kafkaBroker)

		kafkaStatus := "connected"
		if err != nil {
			kafkaStatus = fmt.Sprintf("unreachable: %v", err)
		} else {
			conn.Close()
		}

		c.JSON(200, gin.H{
			"status":   "up",
			"service":  "interaction-service",
			"database": "connected",
			"kafka":    kafkaStatus,
		})
	})

	log.Printf("interaction service starting on port 8082...")
	if err := ginRouter.Run(":8082"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
