package model

type UserInventory struct {
	UserID       string   `json:"user_id"`
	Pantry       []string `json:"pantry"`
	ShoppingList []string `json:"shopping_list"`
	Wishlist     []string `json:"wishlist"`
}

type UpdateInventoryRequest struct {
	Items []string `json:"items" binding:"required"`
}
