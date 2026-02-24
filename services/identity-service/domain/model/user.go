package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	ExternalID   string    `json:"external_id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	DietaryFlags []string  `json:"dietary_flags"`
	CreatedAt    time.Time `json:"created_at"`
}
