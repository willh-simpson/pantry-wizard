package database

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func HandleInteraction(ctx context.Context, db *sql.DB, userID, recipeID, action string) error {
	logQuery, logArgs, err := psql.
		Insert("interactions").
		Columns("id", "user_id", "recipe_id", "event_type").
		Values(squirrel.Expr("uuid_generate_v4()"), userID, recipeID, action).
		ToSql()
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, logQuery, logArgs...); err != nil {
		return err
	}

	var table string
	var isUnaction bool

	switch action {
	case "like":
		table, isUnaction = "recipe_likes", false
	case "unlike":
		table, isUnaction = "recipe_likes", true
	case "save":
		table, isUnaction = "recipe_saves", false
	case "unsave":
		table, isUnaction = "recipe_saves", true
	case "view":
		table, isUnaction = "recipe_views", false
	default:
		table, isUnaction = "recipe_views", false
	}

	if !isUnaction {
		actionQuery, args, err := psql.
			Insert(table).
			Columns("user_id", "recipe_id").
			Values(userID, recipeID).
			Suffix("ON CONFLICT DO NOTHING").
			ToSql()
		if err != nil {
			return err
		}

		tx.ExecContext(ctx, actionQuery, args...)
	} else {
		unactionQuery, args, err := psql.
			Delete(table).
			Where(squirrel.Eq{
				"user_id":   userID,
				"recipe_id": recipeID,
			}).
			ToSql()
		if err != nil {
			return err
		}

		tx.ExecContext(ctx, unactionQuery, args...)
	}

	return tx.Commit()
}
