package media

import (
	"context"
	"errors"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type stubQueries struct {
	getAsset func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error)
}

func (s stubQueries) CreateMediaAsset(context.Context, sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error) {
	return sqlc.AppMediaAsset{}, nil
}

func (s stubQueries) GetMediaAssetByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
	return s.getAsset(ctx, id)
}

func (s stubQueries) ListMediaAssetsByCreatorUserID(context.Context, pgtype.UUID) ([]sqlc.AppMediaAsset, error) {
	return nil, nil
}

func (s stubQueries) UpdateMediaAssetProcessingState(context.Context, sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error) {
	return sqlc.AppMediaAsset{}, nil
}

func TestGetAssetNotFound(t *testing.T) {
	t.Parallel()

	repo := newRepository(stubQueries{
		getAsset: func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error) {
			return sqlc.AppMediaAsset{}, pgx.ErrNoRows
		},
	})

	_, err := repo.GetAsset(context.Background(), uuid.New())
	if !errors.Is(err, ErrAssetNotFound) {
		t.Fatalf("GetAsset() error got %v want %v", err, ErrAssetNotFound)
	}
}

func TestNewRepositoryInitializesQueries(t *testing.T) {
	t.Parallel()

	repository := NewRepository(&pgxpool.Pool{})
	if repository == nil {
		t.Fatal("NewRepository() repository = nil, want non-nil")
	}
	if repository.queries == nil {
		t.Fatal("NewRepository() queries = nil, want initialized")
	}
}
