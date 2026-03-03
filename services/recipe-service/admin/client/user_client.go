package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/willh-simpson/pantry-wizard/services/recipe-service/domain/model"
)

type UserClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewUserClient(baseURL string) *UserClient {
	return &UserClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: time.Second * 5,
		},
	}
}

func (c *UserClient) FetchUserInventory(ctx context.Context, userID string) (model.UserInventory, error) {
	url := fmt.Sprintf("%s/users/%s/inventory", c.BaseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return model.UserInventory{}, fmt.Errorf("could not form HTTP request: %w", err)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return model.UserInventory{}, fmt.Errorf("user service unreachable: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return model.UserInventory{}, fmt.Errorf("user service returned status: %d", res.StatusCode)
	}

	var inventory model.UserInventory
	if err := json.NewDecoder(res.Body).Decode(&inventory); err != nil {
		return model.UserInventory{}, fmt.Errorf("could not decode json response: %w", err)
	}

	return inventory, nil
}
