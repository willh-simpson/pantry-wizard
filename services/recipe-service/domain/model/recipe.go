package model

import "time"

type IngredientInput struct {
	Name     string  `json:"name" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	Unit     string  `json:"unit"`
	Category string  `json:"category"`
}

type CreateRecipeRequest struct {
	Title        string            `json:"title" binding:"required"`
	Description  string            `json:"description"`
	Instructions string            `json:"instructions" binding:"required"`
	AuthorID     string            `json:"author_id" binding:"required"`
	PrepTime     int               `json:"prep_time_min"`
	Calories     int               `json:"calories"`
	BudgetTier   int               `json:"budget_tier" binding:"required,min=1,max=3"`
	Ingredients  []IngredientInput `json:"ingredients" binding:"required,dive"`
}

type RecipeResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}
