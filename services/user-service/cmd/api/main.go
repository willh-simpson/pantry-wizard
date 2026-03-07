package main

import (
	"context"
	"database/sql"
	"flag"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/auth"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/kafka"
	"github.com/willh-simpson/pantry-wizard/services/user-service/config"
	"github.com/willh-simpson/pantry-wizard/services/user-service/domain/api"
	"github.com/willh-simpson/pantry-wizard/services/user-service/domain/database"
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

	kafkaConsumer := kafka.NewConsumer(
		[]string{cfg.KafkaBroker},
		"user-service-group",
		"user-events",
		retryProducer,
	)
	defer kafkaConsumer.Close()

	validator := auth.NewTokenValidator(cfg.AWSRegion, cfg.CognitoPoolID)
	if validator == nil {
		log.Printf("validator is nil")
	} else {
		log.Printf("validator started: JWKS_URL: %s", validator.JWKS_URL)
	}

	handler := api.NewUserHandler(db)

	r := gin.Default()
	r.GET("/health", handler.HealthCheck)

	userGroup := r.Group("/me")
	userGroup.Use(validator.AuthWorker(validator.JWKS_URL))
	{
		userGroup.GET("/inventory", handler.GetUserInventory)

		userGroup.POST("/pantry/add", handler.AddToPantry)
		userGroup.DELETE("/pantry/remove", handler.RemoveFromPantry)
		userGroup.POST("/pantry/move", handler.MoveToPantry)

		userGroup.POST("/shopping-list/add", handler.AddToShoppingList)
		userGroup.DELETE("/shopping-list/remove", handler.RemoveFromShoppingList)

		userGroup.POST("/wishlist/add", handler.AddToWishlist)
		userGroup.DELETE("/wishlist/remove", handler.RemoveFromWishlist)

		userGroup.GET("/profile", handler.GetProfile)
		userGroup.PUT("/profile/preferences", handler.UpdatePreferences)
	}

	log.Printf("user service starting on port %s...", cfg.Port)

	go func() {
		log.Println("user service listening for events...")

		kafkaConsumer.Consume(context.Background(), handler.ProcessUserEvent)
	}()

	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
