package model

import "time"

type User struct {
	ExternalID   string    `json:"external_id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	DietaryFlags []string  `json:"dietary_flags"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserInventory struct {
	UserID       string   `json:"user_id"`
	Pantry       []string `json:"pantry"`
	ShoppingList []string `json:"shopping_list"`
	Wishlist     []string `json:"wishlist"`
}

type ShoppingListSuggestion struct {
	IngredientName string    `json:"ingredient_name"`
	Reason         string    `json:"reason"`
	SuggestedAt    time.Time `json:"suggested_at"`
}

type UpdateInventoryRequest struct {
	Items []string `json:"items" binding:"required"`
}

type UpdatePreferencesRequest struct {
	DietaryFlags []string `json:"dietary_flags" binding:"required"`
}
