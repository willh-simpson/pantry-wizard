package main

import (
	"database/sql"
	"flag"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/auth"
	"github.com/willh-simpson/pantry-wizard/services/identity-service/auth/client"
	"github.com/willh-simpson/pantry-wizard/services/identity-service/config"
	"github.com/willh-simpson/pantry-wizard/services/identity-service/domain/api"
	"github.com/willh-simpson/pantry-wizard/services/identity-service/domain/database"
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

	validator := auth.NewTokenValidator(cfg.AWSRegion, cfg.CognitoPoolID)
	if validator == nil {
		log.Printf("validator is nil")
	} else {
		log.Printf("validator started: JWKS_URL: %s", validator.JWKS_URL)
	}
	cognitoClient, err := client.NewCognitoClient(cfg.AWSRegion, cfg.CognitoAppID, cfg.CognitoPoolID)
	if err != nil {
		log.Fatalf("failed to initialize cognito client: %v", err)
	}

	handler := api.NewIdentityHandler(db, cognitoClient)

	r := gin.Default()
	r.GET("/health", handler.HealthCheck)
	r.POST("/auth/register", handler.Register)
	r.POST("/auth/confirm", handler.ConfirmRegistration)
	r.POST("/auth/login", handler.Login)

	userRoutes := r.Group("/users")
	userRoutes.Use(validator.AuthWorker(validator.JWKS_URL))
	{
		userRoutes.GET("/profile", handler.GetUserProfile)
	}

	log.Printf("identity service starting on port %s...", cfg.Port)
	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
