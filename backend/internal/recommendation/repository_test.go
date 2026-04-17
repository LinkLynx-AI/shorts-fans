package recommendation

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type repositoryStubQueries struct {
	getEvent                    func(context.Context, sqlc.GetRecommendationEventByViewerAndIdempotencyKeyParams) (sqlc.AppRecommendationEvent, error)
	insertEvent                 func(context.Context, sqlc.InsertRecommendationEventParams) (sqlc.AppRecommendationEvent, error)
	listFollowedCreatorIDs      func(context.Context, sqlc.ListRecommendationFollowedCreatorIDsByViewerAndCreatorIDsParams) ([]pgtype.UUID, error)
	listPinnedShortIDs          func(context.Context, sqlc.ListRecommendationPinnedShortIDsByViewerAndShortIDsParams) ([]pgtype.UUID, error)
	listShortGlobalFeatures     func(context.Context, []pgtype.UUID) ([]sqlc.AppRecommendationShortGlobalFeature, error)
	listUnlockedMainIDs         func(context.Context, sqlc.ListRecommendationUnlockedMainIDsByViewerAndMainIDsParams) ([]pgtype.UUID, error)
	listViewerCreatorFeatures   func(context.Context, sqlc.ListRecommendationViewerCreatorFeaturesByViewerAndCreatorIDsParams) ([]sqlc.AppRecommendationViewerCreatorFeature, error)
	listViewerMainFeatures      func(context.Context, sqlc.ListRecommendationViewerMainFeaturesByViewerAndMainIDsParams) ([]sqlc.AppRecommendationViewerMainFeature, error)
	listViewerShortFeatures     func(context.Context, sqlc.ListRecommendationViewerShortFeaturesByViewerAndShortIDsParams) ([]sqlc.AppRecommendationViewerShortFeature, error)
	upsertShortGlobalFeatures   func(context.Context, sqlc.UpsertRecommendationShortGlobalFeaturesParams) (int64, error)
	upsertViewerCreatorFeatures func(context.Context, sqlc.UpsertRecommendationViewerCreatorFeaturesParams) error
	upsertViewerMainFeatures    func(context.Context, sqlc.UpsertRecommendationViewerMainFeaturesParams) (int64, error)
	upsertViewerShortFeatures   func(context.Context, sqlc.UpsertRecommendationViewerShortFeaturesParams) (int64, error)
}

func (s repositoryStubQueries) GetRecommendationEventByViewerAndIdempotencyKey(ctx context.Context, arg sqlc.GetRecommendationEventByViewerAndIdempotencyKeyParams) (sqlc.AppRecommendationEvent, error) {
	return s.getEvent(ctx, arg)
}

func (s repositoryStubQueries) InsertRecommendationEvent(ctx context.Context, arg sqlc.InsertRecommendationEventParams) (sqlc.AppRecommendationEvent, error) {
	return s.insertEvent(ctx, arg)
}

func (s repositoryStubQueries) ListRecommendationFollowedCreatorIDsByViewerAndCreatorIDs(ctx context.Context, arg sqlc.ListRecommendationFollowedCreatorIDsByViewerAndCreatorIDsParams) ([]pgtype.UUID, error) {
	if s.listFollowedCreatorIDs == nil {
		return []pgtype.UUID{}, nil
	}
	return s.listFollowedCreatorIDs(ctx, arg)
}

func (s repositoryStubQueries) ListRecommendationPinnedShortIDsByViewerAndShortIDs(ctx context.Context, arg sqlc.ListRecommendationPinnedShortIDsByViewerAndShortIDsParams) ([]pgtype.UUID, error) {
	if s.listPinnedShortIDs == nil {
		return []pgtype.UUID{}, nil
	}
	return s.listPinnedShortIDs(ctx, arg)
}

func (s repositoryStubQueries) ListRecommendationShortGlobalFeaturesByShortIDs(ctx context.Context, shortIds []pgtype.UUID) ([]sqlc.AppRecommendationShortGlobalFeature, error) {
	if s.listShortGlobalFeatures == nil {
		return []sqlc.AppRecommendationShortGlobalFeature{}, nil
	}
	return s.listShortGlobalFeatures(ctx, shortIds)
}

func (s repositoryStubQueries) ListRecommendationUnlockedMainIDsByViewerAndMainIDs(ctx context.Context, arg sqlc.ListRecommendationUnlockedMainIDsByViewerAndMainIDsParams) ([]pgtype.UUID, error) {
	if s.listUnlockedMainIDs == nil {
		return []pgtype.UUID{}, nil
	}
	return s.listUnlockedMainIDs(ctx, arg)
}

func (s repositoryStubQueries) ListRecommendationViewerCreatorFeaturesByViewerAndCreatorIDs(ctx context.Context, arg sqlc.ListRecommendationViewerCreatorFeaturesByViewerAndCreatorIDsParams) ([]sqlc.AppRecommendationViewerCreatorFeature, error) {
	if s.listViewerCreatorFeatures == nil {
		return []sqlc.AppRecommendationViewerCreatorFeature{}, nil
	}
	return s.listViewerCreatorFeatures(ctx, arg)
}

func (s repositoryStubQueries) ListRecommendationViewerMainFeaturesByViewerAndMainIDs(ctx context.Context, arg sqlc.ListRecommendationViewerMainFeaturesByViewerAndMainIDsParams) ([]sqlc.AppRecommendationViewerMainFeature, error) {
	if s.listViewerMainFeatures == nil {
		return []sqlc.AppRecommendationViewerMainFeature{}, nil
	}
	return s.listViewerMainFeatures(ctx, arg)
}

func (s repositoryStubQueries) ListRecommendationViewerShortFeaturesByViewerAndShortIDs(ctx context.Context, arg sqlc.ListRecommendationViewerShortFeaturesByViewerAndShortIDsParams) ([]sqlc.AppRecommendationViewerShortFeature, error) {
	if s.listViewerShortFeatures == nil {
		return []sqlc.AppRecommendationViewerShortFeature{}, nil
	}
	return s.listViewerShortFeatures(ctx, arg)
}

func (s repositoryStubQueries) UpsertRecommendationShortGlobalFeatures(ctx context.Context, arg sqlc.UpsertRecommendationShortGlobalFeaturesParams) (int64, error) {
	if s.upsertShortGlobalFeatures == nil {
		return 1, nil
	}
	return s.upsertShortGlobalFeatures(ctx, arg)
}

func (s repositoryStubQueries) UpsertRecommendationViewerCreatorFeatures(ctx context.Context, arg sqlc.UpsertRecommendationViewerCreatorFeaturesParams) error {
	if s.upsertViewerCreatorFeatures == nil {
		return nil
	}
	return s.upsertViewerCreatorFeatures(ctx, arg)
}

func (s repositoryStubQueries) UpsertRecommendationViewerMainFeatures(ctx context.Context, arg sqlc.UpsertRecommendationViewerMainFeaturesParams) (int64, error) {
	if s.upsertViewerMainFeatures == nil {
		return 1, nil
	}
	return s.upsertViewerMainFeatures(ctx, arg)
}

func (s repositoryStubQueries) UpsertRecommendationViewerShortFeatures(ctx context.Context, arg sqlc.UpsertRecommendationViewerShortFeaturesParams) (int64, error) {
	if s.upsertViewerShortFeatures == nil {
		return 1, nil
	}
	return s.upsertViewerShortFeatures(ctx, arg)
}

type stubTxBeginner struct {
	tx  pgx.Tx
	err error
}

func (s stubTxBeginner) Begin(context.Context) (pgx.Tx, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.tx, nil
}

type fakeTx struct {
	committed  bool
	rolledBack bool
}

func (f *fakeTx) Begin(context.Context) (pgx.Tx, error) {
	return nil, errors.New("unexpected nested tx")
}
func (f *fakeTx) Commit(context.Context) error {
	f.committed = true
	return nil
}
func (f *fakeTx) Rollback(context.Context) error {
	f.rolledBack = true
	return nil
}
func (f *fakeTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	panic("unexpected CopyFrom")
}
func (f *fakeTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	panic("unexpected SendBatch")
}
func (f *fakeTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }
func (f *fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	panic("unexpected Prepare")
}
func (f *fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("unexpected Exec")
}
func (f *fakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unexpected Query")
}
func (f *fakeTx) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("unexpected QueryRow")
}
func (f *fakeTx) Conn() *pgx.Conn { return nil }

func TestRecordEventImpressionInsertedUpdatesAllFeatures(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.New()
	creatorID := uuid.New()
	mainID := uuid.New()
	shortID := uuid.New()
	tx := &fakeTx{}

	var insertArg sqlc.InsertRecommendationEventParams
	var viewerShortArg sqlc.UpsertRecommendationViewerShortFeaturesParams
	var viewerCreatorArg sqlc.UpsertRecommendationViewerCreatorFeaturesParams
	var viewerMainArg sqlc.UpsertRecommendationViewerMainFeaturesParams
	var shortGlobalArg sqlc.UpsertRecommendationShortGlobalFeaturesParams

	stub := repositoryStubQueries{
		insertEvent: func(_ context.Context, arg sqlc.InsertRecommendationEventParams) (sqlc.AppRecommendationEvent, error) {
			insertArg = arg
			return testInsertRecommendationEventRow(viewerID, now, "impression", "impression-1"), nil
		},
		upsertViewerShortFeatures: func(_ context.Context, arg sqlc.UpsertRecommendationViewerShortFeaturesParams) (int64, error) {
			viewerShortArg = arg
			return 1, nil
		},
		upsertViewerCreatorFeatures: func(_ context.Context, arg sqlc.UpsertRecommendationViewerCreatorFeaturesParams) error {
			viewerCreatorArg = arg
			return nil
		},
		upsertViewerMainFeatures: func(_ context.Context, arg sqlc.UpsertRecommendationViewerMainFeaturesParams) (int64, error) {
			viewerMainArg = arg
			return 1, nil
		},
		upsertShortGlobalFeatures: func(_ context.Context, arg sqlc.UpsertRecommendationShortGlobalFeaturesParams) (int64, error) {
			shortGlobalArg = arg
			return 1, nil
		},
	}

	repo := newRepository(stubTxBeginner{tx: tx}, stub, func(sqlc.DBTX) queries { return stub })
	result, err := repo.RecordEvent(context.Background(), RecordEventInput{
		ViewerUserID:    viewerID,
		EventKind:       EventKindImpression,
		CreatorUserID:   uuidPtr(creatorID),
		CanonicalMainID: uuidPtr(mainID),
		ShortID:         uuidPtr(shortID),
		OccurredAt:      timePtr(now),
		IdempotencyKey:  " impression-1 ",
	})
	if err != nil {
		t.Fatalf("RecordEvent() error = %v, want nil", err)
	}
	if !result.Recorded || result.EventKind != EventKindImpression || result.Idempotency != "impression-1" {
		t.Fatalf("RecordEvent() result got %#v", result)
	}
	if insertArg.IdempotencyKey != "impression-1" {
		t.Fatalf("InsertRecommendationEvent() idempotency got %q want %q", insertArg.IdempotencyKey, "impression-1")
	}
	assertCountAndTime(t, viewerShortArg.ImpressionCount, viewerShortArg.LastImpressionAt, now)
	assertCountAndTime(t, viewerCreatorArg.ImpressionCount, viewerCreatorArg.LastImpressionAt, now)
	assertCountAndTime(t, viewerMainArg.ImpressionCount, viewerMainArg.LastImpressionAt, now)
	assertCountAndTime(t, shortGlobalArg.ImpressionCount, shortGlobalArg.LastImpressionAt, now)
	if viewerShortArg.ShortID != pgUUID(shortID) || viewerMainArg.CanonicalMainID != pgUUID(mainID) || viewerCreatorArg.CreatorUserID != pgUUID(creatorID) {
		t.Fatalf("RecordEvent() aggregate args got short=%v main=%v creator=%v", viewerShortArg.ShortID, viewerMainArg.CanonicalMainID, viewerCreatorArg.CreatorUserID)
	}
	if !tx.committed || tx.rolledBack {
		t.Fatalf("transaction state got committed=%t rolledBack=%t want true false", tx.committed, tx.rolledBack)
	}
}

func TestRecordEventDuplicateSkipsAggregateUpserts(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000100, 0).UTC()
	viewerID := uuid.New()
	creatorID := uuid.New()
	mainID := uuid.New()
	shortID := uuid.New()
	tx := &fakeTx{}

	called := false
	stub := repositoryStubQueries{
		insertEvent: func(context.Context, sqlc.InsertRecommendationEventParams) (sqlc.AppRecommendationEvent, error) {
			return sqlc.AppRecommendationEvent{}, pgx.ErrNoRows
		},
		getEvent: func(context.Context, sqlc.GetRecommendationEventByViewerAndIdempotencyKeyParams) (sqlc.AppRecommendationEvent, error) {
			return testGetRecommendationEventRowWithTargets(
				viewerID,
				creatorID,
				mainID,
				shortID,
				now,
				"view_start",
				"dup-1",
			), nil
		},
		upsertViewerShortFeatures: func(context.Context, sqlc.UpsertRecommendationViewerShortFeaturesParams) (int64, error) {
			called = true
			return 1, nil
		},
		upsertViewerCreatorFeatures: func(context.Context, sqlc.UpsertRecommendationViewerCreatorFeaturesParams) error {
			called = true
			return nil
		},
		upsertViewerMainFeatures: func(context.Context, sqlc.UpsertRecommendationViewerMainFeaturesParams) (int64, error) {
			called = true
			return 1, nil
		},
		upsertShortGlobalFeatures: func(context.Context, sqlc.UpsertRecommendationShortGlobalFeaturesParams) (int64, error) {
			called = true
			return 1, nil
		},
	}

	repo := newRepository(stubTxBeginner{tx: tx}, stub, func(sqlc.DBTX) queries { return stub })
	result, err := repo.RecordEvent(context.Background(), RecordEventInput{
		ViewerUserID:    viewerID,
		EventKind:       EventKindViewStart,
		CreatorUserID:   uuidPtr(creatorID),
		CanonicalMainID: uuidPtr(mainID),
		ShortID:         uuidPtr(shortID),
		IdempotencyKey:  "dup-1",
	})
	if err != nil {
		t.Fatalf("RecordEvent() error = %v, want nil", err)
	}
	if result.Recorded {
		t.Fatalf("RecordEvent() recorded got %t want false", result.Recorded)
	}
	if called {
		t.Fatal("RecordEvent() aggregate upsert called for duplicate event")
	}
}

func TestRecordEventDuplicateWithDifferentPayloadReturnsError(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000150, 0).UTC()
	viewerID := uuid.New()
	creatorID := uuid.New()
	mainID := uuid.New()
	shortID := uuid.New()
	tx := &fakeTx{}

	stub := repositoryStubQueries{
		insertEvent: func(context.Context, sqlc.InsertRecommendationEventParams) (sqlc.AppRecommendationEvent, error) {
			return sqlc.AppRecommendationEvent{}, pgx.ErrNoRows
		},
		getEvent: func(context.Context, sqlc.GetRecommendationEventByViewerAndIdempotencyKeyParams) (sqlc.AppRecommendationEvent, error) {
			return testGetRecommendationEventRowWithTargets(
				viewerID,
				creatorID,
				mainID,
				shortID,
				now,
				string(EventKindViewStart),
				"dup-2",
			), nil
		},
	}

	repo := newRepository(stubTxBeginner{tx: tx}, stub, func(sqlc.DBTX) queries { return stub })
	_, err := repo.RecordEvent(context.Background(), RecordEventInput{
		ViewerUserID:    viewerID,
		EventKind:       EventKindImpression,
		CreatorUserID:   uuidPtr(creatorID),
		CanonicalMainID: uuidPtr(mainID),
		ShortID:         uuidPtr(shortID),
		IdempotencyKey:  "dup-2",
	})
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("RecordEvent() error got %v want %v", err, ErrIdempotencyConflict)
	}
	if tx.committed || !tx.rolledBack {
		t.Fatalf("transaction state got committed=%t rolledBack=%t want false true", tx.committed, tx.rolledBack)
	}
}

func TestRecordEventDuplicateWithDifferentOccurredAtReturnsError(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000160, 0).UTC()
	later := now.Add(time.Minute)
	viewerID := uuid.New()
	creatorID := uuid.New()
	mainID := uuid.New()
	shortID := uuid.New()
	tx := &fakeTx{}

	stub := repositoryStubQueries{
		insertEvent: func(context.Context, sqlc.InsertRecommendationEventParams) (sqlc.AppRecommendationEvent, error) {
			return sqlc.AppRecommendationEvent{}, pgx.ErrNoRows
		},
		getEvent: func(context.Context, sqlc.GetRecommendationEventByViewerAndIdempotencyKeyParams) (sqlc.AppRecommendationEvent, error) {
			return testGetRecommendationEventRowWithTargets(
				viewerID,
				creatorID,
				mainID,
				shortID,
				now,
				string(EventKindViewStart),
				"dup-3",
			), nil
		},
	}

	repo := newRepository(stubTxBeginner{tx: tx}, stub, func(sqlc.DBTX) queries { return stub })
	_, err := repo.RecordEvent(context.Background(), RecordEventInput{
		ViewerUserID:    viewerID,
		EventKind:       EventKindViewStart,
		CreatorUserID:   uuidPtr(creatorID),
		CanonicalMainID: uuidPtr(mainID),
		ShortID:         uuidPtr(shortID),
		OccurredAt:      timePtr(later),
		IdempotencyKey:  "dup-3",
	})
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("RecordEvent() error got %v want %v", err, ErrIdempotencyConflict)
	}
	if tx.committed || !tx.rolledBack {
		t.Fatalf("transaction state got committed=%t rolledBack=%t want false true", tx.committed, tx.rolledBack)
	}
}

func TestRecordEventDuplicateWithExplicitOccurredAtMicrosecondPrecisionAccepted(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000170, 123456789).UTC()
	storedAt := now.Round(time.Microsecond)
	viewerID := uuid.New()
	creatorID := uuid.New()
	mainID := uuid.New()
	shortID := uuid.New()
	tx := &fakeTx{}

	stub := repositoryStubQueries{
		insertEvent: func(context.Context, sqlc.InsertRecommendationEventParams) (sqlc.AppRecommendationEvent, error) {
			return sqlc.AppRecommendationEvent{}, pgx.ErrNoRows
		},
		getEvent: func(context.Context, sqlc.GetRecommendationEventByViewerAndIdempotencyKeyParams) (sqlc.AppRecommendationEvent, error) {
			return testGetRecommendationEventRowWithTargets(
				viewerID,
				creatorID,
				mainID,
				shortID,
				storedAt,
				string(EventKindViewStart),
				"dup-4",
			), nil
		},
	}

	repo := newRepository(stubTxBeginner{tx: tx}, stub, func(sqlc.DBTX) queries { return stub })
	result, err := repo.RecordEvent(context.Background(), RecordEventInput{
		ViewerUserID:    viewerID,
		EventKind:       EventKindViewStart,
		CreatorUserID:   uuidPtr(creatorID),
		CanonicalMainID: uuidPtr(mainID),
		ShortID:         uuidPtr(shortID),
		OccurredAt:      timePtr(now),
		IdempotencyKey:  "dup-4",
	})
	if err != nil {
		t.Fatalf("RecordEvent() error = %v, want nil", err)
	}
	if result.Recorded {
		t.Fatalf("RecordEvent() recorded got %t want false", result.Recorded)
	}
	if !tx.committed || tx.rolledBack {
		t.Fatalf("transaction state got committed=%t rolledBack=%t want true false", tx.committed, tx.rolledBack)
	}
}

func TestRecordEventProfileClickOnlyUpdatesCreatorFeatures(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000200, 0).UTC()
	viewerID := uuid.New()
	creatorID := uuid.New()
	tx := &fakeTx{}

	viewerCreatorCalled := false
	otherCalled := false
	stub := repositoryStubQueries{
		insertEvent: func(context.Context, sqlc.InsertRecommendationEventParams) (sqlc.AppRecommendationEvent, error) {
			return testInsertRecommendationEventRow(viewerID, now, "profile_click", "profile-1"), nil
		},
		upsertViewerCreatorFeatures: func(_ context.Context, arg sqlc.UpsertRecommendationViewerCreatorFeaturesParams) error {
			viewerCreatorCalled = true
			assertCountAndTime(t, arg.ProfileClickCount, arg.LastProfileClickAt, now)
			return nil
		},
		upsertViewerShortFeatures: func(context.Context, sqlc.UpsertRecommendationViewerShortFeaturesParams) (int64, error) {
			otherCalled = true
			return 1, nil
		},
		upsertViewerMainFeatures: func(context.Context, sqlc.UpsertRecommendationViewerMainFeaturesParams) (int64, error) {
			otherCalled = true
			return 1, nil
		},
		upsertShortGlobalFeatures: func(context.Context, sqlc.UpsertRecommendationShortGlobalFeaturesParams) (int64, error) {
			otherCalled = true
			return 1, nil
		},
	}

	repo := newRepository(stubTxBeginner{tx: tx}, stub, func(sqlc.DBTX) queries { return stub })
	_, err := repo.RecordEvent(context.Background(), RecordEventInput{
		ViewerUserID:   viewerID,
		EventKind:      EventKindProfileClick,
		CreatorUserID:  uuidPtr(creatorID),
		IdempotencyKey: "profile-1",
	})
	if err != nil {
		t.Fatalf("RecordEvent() error = %v, want nil", err)
	}
	if !viewerCreatorCalled || otherCalled {
		t.Fatalf("profile click aggregate calls got creator=%t other=%t want true false", viewerCreatorCalled, otherCalled)
	}
}

func TestRecordEventMainClickWithoutShortSkipsShortAggregates(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000300, 0).UTC()
	viewerID := uuid.New()
	creatorID := uuid.New()
	mainID := uuid.New()
	tx := &fakeTx{}

	viewerCreatorCalled := false
	viewerMainCalled := false
	shortCalled := false
	stub := repositoryStubQueries{
		insertEvent: func(context.Context, sqlc.InsertRecommendationEventParams) (sqlc.AppRecommendationEvent, error) {
			return testInsertRecommendationEventRow(viewerID, now, "main_click", "main-1"), nil
		},
		upsertViewerCreatorFeatures: func(_ context.Context, arg sqlc.UpsertRecommendationViewerCreatorFeaturesParams) error {
			viewerCreatorCalled = true
			assertCountAndTime(t, arg.MainClickCount, arg.LastMainClickAt, now)
			return nil
		},
		upsertViewerMainFeatures: func(_ context.Context, arg sqlc.UpsertRecommendationViewerMainFeaturesParams) (int64, error) {
			viewerMainCalled = true
			assertCountAndTime(t, arg.MainClickCount, arg.LastMainClickAt, now)
			return 1, nil
		},
		upsertViewerShortFeatures: func(context.Context, sqlc.UpsertRecommendationViewerShortFeaturesParams) (int64, error) {
			shortCalled = true
			return 1, nil
		},
		upsertShortGlobalFeatures: func(context.Context, sqlc.UpsertRecommendationShortGlobalFeaturesParams) (int64, error) {
			shortCalled = true
			return 1, nil
		},
	}

	repo := newRepository(stubTxBeginner{tx: tx}, stub, func(sqlc.DBTX) queries { return stub })
	_, err := repo.RecordEvent(context.Background(), RecordEventInput{
		ViewerUserID:    viewerID,
		EventKind:       EventKindMainClick,
		CreatorUserID:   uuidPtr(creatorID),
		CanonicalMainID: uuidPtr(mainID),
		IdempotencyKey:  "main-1",
	})
	if err != nil {
		t.Fatalf("RecordEvent() error = %v, want nil", err)
	}
	if !viewerCreatorCalled || !viewerMainCalled || shortCalled {
		t.Fatalf("main click aggregate calls got creator=%t main=%t short=%t want true true false", viewerCreatorCalled, viewerMainCalled, shortCalled)
	}
}

func TestRecordEventValidation(t *testing.T) {
	t.Parallel()

	validViewerID := uuid.New()
	validCreatorID := uuid.New()
	validMainID := uuid.New()
	validShortID := uuid.New()

	tests := []struct {
		name  string
		input RecordEventInput
		want  error
	}{
		{
			name: "missing viewer",
			input: RecordEventInput{
				EventKind:       EventKindProfileClick,
				CreatorUserID:   uuidPtr(validCreatorID),
				IdempotencyKey:  "k",
				CanonicalMainID: uuidPtr(validMainID),
			},
			want: ErrViewerUserIDRequired,
		},
		{
			name: "missing idempotency",
			input: RecordEventInput{
				ViewerUserID:   validViewerID,
				EventKind:      EventKindProfileClick,
				CreatorUserID:  uuidPtr(validCreatorID),
				IdempotencyKey: " ",
			},
			want: ErrIdempotencyKeyRequired,
		},
		{
			name: "missing creator for profile click",
			input: RecordEventInput{
				ViewerUserID:   validViewerID,
				EventKind:      EventKindProfileClick,
				IdempotencyKey: "k",
			},
			want: ErrCreatorUserIDRequired,
		},
		{
			name: "profile click forbids canonical main",
			input: RecordEventInput{
				ViewerUserID:    validViewerID,
				EventKind:       EventKindProfileClick,
				CreatorUserID:   uuidPtr(validCreatorID),
				CanonicalMainID: uuidPtr(validMainID),
				IdempotencyKey:  "k",
			},
			want: ErrCanonicalMainIDForbidden,
		},
		{
			name: "profile click forbids short",
			input: RecordEventInput{
				ViewerUserID:   validViewerID,
				EventKind:      EventKindProfileClick,
				CreatorUserID:  uuidPtr(validCreatorID),
				ShortID:        uuidPtr(validShortID),
				IdempotencyKey: "k",
			},
			want: ErrShortIDForbidden,
		},
		{
			name: "missing short for impression",
			input: RecordEventInput{
				ViewerUserID:    validViewerID,
				EventKind:       EventKindImpression,
				CreatorUserID:   uuidPtr(validCreatorID),
				CanonicalMainID: uuidPtr(validMainID),
				IdempotencyKey:  "k",
			},
			want: ErrShortIDRequired,
		},
		{
			name: "main click rejects nil short pointer",
			input: RecordEventInput{
				ViewerUserID:    validViewerID,
				EventKind:       EventKindMainClick,
				CreatorUserID:   uuidPtr(validCreatorID),
				CanonicalMainID: uuidPtr(validMainID),
				ShortID:         uuidPtr(uuid.Nil),
				IdempotencyKey:  "k",
			},
			want: ErrShortIDInvalid,
		},
		{
			name: "invalid kind",
			input: RecordEventInput{
				ViewerUserID:    validViewerID,
				EventKind:       EventKind("unknown"),
				CreatorUserID:   uuidPtr(validCreatorID),
				CanonicalMainID: uuidPtr(validMainID),
				ShortID:         uuidPtr(validShortID),
				IdempotencyKey:  "k",
			},
			want: ErrEventKindInvalid,
		},
	}

	repo := newRepository(nil, repositoryStubQueries{}, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := repo.RecordEvent(context.Background(), tt.input); !errors.Is(err, tt.want) {
				t.Fatalf("RecordEvent() error got %v want %v", err, tt.want)
			}
		})
	}
}

func TestListRecommendationReads(t *testing.T) {
	t.Parallel()

	viewerID := uuid.New()
	creatorID := uuid.New()
	mainID := uuid.New()
	shortID := uuid.New()
	now := time.Unix(1710000400, 0).UTC()

	var gotShortParams sqlc.ListRecommendationViewerShortFeaturesByViewerAndShortIDsParams
	var gotCreatorParams sqlc.ListRecommendationViewerCreatorFeaturesByViewerAndCreatorIDsParams
	var gotMainParams sqlc.ListRecommendationViewerMainFeaturesByViewerAndMainIDsParams
	var gotPinnedParams sqlc.ListRecommendationPinnedShortIDsByViewerAndShortIDsParams
	var gotFollowedParams sqlc.ListRecommendationFollowedCreatorIDsByViewerAndCreatorIDsParams
	var gotUnlockedParams sqlc.ListRecommendationUnlockedMainIDsByViewerAndMainIDsParams
	var gotShortGlobalIDs []pgtype.UUID

	stub := repositoryStubQueries{
		listViewerShortFeatures: func(_ context.Context, arg sqlc.ListRecommendationViewerShortFeaturesByViewerAndShortIDsParams) ([]sqlc.AppRecommendationViewerShortFeature, error) {
			gotShortParams = arg
			return []sqlc.AppRecommendationViewerShortFeature{testViewerShortFeatureRow(viewerID, creatorID, mainID, shortID, now)}, nil
		},
		listViewerCreatorFeatures: func(_ context.Context, arg sqlc.ListRecommendationViewerCreatorFeaturesByViewerAndCreatorIDsParams) ([]sqlc.AppRecommendationViewerCreatorFeature, error) {
			gotCreatorParams = arg
			return []sqlc.AppRecommendationViewerCreatorFeature{testViewerCreatorFeatureRow(viewerID, creatorID, now)}, nil
		},
		listViewerMainFeatures: func(_ context.Context, arg sqlc.ListRecommendationViewerMainFeaturesByViewerAndMainIDsParams) ([]sqlc.AppRecommendationViewerMainFeature, error) {
			gotMainParams = arg
			return []sqlc.AppRecommendationViewerMainFeature{testViewerMainFeatureRow(viewerID, creatorID, mainID, now)}, nil
		},
		listShortGlobalFeatures: func(_ context.Context, shortIDs []pgtype.UUID) ([]sqlc.AppRecommendationShortGlobalFeature, error) {
			gotShortGlobalIDs = shortIDs
			return []sqlc.AppRecommendationShortGlobalFeature{testShortGlobalFeatureRow(creatorID, mainID, shortID, now)}, nil
		},
		listPinnedShortIDs: func(_ context.Context, arg sqlc.ListRecommendationPinnedShortIDsByViewerAndShortIDsParams) ([]pgtype.UUID, error) {
			gotPinnedParams = arg
			return []pgtype.UUID{pgUUID(shortID)}, nil
		},
		listFollowedCreatorIDs: func(_ context.Context, arg sqlc.ListRecommendationFollowedCreatorIDsByViewerAndCreatorIDsParams) ([]pgtype.UUID, error) {
			gotFollowedParams = arg
			return []pgtype.UUID{pgUUID(creatorID)}, nil
		},
		listUnlockedMainIDs: func(_ context.Context, arg sqlc.ListRecommendationUnlockedMainIDsByViewerAndMainIDsParams) ([]pgtype.UUID, error) {
			gotUnlockedParams = arg
			return []pgtype.UUID{pgUUID(mainID)}, nil
		},
	}

	repo := newRepository(nil, stub, nil)

	shortFeatures, err := repo.ListViewerShortFeatures(context.Background(), viewerID, []uuid.UUID{shortID})
	if err != nil {
		t.Fatalf("ListViewerShortFeatures() error = %v, want nil", err)
	}
	if gotShortParams.ViewerUserID != pgUUID(viewerID) || !reflect.DeepEqual(gotShortParams.ShortIds, []pgtype.UUID{pgUUID(shortID)}) {
		t.Fatalf("ListViewerShortFeatures() params got %#v", gotShortParams)
	}
	if len(shortFeatures) != 1 || shortFeatures[0].ShortID != shortID || shortFeatures[0].ImpressionCount != 2 {
		t.Fatalf("ListViewerShortFeatures() got %#v", shortFeatures)
	}

	creatorFeatures, err := repo.ListViewerCreatorFeatures(context.Background(), viewerID, []uuid.UUID{creatorID})
	if err != nil {
		t.Fatalf("ListViewerCreatorFeatures() error = %v, want nil", err)
	}
	if gotCreatorParams.ViewerUserID != pgUUID(viewerID) || !reflect.DeepEqual(gotCreatorParams.CreatorUserIds, []pgtype.UUID{pgUUID(creatorID)}) {
		t.Fatalf("ListViewerCreatorFeatures() params got %#v", gotCreatorParams)
	}
	if len(creatorFeatures) != 1 || creatorFeatures[0].CreatorUserID != creatorID || creatorFeatures[0].ProfileClickCount != 1 {
		t.Fatalf("ListViewerCreatorFeatures() got %#v", creatorFeatures)
	}

	mainFeatures, err := repo.ListViewerMainFeatures(context.Background(), viewerID, []uuid.UUID{mainID})
	if err != nil {
		t.Fatalf("ListViewerMainFeatures() error = %v, want nil", err)
	}
	if gotMainParams.ViewerUserID != pgUUID(viewerID) || !reflect.DeepEqual(gotMainParams.CanonicalMainIds, []pgtype.UUID{pgUUID(mainID)}) {
		t.Fatalf("ListViewerMainFeatures() params got %#v", gotMainParams)
	}
	if len(mainFeatures) != 1 || mainFeatures[0].CanonicalMainID != mainID || mainFeatures[0].MainClickCount != 1 {
		t.Fatalf("ListViewerMainFeatures() got %#v", mainFeatures)
	}

	globalFeatures, err := repo.ListShortGlobalFeatures(context.Background(), []uuid.UUID{shortID})
	if err != nil {
		t.Fatalf("ListShortGlobalFeatures() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(gotShortGlobalIDs, []pgtype.UUID{pgUUID(shortID)}) {
		t.Fatalf("ListShortGlobalFeatures() ids got %#v", gotShortGlobalIDs)
	}
	if len(globalFeatures) != 1 || globalFeatures[0].ShortID != shortID || globalFeatures[0].UnlockConversionCount != 1 {
		t.Fatalf("ListShortGlobalFeatures() got %#v", globalFeatures)
	}

	pinnedIDs, err := repo.ListViewerPinnedShortIDs(context.Background(), viewerID, []uuid.UUID{shortID})
	if err != nil {
		t.Fatalf("ListViewerPinnedShortIDs() error = %v, want nil", err)
	}
	if gotPinnedParams.ViewerUserID != pgUUID(viewerID) || !reflect.DeepEqual(gotPinnedParams.ShortIds, []pgtype.UUID{pgUUID(shortID)}) {
		t.Fatalf("ListViewerPinnedShortIDs() params got %#v", gotPinnedParams)
	}
	if !reflect.DeepEqual(pinnedIDs, []uuid.UUID{shortID}) {
		t.Fatalf("ListViewerPinnedShortIDs() got %#v want %#v", pinnedIDs, []uuid.UUID{shortID})
	}

	followedIDs, err := repo.ListViewerFollowedCreatorIDs(context.Background(), viewerID, []uuid.UUID{creatorID})
	if err != nil {
		t.Fatalf("ListViewerFollowedCreatorIDs() error = %v, want nil", err)
	}
	if gotFollowedParams.ViewerUserID != pgUUID(viewerID) || !reflect.DeepEqual(gotFollowedParams.CreatorUserIds, []pgtype.UUID{pgUUID(creatorID)}) {
		t.Fatalf("ListViewerFollowedCreatorIDs() params got %#v", gotFollowedParams)
	}
	if !reflect.DeepEqual(followedIDs, []uuid.UUID{creatorID}) {
		t.Fatalf("ListViewerFollowedCreatorIDs() got %#v want %#v", followedIDs, []uuid.UUID{creatorID})
	}

	unlockedIDs, err := repo.ListViewerUnlockedMainIDs(context.Background(), viewerID, []uuid.UUID{mainID})
	if err != nil {
		t.Fatalf("ListViewerUnlockedMainIDs() error = %v, want nil", err)
	}
	if gotUnlockedParams.ViewerUserID != pgUUID(viewerID) || !reflect.DeepEqual(gotUnlockedParams.CanonicalMainIds, []pgtype.UUID{pgUUID(mainID)}) {
		t.Fatalf("ListViewerUnlockedMainIDs() params got %#v", gotUnlockedParams)
	}
	if !reflect.DeepEqual(unlockedIDs, []uuid.UUID{mainID}) {
		t.Fatalf("ListViewerUnlockedMainIDs() got %#v want %#v", unlockedIDs, []uuid.UUID{mainID})
	}
}

func TestListRecommendationReadsEmptyInputSkipsQueries(t *testing.T) {
	t.Parallel()

	repo := newRepository(nil, repositoryStubQueries{
		listViewerShortFeatures: func(context.Context, sqlc.ListRecommendationViewerShortFeaturesByViewerAndShortIDsParams) ([]sqlc.AppRecommendationViewerShortFeature, error) {
			t.Fatal("ListViewerShortFeatures() query called for empty input")
			return nil, nil
		},
		listViewerCreatorFeatures: func(context.Context, sqlc.ListRecommendationViewerCreatorFeaturesByViewerAndCreatorIDsParams) ([]sqlc.AppRecommendationViewerCreatorFeature, error) {
			t.Fatal("ListViewerCreatorFeatures() query called for empty input")
			return nil, nil
		},
		listViewerMainFeatures: func(context.Context, sqlc.ListRecommendationViewerMainFeaturesByViewerAndMainIDsParams) ([]sqlc.AppRecommendationViewerMainFeature, error) {
			t.Fatal("ListViewerMainFeatures() query called for empty input")
			return nil, nil
		},
		listShortGlobalFeatures: func(context.Context, []pgtype.UUID) ([]sqlc.AppRecommendationShortGlobalFeature, error) {
			t.Fatal("ListShortGlobalFeatures() query called for empty input")
			return nil, nil
		},
		listPinnedShortIDs: func(context.Context, sqlc.ListRecommendationPinnedShortIDsByViewerAndShortIDsParams) ([]pgtype.UUID, error) {
			t.Fatal("ListViewerPinnedShortIDs() query called for empty input")
			return nil, nil
		},
		listFollowedCreatorIDs: func(context.Context, sqlc.ListRecommendationFollowedCreatorIDsByViewerAndCreatorIDsParams) ([]pgtype.UUID, error) {
			t.Fatal("ListViewerFollowedCreatorIDs() query called for empty input")
			return nil, nil
		},
		listUnlockedMainIDs: func(context.Context, sqlc.ListRecommendationUnlockedMainIDsByViewerAndMainIDsParams) ([]pgtype.UUID, error) {
			t.Fatal("ListViewerUnlockedMainIDs() query called for empty input")
			return nil, nil
		},
	}, nil)

	viewerID := uuid.New()
	if items, err := repo.ListViewerShortFeatures(context.Background(), viewerID, nil); err != nil || len(items) != 0 {
		t.Fatalf("ListViewerShortFeatures() got items=%#v err=%v want empty nil", items, err)
	}
	if items, err := repo.ListViewerCreatorFeatures(context.Background(), viewerID, nil); err != nil || len(items) != 0 {
		t.Fatalf("ListViewerCreatorFeatures() got items=%#v err=%v want empty nil", items, err)
	}
	if items, err := repo.ListViewerMainFeatures(context.Background(), viewerID, nil); err != nil || len(items) != 0 {
		t.Fatalf("ListViewerMainFeatures() got items=%#v err=%v want empty nil", items, err)
	}
	if items, err := repo.ListShortGlobalFeatures(context.Background(), nil); err != nil || len(items) != 0 {
		t.Fatalf("ListShortGlobalFeatures() got items=%#v err=%v want empty nil", items, err)
	}
	if items, err := repo.ListViewerPinnedShortIDs(context.Background(), viewerID, nil); err != nil || len(items) != 0 {
		t.Fatalf("ListViewerPinnedShortIDs() got items=%#v err=%v want empty nil", items, err)
	}
	if items, err := repo.ListViewerFollowedCreatorIDs(context.Background(), viewerID, nil); err != nil || len(items) != 0 {
		t.Fatalf("ListViewerFollowedCreatorIDs() got items=%#v err=%v want empty nil", items, err)
	}
	if items, err := repo.ListViewerUnlockedMainIDs(context.Background(), viewerID, nil); err != nil || len(items) != 0 {
		t.Fatalf("ListViewerUnlockedMainIDs() got items=%#v err=%v want empty nil", items, err)
	}
}

func TestListRecommendationReadsConversionErrors(t *testing.T) {
	t.Parallel()

	viewerID := uuid.New()
	shortID := uuid.New()
	creatorID := uuid.New()
	mainID := uuid.New()
	repo := newRepository(nil, repositoryStubQueries{
		listViewerShortFeatures: func(context.Context, sqlc.ListRecommendationViewerShortFeaturesByViewerAndShortIDsParams) ([]sqlc.AppRecommendationViewerShortFeature, error) {
			row := testViewerShortFeatureRow(viewerID, creatorID, mainID, shortID, time.Now().UTC())
			row.ShortID = pgtype.UUID{}
			return []sqlc.AppRecommendationViewerShortFeature{row}, nil
		},
		listPinnedShortIDs: func(context.Context, sqlc.ListRecommendationPinnedShortIDsByViewerAndShortIDsParams) ([]pgtype.UUID, error) {
			return []pgtype.UUID{{}}, nil
		},
	}, nil)

	if _, err := repo.ListViewerShortFeatures(context.Background(), viewerID, []uuid.UUID{shortID}); err == nil {
		t.Fatal("ListViewerShortFeatures() error = nil, want conversion error")
	}
	if _, err := repo.ListViewerPinnedShortIDs(context.Background(), viewerID, []uuid.UUID{shortID}); err == nil {
		t.Fatal("ListViewerPinnedShortIDs() error = nil, want conversion error")
	}
}

func testInsertRecommendationEventRow(viewerID uuid.UUID, occurredAt time.Time, kind string, idempotencyKey string) sqlc.AppRecommendationEvent {
	return sqlc.AppRecommendationEvent{
		ID:             pgUUID(uuid.New()),
		ViewerUserID:   pgUUID(viewerID),
		EventKind:      kind,
		OccurredAt:     pgTime(occurredAt),
		IdempotencyKey: idempotencyKey,
		CreatedAt:      pgTime(occurredAt),
		UpdatedAt:      pgTime(occurredAt),
	}
}

func testGetRecommendationEventRowWithTargets(
	viewerID uuid.UUID,
	creatorID uuid.UUID,
	mainID uuid.UUID,
	shortID uuid.UUID,
	occurredAt time.Time,
	kind string,
	idempotencyKey string,
) sqlc.AppRecommendationEvent {
	row := sqlc.AppRecommendationEvent{
		ID:             pgUUID(uuid.New()),
		ViewerUserID:   pgUUID(viewerID),
		EventKind:      kind,
		OccurredAt:     pgTime(occurredAt),
		IdempotencyKey: idempotencyKey,
		CreatedAt:      pgTime(occurredAt),
		UpdatedAt:      pgTime(occurredAt),
	}
	row.CreatorUserID = pgUUID(creatorID)
	row.CanonicalMainID = pgUUID(mainID)
	row.ShortID = pgUUID(shortID)
	return row
}

func testViewerShortFeatureRow(viewerID uuid.UUID, creatorID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID, now time.Time) sqlc.AppRecommendationViewerShortFeature {
	return sqlc.AppRecommendationViewerShortFeature{
		ViewerUserID:           pgUUID(viewerID),
		ShortID:                pgUUID(shortID),
		CreatorUserID:          pgUUID(creatorID),
		CanonicalMainID:        pgUUID(mainID),
		ImpressionCount:        2,
		LastImpressionAt:       pgTime(now),
		ViewStartCount:         1,
		LastViewStartAt:        pgTime(now),
		ViewCompletionCount:    1,
		LastViewCompletionAt:   pgTime(now),
		RewatchLoopCount:       1,
		LastRewatchLoopAt:      pgTime(now),
		MainClickCount:         1,
		LastMainClickAt:        pgTime(now),
		UnlockConversionCount:  1,
		LastUnlockConversionAt: pgTime(now),
	}
}

func testViewerCreatorFeatureRow(viewerID uuid.UUID, creatorID uuid.UUID, now time.Time) sqlc.AppRecommendationViewerCreatorFeature {
	return sqlc.AppRecommendationViewerCreatorFeature{
		ViewerUserID:           pgUUID(viewerID),
		CreatorUserID:          pgUUID(creatorID),
		ImpressionCount:        2,
		LastImpressionAt:       pgTime(now),
		ViewStartCount:         1,
		LastViewStartAt:        pgTime(now),
		ViewCompletionCount:    1,
		LastViewCompletionAt:   pgTime(now),
		RewatchLoopCount:       1,
		LastRewatchLoopAt:      pgTime(now),
		ProfileClickCount:      1,
		LastProfileClickAt:     pgTime(now),
		MainClickCount:         1,
		LastMainClickAt:        pgTime(now),
		UnlockConversionCount:  1,
		LastUnlockConversionAt: pgTime(now),
	}
}

func testViewerMainFeatureRow(viewerID uuid.UUID, creatorID uuid.UUID, mainID uuid.UUID, now time.Time) sqlc.AppRecommendationViewerMainFeature {
	return sqlc.AppRecommendationViewerMainFeature{
		ViewerUserID:           pgUUID(viewerID),
		CanonicalMainID:        pgUUID(mainID),
		CreatorUserID:          pgUUID(creatorID),
		ImpressionCount:        2,
		LastImpressionAt:       pgTime(now),
		ViewStartCount:         1,
		LastViewStartAt:        pgTime(now),
		ViewCompletionCount:    1,
		LastViewCompletionAt:   pgTime(now),
		RewatchLoopCount:       1,
		LastRewatchLoopAt:      pgTime(now),
		MainClickCount:         1,
		LastMainClickAt:        pgTime(now),
		UnlockConversionCount:  1,
		LastUnlockConversionAt: pgTime(now),
	}
}

func testShortGlobalFeatureRow(creatorID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID, now time.Time) sqlc.AppRecommendationShortGlobalFeature {
	return sqlc.AppRecommendationShortGlobalFeature{
		ShortID:                pgUUID(shortID),
		CreatorUserID:          pgUUID(creatorID),
		CanonicalMainID:        pgUUID(mainID),
		ImpressionCount:        2,
		LastImpressionAt:       pgTime(now),
		ViewStartCount:         1,
		LastViewStartAt:        pgTime(now),
		ViewCompletionCount:    1,
		LastViewCompletionAt:   pgTime(now),
		RewatchLoopCount:       1,
		LastRewatchLoopAt:      pgTime(now),
		MainClickCount:         1,
		LastMainClickAt:        pgTime(now),
		UnlockConversionCount:  1,
		LastUnlockConversionAt: pgTime(now),
	}
}

func assertCountAndTime(t *testing.T, gotCount int64, gotTime pgtype.Timestamptz, want time.Time) {
	t.Helper()

	if gotCount != 1 {
		t.Fatalf("count got %d want %d", gotCount, 1)
	}
	if !gotTime.Valid || !gotTime.Time.Equal(want) {
		t.Fatalf("time got %#v want %s", gotTime, want)
	}
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

func pgTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func uuidPtr(value uuid.UUID) *uuid.UUID {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
