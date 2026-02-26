package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

func (h *RecommendationHandler) GetTopRecommendations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	recipes, err := database.GetTopRecipes(c.Request.Context(), h.DB, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch recommendations",
		})

		return
	}

	c.JSON(http.StatusOK, recipes)
}

func (h *RecommendationHandler) ProcessInteraction(ctx context.Context, msg kafka.Message) error {
	var event model.InteractionEvent

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	log.Printf("incrementing popularity score for recipe \"%s\" via %s", event.RecipeID, event.Action)

	err := database.UpdateScore(ctx, h.DB, event.RecipeID, string(event.Action))
	if err != nil {
		log.Printf("could not process interaction for recipe \"%s\": %v", event.RecipeID, err)
	}

	return err
}
