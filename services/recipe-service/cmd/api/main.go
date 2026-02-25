package main

import (
	"database/sql"
	"flag"
	"log"

	"github.com/gin-gonic/gin"
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

	handler := api.NewRecipeHandler(db)

	r := gin.Default()
	r.GET("/health", handler.HealthCheck)
	r.GET("/recipes", handler.ListRecipes)
	r.POST("/recipes", handler.CreateRecipe)

	log.Printf("identity service starting on port %s...", cfg.Port)
	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
