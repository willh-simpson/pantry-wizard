package events

import "time"

type RecipeLikedV1 struct {
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user-id"`
	RecipeID  string    `json:"recipe-id"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  Metadata  `json:"metadata"`
}

type Metadata struct {
	Source    string `json:"source"`
	SessionID string `json:"session_id"`
}
