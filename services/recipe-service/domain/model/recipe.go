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

type Recipe struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	TimesMadeGlobally int64     `json:"times_made_globally"`
	Description       string    `json:"description"`
	Instructions      string    `json:"instructions"`
	AuthorID          string    `json:"author_id"`
	PrepTimeMinutes   string    `json:"prep_time_min"`
	Calories          int32     `json:"calories"`
	BudgetTier        int32     `json:"budget_tier"`
	ImageURL          string    `json:"image_url"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type SearchResult struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	MatchCount  int     `json:"match_count"`
	TotalNeeded int     `json:"total_needed"`
	MatchRatio  float64 `json:"match_ratio"`
	Stars       int     `json:"stars"`
}

const (
	PantryMode       string = "pantry"
	ShoppingListMode string = "shopping_list"

	StrictSearch       string = "strict"
	LessStrictSearch   string = "less_strict"
	UnrestrictedSearch string = "unrestricted"

	StrictThreshold      float64 = 1.0
	LessStrictThreshold  float64 = 0.7
	UnrestrictedTreshold float64 = 0.1
)
