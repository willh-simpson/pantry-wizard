package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/kafka"
	"github.com/willh-simpson/pantry-wizard/services/recommendation-service/domain/database"
	"github.com/willh-simpson/pantry-wizard/services/recommendation-service/domain/model"
)

type RecommendationHandler struct {
	DB       *sql.DB
	Producer kafka.Producer
}

func NewRecommendationHandler(db *sql.DB, prod kafka.Producer) *RecommendationHandler {
	return &RecommendationHandler{
		DB:       db,
		Producer: prod,
	}
}

func (h *RecommendationHandler) HealthCheck(c *gin.Context) {
	err := h.Producer.Ping(c.Request.Context())

	kafkaStatus := "connected"
	if err != nil {
		kafkaStatus = fmt.Sprintf("unreachable: %v", err)
	}

	c.JSON(200, gin.H{
		"status":   "up",
		"service":  "recommendation-service",
		"database": "connected",
		"kafka":    kafkaStatus,
	})
}

func (h *RecommendationHandler) ProcessLike(ctx context.Context, msg kafka.Message) error {
	var event model.LikeEvent

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	log.Printf("incrementing popularity score for recipe: %s", event.RecipeID)

	err := database.UpdatePopularity(ctx, h.DB, event.RecipeID)
	if err != nil {
		log.Printf("could not process like for recipe: %s", event.RecipeID)
	}

	return err
}
