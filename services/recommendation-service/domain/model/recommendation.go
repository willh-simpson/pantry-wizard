package model

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

type RankedRecipe struct {
	RecipeID   string  `json:"recipe_id"`
	TotalScore float64 `json:"total_score"`
}
