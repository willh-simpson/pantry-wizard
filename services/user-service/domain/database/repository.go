package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/willh-simpson/pantry-wizard/services/user-service/domain/model"
	"golang.org/x/sync/errgroup"
)

type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

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

// squirrel does not support dynamic table names.
// separate methods will be needed to keep paramaterized queries. logic is the exact same
func fetchPantry(db *sql.DB, ctx context.Context, userID string) ([]string, error) {
	query, args, err := psql.
		Select("ingredient_name").
		From("user_pantry").
		Where("user_id = ?", userID).
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

func fetchShoppingList(db *sql.DB, ctx context.Context, userID string) ([]string, error) {
	query, args, err := psql.
		Select("ingredient_name").
		From("user_shopping_list").
		Where("user_id = ?", userID).
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

func fetchWishlist(db *sql.DB, ctx context.Context, userID string) ([]string, error) {
	query, args, err := psql.
		Select("ingredient_name").
		From("user_wishlist").
		Where("user_id = ?", userID).
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

func cleanStringValue(str string) string {
	return strings.ToLower(strings.TrimSpace(str))
}
