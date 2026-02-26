package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/kafka"
	"github.com/willh-simpson/pantry-wizard/services/interaction-service/domain/database"
	"github.com/willh-simpson/pantry-wizard/services/interaction-service/domain/model"
)

type InteractionHandler struct {
	DB       *sql.DB
	Producer kafka.Producer
}

func NewInteractionHandler(db *sql.DB, prod kafka.Producer) *InteractionHandler {
	return &InteractionHandler{
		DB:       db,
		Producer: prod,
	}
}

func (h *InteractionHandler) HealthCheck(c *gin.Context) {
	err := h.Producer.Ping(c.Request.Context())

	kafkaStatus := "connected"
	if err != nil {
		kafkaStatus = fmt.Sprintf("unreachable: %v", err)
	}

	c.JSON(200, gin.H{
		"status":   "up",
		"service":  "interaction-service",
		"database": "connected",
		"kafka":    kafkaStatus,
	})
}

func (h *InteractionHandler) PublishInteraction(c *gin.Context, recipeID, userID, action string) {
	event := model.InteractionEvent{
		RecipeID:  recipeID,
		UserID:    userID,
		Timestamp: time.Now().Unix(),
		Action:    model.InteractionType(action),
	}
	payload, _ := json.Marshal(event)

	err := h.Producer.Publish(c.Request.Context(), kafka.Message{
		Topic:      "recipe-interactions",
		Key:        []byte(recipeID),
		Value:      payload,
		RetryCount: 0,
	})

	if err != nil {
		fmt.Printf("kafka publish error: %v\n", err)
	} else {
		log.Printf("published %s to topic \"recipe-interactions\"", action)
	}
}

func (h *InteractionHandler) Interact(c *gin.Context) {
	var req model.InteractionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	if err := database.HandleInteraction(c.Request.Context(), h.DB, req.UserID, req.RecipeID, string(req.Action)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to record " + req.Action,
		})

		return
	}

	h.PublishInteraction(c, req.RecipeID, req.UserID, string(req.Action))

	c.JSON(http.StatusCreated, gin.H{
		"message": "saved " + req.Action,
	})
}
