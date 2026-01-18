package database

import (
	"database/sql"
	"os"
)

func RunMigrations(db *sql.DB) error {
	// read migration file
	schema, err := os.ReadFile("migrations/001_initial_schema.sql")
	if err != nil {
		return err
	}

	// execute migration
	_, err = db.Exec(string(schema))
	return err
}
