package model

import "time"

type InteractionRequest struct {
	Action InteractionType `json:"action" binding:"required"`
}

type InteractionEvent struct {
	RecipeID   string          `json:"recipe_id"`
	ExternalID string          `json:"external_id"`
	Timestamp  int64           `json:"timestamp"`
	Action     InteractionType `json:"action"`
}

type InteractionType string

const (
	View   InteractionType = "view"
	Like   InteractionType = "like"
	Unlike InteractionType = "unlike"
	Save   InteractionType = "save"
	Unsave InteractionType = "unsave"
	Cook   InteractionType = "cook"
)

type RecipeCookedRequest struct {
	Ingredients []string `json:"ingredients" binding:"required"`
}

type RecipeCookedEvent struct {
	ExternalID  string    `json:"external_id"`
	RecipeID    string    `json:"recipe_id"`
	Ingredients []string  `json:"ingredients"`
	CookedAt    time.Time `json:"cooked_at"`
}
