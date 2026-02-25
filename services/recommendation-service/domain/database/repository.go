package database

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func UpdatePopularity(ctx context.Context, db *sql.DB, recipeID string) error {
	query, args, err := psql.
		Insert("recipe_popularity").
		Columns("recipe_id", "like_count").
		Values(recipeID, 1).
		Suffix("ON CONFLICT (recipe_id) DO UPDATE SET like_count = recipe_popularity.like_count + 1").
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, query, args...)

	return err
}
