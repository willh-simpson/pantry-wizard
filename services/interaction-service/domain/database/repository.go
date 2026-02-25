package database

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func SaveLike(ctx context.Context, db *sql.DB, userID, recipeID string) error {
	query, args, err := psql.
		Insert("recipe_likes").
		Values(userID, recipeID).
		Suffix("ON CONFLICT DO NOTHING").
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, query, args...)

	return err
}
