package model

type InteractionRequest struct {
	RecipeID string          `json:"recipe_id" binding:"required,uuid"`
	UserID   string          `json:"user_id" binding:"required,uuid"`
	Action   InteractionType `json:"action" binding:"required"`
}

type InteractionEvent struct {
	RecipeID  string          `json:"recipe_id"`
	UserID    string          `json:"user_id"`
	Timestamp int64           `json:"timestamp"`
	Action    InteractionType `json:"action"`
}

type InteractionType string

const (
	View   InteractionType = "view"
	Like   InteractionType = "like"
	Unlike InteractionType = "unlike"
	Save   InteractionType = "save"
	Unsave InteractionType = "unsave"
)
