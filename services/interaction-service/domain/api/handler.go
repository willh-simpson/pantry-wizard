package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/events"
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

func (h *InteractionHandler) Interact(c *gin.Context) {
	externalID := c.MustGet("user_external_id").(string)
	recipeID := c.Param("recipe_id")

	internalID, err := h.getInternalID(c.Request.Context(), externalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "user sync error",
			"error":   err.Error(),
		})

		return
	}

	var req model.InteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	if err := database.HandleInteraction(h.DB, c.Request.Context(), internalID, recipeID, model.InteractionType(req.Action)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to record " + req.Action,
		})

		return
	}

	h.publishInteraction(c, recipeID, externalID, string(req.Action))

	c.JSON(http.StatusCreated, gin.H{
		"message": "saved " + req.Action,
	})
}

func (h *InteractionHandler) CookRecipe(c *gin.Context) {
	externalID := c.MustGet("user_external_id").(string)
	recipeID := c.Param("recipe_id")

	internalID, err := h.getInternalID(c.Request.Context(), externalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "user sync error",
			"error":   err.Error(),
		})

		return
	}

	var req model.RecipeCookedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ingredients list required",
		})

		return
	}

	err = database.HandleInteraction(h.DB, c.Request.Context(), internalID, recipeID, model.Cook)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not process cook interaction",
			"error":   err.Error(),
		})

		return
	}

	// because logic is handled different across services based on cooking a recipe, 2 different events need to be published
	h.publishInteraction(c, recipeID, externalID, string(model.Cook)) // recipe-service does not have specific logic for cooking
	h.publishCookEvent(c, externalID, recipeID, req.Ingredients)      // user-service does not intercept logic for views/likes/saves

	c.JSON(http.StatusOK, gin.H{
		"message": "recipe execution recorded",
	})
}

func (h *InteractionHandler) ProcessUserEvent(ctx context.Context, msg kafka.Message) error {
	var event events.UserSyncedEvent
	json.Unmarshal(msg.Value, &event)

	log.Printf("syncing user with name '%s'", event.DisplayName)

	err := database.UpsertUser(h.DB, ctx, event.ExternalID)

	return err
}

func (h *InteractionHandler) getInternalID(ctx context.Context, externalID string) (string, error) {
	return database.GetInternalID(h.DB, ctx, externalID)
}

func (h *InteractionHandler) publishInteraction(c *gin.Context, recipeID, externalID, action string) {
	event := model.InteractionEvent{
		RecipeID:   recipeID,
		ExternalID: externalID,
		Timestamp:  time.Now().Unix(),
		Action:     model.InteractionType(action),
	}
	payload, _ := json.Marshal(event)

	err := h.Producer.Publish(c.Request.Context(), kafka.Message{
		Topic:      "interactions.recipes",
		Key:        []byte(recipeID),
		Value:      payload,
		RetryCount: 0,
	})

	if err != nil {
		fmt.Printf("kafka publish error: %v", err)
	} else {
		log.Printf("published %s to topic \"interactions.recipe\"", action)
	}
}

func (h *InteractionHandler) publishCookEvent(c *gin.Context, externalID, recipeID string, ingredients []string) {
	event := events.RecipeCookedEvent{
		ExternalID:  externalID,
		RecipeID:    recipeID,
		Ingredients: ingredients,
		CookedAt:    time.Now(),
	}

	payload, _ := json.Marshal(event)

	h.Producer.Publish(c.Request.Context(), kafka.Message{
		Topic: "interactions.recipes.cooked",
		Key:   []byte(externalID),
		Value: payload,
	})
}
