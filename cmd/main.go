package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq" // <- registers the "postgres" driver
	"github.com/pressly/goose/v3"
	"github.com/sikozonpc/ecom/internal/adapters/postgresql/migrations"
	"github.com/sikozonpc/ecom/internal/env"
)

func main() {
	ctx := context.Background()

	cfg := config{
		addr: ":8080",
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING", "host=localhost user=postgres password=postgres dbname=ecom sslmode=disable"),
		},
	}

	// Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Run Goose Migrations
	sqlDB, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		slog.Error("failed to open db for migrations", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("failed to set goose dialect", "error", err)
		os.Exit(1)
	}

	goose.SetBaseFS(migrations.Files)

	if err := goose.Up(sqlDB, "."); err != nil {
		slog.Error("failed to apply migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("migrations applied successfully")

	// Database
	conn, err := pgx.Connect(ctx, cfg.db.dsn)
	if err != nil {
		panic(err)
	}

	defer conn.Close(ctx)

	logger.Info("connected to database", "dsn", cfg.db.dsn)

	api := application{
		config: cfg,
		db:     conn,
	}

	if err := api.run(api.mount()); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
