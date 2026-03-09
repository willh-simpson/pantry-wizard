package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/events"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/kafka"
	"github.com/willh-simpson/pantry-wizard/services/user-service/domain/database"
	"github.com/willh-simpson/pantry-wizard/services/user-service/domain/model"
)

type UserHandler struct {
	DB *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{
		DB: db,
	}
}

func (h *UserHandler) ProcessUserEvent(ctx context.Context, msg kafka.Message) error {
	var event events.UserSyncedEvent

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	log.Printf("syncing user with name '%s'", event.DisplayName)

	err := database.UpsertUser(h.DB, ctx, event.ExternalID, event.Email, event.DisplayName)
	if err != nil {
		log.Printf("could not sync user with name '%s': %v", event.DisplayName, err)
	}

	return err
}

func (h *UserHandler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":   "up",
		"service":  "user-service",
		"database": "connected",
	})
}

func (h *UserHandler) GetUserInventory(c *gin.Context) {
	userID, exists := c.Get("user_external_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user identity not found",
		})

		return
	}

	inventory, err := database.GetFullInventory(h.DB, c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to fetch user inventory",
			"error":   err.Error(),
		})
	}

	c.JSON(http.StatusOK, inventory)
}

func (h *UserHandler) AddToPantry(c *gin.Context) {
	h.handleListUpdate(c, "user_pantry", true)
}

func (h *UserHandler) RemoveFromPantry(c *gin.Context) {
	h.handleListUpdate(c, "user_pantry", false)
}

func (h *UserHandler) AddToShoppingList(c *gin.Context) {
	h.handleListUpdate(c, "user_shopping_list", true)
}

func (h *UserHandler) RemoveFromShoppingList(c *gin.Context) {
	h.handleListUpdate(c, "user_shopping_list", false)
}

func (h *UserHandler) AddToWishlist(c *gin.Context) {
	h.handleListUpdate(c, "user_wishlist", true)
}

func (h *UserHandler) RemoveFromWishlist(c *gin.Context) {
	h.handleListUpdate(c, "user_wishlist", false)
}

func (h *UserHandler) MoveToPantry(c *gin.Context) {
	externalID := c.MustGet("user_external_id").(string)

	internalID, err := database.GetInternalID(h.DB, c.Request.Context(), externalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "user not found in local records",
			"error":   err.Error(),
		})

		return
	}

	var req model.UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
			"error":   err.Error(),
		})

		return
	}

	err = database.MoveToPantry(h.DB, c.Request.Context(), internalID, req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to move items to pantry",
			"error":   err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "items moved to pantry successfully",
	})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	externalID := c.MustGet("user_external_id").(string)

	user, err := database.GetUserByExternalID(h.DB, c.Request.Context(), externalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "profile not found",
			"error":   err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdatePreferences(c *gin.Context) {
	externalID := c.MustGet("user_external_id").(string)

	var input model.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid input",
			"error":   err.Error(),
		})

		return
	}

	err := database.UpdateUserPreferences(h.DB, c.Request.Context(), externalID, input.DietaryFlags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not update user preferences",
			"error":   err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user preferences updated successfully",
	})
}

func (h *UserHandler) GetShoppingListSuggestions(c *gin.Context) {
	externalID := c.MustGet("user_external_id").(string)

	userID, err := h.getInternalID(c.Request.Context(), externalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "user not found in internal records",
			"error":   err.Error(),
		})
	}

	suggestions, err := database.GetShoppingListSuggestions(h.DB, c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get shopping list suggestions",
			"error":   err.Error(),
		})

		return
	}

	if suggestions == nil {
		suggestions = []model.ShoppingListSuggestion{} // return empty array instead of nil
	}

	c.JSON(http.StatusOK, gin.H{
		"shopping_list_suggestions": suggestions,
	})
}

func (h *UserHandler) ProcessRecipeCookedEvent(ctx context.Context, msg kafka.Message) error {
	var event events.RecipeCookedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	log.Printf("processing meal execution for user %s, recipe %s", event.ExternalID, event.RecipeID)

	internalID, err := h.getInternalID(ctx, event.ExternalID)
	if err != nil {
		return fmt.Errorf("user not found for event: %w", err)
	}

	err = database.RecordMealExecution(h.DB, ctx, internalID, event.RecipeID, event.Ingredients)
	if err != nil {
		log.Printf("error recording meal execution: %v", err)

		return err
	}

	return nil
}

func (h *UserHandler) getInternalID(ctx context.Context, externalID string) (string, error) {
	return database.GetInternalID(h.DB, ctx, externalID)
}

func (h *UserHandler) handleListUpdate(c *gin.Context, table string, isAddOperation bool) {
	externalID := c.MustGet("user_external_id").(string)

	internalID, err := h.getInternalID(c.Request.Context(), externalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "user not found in local records",
			"error":   err.Error(),
		})

		return
	}

	var req model.UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body. 'items []string' field is required",
		})

		return
	}

	var httpStatus int
	if isAddOperation {
		err = database.AddItemsToList(h.DB, c.Request.Context(), table, internalID, req.Items)
		httpStatus = http.StatusCreated
	} else {
		err = database.RemoveItemsFromList(h.DB, c.Request.Context(), table, internalID, req.Items)
		httpStatus = http.StatusOK
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not modify items in " + table,
			"error":   err.Error(),
		})

		return
	}

	c.JSON(httpStatus, gin.H{
		"message": "transaction completed successfully",
	})
}
