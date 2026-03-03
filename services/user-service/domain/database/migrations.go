package database

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// applies all 'up' migrations in migrationsPath
func RunMigrations(dbURL string, migrationsPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dbURL,
	)

	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("database is already up to date")

			return nil
		}

		return fmt.Errorf("could not run up migrations: %w", err)
	}

	log.Println("migrations applied successfully")

	return nil
}

func ForceMigration(dbURL string, version int) error {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return err
	}

	return m.Force(version)
}
