package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/willh-simpson/pantry-wizard/services/user-service/domain/model"
	"golang.org/x/sync/errgroup"
)

type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func GetUserByExternalID(db *sql.DB, ctx context.Context, externalID string) (*model.User, error) {
	query, args, err := psql.
		Select("external_id", "email", "display_name", "dietary_flags", "created_at", "updated_at").
		From("users").
		Where("external_id = ?", externalID).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("could not build query: %w", err)
	}

	var user model.User
	err = db.
		QueryRowContext(ctx, query, args...).
		Scan(&user.ExternalID, &user.Email, &user.DisplayName, pq.Array(&user.DietaryFlags), &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("could not fetch user: %w", err)
	}

	return &user, nil
}

func UpdateUserPreferences(db *sql.DB, ctx context.Context, externalID string, dietaryFlags []string) error {
	query, args, err := psql.
		Update("users").
		Set("dietary_flags", pq.Array(dietaryFlags)).
		Where("external_id = ?", externalID).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	_, err = db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating user preferences: %w", err)
	}

	return nil
}

func GetFullInventory(db *sql.DB, ctx context.Context, userID string) (*model.UserInventory, error) {
	var inventory model.UserInventory
	inventory.UserID = userID

	g, ctx := errgroup.WithContext(ctx)

	// fetch each list in parallel to optimize transaction speed
	g.Go(func() error {
		pantry, err := fetchPantry(db, ctx, userID)
		if err != nil {
			return fmt.Errorf("pantry fetch failed: %w", err)
		}

		inventory.Pantry = pantry

		return nil
	})

	g.Go(func() error {
		shoppingList, err := fetchShoppingList(db, ctx, userID)
		if err != nil {
			return fmt.Errorf("shopping list fetch failed: %w", err)
		}

		inventory.ShoppingList = shoppingList

		return nil
	})

	g.Go(func() error {
		wishlist, err := fetchWishlist(db, ctx, userID)
		if err != nil {
			return fmt.Errorf("wishlist fetch failed: %w", err)
		}

		inventory.Wishlist = wishlist

		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	return &inventory, nil
}

func AddItemsToList(ex DBExecutor, ctx context.Context, table, userID string, items []string) error {
	if len(items) == 0 {
		return nil
	}

	queryBuilder := psql.
		Insert(table).
		Columns("user_id", "ingredient_name")

	for _, item := range items {
		queryBuilder = queryBuilder.
			Values(userID, cleanStringValue(item))
	}

	query, args, err := queryBuilder.
		Suffix("ON CONFLICT (user_id, ingredient_name) DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build bulk insert query: %w", err)
	}

	_, err = ex.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to bulk insert into %s: %w", table, err)
	}

	return nil
}

func RemoveItemsFromList(ex DBExecutor, ctx context.Context, table, userID string, items []string) error {
	if len(items) == 0 {
		return nil
	}

	cleanItems := make([]string, len(items))
	for i, item := range items {
		cleanItems[i] = cleanStringValue(item)
	}

	query, args, err := psql.
		Delete(table).
		Where(squirrel.Eq{
			"user_id":         userID,
			"ingredient_name": cleanItems,
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build bulk delete query: %w", err)
	}

	_, err = ex.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to bulk delete from %s: %w", table, err)
	}

	return nil
}

func MoveToPantry(db *sql.DB, ctx context.Context, userID string, items []string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := AddItemsToList(tx, ctx, "user_pantry", userID, items); err != nil {
		return fmt.Errorf("move failed while adding to pantry: %w", err)
	}

	if err := RemoveItemsFromList(tx, ctx, "user_shopping_list", userID, items); err != nil {
		return fmt.Errorf("move failed while removing from shopping list: %w", err)
	}

	return tx.Commit()
}

func UpsertUser(db *sql.DB, ctx context.Context, externalID, email, displayName string) error {
	query, args, err := psql.
		Insert("users").
		Columns("external_id", "email", "display_name").
		Values(externalID, email, displayName).
		Suffix(`
		ON CONFLICT (external_id)
		DO UPDATE SET display_name = EXCLUDED.display_name
	`).
		ToSql()
	if err != nil {
		return fmt.Errorf("could not build query: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("could not execute transaction: %w", err)
	}

	return tx.Commit()
}

func RecordMealExecution(db *sql.DB, ctx context.Context, userID string, recipeID string, ingredients []string) error {
	recipeQuery, recipeArgs, err := psql.
		Insert("consumed_recipes").
		Columns("user_id", "recipe_id", "times_made").
		Values(userID, recipeID, 1).
		Suffix(`
		ON CONFLICT (user_id, recipe_id)
		DO UPDATE SET times_made = consumed_recipes.times_made + 1, last_made_at = NOW()
	`).
		ToSql()
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, recipeQuery, recipeArgs...); err != nil {
		return fmt.Errorf("failed to update consumed recipes: %w", err)
	}

	for _, name := range ingredients {
		ingredientQuery, ingredientArgs, err := psql.
			Insert("consumed_ingredients").
			Columns("user_id", "ingredient_name", "times_used").
			Values(userID, name, 1).
			Suffix(`
			ON CONFLICT (user_id, ingredient_name)
			DO UPDATE SET times_used = consumed_ingredients.times_used + 1
		`).
			ToSql()
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, ingredientQuery, ingredientArgs...); err != nil {
			return fmt.Errorf("failed to update consumed ingredients: %w", err)
		}

		suggestQuery, suggestArgs, err := psql.
			Insert("shopping_list_suggestions").
			Columns("user_id", "ingredient_name", "reason").
			Values(userID, name, "Used in recently cooked recipe").
			Suffix(`
			ON CONFLICT (user_id, ingredient_name) DO NOTHING
		`).
			ToSql()
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, suggestQuery, suggestArgs...); err != nil {
			return fmt.Errorf("failed to update shopping list suggestions: %w", err)
		}
	}

	return tx.Commit()
}

func GetShoppingListSuggestions(db *sql.DB, ctx context.Context, userID string) ([]model.ShoppingListSuggestion, error) {
	query, args, err := psql.
		Select("ingredient_name", "reason", "suggested_at").
		From("shopping_list_suggestions").
		Where("user_id = ?", userID).
		OrderBy("suggested_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []model.ShoppingListSuggestion
	for rows.Next() {
		var s model.ShoppingListSuggestion
		if err := rows.Scan(&s.IngredientName, &s.Reason, &s.SuggestedAt); err != nil {
			return nil, err
		}

		suggestions = append(suggestions, s)
	}

	return suggestions, nil
}

func cleanStringValue(str string) string {
	return strings.ToLower(strings.TrimSpace(str))
}

// squirrel does not support dynamic table names.
// separate methods will be needed to keep paramaterized queries. logic is the exact same
func fetchPantry(db *sql.DB, ctx context.Context, externalID string) ([]string, error) {
	query, args, err := psql.
		Select("p.ingredient_name").
		From("user_pantry p").
		Join("users u ON p.user_id = u.id").
		Where("u.external_id = ?", externalID).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("fetch pantry - failed to build query: %w", err)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetch pantry - failed to execute query: %w", err)
	}
	defer rows.Close()

	var ingredients []string
	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("fetch pantry - failed to map query: %w", err)
		}

		ingredients = append(ingredients, name)
	}

	// returning empty slice instead of nil object makes for cleaner JSON
	if ingredients == nil {
		return []string{}, nil
	}

	return ingredients, nil
}

func fetchShoppingList(db *sql.DB, ctx context.Context, externalID string) ([]string, error) {
	query, args, err := psql.
		Select("s.ingredient_name").
		From("user_shopping_list s").
		Join("users u ON s.user_id = u.id").
		Where("u.external_id = ?", externalID).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("fetch shopping list - failed to build query: %w", err)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetch shopping list - failed to execute query: %w", err)
	}
	defer rows.Close()

	var ingredients []string
	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("fetch shopping list - failed to map query: %w", err)
		}

		ingredients = append(ingredients, name)
	}

	// returning empty slice instead of nil object makes for cleaner JSON
	if ingredients == nil {
		return []string{}, nil
	}

	return ingredients, nil
}

func fetchWishlist(db *sql.DB, ctx context.Context, externalID string) ([]string, error) {
	query, args, err := psql.
		Select("w.ingredient_name").
		From("user_wishlist w").
		Join("users u ON w.user_id = u.id").
		Where("u.external_id = ?", externalID).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("fetch wishlist - failed to build query: %w", err)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetch wishlist - failed to execute query: %w", err)
	}
	defer rows.Close()

	var ingredients []string
	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("fetch wishlist - failed to map query: %w", err)
		}

		ingredients = append(ingredients, name)
	}

	// returning empty slice instead of nil object makes for cleaner JSON
	if ingredients == nil {
		return []string{}, nil
	}

	return ingredients, nil
}

func GetInternalID(db *sql.DB, ctx context.Context, externalID string) (string, error) {
	query, args, err := psql.
		Select("id").
		From("users").
		Where("external_id = ?", externalID).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("error building query: %w", err)
	}

	var internalID string
	err = db.
		QueryRowContext(ctx, query, args...).
		Scan(&internalID)
	if err != nil {
		return "", fmt.Errorf("error fetching internal user id: %w", err)
	}

	return internalID, err
}
