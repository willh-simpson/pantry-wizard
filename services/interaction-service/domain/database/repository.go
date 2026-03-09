package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/willh-simpson/pantry-wizard/services/interaction-service/domain/model"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func HandleInteraction(db *sql.DB, ctx context.Context, userID, recipeID string, action model.InteractionType) error {
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
	case model.Like:
		table, isUnaction = "recipe_likes", false
	case model.Unlike:
		table, isUnaction = "recipe_likes", true
	case model.Save:
		table, isUnaction = "recipe_saves", false
	case model.Unsave:
		table, isUnaction = "recipe_saves", true
	case model.View:
		table, isUnaction = "recipe_views", false
	case model.Cook:
		table, isUnaction = "recipe_cooks", false
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
		return "", fmt.Errorf("error fetching user: %w", err)
	}

	return internalID, nil
}

func UpsertUser(db *sql.DB, ctx context.Context, externalID string) error {
	query, args, err := psql.
		Insert("users").
		Columns("external_id").
		Values(externalID).
		Suffix(`
		ON CONFLICT (external_id)
		DO NOTHING
	`).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("error trying to upsert user: %w", err)
	}

	return tx.Commit()
}
