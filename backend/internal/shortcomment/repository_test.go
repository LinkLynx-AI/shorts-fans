package shortcomment

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type stubCommentQueries struct {
	createParams sqlc.CreateShortCommentParams
	createRow    sqlc.CreateShortCommentRow
	createErr    error
	createCalled bool

	getShortID     pgtype.UUID
	getShortErr    error
	getShortCalled bool

	firstPageParams sqlc.ListShortCommentsFirstPageParams
	firstPageRows   []sqlc.ListShortCommentsFirstPageRow
	firstPageErr    error
	firstPageCalled bool

	afterCursorParams sqlc.ListShortCommentsAfterCursorParams
	afterCursorRows   []sqlc.ListShortCommentsAfterCursorRow
	afterCursorErr    error
	afterCursorCalled bool
}

func (s *stubCommentQueries) CreateShortComment(_ context.Context, arg sqlc.CreateShortCommentParams) (sqlc.CreateShortCommentRow, error) {
	s.createCalled = true
	s.createParams = arg
	if s.createErr != nil {
		return sqlc.CreateShortCommentRow{}, s.createErr
	}

	return s.createRow, nil
}

func (s *stubCommentQueries) GetPublicShortForComments(_ context.Context, shortID pgtype.UUID) (pgtype.UUID, error) {
	s.getShortCalled = true
	s.getShortID = shortID
	if s.getShortErr != nil {
		return pgtype.UUID{}, s.getShortErr
	}

	return shortID, nil
}

func (s *stubCommentQueries) ListShortCommentsAfterCursor(_ context.Context, arg sqlc.ListShortCommentsAfterCursorParams) ([]sqlc.ListShortCommentsAfterCursorRow, error) {
	s.afterCursorCalled = true
	s.afterCursorParams = arg
	if s.afterCursorErr != nil {
		return nil, s.afterCursorErr
	}

	return s.afterCursorRows, nil
}

func (s *stubCommentQueries) ListShortCommentsFirstPage(_ context.Context, arg sqlc.ListShortCommentsFirstPageParams) ([]sqlc.ListShortCommentsFirstPageRow, error) {
	s.firstPageCalled = true
	s.firstPageParams = arg
	if s.firstPageErr != nil {
		return nil, s.firstPageErr
	}

	return s.firstPageRows, nil
}

func textPG(value string) pgtype.Text {
	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

func TestRepositoryListCommentsReturnsPageAndNextCursor(t *testing.T) {
	t.Parallel()

	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	firstCommentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	secondCommentID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	authorID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	createdAt := time.Date(2026, 1, 2, 14, 5, 0, 0, time.UTC)
	olderCreatedAt := createdAt.Add(-time.Minute)
	avatarURL := "https://cdn.example.com/avatar.jpg"
	queries := &stubCommentQueries{
		firstPageRows: []sqlc.ListShortCommentsFirstPageRow{
			{
				ID:           postgres.UUIDToPG(firstCommentID),
				ShortID:      postgres.UUIDToPG(shortID),
				AuthorUserID: postgres.UUIDToPG(authorID),
				Body:         textPG("first"),
				CreatedAt:    postgres.TimeToPG(&createdAt),
				DisplayName:  textPG(" Kana Mori "),
				Handle:       textPG(" kanamori "),
				AvatarUrl: pgtype.Text{
					String: avatarURL,
					Valid:  true,
				},
			},
			{
				ID:           postgres.UUIDToPG(secondCommentID),
				ShortID:      postgres.UUIDToPG(shortID),
				AuthorUserID: postgres.UUIDToPG(authorID),
				Body:         textPG("second"),
				CreatedAt:    postgres.TimeToPG(&olderCreatedAt),
				DisplayName:  textPG("Kana Mori"),
				Handle:       textPG("kanamori"),
			},
		},
	}

	items, nextCursor, err := newRepository(queries).ListComments(context.Background(), shortID, nil, 1)
	if err != nil {
		t.Fatalf("ListComments() error = %v, want nil", err)
	}
	if queries.getShortCalled {
		t.Fatal("ListComments() get public short called = true, want false for non-empty gated page")
	}
	if !queries.firstPageCalled {
		t.Fatal("ListComments() list called = false, want true")
	}
	if queries.afterCursorCalled {
		t.Fatal("ListComments() after cursor list called = true, want false")
	}
	if queries.firstPageParams.ShortID != postgres.UUIDToPG(shortID) {
		t.Fatalf("ListComments() short id param got %v want %v", queries.firstPageParams.ShortID, postgres.UUIDToPG(shortID))
	}
	if queries.firstPageParams.LimitCount != 2 {
		t.Fatalf("ListComments() limit param got %d want 2", queries.firstPageParams.LimitCount)
	}
	if len(items) != 1 {
		t.Fatalf("ListComments() item count got %d want 1", len(items))
	}
	if items[0].ID != firstCommentID {
		t.Fatalf("ListComments() first id got %s want %s", items[0].ID, firstCommentID)
	}
	if items[0].Author.DisplayName != "Kana Mori" {
		t.Fatalf("ListComments() author display name got %q want %q", items[0].Author.DisplayName, "Kana Mori")
	}
	if items[0].Author.Handle != "kanamori" {
		t.Fatalf("ListComments() author handle got %q want %q", items[0].Author.Handle, "kanamori")
	}
	if items[0].Author.AvatarURL == nil || *items[0].Author.AvatarURL != avatarURL {
		t.Fatalf("ListComments() avatar got %#v want %q", items[0].Author.AvatarURL, avatarURL)
	}
	if nextCursor == nil {
		t.Fatal("ListComments() next cursor = nil, want cursor")
	}
	if nextCursor.CommentID != firstCommentID {
		t.Fatalf("ListComments() next cursor id got %s want %s", nextCursor.CommentID, firstCommentID)
	}
	if !nextCursor.CreatedAt.Equal(createdAt) {
		t.Fatalf("ListComments() next cursor created_at got %s want %s", nextCursor.CreatedAt, createdAt)
	}
}

func TestRepositoryListCommentsAppliesCursorAndMapsMissingShort(t *testing.T) {
	t.Parallel()

	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cursorID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	cursorAt := time.Date(2026, 1, 2, 14, 5, 0, 0, time.UTC)
	queries := &stubCommentQueries{
		afterCursorRows: []sqlc.ListShortCommentsAfterCursorRow{
			{PublicShortID: postgres.UUIDToPG(shortID)},
		},
	}
	if _, _, err := newRepository(queries).ListComments(context.Background(), shortID, &Cursor{CommentID: cursorID, CreatedAt: cursorAt}, 0); err != nil {
		t.Fatalf("ListComments() error = %v, want nil", err)
	}
	if queries.getShortCalled {
		t.Fatal("ListComments() get public short called = true, want false for empty public page")
	}
	if !queries.afterCursorCalled {
		t.Fatal("ListComments() after cursor list called = false, want true")
	}
	if queries.firstPageCalled {
		t.Fatal("ListComments() first page list called = true, want false")
	}
	if queries.afterCursorParams.LimitCount != DefaultPageSize+1 {
		t.Fatalf("ListComments() default limit got %d want %d", queries.afterCursorParams.LimitCount, DefaultPageSize+1)
	}
	if queries.afterCursorParams.ShortID != postgres.UUIDToPG(shortID) {
		t.Fatalf("ListComments() short id param got %v want %v", queries.afterCursorParams.ShortID, postgres.UUIDToPG(shortID))
	}
	if queries.afterCursorParams.CursorCommentID != postgres.UUIDToPG(cursorID) {
		t.Fatalf("ListComments() cursor id got %v want %v", queries.afterCursorParams.CursorCommentID, postgres.UUIDToPG(cursorID))
	}
	if !queries.afterCursorParams.CursorCreatedAt.Time.Equal(cursorAt) {
		t.Fatalf("ListComments() cursor created_at got %s want %s", queries.afterCursorParams.CursorCreatedAt.Time, cursorAt)
	}

	queries = &stubCommentQueries{}
	if _, _, err := newRepository(queries).ListComments(context.Background(), shortID, nil, 0); !errors.Is(err, ErrShortNotFound) {
		t.Fatalf("ListComments() missing short error got %v want ErrShortNotFound", err)
	}
}

func TestRepositoryListCommentsAfterCursorReturnsPageAndNextCursor(t *testing.T) {
	t.Parallel()

	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cursorID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	firstCommentID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	secondCommentID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	authorID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	cursorAt := time.Date(2026, 1, 2, 14, 5, 0, 0, time.UTC)
	createdAt := cursorAt.Add(-time.Minute)
	olderCreatedAt := createdAt.Add(-time.Minute)
	queries := &stubCommentQueries{
		afterCursorRows: []sqlc.ListShortCommentsAfterCursorRow{
			{
				ID:           postgres.UUIDToPG(firstCommentID),
				ShortID:      postgres.UUIDToPG(shortID),
				AuthorUserID: postgres.UUIDToPG(authorID),
				Body:         textPG("after cursor"),
				CreatedAt:    postgres.TimeToPG(&createdAt),
				DisplayName:  textPG("Kana Mori"),
				Handle:       textPG("kanamori"),
			},
			{
				ID:           postgres.UUIDToPG(secondCommentID),
				ShortID:      postgres.UUIDToPG(shortID),
				AuthorUserID: postgres.UUIDToPG(authorID),
				Body:         textPG("older"),
				CreatedAt:    postgres.TimeToPG(&olderCreatedAt),
				DisplayName:  textPG("Kana Mori"),
				Handle:       textPG("kanamori"),
			},
		},
	}

	items, nextCursor, err := newRepository(queries).ListComments(context.Background(), shortID, &Cursor{CommentID: cursorID, CreatedAt: cursorAt}, 1)
	if err != nil {
		t.Fatalf("ListComments() error = %v, want nil", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListComments() item count got %d want 1", len(items))
	}
	if queries.getShortCalled {
		t.Fatal("ListComments() get public short called = true, want false for non-empty gated page")
	}
	if items[0].ID != firstCommentID {
		t.Fatalf("ListComments() first id got %s want %s", items[0].ID, firstCommentID)
	}
	if items[0].Body != "after cursor" {
		t.Fatalf("ListComments() first body got %q want after cursor", items[0].Body)
	}
	if nextCursor == nil {
		t.Fatal("ListComments() next cursor = nil, want cursor")
	}
	if nextCursor.CommentID != firstCommentID {
		t.Fatalf("ListComments() next cursor id got %s want %s", nextCursor.CommentID, firstCommentID)
	}
	if !nextCursor.CreatedAt.Equal(createdAt) {
		t.Fatalf("ListComments() next cursor created_at got %s want %s", nextCursor.CreatedAt, createdAt)
	}
}

func TestRepositoryCreateCommentNormalizesBody(t *testing.T) {
	t.Parallel()

	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	commentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	authorID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	createdAt := time.Date(2026, 1, 2, 14, 5, 0, 0, time.UTC)
	queries := &stubCommentQueries{
		createRow: sqlc.CreateShortCommentRow{
			ID:           postgres.UUIDToPG(commentID),
			ShortID:      postgres.UUIDToPG(shortID),
			AuthorUserID: postgres.UUIDToPG(authorID),
			Body:         "hello",
			CreatedAt:    postgres.TimeToPG(&createdAt),
			DisplayName:  "Kana Mori",
			Handle:       "kanamori",
		},
	}

	comment, err := newRepository(queries).CreateComment(context.Background(), CreateInput{
		AuthorUserID: authorID,
		Body:         "  hello  ",
		ShortID:      shortID,
	})
	if err != nil {
		t.Fatalf("CreateComment() error = %v, want nil", err)
	}
	if !queries.createCalled {
		t.Fatal("CreateComment() create called = false, want true")
	}
	if queries.createParams.Body != "hello" {
		t.Fatalf("CreateComment() body param got %q want %q", queries.createParams.Body, "hello")
	}
	if queries.createParams.ShortID != postgres.UUIDToPG(shortID) {
		t.Fatalf("CreateComment() short id param got %v want %v", queries.createParams.ShortID, postgres.UUIDToPG(shortID))
	}
	if queries.createParams.AuthorUserID != postgres.UUIDToPG(authorID) {
		t.Fatalf("CreateComment() author id param got %v want %v", queries.createParams.AuthorUserID, postgres.UUIDToPG(authorID))
	}
	if comment.ID != commentID {
		t.Fatalf("CreateComment() comment id got %s want %s", comment.ID, commentID)
	}
	if comment.Body != "hello" {
		t.Fatalf("CreateComment() body got %q want %q", comment.Body, "hello")
	}
}

func TestRepositoryCreateCommentValidatesInputAndMapsMissingShort(t *testing.T) {
	t.Parallel()

	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	authorID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	for _, body := range []string{"   ", strings.Repeat("a", MaxBodyLength+1)} {
		if _, err := newRepository(&stubCommentQueries{}).CreateComment(context.Background(), CreateInput{
			AuthorUserID: authorID,
			Body:         body,
			ShortID:      shortID,
		}); err == nil {
			t.Fatalf("CreateComment(%q) error = nil, want validation error", body)
		} else if !errors.Is(err, ErrInvalidCommentBody) {
			t.Fatalf("CreateComment(%q) error got %v want ErrInvalidCommentBody", body, err)
		}
	}

	queries := &stubCommentQueries{
		createErr:   pgx.ErrNoRows,
		getShortErr: pgx.ErrNoRows,
	}
	if _, err := newRepository(queries).CreateComment(context.Background(), CreateInput{
		AuthorUserID: authorID,
		Body:         "hello",
		ShortID:      shortID,
	}); !errors.Is(err, ErrShortNotFound) {
		t.Fatalf("CreateComment() missing short error got %v want ErrShortNotFound", err)
	}
}

func TestRepositoryCreateCommentDistinguishesMissingAuthorProfile(t *testing.T) {
	t.Parallel()

	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	authorID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	queries := &stubCommentQueries{createErr: pgx.ErrNoRows}

	if _, err := newRepository(queries).CreateComment(context.Background(), CreateInput{
		AuthorUserID: authorID,
		Body:         "hello",
		ShortID:      shortID,
	}); err == nil {
		t.Fatal("CreateComment() missing profile error = nil, want error")
	} else if errors.Is(err, ErrShortNotFound) {
		t.Fatalf("CreateComment() missing profile error got ErrShortNotFound, want internal data error: %v", err)
	}
	if !queries.getShortCalled {
		t.Fatal("CreateComment() missing profile did not verify public short")
	}
}
