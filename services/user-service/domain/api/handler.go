package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
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

func (h *UserHandler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":   "up",
		"service":  "user-service",
		"database": "connected",
	})
}

func (h *UserHandler) GetUserInventory(c *gin.Context) {
	userID := c.Param("user_id")

	inventory, err := database.GetFullInventory(h.DB, c.Request.Context(), userID)
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
	userID := c.Param("user_id")

	var req model.UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
			"error":   err.Error(),
		})

		return
	}

	err := database.MoveToPantry(h.DB, c.Request.Context(), userID, req.Items)
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

func (h *UserHandler) handleListUpdate(c *gin.Context, table string, isAddOperation bool) {
	userID := c.Param("user_id")

	var req model.UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body. 'items []string' field is required",
		})

		return
	}

	var err error
	var httpStatus int
	if isAddOperation {
		err = database.AddItemsToList(h.DB, c.Request.Context(), table, userID, req.Items)
		httpStatus = http.StatusCreated
	} else {
		err = database.RemoveItemsFromList(h.DB, c.Request.Context(), table, userID, req.Items)
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
