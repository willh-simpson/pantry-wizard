package main

import (
	"context"
	"database/sql"
	"flag"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/kafka"
	"github.com/willh-simpson/pantry-wizard/services/recommendation-service/config"
	"github.com/willh-simpson/pantry-wizard/services/recommendation-service/domain/api"
	"github.com/willh-simpson/pantry-wizard/services/recommendation-service/domain/database"
)

func main() {
	forceVersion := flag.Int("force", -1, "force a specific migration version to clear dirty state")
	flag.Parse()

	cfg := config.LoadConfig()

	if *forceVersion != -1 {
		log.Printf("maintenance: forcing database version to %d...", *forceVersion)

		if err := database.ForceMigration(cfg.DB_DSN, *forceVersion); err != nil {
			log.Fatalf("force migration failed: %v", err)
		}

		log.Println("force migration successful. exiting")

		return
	}

	if err := database.RunMigrations(cfg.DB_DSN, "migrations"); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DB_DSN)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	retryProducer := kafka.NewProducer([]string{cfg.KafkaBroker})
	defer retryProducer.Close()

	consumer := kafka.NewConsumer(
		[]string{cfg.KafkaBroker},
		"recommendation-service-group",
		"recipe-likes",
		retryProducer,
	)
	defer consumer.Close()

	handler := api.NewRecommendationHandler(db, retryProducer)

	r := gin.Default()
	r.GET("/health", handler.HealthCheck)

	log.Printf("recommendation service starting on port %s...", cfg.Port)

	go func() {
		log.Println("recommendation service listening for events...")
		consumer.Consume(context.Background(), handler.ProcessLike)
	}()

	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
