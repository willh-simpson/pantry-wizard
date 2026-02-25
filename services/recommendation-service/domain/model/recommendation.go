package model

type LikeEvent struct {
	RecipeID  string     `json:"recipe_id"`
	UserID    string     `json:"user_id"`
	Timestamp int64      `json:"timestamp"`
	Action    LikeAction `json:"action"`
}

type LikeAction string

const (
	Like   LikeAction = "like"
	Unlike LikeAction = "unlike"
)
