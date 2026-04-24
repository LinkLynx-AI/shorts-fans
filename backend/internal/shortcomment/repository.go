// Package shortcomment manages public short comments.
package shortcomment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// DefaultPageSize is the default short comment page size.
	DefaultPageSize = 20
	// MaxBodyLength is the maximum comment body length after trimming.
	MaxBodyLength = 500
)

var (
	// ErrInvalidCommentBody indicates the comment body violates domain constraints.
	ErrInvalidCommentBody = errors.New("short comment body が不正です")
	// ErrShortNotFound indicates the target public short was not found.
	ErrShortNotFound = errors.New("short comment target short が見つかりません")
)

type queries interface {
	CreateShortComment(ctx context.Context, arg sqlc.CreateShortCommentParams) (sqlc.CreateShortCommentRow, error)
	GetPublicShortForComments(ctx context.Context, shortID pgtype.UUID) (pgtype.UUID, error)
	ListShortCommentsAfterCursor(ctx context.Context, arg sqlc.ListShortCommentsAfterCursorParams) ([]sqlc.ListShortCommentsAfterCursorRow, error)
	ListShortCommentsFirstPage(ctx context.Context, arg sqlc.ListShortCommentsFirstPageParams) ([]sqlc.ListShortCommentsFirstPageRow, error)
}

// Repository provides public short comment read and write operations.
type Repository struct {
	queries queries
}

// Cursor is the short comment keyset pagination cursor.
type Cursor struct {
	CommentID uuid.UUID
	CreatedAt time.Time
}

// Author is the display-only comment author summary.
type Author struct {
	AvatarURL   *string
	DisplayName string
	Handle      string
	ID          uuid.UUID
}

// Comment is the public short comment read model.
type Comment struct {
	Author    Author
	Body      string
	CreatedAt time.Time
	ID        uuid.UUID
	ShortID   uuid.UUID
}

// CreateInput contains the fields required to create a short comment.
type CreateInput struct {
	AuthorUserID uuid.UUID
	Body         string
	ShortID      uuid.UUID
}

// NewRepository creates a short comment repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return newRepository(sqlc.New(pool))
}

func newRepository(queries queries) *Repository {
	return &Repository{queries: queries}
}

// ListComments returns a newest-first page of public short comments.
func (r *Repository) ListComments(ctx context.Context, shortID uuid.UUID, cursor *Cursor, limit int) ([]Comment, *Cursor, error) {
	if r == nil || r.queries == nil {
		return nil, nil, fmt.Errorf("short comment repository が初期化されていません")
	}

	pageLimit, queryLimit := resolveListLimit(limit)
	if cursor == nil {
		rows, err := r.queries.ListShortCommentsFirstPage(ctx, sqlc.ListShortCommentsFirstPageParams{
			LimitCount: queryLimit,
			ShortID:    postgres.UUIDToPG(shortID),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("short comment list first page short=%s: %w", shortID, err)
		}
		if len(rows) == 0 {
			return nil, nil, fmt.Errorf("short comment public short lookup short=%s: %w", shortID, ErrShortNotFound)
		}

		return mapFirstPageRows(rows, pageLimit)
	}

	rows, err := r.queries.ListShortCommentsAfterCursor(ctx, sqlc.ListShortCommentsAfterCursorParams{
		CursorCommentID: postgres.UUIDToPG(cursor.CommentID),
		CursorCreatedAt: postgres.TimeToPG(&cursor.CreatedAt),
		LimitCount:      queryLimit,
		ShortID:         postgres.UUIDToPG(shortID),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("short comment list after cursor short=%s cursor_comment=%s: %w", shortID, cursor.CommentID, err)
	}
	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("short comment public short lookup short=%s: %w", shortID, ErrShortNotFound)
	}

	return mapAfterCursorRows(rows, pageLimit)
}

// CreateComment creates an immediately visible public short comment.
func (r *Repository) CreateComment(ctx context.Context, input CreateInput) (Comment, error) {
	if r == nil || r.queries == nil {
		return Comment{}, fmt.Errorf("short comment repository が初期化されていません")
	}

	body, err := normalizeCommentBody(input.Body)
	if err != nil {
		return Comment{}, err
	}

	row, err := r.queries.CreateShortComment(ctx, sqlc.CreateShortCommentParams{
		AuthorUserID: postgres.UUIDToPG(input.AuthorUserID),
		Body:         body,
		ShortID:      postgres.UUIDToPG(input.ShortID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if lookupErr := r.ensurePublicShort(ctx, input.ShortID); lookupErr != nil {
				return Comment{}, fmt.Errorf("short comment create short=%s author=%s: %w", input.ShortID, input.AuthorUserID, lookupErr)
			}

			return Comment{}, fmt.Errorf("short comment create author profile missing short=%s author=%s", input.ShortID, input.AuthorUserID)
		}

		return Comment{}, fmt.Errorf("short comment create short=%s author=%s: %w", input.ShortID, input.AuthorUserID, err)
	}

	comment, err := mapCreatedCommentRow(row)
	if err != nil {
		return Comment{}, fmt.Errorf("short comment create result short=%s: %w", input.ShortID, err)
	}

	return comment, nil
}

func (r *Repository) ensurePublicShort(ctx context.Context, shortID uuid.UUID) error {
	if _, err := r.queries.GetPublicShortForComments(ctx, postgres.UUIDToPG(shortID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("short comment public short lookup short=%s: %w", shortID, ErrShortNotFound)
		}

		return fmt.Errorf("short comment public short lookup short=%s: %w", shortID, err)
	}

	return nil
}

func resolveListLimit(limit int) (int, int32) {
	pageLimit := limit
	if pageLimit <= 0 {
		pageLimit = DefaultPageSize
	}

	return pageLimit, int32(pageLimit + 1)
}

func mapFirstPageRows(rows []sqlc.ListShortCommentsFirstPageRow, limit int) ([]Comment, *Cursor, error) {
	items := make([]Comment, 0, min(limit, len(rows)))
	for _, row := range rows {
		if !row.ID.Valid {
			continue
		}
		if len(items) >= limit {
			break
		}

		item, err := mapFirstPageRow(row)
		if err != nil {
			return nil, nil, err
		}

		items = append(items, item)
	}

	if len(rows) <= limit {
		if len(items) == 0 {
			return []Comment{}, nil, nil
		}

		return items, nil, nil
	}

	lastItem := items[len(items)-1]

	return items, &Cursor{
		CommentID: lastItem.ID,
		CreatedAt: lastItem.CreatedAt,
	}, nil
}

func mapAfterCursorRows(rows []sqlc.ListShortCommentsAfterCursorRow, limit int) ([]Comment, *Cursor, error) {
	items := make([]Comment, 0, min(limit, len(rows)))
	for _, row := range rows {
		if !row.ID.Valid {
			continue
		}
		if len(items) >= limit {
			break
		}

		item, err := mapAfterCursorRow(row)
		if err != nil {
			return nil, nil, err
		}

		items = append(items, item)
	}

	if len(rows) <= limit {
		if len(items) == 0 {
			return []Comment{}, nil, nil
		}

		return items, nil, nil
	}

	lastItem := items[len(items)-1]

	return items, &Cursor{
		CommentID: lastItem.ID,
		CreatedAt: lastItem.CreatedAt,
	}, nil
}

func mapFirstPageRow(row sqlc.ListShortCommentsFirstPageRow) (Comment, error) {
	body, err := requiredTextFromPG(row.Body, "comment body")
	if err != nil {
		return Comment{}, err
	}
	displayName, err := requiredTextFromPG(row.DisplayName, "comment author display name")
	if err != nil {
		return Comment{}, err
	}
	handle, err := requiredTextFromPG(row.Handle, "comment author handle")
	if err != nil {
		return Comment{}, err
	}

	return mapCommentFields(
		row.ID,
		row.ShortID,
		row.AuthorUserID,
		body,
		row.CreatedAt,
		displayName,
		handle,
		row.AvatarUrl,
	)
}

func mapAfterCursorRow(row sqlc.ListShortCommentsAfterCursorRow) (Comment, error) {
	body, err := requiredTextFromPG(row.Body, "comment body")
	if err != nil {
		return Comment{}, err
	}
	displayName, err := requiredTextFromPG(row.DisplayName, "comment author display name")
	if err != nil {
		return Comment{}, err
	}
	handle, err := requiredTextFromPG(row.Handle, "comment author handle")
	if err != nil {
		return Comment{}, err
	}

	return mapCommentFields(
		row.ID,
		row.ShortID,
		row.AuthorUserID,
		body,
		row.CreatedAt,
		displayName,
		handle,
		row.AvatarUrl,
	)
}

func mapCreatedCommentRow(row sqlc.CreateShortCommentRow) (Comment, error) {
	return mapCommentFields(
		row.ID,
		row.ShortID,
		row.AuthorUserID,
		row.Body,
		row.CreatedAt,
		row.DisplayName,
		row.Handle,
		row.AvatarUrl,
	)
}

func mapCommentFields(
	rawID pgtype.UUID,
	rawShortID pgtype.UUID,
	rawAuthorUserID pgtype.UUID,
	body string,
	rawCreatedAt pgtype.Timestamptz,
	displayName string,
	handle string,
	avatarURL pgtype.Text,
) (Comment, error) {
	id, err := postgres.UUIDFromPG(rawID)
	if err != nil {
		return Comment{}, fmt.Errorf("comment id 変換: %w", err)
	}
	shortID, err := postgres.UUIDFromPG(rawShortID)
	if err != nil {
		return Comment{}, fmt.Errorf("comment short id 変換: %w", err)
	}
	authorUserID, err := postgres.UUIDFromPG(rawAuthorUserID)
	if err != nil {
		return Comment{}, fmt.Errorf("comment author user id 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(rawCreatedAt)
	if err != nil {
		return Comment{}, fmt.Errorf("comment created_at 変換: %w", err)
	}

	trimmedDisplayName := strings.TrimSpace(displayName)
	trimmedHandle := strings.TrimSpace(handle)
	if trimmedDisplayName == "" || trimmedHandle == "" {
		return Comment{}, fmt.Errorf("comment author display name または handle がありません")
	}

	return Comment{
		Author: Author{
			AvatarURL:   postgres.OptionalTextFromPG(avatarURL),
			DisplayName: trimmedDisplayName,
			Handle:      trimmedHandle,
			ID:          authorUserID,
		},
		Body:      body,
		CreatedAt: createdAt,
		ID:        id,
		ShortID:   shortID,
	}, nil
}

func requiredTextFromPG(value pgtype.Text, fieldName string) (string, error) {
	if !value.Valid {
		return "", fmt.Errorf("%s が null です", fieldName)
	}

	return value.String, nil
}

func normalizeCommentBody(body string) (string, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" || exceedsCommentBodyLength(trimmed, MaxBodyLength) {
		return "", ErrInvalidCommentBody
	}

	return trimmed, nil
}

func exceedsCommentBodyLength(body string, maxLength int) bool {
	count := 0
	for range body {
		count++
		if count > maxLength {
			return true
		}
	}

	return false
}
