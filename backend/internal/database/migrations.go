package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations(migrationsDir string) error {
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	// Ensure schema_migrations table exists to prevent re-running migrations.
	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY, applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`)
	if err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}

	var upMigrations []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			upMigrations = append(upMigrations, file.Name())
		}
	}

	sort.Strings(upMigrations)

	for _, migrationFile := range upMigrations {
		var alreadyApplied bool
		err = DB.QueryRow(`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE filename = $1)`, migrationFile).Scan(&alreadyApplied)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", migrationFile, err)
		}
		if alreadyApplied {
			log.Printf("Skipping migration (already applied): %s", migrationFile)
			continue
		}

		log.Printf("Running migration: %s", migrationFile)
		content, err := os.ReadFile(filepath.Join(migrationsDir, migrationFile))
		if err != nil {
			return err
		}

		tx, err := DB.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin tx for migration %s: %w", migrationFile, err)
		}

		if _, err = tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed: %w", migrationFile, err)
		}

		if _, err = tx.Exec(`INSERT INTO schema_migrations (filename) VALUES ($1)`, migrationFile); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migrationFile, err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migrationFile, err)
		}
	}

	log.Println("Migrations completed successfully")
	return nil
}
