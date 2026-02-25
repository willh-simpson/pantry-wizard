package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/domain/model"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func CreateFullRecipe(db *sql.DB, req model.CreateRecipeRequest) (string, error) {
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	recipeQuery, recipeArgs, err := psql.
		Insert("recipes").
		Columns("title", "description", "instructions", "author_id", "prep_time_min", "calories", "budget_tier").
		Values(req.Title, req.Description, req.Instructions, req.AuthorID, req.PrepTime, req.Calories, req.BudgetTier).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return "", fmt.Errorf("failed to build recipy query: %v", err)
	}

	var recipeID string

	err = tx.
		QueryRowContext(ctx, recipeQuery, recipeArgs...).
		Scan(&recipeID)

	for _, ingredient := range req.Ingredients {
		var ingredientID string

		ingredientQuery, ingredientArgs, err := psql.
			Insert("ingredients").
			Columns("name", "category").
			Values(ingredient.Name, ingredient.Category).
			Suffix("ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id").
			ToSql()
		if err != nil {
			return "", fmt.Errorf("failed to build ingredient query: %v", err)
		}

		err = tx.
			QueryRowContext(ctx, ingredientQuery, ingredientArgs...).
			Scan(&ingredientID)
		if err != nil {
			return "", fmt.Errorf("failed to upsert ingredient \"%s\": %v", ingredient.Name, err)
		}

		linkQuery, linkArgs, err := psql.
			Insert("recipe_ingredients").
			Columns("recipe_id", "ingredient_id", "amount", "unit").
			Values(recipeID, ingredientID, ingredient.Amount, ingredient.Unit).
			ToSql()
		if err != nil {
			return "", fmt.Errorf("failed to build link query for ingredient \"%s\": %v", ingredient.Name, err)
		}

		_, err = tx.ExecContext(ctx, linkQuery, linkArgs...)
		if err != nil {
			return "", fmt.Errorf("failed to link ingredient \"%s\": %v", ingredient.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %v", err)
	}

	return recipeID, nil
}

func SearchRecipes(db *sql.DB, title string, maxBudget int, maxPrepTime int) ([]model.RecipeResponse, error) {
	queryBuilder := psql.
		Select("id", "title", "created_at").
		From("recipes")

	if title != "" {
		queryBuilder = queryBuilder.Where(squirrel.ILike{
			"title": fmt.Sprintf("%%%s%%", title),
		})
	}

	if maxBudget > 0 {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{
			"budget_tier": maxBudget,
		})
	}

	if maxPrepTime > 0 {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{
			"prep_time_min": maxPrepTime,
		})
	}

	queryBuilder = queryBuilder.OrderBy("created_at DESC")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %v", err)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %v", err)
	}
	defer rows.Close()

	var recipes []model.RecipeResponse
	for rows.Next() {
		var r model.RecipeResponse

		if err := rows.Scan(&r.ID, &r.Title, &r.CreatedAt); err != nil {
			return nil, err
		}

		recipes = append(recipes, r)
	}

	return recipes, nil
}
