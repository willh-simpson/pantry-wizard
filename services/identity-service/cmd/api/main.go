package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"
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

	handler := api.NewIdentityHandler()

	r := gin.Default()
	r.GET("/health", handler.HealthCheck)

	log.Printf("identity service starting on port %s...", cfg.Port)
	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
