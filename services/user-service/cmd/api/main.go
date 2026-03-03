package main

import (
	"database/sql"
	"flag"
	"log"

	"github.com/gin-gonic/gin"
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

	handler := api.NewUserHandler(db)

	r := gin.Default()
	r.GET("/health", handler.HealthCheck)

	r.GET("/users/:user_id/inventory", handler.GetUserInventory)

	r.POST("/users/:user_id/pantry/add", handler.AddToPantry)
	r.DELETE("/users/:user_id/pantry/remove", handler.RemoveFromPantry)
	r.POST("/users/:user_id/pantry/move", handler.MoveToPantry)

	r.POST("/users/:user_id/shopping-list/add", handler.AddToShoppingList)
	r.DELETE("/users/:user_id/shopping-list/remove", handler.RemoveFromShoppingList)

	r.POST("/users/:user_id/wishlist/add", handler.AddToWishlist)
	r.DELETE("/users/:user_id/wishlist/remove", handler.RemoveFromWishlist)

	log.Printf("user service starting on port %s...", cfg.Port)
	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
