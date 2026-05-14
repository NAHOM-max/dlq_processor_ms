package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found, falling back to environment variables")
	}

	dsn := mustEnv("POSTGRES_DSN")

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	sql, err := os.ReadFile("internal/infrastructure/storage/migrations/001_create_dlq_events.sql")
	if err != nil {
		slog.Error("failed to read migration file", "error", err)
		os.Exit(1)
	}

	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}

	slog.Info("migration applied successfully")
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("missing required env var", "key", key)
		os.Exit(1)
	}
	return v
}
