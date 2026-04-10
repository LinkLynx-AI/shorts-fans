package shorts

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type repositoryStubQueries struct {
	createMain          func(context.Context, sqlc.CreateMainParams) (sqlc.AppMain, error)
	getMain             func(context.Context, pgtype.UUID) (sqlc.AppMain, error)
	listMains           func(context.Context, pgtype.UUID) ([]sqlc.AppMain, error)
	updateMain          func(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error)
	getUnlockableMain   func(context.Context, pgtype.UUID) (sqlc.AppUnlockableMain, error)
	createShort         func(context.Context, sqlc.CreateShortParams) (sqlc.AppShort, error)
	getShort            func(context.Context, pgtype.UUID) (sqlc.AppShort, error)
	listShorts          func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error)
	updateShort         func(context.Context, sqlc.UpdateShortStateParams) (sqlc.AppShort, error)
	publishShort        func(context.Context, pgtype.UUID) (sqlc.AppShort, error)
	listPublicShorts    func(context.Context, pgtype.UUID) ([]sqlc.AppPublicShort, error)
	getPublicShort      func(context.Context, pgtype.UUID) (sqlc.AppPublicShort, error)
	listCanonicalShorts func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error)
	getCanonicalMainID  func(context.Context, pgtype.UUID) (pgtype.UUID, error)
}

func (s repositoryStubQueries) CreateMain(ctx context.Context, arg sqlc.CreateMainParams) (sqlc.AppMain, error) {
	if s.createMain == nil {
		return sqlc.AppMain{}, nil
	}
	return s.createMain(ctx, arg)
}

func (s repositoryStubQueries) GetMainByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
	if s.getMain == nil {
		return sqlc.AppMain{}, nil
	}
	return s.getMain(ctx, id)
}

func (s repositoryStubQueries) ListMainsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMain, error) {
	if s.listMains == nil {
		return nil, nil
	}
	return s.listMains(ctx, creatorUserID)
}

func (s repositoryStubQueries) UpdateMainState(ctx context.Context, arg sqlc.UpdateMainStateParams) (sqlc.AppMain, error) {
	if s.updateMain == nil {
		return sqlc.AppMain{}, nil
	}
	return s.updateMain(ctx, arg)
}

func (s repositoryStubQueries) GetUnlockableMainByID(ctx context.Context, id pgtype.UUID) (sqlc.AppUnlockableMain, error) {
	if s.getUnlockableMain == nil {
		return sqlc.AppUnlockableMain{}, nil
	}
	return s.getUnlockableMain(ctx, id)
}

func (s repositoryStubQueries) CreateShort(ctx context.Context, arg sqlc.CreateShortParams) (sqlc.AppShort, error) {
	if s.createShort == nil {
		return sqlc.AppShort{}, nil
	}
	return s.createShort(ctx, arg)
}

func (s repositoryStubQueries) GetShortByID(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
	if s.getShort == nil {
		return sqlc.AppShort{}, nil
	}
	return s.getShort(ctx, id)
}

func (s repositoryStubQueries) ListShortsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppShort, error) {
	if s.listShorts == nil {
		return nil, nil
	}
	return s.listShorts(ctx, creatorUserID)
}

func (s repositoryStubQueries) UpdateShortState(ctx context.Context, arg sqlc.UpdateShortStateParams) (sqlc.AppShort, error) {
	if s.updateShort == nil {
		return sqlc.AppShort{}, nil
	}
	return s.updateShort(ctx, arg)
}

func (s repositoryStubQueries) PublishShort(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
	if s.publishShort == nil {
		return sqlc.AppShort{}, nil
	}
	return s.publishShort(ctx, id)
}

func (s repositoryStubQueries) ListPublicShortsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppPublicShort, error) {
	if s.listPublicShorts == nil {
		return nil, nil
	}
	return s.listPublicShorts(ctx, creatorUserID)
}

func (s repositoryStubQueries) GetPublicShortByID(ctx context.Context, id pgtype.UUID) (sqlc.AppPublicShort, error) {
	if s.getPublicShort == nil {
		return sqlc.AppPublicShort{}, nil
	}
	return s.getPublicShort(ctx, id)
}

func (s repositoryStubQueries) ListShortsByCanonicalMainID(ctx context.Context, canonicalMainID pgtype.UUID) ([]sqlc.AppShort, error) {
	if s.listCanonicalShorts == nil {
		return nil, nil
	}
	return s.listCanonicalShorts(ctx, canonicalMainID)
}

func (s repositoryStubQueries) GetCanonicalMainIDByShortID(ctx context.Context, id pgtype.UUID) (pgtype.UUID, error) {
	if s.getCanonicalMainID == nil {
		return pgtype.UUID{}, nil
	}
	return s.getCanonicalMainID(ctx, id)
}

func TestBuildCreateParams(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	reviewReason := stringPtr("needs_review")
	postReportState := stringPtr("clean")
	priceMinor := int64(1200)
	currencyCode := "JPY"

	gotMain := buildCreateMainParams(CreateMainInput{
		CreatorUserID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		MediaAssetID:        uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		State:               "draft",
		ReviewReasonCode:    reviewReason,
		PostReportState:     postReportState,
		PriceMinor:          priceMinor,
		CurrencyCode:        currencyCode,
		OwnershipConfirmed:  true,
		ConsentConfirmed:    true,
		ApprovedForUnlockAt: &now,
	})
	wantMain := sqlc.CreateMainParams{
		CreatorUserID:       postgresUUID(uuid.MustParse("00000000-0000-0000-0000-000000000001")),
		MediaAssetID:        postgresUUID(uuid.MustParse("00000000-0000-0000-0000-000000000002")),
		State:               "draft",
		ReviewReasonCode:    textValue(reviewReason),
		PostReportState:     textValue(postReportState),
		PriceMinor:          priceMinor,
		CurrencyCode:        currencyCode,
		OwnershipConfirmed:  true,
		ConsentConfirmed:    true,
		ApprovedForUnlockAt: timestamp(now),
	}
	if !reflect.DeepEqual(gotMain, wantMain) {
		t.Fatalf("buildCreateMainParams() got %#v want %#v", gotMain, wantMain)
	}

	gotShort := buildCreateShortParams(CreateShortInput{
		CreatorUserID:        uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		CanonicalMainID:      uuid.MustParse("00000000-0000-0000-0000-000000000004"),
		MediaAssetID:         uuid.MustParse("00000000-0000-0000-0000-000000000005"),
		State:                "reviewed",
		ReviewReasonCode:     reviewReason,
		PostReportState:      postReportState,
		ApprovedForPublishAt: &now,
		PublishedAt:          &now,
	})
	wantShort := sqlc.CreateShortParams{
		CreatorUserID:        postgresUUID(uuid.MustParse("00000000-0000-0000-0000-000000000003")),
		CanonicalMainID:      postgresUUID(uuid.MustParse("00000000-0000-0000-0000-000000000004")),
		MediaAssetID:         postgresUUID(uuid.MustParse("00000000-0000-0000-0000-000000000005")),
		State:                "reviewed",
		ReviewReasonCode:     textValue(reviewReason),
		PostReportState:      textValue(postReportState),
		ApprovedForPublishAt: timestamp(now),
		PublishedAt:          timestamp(now),
	}
	if !reflect.DeepEqual(gotShort, wantShort) {
		t.Fatalf("buildCreateShortParams() got %#v want %#v", gotShort, wantShort)
	}
}

func TestRepositorySuccessPaths(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.New()
	mainID := uuid.New()
	mainMediaID := uuid.New()
	shortID := uuid.New()
	shortMediaID := uuid.New()
	reviewReason := stringPtr("needs_review")
	postReportState := stringPtr("clean")
	priceMinor := int64(1200)
	currencyCode := "JPY"
	approvedForUnlockAt := timePtr(now.Add(time.Hour))
	approvedForPublishAt := timePtr(now.Add(2 * time.Hour))
	publishedAt := timePtr(now.Add(3 * time.Hour))

	mainRow := testMainRow(mainID, creatorID, mainMediaID, now, reviewReason, postReportState, priceMinor, currencyCode, approvedForUnlockAt)
	unlockableRow := testUnlockableMainRow(mainID, creatorID, mainMediaID, now, reviewReason, postReportState, priceMinor, currencyCode, approvedForUnlockAt)
	shortRow := testShortRow(shortID, creatorID, mainID, shortMediaID, now, reviewReason, postReportState, approvedForPublishAt, publishedAt)
	publicShortRow := testPublicShortRow(shortID, creatorID, mainID, shortMediaID, now, reviewReason, postReportState, approvedForPublishAt, publishedAt)
	tx := &stubTx{}

	var createMainArg sqlc.CreateMainParams
	var updateMainArg sqlc.UpdateMainStateParams
	var createShortArg sqlc.CreateShortParams
	var updateShortArg sqlc.UpdateShortStateParams
	var txShortArgs []sqlc.CreateShortParams

	queryStub := repositoryStubQueries{
		createMain: func(_ context.Context, arg sqlc.CreateMainParams) (sqlc.AppMain, error) {
			createMainArg = arg
			return mainRow, nil
		},
		getMain: func(_ context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
			if id != postgresUUID(mainID) {
				t.Fatalf("GetMainByID() id got %v want %v", id, postgresUUID(mainID))
			}
			return mainRow, nil
		},
		listMains: func(_ context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMain, error) {
			if creatorUserID != postgresUUID(creatorID) {
				t.Fatalf("ListMainsByCreatorUserID() creator got %v want %v", creatorUserID, postgresUUID(creatorID))
			}
			return []sqlc.AppMain{mainRow}, nil
		},
		updateMain: func(_ context.Context, arg sqlc.UpdateMainStateParams) (sqlc.AppMain, error) {
			updateMainArg = arg
			return mainRow, nil
		},
		getUnlockableMain: func(_ context.Context, id pgtype.UUID) (sqlc.AppUnlockableMain, error) {
			if id != postgresUUID(mainID) {
				t.Fatalf("GetUnlockableMainByID() id got %v want %v", id, postgresUUID(mainID))
			}
			return unlockableRow, nil
		},
		createShort: func(_ context.Context, arg sqlc.CreateShortParams) (sqlc.AppShort, error) {
			createShortArg = arg
			return shortRow, nil
		},
		getShort: func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
			if id != postgresUUID(shortID) {
				t.Fatalf("GetShortByID() id got %v want %v", id, postgresUUID(shortID))
			}
			return shortRow, nil
		},
		listShorts: func(_ context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppShort, error) {
			if creatorUserID != postgresUUID(creatorID) {
				t.Fatalf("ListShortsByCreatorUserID() creator got %v want %v", creatorUserID, postgresUUID(creatorID))
			}
			return []sqlc.AppShort{shortRow}, nil
		},
		updateShort: func(_ context.Context, arg sqlc.UpdateShortStateParams) (sqlc.AppShort, error) {
			updateShortArg = arg
			return shortRow, nil
		},
		publishShort: func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
			if id != postgresUUID(shortID) {
				t.Fatalf("PublishShort() id got %v want %v", id, postgresUUID(shortID))
			}
			return shortRow, nil
		},
		listPublicShorts: func(_ context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppPublicShort, error) {
			if creatorUserID != postgresUUID(creatorID) {
				t.Fatalf("ListPublicShortsByCreatorUserID() creator got %v want %v", creatorUserID, postgresUUID(creatorID))
			}
			return []sqlc.AppPublicShort{publicShortRow}, nil
		},
		getPublicShort: func(_ context.Context, id pgtype.UUID) (sqlc.AppPublicShort, error) {
			if id != postgresUUID(shortID) {
				t.Fatalf("GetPublicShortByID() id got %v want %v", id, postgresUUID(shortID))
			}
			return publicShortRow, nil
		},
		listCanonicalShorts: func(_ context.Context, canonicalMainID pgtype.UUID) ([]sqlc.AppShort, error) {
			if canonicalMainID != postgresUUID(mainID) {
				t.Fatalf("ListShortsByCanonicalMainID() id got %v want %v", canonicalMainID, postgresUUID(mainID))
			}
			return []sqlc.AppShort{shortRow}, nil
		},
		getCanonicalMainID: func(_ context.Context, id pgtype.UUID) (pgtype.UUID, error) {
			if id != postgresUUID(shortID) {
				t.Fatalf("GetCanonicalMainIDByShortID() id got %v want %v", id, postgresUUID(shortID))
			}
			return postgresUUID(mainID), nil
		},
	}

	txQueries := repositoryStubQueries{
		createMain: func(_ context.Context, arg sqlc.CreateMainParams) (sqlc.AppMain, error) {
			if arg.CreatorUserID != postgresUUID(creatorID) {
				t.Fatalf("tx CreateMain() creator got %v want %v", arg.CreatorUserID, postgresUUID(creatorID))
			}
			return mainRow, nil
		},
		createShort: func(_ context.Context, arg sqlc.CreateShortParams) (sqlc.AppShort, error) {
			txShortArgs = append(txShortArgs, arg)
			return shortRow, nil
		},
	}

	repo := newRepository(stubBeginner{tx: tx}, queryStub, func(db sqlc.DBTX) queries {
		if db != tx {
			t.Fatalf("newQueries() db got %v want %v", db, tx)
		}
		return txQueries
	})

	mainInput := CreateMainInput{
		CreatorUserID:       creatorID,
		MediaAssetID:        mainMediaID,
		State:               "draft",
		ReviewReasonCode:    reviewReason,
		PostReportState:     postReportState,
		PriceMinor:          priceMinor,
		CurrencyCode:        currencyCode,
		OwnershipConfirmed:  true,
		ConsentConfirmed:    true,
		ApprovedForUnlockAt: approvedForUnlockAt,
	}
	shortInput := CreateShortInput{
		CreatorUserID:        creatorID,
		CanonicalMainID:      mainID,
		MediaAssetID:         shortMediaID,
		State:                "reviewed",
		ReviewReasonCode:     reviewReason,
		PostReportState:      postReportState,
		ApprovedForPublishAt: approvedForPublishAt,
		PublishedAt:          publishedAt,
	}

	createdMain, err := repo.CreateMain(context.Background(), mainInput)
	if err != nil {
		t.Fatalf("CreateMain() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(createdMain, wantMain(mainID, creatorID, mainMediaID, now, reviewReason, postReportState, priceMinor, currencyCode, approvedForUnlockAt)) {
		t.Fatalf("CreateMain() got %#v want %#v", createdMain, wantMain(mainID, creatorID, mainMediaID, now, reviewReason, postReportState, priceMinor, currencyCode, approvedForUnlockAt))
	}
	if createMainArg.CreatorUserID != postgresUUID(creatorID) {
		t.Fatalf("CreateMain() creator arg got %v want %v", createMainArg.CreatorUserID, postgresUUID(creatorID))
	}

	gotMain, err := repo.GetMain(context.Background(), mainID)
	if err != nil {
		t.Fatalf("GetMain() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(gotMain, createdMain) {
		t.Fatalf("GetMain() got %#v want %#v", gotMain, createdMain)
	}

	gotUnlockableMain, err := repo.GetUnlockableMain(context.Background(), mainID)
	if err != nil {
		t.Fatalf("GetUnlockableMain() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(gotUnlockableMain, createdMain) {
		t.Fatalf("GetUnlockableMain() got %#v want %#v", gotUnlockableMain, createdMain)
	}

	listedMains, err := repo.ListMainsByCreator(context.Background(), creatorID)
	if err != nil {
		t.Fatalf("ListMainsByCreator() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(listedMains, []Main{createdMain}) {
		t.Fatalf("ListMainsByCreator() got %#v want %#v", listedMains, []Main{createdMain})
	}

	updatedMain, err := repo.UpdateMain(context.Background(), UpdateMainInput{
		ID:                  mainID,
		State:               "reviewed",
		ReviewReasonCode:    reviewReason,
		PostReportState:     postReportState,
		PriceMinor:          priceMinor,
		CurrencyCode:        currencyCode,
		OwnershipConfirmed:  true,
		ConsentConfirmed:    true,
		ApprovedForUnlockAt: approvedForUnlockAt,
	})
	if err != nil {
		t.Fatalf("UpdateMain() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(updatedMain, createdMain) {
		t.Fatalf("UpdateMain() got %#v want %#v", updatedMain, createdMain)
	}
	if updateMainArg.ID != postgresUUID(mainID) {
		t.Fatalf("UpdateMain() id arg got %v want %v", updateMainArg.ID, postgresUUID(mainID))
	}

	createdShort, err := repo.CreateShort(context.Background(), shortInput)
	if err != nil {
		t.Fatalf("CreateShort() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(createdShort, wantShort(shortID, creatorID, mainID, shortMediaID, now, reviewReason, postReportState, approvedForPublishAt, publishedAt)) {
		t.Fatalf("CreateShort() got %#v want %#v", createdShort, wantShort(shortID, creatorID, mainID, shortMediaID, now, reviewReason, postReportState, approvedForPublishAt, publishedAt))
	}
	if createShortArg.CanonicalMainID != postgresUUID(mainID) {
		t.Fatalf("CreateShort() canonical main arg got %v want %v", createShortArg.CanonicalMainID, postgresUUID(mainID))
	}

	gotShort, err := repo.GetShort(context.Background(), shortID)
	if err != nil {
		t.Fatalf("GetShort() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(gotShort, createdShort) {
		t.Fatalf("GetShort() got %#v want %#v", gotShort, createdShort)
	}

	gotPublicShort, err := repo.GetPublicShort(context.Background(), shortID)
	if err != nil {
		t.Fatalf("GetPublicShort() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(gotPublicShort, createdShort) {
		t.Fatalf("GetPublicShort() got %#v want %#v", gotPublicShort, createdShort)
	}

	listedShorts, err := repo.ListShortsByCreator(context.Background(), creatorID)
	if err != nil {
		t.Fatalf("ListShortsByCreator() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(listedShorts, []Short{createdShort}) {
		t.Fatalf("ListShortsByCreator() got %#v want %#v", listedShorts, []Short{createdShort})
	}

	listedPublicShorts, err := repo.ListPublicShortsByCreator(context.Background(), creatorID)
	if err != nil {
		t.Fatalf("ListPublicShortsByCreator() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(listedPublicShorts, []Short{createdShort}) {
		t.Fatalf("ListPublicShortsByCreator() got %#v want %#v", listedPublicShorts, []Short{createdShort})
	}

	listedCanonicalShorts, err := repo.ListShortsByCanonicalMain(context.Background(), mainID)
	if err != nil {
		t.Fatalf("ListShortsByCanonicalMain() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(listedCanonicalShorts, []Short{createdShort}) {
		t.Fatalf("ListShortsByCanonicalMain() got %#v want %#v", listedCanonicalShorts, []Short{createdShort})
	}

	canonicalMainID, err := repo.GetCanonicalMainIDByShort(context.Background(), shortID)
	if err != nil {
		t.Fatalf("GetCanonicalMainIDByShort() error = %v, want nil", err)
	}
	if canonicalMainID != mainID {
		t.Fatalf("GetCanonicalMainIDByShort() got %s want %s", canonicalMainID, mainID)
	}

	updatedShort, err := repo.UpdateShort(context.Background(), UpdateShortInput{
		ID:                   shortID,
		State:                "published",
		ReviewReasonCode:     reviewReason,
		PostReportState:      postReportState,
		ApprovedForPublishAt: approvedForPublishAt,
		PublishedAt:          publishedAt,
	})
	if err != nil {
		t.Fatalf("UpdateShort() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(updatedShort, createdShort) {
		t.Fatalf("UpdateShort() got %#v want %#v", updatedShort, createdShort)
	}
	if updateShortArg.ID != postgresUUID(shortID) {
		t.Fatalf("UpdateShort() id arg got %v want %v", updateShortArg.ID, postgresUUID(shortID))
	}

	publishedShort, err := repo.PublishShort(context.Background(), shortID)
	if err != nil {
		t.Fatalf("PublishShort() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(publishedShort, createdShort) {
		t.Fatalf("PublishShort() got %#v want %#v", publishedShort, createdShort)
	}

	mainWithShorts, err := repo.CreateMainWithShorts(context.Background(), CreateMainWithShortsInput{
		Main: mainInput,
		Shorts: []CreateLinkedShortInput{
			{
				MediaAssetID:         shortMediaID,
				State:                "published",
				ReviewReasonCode:     reviewReason,
				PostReportState:      postReportState,
				ApprovedForPublishAt: approvedForPublishAt,
				PublishedAt:          publishedAt,
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateMainWithShorts() error = %v, want nil", err)
	}
	wantMainWithShorts := MainWithShorts{
		Main:   createdMain,
		Shorts: []Short{createdShort},
	}
	if !reflect.DeepEqual(mainWithShorts, wantMainWithShorts) {
		t.Fatalf("CreateMainWithShorts() got %#v want %#v", mainWithShorts, wantMainWithShorts)
	}
	if !tx.committed {
		t.Fatal("CreateMainWithShorts() committed = false, want true")
	}
	if tx.rolledBack {
		t.Fatal("CreateMainWithShorts() rolledBack = true, want false")
	}
	if len(txShortArgs) != 1 {
		t.Fatalf("CreateMainWithShorts() short arg count got %d want 1", len(txShortArgs))
	}
	if txShortArgs[0].CanonicalMainID != mainRow.ID {
		t.Fatalf("CreateMainWithShorts() canonical main arg got %v want %v", txShortArgs[0].CanonicalMainID, mainRow.ID)
	}
}

func TestRepositoryNotFoundPaths(t *testing.T) {
	t.Parallel()

	mainID := uuid.New()
	shortID := uuid.New()

	tests := []struct {
		name string
		run  func(*Repository) error
		want error
	}{
		{
			name: "get main",
			run: func(repo *Repository) error {
				_, err := repo.GetMain(context.Background(), mainID)
				return err
			},
			want: ErrMainNotFound,
		},
		{
			name: "get unlockable main",
			run: func(repo *Repository) error {
				_, err := repo.GetUnlockableMain(context.Background(), mainID)
				return err
			},
			want: ErrUnlockableMainNotFound,
		},
		{
			name: "update main",
			run: func(repo *Repository) error {
				_, err := repo.UpdateMain(context.Background(), UpdateMainInput{ID: mainID})
				return err
			},
			want: ErrMainNotFound,
		},
		{
			name: "get short",
			run: func(repo *Repository) error {
				_, err := repo.GetShort(context.Background(), shortID)
				return err
			},
			want: ErrShortNotFound,
		},
		{
			name: "get public short",
			run: func(repo *Repository) error {
				_, err := repo.GetPublicShort(context.Background(), shortID)
				return err
			},
			want: ErrShortNotFound,
		},
		{
			name: "get canonical main id",
			run: func(repo *Repository) error {
				_, err := repo.GetCanonicalMainIDByShort(context.Background(), shortID)
				return err
			},
			want: ErrShortNotFound,
		},
		{
			name: "update short",
			run: func(repo *Repository) error {
				_, err := repo.UpdateShort(context.Background(), UpdateShortInput{ID: shortID})
				return err
			},
			want: ErrShortNotFound,
		},
		{
			name: "publish short",
			run: func(repo *Repository) error {
				_, err := repo.PublishShort(context.Background(), shortID)
				return err
			},
			want: ErrShortNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newRepository(nil, repositoryStubQueries{
				getMain: func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
					return sqlc.AppMain{}, pgx.ErrNoRows
				},
				getUnlockableMain: func(context.Context, pgtype.UUID) (sqlc.AppUnlockableMain, error) {
					return sqlc.AppUnlockableMain{}, pgx.ErrNoRows
				},
				updateMain: func(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error) {
					return sqlc.AppMain{}, pgx.ErrNoRows
				},
				getShort: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
					return sqlc.AppShort{}, pgx.ErrNoRows
				},
				getPublicShort: func(context.Context, pgtype.UUID) (sqlc.AppPublicShort, error) {
					return sqlc.AppPublicShort{}, pgx.ErrNoRows
				},
				getCanonicalMainID: func(context.Context, pgtype.UUID) (pgtype.UUID, error) {
					return pgtype.UUID{}, pgx.ErrNoRows
				},
				updateShort: func(context.Context, sqlc.UpdateShortStateParams) (sqlc.AppShort, error) {
					return sqlc.AppShort{}, pgx.ErrNoRows
				},
				publishShort: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
					return sqlc.AppShort{}, pgx.ErrNoRows
				},
			}, nil)

			if err := tt.run(repo); !errors.Is(err, tt.want) {
				t.Fatalf("%s error got %v want %v", tt.name, err, tt.want)
			}
		})
	}
}

func TestRepositoryConversionErrors(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.New()
	mainID := uuid.New()
	mainMediaID := uuid.New()
	shortID := uuid.New()
	shortMediaID := uuid.New()

	invalidMain := testMainRow(mainID, creatorID, mainMediaID, now, nil, nil, 1200, "JPY", nil)
	invalidMain.ID = pgtype.UUID{}
	invalidUnlockableMain := testUnlockableMainRow(mainID, creatorID, mainMediaID, now, nil, nil, 1200, "JPY", nil)
	invalidUnlockableMain.ID = pgtype.UUID{}
	invalidShort := testShortRow(shortID, creatorID, mainID, shortMediaID, now, nil, nil, nil, nil)
	invalidShort.ID = pgtype.UUID{}
	invalidPublicShort := testPublicShortRow(shortID, creatorID, mainID, shortMediaID, now, nil, nil, nil, nil)
	invalidPublicShort.ID = pgtype.UUID{}

	repo := newRepository(nil, repositoryStubQueries{
		createMain: func(context.Context, sqlc.CreateMainParams) (sqlc.AppMain, error) {
			return invalidMain, nil
		},
		getUnlockableMain: func(context.Context, pgtype.UUID) (sqlc.AppUnlockableMain, error) {
			return invalidUnlockableMain, nil
		},
		listMains: func(context.Context, pgtype.UUID) ([]sqlc.AppMain, error) {
			return []sqlc.AppMain{invalidMain}, nil
		},
		createShort: func(context.Context, sqlc.CreateShortParams) (sqlc.AppShort, error) {
			return invalidShort, nil
		},
		getPublicShort: func(context.Context, pgtype.UUID) (sqlc.AppPublicShort, error) {
			return invalidPublicShort, nil
		},
		listShorts: func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) {
			return []sqlc.AppShort{invalidShort}, nil
		},
		listPublicShorts: func(context.Context, pgtype.UUID) ([]sqlc.AppPublicShort, error) {
			return []sqlc.AppPublicShort{invalidPublicShort}, nil
		},
		listCanonicalShorts: func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) {
			return []sqlc.AppShort{invalidShort}, nil
		},
		getCanonicalMainID: func(context.Context, pgtype.UUID) (pgtype.UUID, error) {
			return pgtype.UUID{}, nil
		},
	}, nil)

	if _, err := repo.CreateMain(context.Background(), CreateMainInput{}); err == nil {
		t.Fatal("CreateMain() error = nil, want conversion error")
	}
	if _, err := repo.GetUnlockableMain(context.Background(), mainID); err == nil {
		t.Fatal("GetUnlockableMain() error = nil, want conversion error")
	}
	if _, err := repo.ListMainsByCreator(context.Background(), creatorID); err == nil {
		t.Fatal("ListMainsByCreator() error = nil, want conversion error")
	}
	if _, err := repo.CreateShort(context.Background(), CreateShortInput{}); err == nil {
		t.Fatal("CreateShort() error = nil, want conversion error")
	}
	if _, err := repo.GetPublicShort(context.Background(), shortID); err == nil {
		t.Fatal("GetPublicShort() error = nil, want conversion error")
	}
	if _, err := repo.ListShortsByCreator(context.Background(), creatorID); err == nil {
		t.Fatal("ListShortsByCreator() error = nil, want conversion error")
	}
	if _, err := repo.ListPublicShortsByCreator(context.Background(), creatorID); err == nil {
		t.Fatal("ListPublicShortsByCreator() error = nil, want conversion error")
	}
	if _, err := repo.ListShortsByCanonicalMain(context.Background(), mainID); err == nil {
		t.Fatal("ListShortsByCanonicalMain() error = nil, want conversion error")
	}
	if _, err := repo.GetCanonicalMainIDByShort(context.Background(), shortID); err == nil {
		t.Fatal("GetCanonicalMainIDByShort() error = nil, want conversion error")
	}
}

func TestMapFunctionsRejectInvalidRows(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.New()
	mainID := uuid.New()
	mainMediaID := uuid.New()
	shortID := uuid.New()
	shortMediaID := uuid.New()

	invalidMain := testMainRow(mainID, creatorID, mainMediaID, now, nil, nil, 1200, "JPY", nil)
	invalidMain.CreatedAt = pgtype.Timestamptz{}
	if _, err := mapMain(invalidMain); err == nil {
		t.Fatal("mapMain() error = nil, want conversion error")
	}

	invalidUnlockableMain := testUnlockableMainRow(mainID, creatorID, mainMediaID, now, nil, nil, 1200, "JPY", nil)
	invalidUnlockableMain.CreatedAt = pgtype.Timestamptz{}
	if _, err := mapUnlockableMain(invalidUnlockableMain); err == nil {
		t.Fatal("mapUnlockableMain() error = nil, want conversion error")
	}

	invalidShort := testShortRow(shortID, creatorID, mainID, shortMediaID, now, nil, nil, nil, nil)
	invalidShort.CreatedAt = pgtype.Timestamptz{}
	if _, err := mapShort(invalidShort); err == nil {
		t.Fatal("mapShort() error = nil, want conversion error")
	}

	invalidPublicShort := testPublicShortRow(shortID, creatorID, mainID, shortMediaID, now, nil, nil, nil, nil)
	invalidPublicShort.CreatedAt = pgtype.Timestamptz{}
	if _, err := mapPublicShort(invalidPublicShort); err == nil {
		t.Fatal("mapPublicShort() error = nil, want conversion error")
	}
}

func testMainRow(id uuid.UUID, creatorID uuid.UUID, mediaAssetID uuid.UUID, now time.Time, reviewReason *string, postReportState *string, priceMinor int64, currencyCode string, approvedForUnlockAt *time.Time) sqlc.AppMain {
	return sqlc.AppMain{
		ID:                  postgresUUID(id),
		CreatorUserID:       postgresUUID(creatorID),
		MediaAssetID:        postgresUUID(mediaAssetID),
		State:               "draft",
		ReviewReasonCode:    textValue(reviewReason),
		PostReportState:     textValue(postReportState),
		PriceMinor:          priceMinor,
		CurrencyCode:        currencyCode,
		OwnershipConfirmed:  true,
		ConsentConfirmed:    true,
		ApprovedForUnlockAt: optionalTimestamp(approvedForUnlockAt),
		CreatedAt:           timestamp(now),
		UpdatedAt:           timestamp(now.Add(time.Minute)),
	}
}

func testUnlockableMainRow(id uuid.UUID, creatorID uuid.UUID, mediaAssetID uuid.UUID, now time.Time, reviewReason *string, postReportState *string, priceMinor int64, currencyCode string, approvedForUnlockAt *time.Time) sqlc.AppUnlockableMain {
	return sqlc.AppUnlockableMain{
		ID:                  postgresUUID(id),
		CreatorUserID:       postgresUUID(creatorID),
		MediaAssetID:        postgresUUID(mediaAssetID),
		State:               "draft",
		ReviewReasonCode:    textValue(reviewReason),
		PostReportState:     textValue(postReportState),
		PriceMinor:          pgtype.Int8{Int64: priceMinor, Valid: true},
		CurrencyCode:        pgtype.Text{String: currencyCode, Valid: true},
		OwnershipConfirmed:  true,
		ConsentConfirmed:    true,
		ApprovedForUnlockAt: optionalTimestamp(approvedForUnlockAt),
		CreatedAt:           timestamp(now),
		UpdatedAt:           timestamp(now.Add(time.Minute)),
	}
}

func testShortRow(id uuid.UUID, creatorID uuid.UUID, canonicalMainID uuid.UUID, mediaAssetID uuid.UUID, now time.Time, reviewReason *string, postReportState *string, approvedForPublishAt *time.Time, publishedAt *time.Time) sqlc.AppShort {
	return sqlc.AppShort{
		ID:                   postgresUUID(id),
		CreatorUserID:        postgresUUID(creatorID),
		CanonicalMainID:      postgresUUID(canonicalMainID),
		MediaAssetID:         postgresUUID(mediaAssetID),
		State:                "reviewed",
		ReviewReasonCode:     textValue(reviewReason),
		PostReportState:      textValue(postReportState),
		ApprovedForPublishAt: optionalTimestamp(approvedForPublishAt),
		PublishedAt:          optionalTimestamp(publishedAt),
		CreatedAt:            timestamp(now),
		UpdatedAt:            timestamp(now.Add(time.Minute)),
	}
}

func testPublicShortRow(id uuid.UUID, creatorID uuid.UUID, canonicalMainID uuid.UUID, mediaAssetID uuid.UUID, now time.Time, reviewReason *string, postReportState *string, approvedForPublishAt *time.Time, publishedAt *time.Time) sqlc.AppPublicShort {
	return sqlc.AppPublicShort{
		ID:                   postgresUUID(id),
		CreatorUserID:        postgresUUID(creatorID),
		CanonicalMainID:      postgresUUID(canonicalMainID),
		MediaAssetID:         postgresUUID(mediaAssetID),
		State:                "reviewed",
		ReviewReasonCode:     textValue(reviewReason),
		PostReportState:      textValue(postReportState),
		ApprovedForPublishAt: optionalTimestamp(approvedForPublishAt),
		PublishedAt:          optionalTimestamp(publishedAt),
		CreatedAt:            timestamp(now),
		UpdatedAt:            timestamp(now.Add(time.Minute)),
	}
}

func wantMain(id uuid.UUID, creatorID uuid.UUID, mediaAssetID uuid.UUID, now time.Time, reviewReason *string, postReportState *string, priceMinor int64, currencyCode string, approvedForUnlockAt *time.Time) Main {
	return Main{
		ID:                  id,
		CreatorUserID:       creatorID,
		MediaAssetID:        mediaAssetID,
		State:               "draft",
		ReviewReasonCode:    reviewReason,
		PostReportState:     postReportState,
		PriceMinor:          priceMinor,
		CurrencyCode:        currencyCode,
		OwnershipConfirmed:  true,
		ConsentConfirmed:    true,
		ApprovedForUnlockAt: approvedForUnlockAt,
		CreatedAt:           now,
		UpdatedAt:           now.Add(time.Minute),
	}
}

func wantShort(id uuid.UUID, creatorID uuid.UUID, canonicalMainID uuid.UUID, mediaAssetID uuid.UUID, now time.Time, reviewReason *string, postReportState *string, approvedForPublishAt *time.Time, publishedAt *time.Time) Short {
	return Short{
		ID:                   id,
		CreatorUserID:        creatorID,
		CanonicalMainID:      canonicalMainID,
		MediaAssetID:         mediaAssetID,
		State:                "reviewed",
		ReviewReasonCode:     reviewReason,
		PostReportState:      postReportState,
		ApprovedForPublishAt: approvedForPublishAt,
		PublishedAt:          publishedAt,
		CreatedAt:            now,
		UpdatedAt:            now.Add(time.Minute),
	}
}

func textValue(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func optionalTimestamp(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return timestamp(*value)
}

func stringPtr(value string) *string {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
