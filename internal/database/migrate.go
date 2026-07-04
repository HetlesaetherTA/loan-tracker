package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"log/slog"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Up(ctx context.Context, db *sql.DB) error {
	slog.Info("Database migration UP started")

	sqlFiles, err := getMigrationFiles(".up.sql")
	if err != nil {
		return err
	}

	sort.Slice(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Name() < sqlFiles[j].Name()
	})

	fmt.Printf("Found %d UP migrations to execute...\n", len(sqlFiles))
	return executeSQLFiles(ctx, db, sqlFiles)
}

func Down(ctx context.Context, db *sql.DB) error {
	slog.Warn("Database migration DOWN started")

	if strings.ToLower(os.Getenv("APP_ENV")) != "dev" {
		return fmt.Errorf("you may only DOWN migration when APP_ENV=dev (APP_ENV=%s)", os.Getenv("APP_ENV"))
	}

	sqlFiles, err := getMigrationFiles(".down.sql")
	if err != nil {
		return err
	}

	sort.Slice(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Name() > sqlFiles[j].Name()
	})

	slog.Info(fmt.Sprintf("Found %d DOWN migrations to execute...", len(sqlFiles)))
	return executeSQLFiles(ctx, db, sqlFiles)
}

func getMigrationFiles(suffix string) ([]fs.DirEntry, error) {
	files, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	var sqlFiles []fs.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
			sqlFiles = append(sqlFiles, file)
		}
	}
	return sqlFiles, nil
}

func executeSQLFiles(ctx context.Context, db *sql.DB, files []fs.DirEntry) error {
	for _, file := range files {
		fileName := file.Name()

		fmt.Printf("Executing migration: %s\n", fileName)

		embeddedPath := "migrations/" + fileName

		content, err := migrationFiles.ReadFile(embeddedPath)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", fileName, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		_, err = tx.ExecContext(ctx, string(content))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("migration failed in file %s: %w", fileName, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction for %s: %w", fileName, err)
		}
	}
	return nil
}
