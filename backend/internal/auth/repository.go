package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrCurrentViewerNotFound は有効な session に紐づく viewer が見つからないことを表します。
var ErrCurrentViewerNotFound = errors.New("current viewer が見つかりません")

type queries interface {
	GetCurrentViewerBySessionTokenHash(ctx context.Context, sessionTokenHash string) (sqlc.GetCurrentViewerBySessionTokenHashRow, error)
}

// Repository は auth bootstrap 向けの永続化 read を包みます。
type Repository struct {
	queries queries
}

// NewRepository は pgxpool ベースの auth repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	return newRepository(sqlc.New(pool))
}

func newRepository(q queries) *Repository {
	return &Repository{queries: q}
}

// GetCurrentViewerBySessionTokenHash は session token hash から current viewer を取得します。
func (r *Repository) GetCurrentViewerBySessionTokenHash(ctx context.Context, sessionTokenHash string) (CurrentViewer, error) {
	row, err := r.queries.GetCurrentViewerBySessionTokenHash(ctx, sessionTokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CurrentViewer{}, fmt.Errorf("current viewer 取得 session=%s: %w", sessionTokenHash, ErrCurrentViewerNotFound)
		}

		return CurrentViewer{}, fmt.Errorf("current viewer 取得 session=%s: %w", sessionTokenHash, err)
	}

	viewer, err := mapCurrentViewer(row)
	if err != nil {
		return CurrentViewer{}, fmt.Errorf("current viewer 取得結果の変換 session=%s: %w", sessionTokenHash, err)
	}

	return viewer, nil
}

func mapCurrentViewer(row sqlc.GetCurrentViewerBySessionTokenHashRow) (CurrentViewer, error) {
	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return CurrentViewer{}, fmt.Errorf("current viewer の user id 変換: %w", err)
	}

	return CurrentViewer{
		ID:                   userID,
		ActiveMode:           ActiveMode(row.ActiveMode),
		CanAccessCreatorMode: row.CanAccessCreatorMode,
	}, nil
}
