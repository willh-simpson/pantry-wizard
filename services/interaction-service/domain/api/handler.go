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

func (h *InteractionHandler) LikeRecipe(c *gin.Context) {
	var req model.LikeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	if err := database.SaveLike(c.Request.Context(), h.DB, req.UserID, req.RecipeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to record like",
		})

		return
	}

	event := model.LikeEvent{
		RecipeID:  req.RecipeID,
		UserID:    req.UserID,
		Timestamp: time.Now().Unix(),
		Action:    model.Like,
	}
	payload, _ := json.Marshal(event)

	err := h.Producer.Publish(c.Request.Context(), kafka.Message{
		Topic:      "recipe-likes",
		Key:        []byte(req.RecipeID),
		Value:      payload,
		RetryCount: 0,
	})

	if err != nil {
		fmt.Printf("kafka publish error: %v\n", err)
	} else {
		log.Println("published like to topic \"recipe-likes\"")
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "recipe liked",
	})
}
