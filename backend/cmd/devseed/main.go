package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/config"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/devseed"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/google/uuid"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()
	if cfg.PostgresDSN == "" {
		logger.Error("missing postgres configuration", "error", "POSTGRES_DSN is required")
		os.Exit(1)
	}

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Error("failed to initialize postgres pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	summary, err := devseed.Run(ctx, pool)
	if err != nil {
		logger.Error("failed to apply dev seed", "error", err)
		os.Exit(1)
	}

	logger.Info(
		"dev seed applied",
		"creator_user_id", summary.CreatorUserID.String(),
		"fan_user_id", summary.FanUserID.String(),
		"main_id", summary.MainID.String(),
		"short_ids", uuidStrings(summary.ShortIDs),
	)
}

func uuidStrings(ids []uuid.UUID) []string {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, id.String())
	}

	return values
}
