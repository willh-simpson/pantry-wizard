package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/events"
	"github.com/willh-simpson/pantry-wizard/libs/go/common/kafka"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/admin/client"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/admin/ingest"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/domain/database"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/domain/model"
)

type RecipeHandler struct {
	DB         *sql.DB
	UserClient *client.UserClient
}

var MEAL_DB_URI = "https://www.themealdb.com/api/json/v1/1/search.php?s="

func NewRecipeHandler(db *sql.DB, userClient client.UserClient) *RecipeHandler {
	return &RecipeHandler{
		DB:         db,
		UserClient: &userClient,
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

func (h *RecipeHandler) GetRecipe(c *gin.Context) {
	recipeID := c.Param("recipe_id")

	recipe, err := database.GetRecipeByID(h.DB, c.Request.Context(), recipeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "recipe not found",
			"error":   err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, recipe)
}

func (h *RecipeHandler) AdminIngest(c *gin.Context) {
	searchQuery := c.DefaultQuery("s", "chicken")
	res, err := http.Get(MEAL_DB_URI + searchQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "API fetch failed",
		})

		return
	}
	defer res.Body.Close()

	var data ingest.MealDBResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		log.Printf("JSON decode error: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to decode API response",
		})

		return
	}
	if data.Meals == nil {
		log.Printf("no meals found for query: %s", searchQuery)

		c.JSON(http.StatusOK, gin.H{
			"message":          "no recipes found",
			"recipes_imported": 0,
		})

		return
	}

	count := 0
	for _, meal := range data.Meals {
		if err := database.IngredientFromMealDB(h.DB, c.Request.Context(), meal); err != nil {
			log.Printf("failed to ingest meal %v: %v", meal["strMeal"], err)
		}

		count++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "ingestion complete",
		"recipes_imported": count,
	})
}

func (h *RecipeHandler) SearchRecipes(c *gin.Context) {
	mode := c.DefaultQuery("mode", model.PantryMode)
	strictness := c.DefaultQuery("strictness", model.UnrestrictedSearch)

	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "'user_id' is a required field",
		})

		return
	}

	var inventory model.UserInventory
	inventory, err := h.UserClient.FetchUserInventory(c.Request.Context(), userID)
	if err != nil {
		log.Printf("failed to fetch user inventory for user %s: %v", userID, err)

		inventory = model.UserInventory{
			Pantry:       []string{},
			ShoppingList: []string{},
			Wishlist:     []string{},
		}
	}

	log.Printf("DEBUG: received pantry: %v", inventory.Pantry)
	log.Printf("DEBUG: received wishlist: %v", inventory.Wishlist)
	log.Printf("DEBUG: search mode: %s", c.Query("mode"))

	var results []model.SearchResult
	if mode == model.ShoppingListMode {
		results, err = database.AdvancedShoppingListSearch(
			h.DB,
			c.Request.Context(),
			inventory.ShoppingList,
			inventory.Wishlist,
			strictness,
		)
	} else {
		results, err = database.AdvancedPantrySearch(
			h.DB, c.Request.Context(),
			inventory.Pantry,
			inventory.Wishlist,
			strictness,
		)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "search failed",
			"error":   err.Error(),
		})

		return
	}

	for i := range results {
		ratio := float64(results[i].MatchCount) / float64(results[i].TotalNeeded)

		results[i].MatchRatio = math.Round(ratio*100) / 100 // round to two decimal places for cleaner ratio formatting
		results[i].Stars = calculateStars(ratio)
	}

	c.JSON(http.StatusOK, results)
}

func (h *RecipeHandler) ProcessRecipeCookedEvent(ctx context.Context, msg kafka.Message) error {
	var event events.RecipeCookedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	log.Printf("UPDATE: recipe %s cooked by a user", event.RecipeID)

	return database.IncrementTimesMadeGlobally(h.DB, ctx, event.RecipeID)
}

func calculateStars(ratio float64) int {
	switch {
	case ratio >= 1.0:
		return 5
	case ratio >= 0.8:
		return 4
	case ratio >= 0.6:
		return 3
	case ratio >= 0.4:
		return 2
	case ratio >= 0.1:
		return 1
	default:
		return 0
	}
}
