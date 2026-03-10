package main

import (
	"context"
	"database/sql"
	"flag"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/auth"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/kafka"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/admin/client"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/config"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/domain/api"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/domain/database"
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

	kafkaProducer := kafka.NewProducer([]string{cfg.KafkaBroker})
	defer kafkaProducer.Close()

	retryProducer := kafka.NewProducer([]string{cfg.KafkaBroker})
	defer retryProducer.Close()

	userClient := *client.NewUserClient(cfg.UserServiceURL)
	handler := api.NewRecipeHandler(db, kafkaProducer, userClient)
	validator := auth.NewTokenValidator(cfg.AWSRegion, cfg.CognitoPoolID)

	r := gin.Default()
	r.GET("/health", handler.HealthCheck)
	r.GET("/admin/ingest", handler.AdminIngest)

	protected := r.Group("/recipes")
	protected.Use(validator.AuthWorker(validator.JWKS_URL))
	{
		protected.GET("/search", handler.SearchRecipes)
		protected.GET("/", handler.ListRecipes)
		protected.GET("/:recipe_id", handler.GetRecipe)
		protected.POST("/", handler.CreateRecipe)
	}

	log.Printf("identity service starting on port %s...", cfg.Port)

	go func() {
		kafkaConsumer := kafka.NewConsumer(
			[]string{cfg.KafkaBroker},
			"recipe-service-stats-group",
			"recipe-cook-interactions",
			retryProducer,
		)
		defer kafkaConsumer.Close()

		log.Println("recipe service listening for recipe stats...")

		kafkaConsumer.Consume(context.Background(), handler.ProcessRecipeCookedEvent)
	}()

	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
