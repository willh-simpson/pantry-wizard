package database

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/willh-simpson/pantry-wizard/services/recommendation-service/domain/model"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func UpdateScore(ctx context.Context, db *sql.DB, recipeID, interactionType string) error {
	var weight float64
	var countDelta int
	var column string

	switch interactionType {
	case "like":
		column, weight, countDelta = "like_count", 1.0, 1
	case "unlike":
		column, weight, countDelta = "like_count", -1.0, -1
	case "save":
		column, weight, countDelta = "save_count", 3.0, 1
	case "unsave":
		column, weight, countDelta = "save_count", -3.0, 1
	case "view":
		column, weight, countDelta = "view_count", 0.1, 1
	default:
		column, weight, countDelta = "view_count", 0.0, 0
	}

	query, args, err := psql.
		Insert("recipe_scores").
		Columns("recipe_id", column, "total_score").
		Values(recipeID, countDelta, weight).
		Suffix(`
			ON CONFLICT (recipe_id) DO UPDATE SET
			`+column+` = GREATEST(0, recipe_scores.`+column+` + ?),
			total_score = recipe_scores.total_score + ?,
			updated_at = CURRENT_TIMESTAMP
		`, countDelta, weight).
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, query, args...)

	return err
}

func GetTopRecipes(ctx context.Context, db *sql.DB, limit int) ([]model.RankedRecipe, error) {
	query, args, err := psql.
		Select("recipe_id", "total_score").
		From("recipe_scores").
		OrderBy("total_score DESC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []model.RankedRecipe
	for rows.Next() {
		var r model.RankedRecipe

		if err := rows.Scan(&r.RecipeID, &r.TotalScore); err != nil {
			return nil, err
		}

		recipes = append(recipes, r)
	}

	return recipes, nil
}
