package events

import "time"

type RecipeLikedV1 struct {
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	RecipeID  string    `json:"recipe_id"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  Metadata  `json:"metadata"`
}

type Metadata struct {
	Source    string `json:"source"`
	SessionID string `json:"session_id"`
}

type UserSyncedEvent struct {
	ExternalID  string `json:"external_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Action      string `json:"action"`
}

type UserSyncAction string

const (
	UserCreated UserSyncAction = "CREATED"
	UserUpdated UserSyncAction = "UPDATED"
)
