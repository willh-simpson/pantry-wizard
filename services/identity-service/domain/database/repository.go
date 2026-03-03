package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/willh-simpson/pantry-wizard/services/identity-service/domain/model"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func CreateOrUpdateUser(db *sql.DB, ctx context.Context, email, externalID, displayName string) (*model.User, error) {
	query, args, err := psql.
		Insert("users").
		Columns("email", "external_id", "display_name", "updated_at").
		Values(email, externalID, displayName, time.Now()).
		Suffix(`
		ON CONFLICT (external_id)
		DO UPDATE SET
			display_name = EXCLUDED.display_name,
			updated_at = NOW(),
		RETURNING id, external_id, email, display_name, dietary_flags, created_at
	`).
		ToSql()

	var user model.User
	err = db.
		QueryRowContext(ctx, query, args...).
		Scan(&user.ID, &user.ExternalID, &user.Email, &user.DisplayName, &user.DietaryFlags, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to sync user to local db: %w", err)
	}

	return &user, nil
}

func GetUserByEmail(db *sql.DB, ctx context.Context, email string) (*model.User, error) {
	query, args, err := psql.
		Select("id", "external_id", "email", "display_name", "dietary_flags").
		From("users").
		Where("email = ?", email).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build user select query: %w", err)
	}

	var user model.User
	err = db.
		QueryRowContext(ctx, query, args...).
		Scan(&user.ID, &user.ExternalID, &user.Email, &user.DisplayName, &user.DietaryFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to select user: %w", err)
	}

	return &user, nil
}

func GetUserByExternalID(db *sql.DB, ctx context.Context, externalID string) (*model.User, error) {
	query, args, err := psql.
		Select("id", "external_id", "email", "display_name", "dietary_flags", "created_at").
		From("users").
		Where("external_id = ?", externalID).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build user select query: %w", err)
	}

	var user model.User
	err = db.
		QueryRowContext(ctx, query, args...).
		Scan(&user.ID, &user.ExternalID, &user.Email, &user.DisplayName, &user.DietaryFlags, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to select user: %w", err)
	}

	return &user, nil
}
