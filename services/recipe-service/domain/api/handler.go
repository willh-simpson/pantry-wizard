package api

import (
	"database/sql"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/domain/database"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/domain/model"
)

type RecipeHandler struct {
	DB *sql.DB
}

func NewRecipeHandler(db *sql.DB) *RecipeHandler {
	return &RecipeHandler{
		DB: db,
	}
}

func (h *RecipeHandler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":   "up",
		"service":  "recipe-service",
		"database": "connected",
	})
}

func (h *RecipeHandler) CreateRecipe(c *gin.Context) {
	var req model.CreateRecipeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})

		return
	}

	id, err := database.CreateFullRecipe(h.DB, req)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to save recipe",
		})

		return
	}

	c.JSON(201, gin.H{
		"id":      id,
		"message": "recipe created successfully",
	})
}

func (h *RecipeHandler) ListRecipes(c *gin.Context) {
	title := c.Query("title")
	maxBudget, _ := strconv.Atoi(c.DefaultQuery("budget", "0"))
	maxPrepTime, _ := strconv.Atoi(c.DefaultQuery("prep_time", "0"))

	recipes, err := database.SearchRecipes(h.DB, title, maxBudget, maxPrepTime)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to fetch recipes",
		})
	}

	c.JSON(200, recipes)
}
