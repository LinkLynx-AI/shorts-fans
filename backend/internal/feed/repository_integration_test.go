package feed

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/devseed"
	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

const feedIntegrationPostgresDSNEnv = "POSTGRES_DSN"

func TestRepositoryListFollowingRanksPersonalizedCandidates(t *testing.T) {
	t.Parallel()

	ctx, pool, cleanup := newFeedTestDatabase(t)
	defer cleanup()

	summary, err := devseed.Run(ctx, pool)
	if err != nil {
		t.Fatalf("devseed.Run() error = %v, want nil", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	creatorB := creatorScenario{
		userID:      uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		mainID:      uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		mainAssetID: uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
		shorts: []scenarioShort{
			{
				id:          uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
				assetID:     uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"),
				caption:     "creator b first short",
				publishedAt: now.Add(-1 * time.Hour),
			},
			{
				id:          uuid.MustParse("f0f0f0f0-f0f0-f0f0-f0f0-f0f0f0f0f0f0"),
				assetID:     uuid.MustParse("abababab-abab-abab-abab-abababababab"),
				caption:     "creator b second short",
				publishedAt: now.Add(-2 * time.Hour),
			},
		},
		displayName: "Creator B",
		handle:      "creatorb",
	}
	creatorC := creatorScenario{
		userID:      uuid.MustParse("cdcdcdcd-cdcd-cdcd-cdcd-cdcdcdcdcdcd"),
		mainID:      uuid.MustParse("edededed-eded-eded-eded-edededededed"),
		mainAssetID: uuid.MustParse("12121212-1212-1212-1212-121212121212"),
		shorts: []scenarioShort{
			{
				id:          uuid.MustParse("34343434-3434-3434-3434-343434343434"),
				assetID:     uuid.MustParse("56565656-5656-5656-5656-565656565656"),
				caption:     "creator c short",
				publishedAt: now.Add(-90 * time.Minute),
			},
		},
		displayName: "Creator C",
		handle:      "creatorc",
	}
	creatorD := creatorScenario{
		userID:      uuid.MustParse("78787878-7878-7878-7878-787878787878"),
		mainID:      uuid.MustParse("90909090-9090-9090-9090-909090909090"),
		mainAssetID: uuid.MustParse("13131313-1313-1313-1313-131313131313"),
		shorts: []scenarioShort{
			{
				id:          uuid.MustParse("24242424-2424-2424-2424-242424242424"),
				assetID:     uuid.MustParse("35353535-3535-3535-3535-353535353535"),
				caption:     "creator d not-followed short",
				publishedAt: now.Add(-30 * time.Minute),
			},
		},
		displayName: "Creator D",
		handle:      "creatord",
	}

	if err := insertCreatorScenario(ctx, pool, creatorB, true, summary.FanUserID); err != nil {
		t.Fatalf("insertCreatorScenario(creatorB) error = %v, want nil", err)
	}
	if err := insertCreatorScenario(ctx, pool, creatorC, true, summary.FanUserID); err != nil {
		t.Fatalf("insertCreatorScenario(creatorC) error = %v, want nil", err)
	}
	if err := insertCreatorScenario(ctx, pool, creatorD, false, summary.FanUserID); err != nil {
		t.Fatalf("insertCreatorScenario(creatorD) error = %v, want nil", err)
	}

	if err := upsertViewerCreatorFeatures(ctx, pool, summary.FanUserID, creatorB.userID, 1, 1); err != nil {
		t.Fatalf("upsertViewerCreatorFeatures(creatorB) error = %v, want nil", err)
	}
	if err := upsertViewerMainFeatures(ctx, pool, summary.FanUserID, creatorB.mainID, creatorB.userID, 1); err != nil {
		t.Fatalf("upsertViewerMainFeatures(creatorB) error = %v, want nil", err)
	}
	for _, short := range creatorB.shorts {
		if err := upsertShortGlobalFeatures(ctx, pool, short.id, creatorB.userID, creatorB.mainID, 2, 5); err != nil {
			t.Fatalf("upsertShortGlobalFeatures(creatorB short=%s) error = %v, want nil", short.id, err)
		}
	}
	if err := upsertViewerCreatorFeatures(ctx, pool, summary.FanUserID, creatorC.userID, 1, 0); err != nil {
		t.Fatalf("upsertViewerCreatorFeatures(creatorC) error = %v, want nil", err)
	}
	if err := upsertViewerMainFeatures(ctx, pool, summary.FanUserID, creatorC.mainID, creatorC.userID, 1); err != nil {
		t.Fatalf("upsertViewerMainFeatures(creatorC) error = %v, want nil", err)
	}
	if err := upsertShortGlobalFeatures(ctx, pool, creatorC.shorts[0].id, creatorC.userID, creatorC.mainID, 1, 0); err != nil {
		t.Fatalf("upsertShortGlobalFeatures(creatorC) error = %v, want nil", err)
	}
	if err := upsertViewerCreatorFeatures(ctx, pool, summary.FanUserID, creatorD.userID, 3, 3); err != nil {
		t.Fatalf("upsertViewerCreatorFeatures(creatorD) error = %v, want nil", err)
	}
	if err := upsertViewerMainFeatures(ctx, pool, summary.FanUserID, creatorD.mainID, creatorD.userID, 3); err != nil {
		t.Fatalf("upsertViewerMainFeatures(creatorD) error = %v, want nil", err)
	}
	if err := upsertShortGlobalFeatures(ctx, pool, creatorD.shorts[0].id, creatorD.userID, creatorD.mainID, 5, 5); err != nil {
		t.Fatalf("upsertShortGlobalFeatures(creatorD) error = %v, want nil", err)
	}

	repo := NewRepository(pool)

	items, nextCursor, err := repo.ListFollowing(ctx, summary.FanUserID, nil, 3)
	if err != nil {
		t.Fatalf("ListFollowing() error = %v, want nil", err)
	}
	if len(items) != 3 {
		t.Fatalf("ListFollowing() item len got %d want %d", len(items), 3)
	}
	wantFirstPage := []uuid.UUID{
		creatorB.shorts[0].id,
		creatorC.shorts[0].id,
		creatorB.shorts[1].id,
	}
	for index, wantShortID := range wantFirstPage {
		if items[index].Short.ID != wantShortID {
			t.Fatalf("ListFollowing() first page short[%d] got %s want %s", index, items[index].Short.ID, wantShortID)
		}
	}
	for _, item := range items {
		if item.Short.ID == creatorD.shorts[0].id {
			t.Fatalf("ListFollowing() returned not-followed short %s", item.Short.ID)
		}
	}
	if nextCursor == nil || len(nextCursor.FollowingRemainingShortIDs) != 2 {
		t.Fatalf("ListFollowing() next cursor got %#v want snapshot cursor with 2 remaining ids", nextCursor)
	}
	latePublishedAt := time.Now().UTC().Add(1 * time.Minute).Truncate(time.Second)
	lateCreator := creatorScenario{
		userID:      uuid.MustParse("45454545-4545-4545-4545-454545454545"),
		mainID:      uuid.MustParse("46464646-4646-4646-4646-464646464646"),
		mainAssetID: uuid.MustParse("47474747-4747-4747-4747-474747474747"),
		shorts: []scenarioShort{
			{
				id:          uuid.MustParse("48484848-4848-4848-4848-484848484848"),
				assetID:     uuid.MustParse("49494949-4949-4949-4949-494949494949"),
				caption:     "late followed short",
				publishedAt: latePublishedAt,
			},
		},
		displayName: "Late Creator",
		handle:      "latecreator",
	}
	if err := insertCreatorScenario(ctx, pool, lateCreator, true, summary.FanUserID); err != nil {
		t.Fatalf("insertCreatorScenario(lateCreator) error = %v, want nil", err)
	}
	if err := upsertViewerCreatorFeatures(ctx, pool, summary.FanUserID, lateCreator.userID, 2, 2); err != nil {
		t.Fatalf("upsertViewerCreatorFeatures(lateCreator) error = %v, want nil", err)
	}
	if err := upsertViewerMainFeatures(ctx, pool, summary.FanUserID, lateCreator.mainID, lateCreator.userID, 2); err != nil {
		t.Fatalf("upsertViewerMainFeatures(lateCreator) error = %v, want nil", err)
	}
	if err := upsertShortGlobalFeatures(ctx, pool, lateCreator.shorts[0].id, lateCreator.userID, lateCreator.mainID, 3, 3); err != nil {
		t.Fatalf("upsertShortGlobalFeatures(lateCreator) error = %v, want nil", err)
	}

	nextPage, finalCursor, err := repo.ListFollowing(ctx, summary.FanUserID, nextCursor, 3)
	if err != nil {
		t.Fatalf("ListFollowing(next page) error = %v, want nil", err)
	}
	if len(nextPage) != 2 {
		t.Fatalf("ListFollowing(next page) item len got %d want %d", len(nextPage), 2)
	}
	if finalCursor != nil {
		t.Fatalf("ListFollowing(next page) cursor got %#v want nil", finalCursor)
	}
	if nextPage[0].Short.ID != summary.ShortIDs[1] || nextPage[1].Short.ID != summary.ShortIDs[0] {
		t.Fatalf(
			"ListFollowing(next page) got [%s %s] want [%s %s]",
			nextPage[0].Short.ID,
			nextPage[1].Short.ID,
			summary.ShortIDs[1],
			summary.ShortIDs[0],
		)
	}
	for _, item := range nextPage {
		if item.Short.ID == lateCreator.shorts[0].id {
			t.Fatalf("ListFollowing(next page) returned short published after ranking reference %s", item.Short.ID)
		}
	}
}

func TestRepositoryListFollowingIncludesRecentFollowedShortOnFirstPage(t *testing.T) {
	t.Parallel()

	ctx, pool, cleanup := newFeedTestDatabase(t)
	defer cleanup()

	viewerID := uuid.MustParse("10101010-1010-1010-1010-101010101010")
	if err := insertFeedViewer(ctx, pool, viewerID); err != nil {
		t.Fatalf("insertFeedViewer() error = %v, want nil", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	recentCreator := creatorScenario{
		userID:      uuid.MustParse("11111111-2222-3333-4444-555555555555"),
		mainID:      uuid.MustParse("66666666-7777-8888-9999-aaaaaaaaaaaa"),
		mainAssetID: uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff"),
		shorts: []scenarioShort{
			{
				id:          uuid.MustParse("12121212-3434-5656-7878-909090909090"),
				assetID:     uuid.MustParse("13131313-3535-5757-7979-919191919191"),
				caption:     "recent followed short",
				publishedAt: now.Add(-2 * time.Minute),
			},
		},
		displayName: "Recent Creator",
		handle:      "recentcreator",
	}
	if err := insertCreatorScenario(ctx, pool, recentCreator, true, viewerID); err != nil {
		t.Fatalf("insertCreatorScenario(recentCreator) error = %v, want nil", err)
	}

	repo := NewRepository(pool)
	items, _, err := repo.ListFollowing(ctx, viewerID, nil, 10)
	if err != nil {
		t.Fatalf("ListFollowing() error = %v, want nil", err)
	}

	foundRecent := false
	for _, item := range items {
		if item.Short.ID == recentCreator.shorts[0].id {
			foundRecent = true
			break
		}
	}
	if !foundRecent {
		t.Fatalf("ListFollowing() missing recent followed short %s on first page", recentCreator.shorts[0].id)
	}
}

func TestRepositoryListFollowingUsesShortIDAsFinalTieBreaker(t *testing.T) {
	t.Parallel()

	ctx, pool, cleanup := newFeedTestDatabase(t)
	defer cleanup()

	viewerID := uuid.MustParse("20202020-2020-2020-2020-202020202020")
	if err := insertFeedViewer(ctx, pool, viewerID); err != nil {
		t.Fatalf("insertFeedViewer() error = %v, want nil", err)
	}

	publishedAt := time.Now().UTC().Truncate(time.Second).Add(-10 * time.Minute)
	creatorA := creatorScenario{
		userID:      uuid.MustParse("30303030-3030-3030-3030-303030303030"),
		mainID:      uuid.MustParse("31313131-3131-3131-3131-313131313131"),
		mainAssetID: uuid.MustParse("32323232-3232-3232-3232-323232323232"),
		shorts: []scenarioShort{
			{
				id:          uuid.MustParse("33333333-3333-3333-3333-333333333333"),
				assetID:     uuid.MustParse("34343434-3434-3434-3434-343434343434"),
				caption:     "tie short a",
				publishedAt: publishedAt,
			},
		},
		displayName: "Tie Creator A",
		handle:      "tiecreatora",
	}
	creatorB := creatorScenario{
		userID:      uuid.MustParse("40404040-4040-4040-4040-404040404040"),
		mainID:      uuid.MustParse("41414141-4141-4141-4141-414141414141"),
		mainAssetID: uuid.MustParse("42424242-4242-4242-4242-424242424242"),
		shorts: []scenarioShort{
			{
				id:          uuid.MustParse("43434343-4343-4343-4343-434343434343"),
				assetID:     uuid.MustParse("44444444-4444-4444-4444-444444444444"),
				caption:     "tie short b",
				publishedAt: publishedAt,
			},
		},
		displayName: "Tie Creator B",
		handle:      "tiecreatorb",
	}
	if err := insertCreatorScenario(ctx, pool, creatorA, true, viewerID); err != nil {
		t.Fatalf("insertCreatorScenario(creatorA) error = %v, want nil", err)
	}
	if err := insertCreatorScenario(ctx, pool, creatorB, true, viewerID); err != nil {
		t.Fatalf("insertCreatorScenario(creatorB) error = %v, want nil", err)
	}

	repo := NewRepository(pool)
	firstPage, nextCursor, err := repo.ListFollowing(ctx, viewerID, nil, 1)
	if err != nil {
		t.Fatalf("ListFollowing(first page) error = %v, want nil", err)
	}
	if len(firstPage) != 1 || nextCursor == nil {
		t.Fatalf("ListFollowing(first page) got items=%d cursor=%#v want 1 non-nil", len(firstPage), nextCursor)
	}

	secondPage, finalCursor, err := repo.ListFollowing(ctx, viewerID, nextCursor, 1)
	if err != nil {
		t.Fatalf("ListFollowing(second page) error = %v, want nil", err)
	}
	if len(secondPage) != 1 || finalCursor != nil {
		t.Fatalf("ListFollowing(second page) got items=%d cursor=%#v want 1 nil", len(secondPage), finalCursor)
	}

	gotIDs := []uuid.UUID{firstPage[0].Short.ID, secondPage[0].Short.ID}
	wantIDs := []uuid.UUID{creatorB.shorts[0].id, creatorA.shorts[0].id}
	for index, wantID := range wantIDs {
		if gotIDs[index] != wantID {
			t.Fatalf("ListFollowing() tied order[%d] got %s want %s", index, gotIDs[index], wantID)
		}
	}
}

func TestRepositoryListFollowingContinuationPreservesSnapshotAfterFeatureMutation(t *testing.T) {
	t.Parallel()

	ctx, pool, cleanup := newFeedTestDatabase(t)
	defer cleanup()

	viewerID := uuid.MustParse("50505050-5050-5050-5050-505050505050")
	if err := insertFeedViewer(ctx, pool, viewerID); err != nil {
		t.Fatalf("insertFeedViewer() error = %v, want nil", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	creatorA := creatorScenario{
		userID:      uuid.MustParse("51515151-5151-5151-5151-515151515151"),
		mainID:      uuid.MustParse("52525252-5252-5252-5252-525252525252"),
		mainAssetID: uuid.MustParse("53535353-5353-5353-5353-535353535353"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("54545454-5454-5454-5454-545454545454"),
			assetID:     uuid.MustParse("55555555-5555-5555-5555-555555555555"),
			caption:     "snapshot first page",
			publishedAt: now.Add(-10 * time.Minute),
		}},
		displayName: "Snapshot Creator A",
		handle:      "snapshota",
	}
	creatorB := creatorScenario{
		userID:      uuid.MustParse("56565656-5656-5656-5656-565656565656"),
		mainID:      uuid.MustParse("57575757-5757-5757-5757-575757575757"),
		mainAssetID: uuid.MustParse("58585858-5858-5858-5858-585858585858"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("59595959-5959-5959-5959-595959595959"),
			assetID:     uuid.MustParse("5a5a5a5a-5a5a-5a5a-5a5a-5a5a5a5a5a5a"),
			caption:     "snapshot second page",
			publishedAt: now.Add(-20 * time.Minute),
		}},
		displayName: "Snapshot Creator B",
		handle:      "snapshotb",
	}
	creatorC := creatorScenario{
		userID:      uuid.MustParse("5b5b5b5b-5b5b-5b5b-5b5b-5b5b5b5b5b5b"),
		mainID:      uuid.MustParse("5c5c5c5c-5c5c-5c5c-5c5c-5c5c5c5c5c5c"),
		mainAssetID: uuid.MustParse("5d5d5d5d-5d5d-5d5d-5d5d-5d5d5d5d5d5d"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("5e5e5e5e-5e5e-5e5e-5e5e-5e5e5e5e5e5e"),
			assetID:     uuid.MustParse("5f5f5f5f-5f5f-5f5f-5f5f-5f5f5f5f5f5f"),
			caption:     "snapshot third page",
			publishedAt: now.Add(-30 * time.Minute),
		}},
		displayName: "Snapshot Creator C",
		handle:      "snapshotc",
	}

	for _, creator := range []creatorScenario{creatorA, creatorB, creatorC} {
		if err := insertCreatorScenario(ctx, pool, creator, true, viewerID); err != nil {
			t.Fatalf("insertCreatorScenario(%s) error = %v, want nil", creator.handle, err)
		}
	}

	repo := NewRepository(pool)

	firstPage, nextCursor, err := repo.ListFollowing(ctx, viewerID, nil, 1)
	if err != nil {
		t.Fatalf("ListFollowing(first page) error = %v, want nil", err)
	}
	if len(firstPage) != 1 || nextCursor == nil {
		t.Fatalf("ListFollowing(first page) got items=%d cursor=%#v want 1 snapshot cursor", len(firstPage), nextCursor)
	}
	if firstPage[0].Short.ID != creatorA.shorts[0].id {
		t.Fatalf("ListFollowing(first page) got short %s want %s", firstPage[0].Short.ID, creatorA.shorts[0].id)
	}
	wantRemaining := []uuid.UUID{creatorB.shorts[0].id, creatorC.shorts[0].id}
	if len(nextCursor.FollowingRemainingShortIDs) != len(wantRemaining) {
		t.Fatalf(
			"ListFollowing(first page) remaining len got %d want %d",
			len(nextCursor.FollowingRemainingShortIDs),
			len(wantRemaining),
		)
	}
	for index, wantShortID := range wantRemaining {
		if nextCursor.FollowingRemainingShortIDs[index] != wantShortID {
			t.Fatalf(
				"ListFollowing(first page) remaining[%d] got %s want %s",
				index,
				nextCursor.FollowingRemainingShortIDs[index],
				wantShortID,
			)
		}
	}

	if err := upsertViewerCreatorFeatures(ctx, pool, viewerID, creatorB.userID, 1, 1); err != nil {
		t.Fatalf("upsertViewerCreatorFeatures(creatorB) error = %v, want nil", err)
	}
	if err := upsertViewerMainFeatures(ctx, pool, viewerID, creatorB.mainID, creatorB.userID, 1); err != nil {
		t.Fatalf("upsertViewerMainFeatures(creatorB) error = %v, want nil", err)
	}

	secondPage, finalCursor, err := repo.ListFollowing(ctx, viewerID, nextCursor, 1)
	if err != nil {
		t.Fatalf("ListFollowing(second page) error = %v, want nil", err)
	}
	if len(secondPage) != 1 || finalCursor == nil {
		t.Fatalf("ListFollowing(second page) got items=%d cursor=%#v want 1 snapshot cursor", len(secondPage), finalCursor)
	}
	if secondPage[0].Short.ID != creatorB.shorts[0].id {
		t.Fatalf(
			"ListFollowing(second page) got short %s want snapshot short %s",
			secondPage[0].Short.ID,
			creatorB.shorts[0].id,
		)
	}
	if len(finalCursor.FollowingRemainingShortIDs) != 1 || finalCursor.FollowingRemainingShortIDs[0] != creatorC.shorts[0].id {
		t.Fatalf(
			"ListFollowing(second page) remaining got %#v want [%s]",
			finalCursor.FollowingRemainingShortIDs,
			creatorC.shorts[0].id,
		)
	}
}

func TestRepositoryListFollowingContinuationPreservesSnapshotAfterFollowMutation(t *testing.T) {
	t.Parallel()

	ctx, pool, cleanup := newFeedTestDatabase(t)
	defer cleanup()

	viewerID := uuid.MustParse("60606060-6060-6060-6060-606060606060")
	if err := insertFeedViewer(ctx, pool, viewerID); err != nil {
		t.Fatalf("insertFeedViewer() error = %v, want nil", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	creatorA := creatorScenario{
		userID:      uuid.MustParse("61616161-6161-6161-6161-616161616161"),
		mainID:      uuid.MustParse("62626262-6262-6262-6262-626262626262"),
		mainAssetID: uuid.MustParse("63636363-6363-6363-6363-636363636363"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("64646464-6464-6464-6464-646464646464"),
			assetID:     uuid.MustParse("65656565-6565-6565-6565-656565656565"),
			caption:     "follow snapshot first page",
			publishedAt: now.Add(-10 * time.Minute),
		}},
		displayName: "Follow Snapshot A",
		handle:      "followsnapa",
	}
	creatorB := creatorScenario{
		userID:      uuid.MustParse("66666666-6666-6666-6666-666666666666"),
		mainID:      uuid.MustParse("67676767-6767-6767-6767-676767676767"),
		mainAssetID: uuid.MustParse("68686868-6868-6868-6868-686868686868"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("69696969-6969-6969-6969-696969696969"),
			assetID:     uuid.MustParse("6a6a6a6a-6a6a-6a6a-6a6a-6a6a6a6a6a6a"),
			caption:     "follow snapshot second page",
			publishedAt: now.Add(-20 * time.Minute),
		}},
		displayName: "Follow Snapshot B",
		handle:      "followsnapb",
	}
	creatorC := creatorScenario{
		userID:      uuid.MustParse("6b6b6b6b-6b6b-6b6b-6b6b-6b6b6b6b6b6b"),
		mainID:      uuid.MustParse("6c6c6c6c-6c6c-6c6c-6c6c-6c6c6c6c6c6c"),
		mainAssetID: uuid.MustParse("6d6d6d6d-6d6d-6d6d-6d6d-6d6d6d6d6d6d"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("6e6e6e6e-6e6e-6e6e-6e6e-6e6e6e6e6e6e"),
			assetID:     uuid.MustParse("6f6f6f6f-6f6f-6f6f-6f6f-6f6f6f6f6f6f"),
			caption:     "follow snapshot third page",
			publishedAt: now.Add(-30 * time.Minute),
		}},
		displayName: "Follow Snapshot C",
		handle:      "followsnapc",
	}
	creatorD := creatorScenario{
		userID:      uuid.MustParse("70707070-7070-7070-7070-707070707070"),
		mainID:      uuid.MustParse("71717171-7171-7171-7171-717171717171"),
		mainAssetID: uuid.MustParse("72727272-7272-7272-7272-727272727272"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("73737373-7373-7373-7373-737373737373"),
			assetID:     uuid.MustParse("74747474-7474-7474-7474-747474747474"),
			caption:     "follow mutation should not inject",
			publishedAt: now.Add(-15 * time.Minute),
		}},
		displayName: "Follow Snapshot D",
		handle:      "followsnapd",
	}

	for _, creator := range []creatorScenario{creatorA, creatorB, creatorC} {
		if err := insertCreatorScenario(ctx, pool, creator, true, viewerID); err != nil {
			t.Fatalf("insertCreatorScenario(%s) error = %v, want nil", creator.handle, err)
		}
	}
	if err := insertCreatorScenario(ctx, pool, creatorD, false, viewerID); err != nil {
		t.Fatalf("insertCreatorScenario(%s) error = %v, want nil", creatorD.handle, err)
	}

	repo := NewRepository(pool)

	firstPage, nextCursor, err := repo.ListFollowing(ctx, viewerID, nil, 1)
	if err != nil {
		t.Fatalf("ListFollowing(first page) error = %v, want nil", err)
	}
	if len(firstPage) != 1 || nextCursor == nil {
		t.Fatalf("ListFollowing(first page) got items=%d cursor=%#v want 1 snapshot cursor", len(firstPage), nextCursor)
	}
	if firstPage[0].Short.ID != creatorA.shorts[0].id {
		t.Fatalf("ListFollowing(first page) got short %s want %s", firstPage[0].Short.ID, creatorA.shorts[0].id)
	}

	if _, err := pool.Exec(
		ctx,
		`DELETE FROM app.creator_follows WHERE user_id = $1 AND creator_user_id = $2`,
		viewerID,
		creatorB.userID,
	); err != nil {
		t.Fatalf("DELETE creator_follows(creatorB) error = %v, want nil", err)
	}
	if _, err := pool.Exec(
		ctx,
		`INSERT INTO app.creator_follows (user_id, creator_user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		viewerID,
		creatorD.userID,
	); err != nil {
		t.Fatalf("INSERT creator_follows(creatorD) error = %v, want nil", err)
	}

	secondPage, finalCursor, err := repo.ListFollowing(ctx, viewerID, nextCursor, 2)
	if err != nil {
		t.Fatalf("ListFollowing(second page) error = %v, want nil", err)
	}
	if len(secondPage) != 2 || finalCursor != nil {
		t.Fatalf("ListFollowing(second page) got items=%d cursor=%#v want 2 nil", len(secondPage), finalCursor)
	}
	if secondPage[0].Short.ID != creatorB.shorts[0].id || secondPage[1].Short.ID != creatorC.shorts[0].id {
		t.Fatalf(
			"ListFollowing(second page) got [%s %s] want snapshot [%s %s]",
			secondPage[0].Short.ID,
			secondPage[1].Short.ID,
			creatorB.shorts[0].id,
			creatorC.shorts[0].id,
		)
	}
	if secondPage[0].Viewer.IsFollowingCreator {
		t.Fatalf("ListFollowing(second page) creatorB follow state got true want false after unfollow")
	}
	for _, item := range secondPage {
		if item.Short.ID == creatorD.shorts[0].id {
			t.Fatalf("ListFollowing(second page) injected newly followed short %s", item.Short.ID)
		}
	}
}

func TestRepositoryListRecommendedRanksPersonalizedCandidates(t *testing.T) {
	t.Parallel()

	ctx, pool, cleanup := newFeedTestDatabase(t)
	defer cleanup()

	viewerID := uuid.MustParse("80808080-8080-8080-8080-808080808080")
	if err := insertFeedViewer(ctx, pool, viewerID); err != nil {
		t.Fatalf("insertFeedViewer() error = %v, want nil", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	creatorA := creatorScenario{
		userID:      uuid.MustParse("81818181-8181-8181-8181-818181818181"),
		mainID:      uuid.MustParse("82828282-8282-8282-8282-828282828282"),
		mainAssetID: uuid.MustParse("83838383-8383-8383-8383-838383838383"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("84848484-8484-8484-8484-848484848484"),
			assetID:     uuid.MustParse("85858585-8585-8585-8585-858585858585"),
			caption:     "cold but recent short",
			publishedAt: now.Add(-10 * time.Minute),
		}},
		displayName: "Recommended A",
		handle:      "recommendeda",
	}
	creatorB := creatorScenario{
		userID:      uuid.MustParse("86868686-8686-8686-8686-868686868686"),
		mainID:      uuid.MustParse("87878787-8787-8787-8787-878787878787"),
		mainAssetID: uuid.MustParse("88888888-8888-8888-8888-888888888888"),
		shorts: []scenarioShort{
			{
				id:          uuid.MustParse("89898989-8989-8989-8989-898989898989"),
				assetID:     uuid.MustParse("8a8a8a8a-8a8a-8a8a-8a8a-8a8a8a8a8a8a"),
				caption:     "preferred creator short one",
				publishedAt: now.Add(-1 * time.Hour),
			},
			{
				id:          uuid.MustParse("8b8b8b8b-8b8b-8b8b-8b8b-8b8b8b8b8b8b"),
				assetID:     uuid.MustParse("8c8c8c8c-8c8c-8c8c-8c8c-8c8c8c8c8c8c"),
				caption:     "preferred creator short two",
				publishedAt: now.Add(-2 * time.Hour),
			},
		},
		displayName: "Recommended B",
		handle:      "recommendedb",
	}
	creatorC := creatorScenario{
		userID:      uuid.MustParse("8d8d8d8d-8d8d-8d8d-8d8d-8d8d8d8d8d8d"),
		mainID:      uuid.MustParse("8e8e8e8e-8e8e-8e8e-8e8e-8e8e8e8e8e8e"),
		mainAssetID: uuid.MustParse("8f8f8f8f-8f8f-8f8f-8f8f-8f8f8f8f8f8f"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("90909090-9090-9090-9090-909090909091"),
			assetID:     uuid.MustParse("90909090-9090-9090-9090-909090909092"),
			caption:     "secondary creator short",
			publishedAt: now.Add(-30 * time.Minute),
		}},
		displayName: "Recommended C",
		handle:      "recommendedc",
	}

	for _, creator := range []creatorScenario{creatorA, creatorB, creatorC} {
		followViewer := creator.userID == creatorB.userID
		if err := insertCreatorScenario(ctx, pool, creator, followViewer, viewerID); err != nil {
			t.Fatalf("insertCreatorScenario(%s) error = %v, want nil", creator.handle, err)
		}
	}

	if err := upsertViewerCreatorFeatures(ctx, pool, viewerID, creatorB.userID, 2, 1); err != nil {
		t.Fatalf("upsertViewerCreatorFeatures(creatorB) error = %v, want nil", err)
	}
	if err := upsertViewerMainFeatures(ctx, pool, viewerID, creatorB.mainID, creatorB.userID, 2); err != nil {
		t.Fatalf("upsertViewerMainFeatures(creatorB) error = %v, want nil", err)
	}
	for _, short := range creatorB.shorts {
		if err := upsertShortGlobalFeatures(ctx, pool, short.id, creatorB.userID, creatorB.mainID, 2, 5); err != nil {
			t.Fatalf("upsertShortGlobalFeatures(creatorB short=%s) error = %v, want nil", short.id, err)
		}
	}
	if err := upsertViewerCreatorFeatures(ctx, pool, viewerID, creatorC.userID, 2, 0); err != nil {
		t.Fatalf("upsertViewerCreatorFeatures(creatorC) error = %v, want nil", err)
	}
	if err := upsertViewerMainFeatures(ctx, pool, viewerID, creatorC.mainID, creatorC.userID, 2); err != nil {
		t.Fatalf("upsertViewerMainFeatures(creatorC) error = %v, want nil", err)
	}
	if err := upsertShortGlobalFeatures(ctx, pool, creatorC.shorts[0].id, creatorC.userID, creatorC.mainID, 1, 1); err != nil {
		t.Fatalf("upsertShortGlobalFeatures(creatorC) error = %v, want nil", err)
	}

	repo := NewRepository(pool)

	items, nextCursor, err := repo.ListRecommended(ctx, &viewerID, nil, 3)
	if err != nil {
		t.Fatalf("ListRecommended() error = %v, want nil", err)
	}
	if len(items) != 3 {
		t.Fatalf("ListRecommended() item len got %d want %d", len(items), 3)
	}
	wantFirstPage := []uuid.UUID{
		creatorB.shorts[0].id,
		creatorC.shorts[0].id,
		creatorB.shorts[1].id,
	}
	for index, wantShortID := range wantFirstPage {
		if items[index].Short.ID != wantShortID {
			t.Fatalf("ListRecommended() first page short[%d] got %s want %s", index, items[index].Short.ID, wantShortID)
		}
	}
	if nextCursor == nil || len(nextCursor.RecommendedRemainingShortIDs) != 1 || nextCursor.RecommendedRemainingShortIDs[0] != creatorA.shorts[0].id {
		t.Fatalf("ListRecommended() next cursor got %#v want remaining [%s]", nextCursor, creatorA.shorts[0].id)
	}
}

func TestRepositoryListRecommendedUsesGlobalFallbackForPublicViewer(t *testing.T) {
	t.Parallel()

	ctx, pool, cleanup := newFeedTestDatabase(t)
	defer cleanup()

	now := time.Now().UTC().Truncate(time.Second)
	dummyViewerID := uuid.MustParse("91919191-9191-9191-9191-919191919191")
	creatorA := creatorScenario{
		userID:      uuid.MustParse("92929292-9292-9292-9292-929292929292"),
		mainID:      uuid.MustParse("93939393-9393-9393-9393-939393939393"),
		mainAssetID: uuid.MustParse("94949494-9494-9494-9494-949494949494"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("95959595-9595-9595-9595-959595959595"),
			assetID:     uuid.MustParse("96969696-9696-9696-9696-969696969696"),
			caption:     "recent low-prior short",
			publishedAt: now.Add(-10 * time.Minute),
		}},
		displayName: "Fallback A",
		handle:      "fallbacka",
	}
	creatorB := creatorScenario{
		userID:      uuid.MustParse("97979797-9797-9797-9797-979797979797"),
		mainID:      uuid.MustParse("98989898-9898-9898-9898-989898989898"),
		mainAssetID: uuid.MustParse("99999999-9999-9999-9999-999999999999"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("aaaaaaaa-1111-2222-3333-bbbbbbbbbbbb"),
			assetID:     uuid.MustParse("cccccccc-1111-2222-3333-dddddddddddd"),
			caption:     "older but high-prior short",
			publishedAt: now.Add(-2 * time.Hour),
		}},
		displayName: "Fallback B",
		handle:      "fallbackb",
	}

	for _, creator := range []creatorScenario{creatorA, creatorB} {
		if err := insertCreatorScenario(ctx, pool, creator, false, dummyViewerID); err != nil {
			t.Fatalf("insertCreatorScenario(%s) error = %v, want nil", creator.handle, err)
		}
	}
	if err := upsertShortGlobalFeatures(ctx, pool, creatorB.shorts[0].id, creatorB.userID, creatorB.mainID, 6, 8); err != nil {
		t.Fatalf("upsertShortGlobalFeatures(creatorB) error = %v, want nil", err)
	}

	repo := NewRepository(pool)

	items, nextCursor, err := repo.ListRecommended(ctx, nil, nil, 10)
	if err != nil {
		t.Fatalf("ListRecommended() error = %v, want nil", err)
	}
	if len(items) != 2 {
		t.Fatalf("ListRecommended() item len got %d want %d", len(items), 2)
	}
	if items[0].Short.ID != creatorB.shorts[0].id || items[1].Short.ID != creatorA.shorts[0].id {
		t.Fatalf("ListRecommended() got [%s %s] want [%s %s]", items[0].Short.ID, items[1].Short.ID, creatorB.shorts[0].id, creatorA.shorts[0].id)
	}
	if items[0].Viewer.IsFollowingCreator || items[0].Viewer.IsPinned || items[0].Unlock.IsUnlocked {
		t.Fatalf("ListRecommended() public viewer state got %#v want false relation state", items[0])
	}
	if nextCursor != nil {
		t.Fatalf("ListRecommended() cursor got %#v want nil", nextCursor)
	}
}

func TestRepositoryListRecommendedContinuationPreservesSnapshotAfterMutation(t *testing.T) {
	t.Parallel()

	ctx, pool, cleanup := newFeedTestDatabase(t)
	defer cleanup()

	viewerID := uuid.MustParse("10101010-2020-3030-4040-505050505050")
	if err := insertFeedViewer(ctx, pool, viewerID); err != nil {
		t.Fatalf("insertFeedViewer() error = %v, want nil", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	creatorA := creatorScenario{
		userID:      uuid.MustParse("11111111-2222-3333-4444-555555555556"),
		mainID:      uuid.MustParse("66666666-7777-8888-9999-aaaaaaaaaaab"),
		mainAssetID: uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-fffffffffff0"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("12121212-3434-5656-7878-909090909091"),
			assetID:     uuid.MustParse("13131313-3535-5757-7979-919191919192"),
			caption:     "recommended first page",
			publishedAt: now.Add(-10 * time.Minute),
		}},
		displayName: "Recommended Snapshot A",
		handle:      "recsnapa",
	}
	creatorB := creatorScenario{
		userID:      uuid.MustParse("14141414-2424-3434-4444-545454545454"),
		mainID:      uuid.MustParse("15151515-2525-3535-4545-555555555555"),
		mainAssetID: uuid.MustParse("16161616-2626-3636-4646-565656565656"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("17171717-2727-3737-4747-575757575757"),
			assetID:     uuid.MustParse("18181818-2828-3838-4848-585858585858"),
			caption:     "recommended second page",
			publishedAt: now.Add(-20 * time.Minute),
		}},
		displayName: "Recommended Snapshot B",
		handle:      "recsnapb",
	}
	creatorC := creatorScenario{
		userID:      uuid.MustParse("19191919-2929-3939-4949-595959595959"),
		mainID:      uuid.MustParse("1a1a1a1a-2a2a-3a3a-4a4a-5a5a5a5a5a5a"),
		mainAssetID: uuid.MustParse("1b1b1b1b-2b2b-3b3b-4b4b-5b5b5b5b5b5b"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("1c1c1c1c-2c2c-3c3c-4c4c-5c5c5c5c5c5c"),
			assetID:     uuid.MustParse("1d1d1d1d-2d2d-3d3d-4d4d-5d5d5d5d5d5d"),
			caption:     "recommended third page",
			publishedAt: now.Add(-30 * time.Minute),
		}},
		displayName: "Recommended Snapshot C",
		handle:      "recsnapc",
	}

	for _, creator := range []creatorScenario{creatorA, creatorB, creatorC} {
		if err := insertCreatorScenario(ctx, pool, creator, false, viewerID); err != nil {
			t.Fatalf("insertCreatorScenario(%s) error = %v, want nil", creator.handle, err)
		}
	}

	if err := upsertViewerCreatorFeatures(ctx, pool, viewerID, creatorA.userID, 3, 0); err != nil {
		t.Fatalf("upsertViewerCreatorFeatures(creatorA) error = %v, want nil", err)
	}
	if err := upsertViewerMainFeatures(ctx, pool, viewerID, creatorA.mainID, creatorA.userID, 3); err != nil {
		t.Fatalf("upsertViewerMainFeatures(creatorA) error = %v, want nil", err)
	}
	if err := upsertViewerCreatorFeatures(ctx, pool, viewerID, creatorB.userID, 2, 0); err != nil {
		t.Fatalf("upsertViewerCreatorFeatures(creatorB) error = %v, want nil", err)
	}
	if err := upsertViewerMainFeatures(ctx, pool, viewerID, creatorB.mainID, creatorB.userID, 2); err != nil {
		t.Fatalf("upsertViewerMainFeatures(creatorB) error = %v, want nil", err)
	}
	if err := upsertViewerCreatorFeatures(ctx, pool, viewerID, creatorC.userID, 1, 0); err != nil {
		t.Fatalf("upsertViewerCreatorFeatures(creatorC) error = %v, want nil", err)
	}
	if err := upsertViewerMainFeatures(ctx, pool, viewerID, creatorC.mainID, creatorC.userID, 1); err != nil {
		t.Fatalf("upsertViewerMainFeatures(creatorC) error = %v, want nil", err)
	}

	repo := NewRepository(pool)

	firstPage, nextCursor, err := repo.ListRecommended(ctx, &viewerID, nil, 1)
	if err != nil {
		t.Fatalf("ListRecommended(first page) error = %v, want nil", err)
	}
	if len(firstPage) != 1 || nextCursor == nil {
		t.Fatalf("ListRecommended(first page) got items=%d cursor=%#v want 1 snapshot cursor", len(firstPage), nextCursor)
	}
	if firstPage[0].Short.ID != creatorA.shorts[0].id {
		t.Fatalf("ListRecommended(first page) got short %s want %s", firstPage[0].Short.ID, creatorA.shorts[0].id)
	}

	hotCreator := creatorScenario{
		userID:      uuid.MustParse("1e1e1e1e-2e2e-3e3e-4e4e-5e5e5e5e5e5e"),
		mainID:      uuid.MustParse("1f1f1f1f-2f2f-3f3f-4f4f-5f5f5f5f5f5f"),
		mainAssetID: uuid.MustParse("20202020-3030-4040-5050-606060606060"),
		shorts: []scenarioShort{{
			id:          uuid.MustParse("21212121-3131-4141-5151-616161616161"),
			assetID:     uuid.MustParse("22222222-3232-4242-5252-626262626262"),
			caption:     "mutation should not inject",
			publishedAt: now.Add(-5 * time.Minute),
		}},
		displayName: "Recommended Snapshot D",
		handle:      "recsnapd",
	}
	if err := insertCreatorScenario(ctx, pool, hotCreator, false, viewerID); err != nil {
		t.Fatalf("insertCreatorScenario(%s) error = %v, want nil", hotCreator.handle, err)
	}
	if err := upsertShortGlobalFeatures(ctx, pool, hotCreator.shorts[0].id, hotCreator.userID, hotCreator.mainID, 10, 10); err != nil {
		t.Fatalf("upsertShortGlobalFeatures(hotCreator) error = %v, want nil", err)
	}
	if _, err := pool.Exec(
		ctx,
		`INSERT INTO app.main_unlocks (user_id, main_id, payment_provider_purchase_ref, purchased_at) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, main_id) DO NOTHING`,
		viewerID,
		creatorB.mainID,
		"snapshot-b-unlock",
		now,
	); err != nil {
		t.Fatalf("INSERT main_unlocks(creatorB) error = %v, want nil", err)
	}

	secondPage, finalCursor, err := repo.ListRecommended(ctx, &viewerID, nextCursor, 2)
	if err != nil {
		t.Fatalf("ListRecommended(second page) error = %v, want nil", err)
	}
	if len(secondPage) != 2 || finalCursor != nil {
		t.Fatalf("ListRecommended(second page) got items=%d cursor=%#v want 2 nil", len(secondPage), finalCursor)
	}
	if secondPage[0].Short.ID != creatorB.shorts[0].id || secondPage[1].Short.ID != creatorC.shorts[0].id {
		t.Fatalf(
			"ListRecommended(second page) got [%s %s] want snapshot [%s %s]",
			secondPage[0].Short.ID,
			secondPage[1].Short.ID,
			creatorB.shorts[0].id,
			creatorC.shorts[0].id,
		)
	}
	if !secondPage[0].Unlock.IsUnlocked {
		t.Fatalf("ListRecommended(second page) creatorB unlock got %t want true after mutation", secondPage[0].Unlock.IsUnlocked)
	}
	for _, item := range secondPage {
		if item.Short.ID == hotCreator.shorts[0].id {
			t.Fatalf("ListRecommended(second page) injected new hot short %s", item.Short.ID)
		}
	}
}

type scenarioShort struct {
	id          uuid.UUID
	assetID     uuid.UUID
	caption     string
	publishedAt time.Time
}

type creatorScenario struct {
	userID      uuid.UUID
	mainID      uuid.UUID
	mainAssetID uuid.UUID
	shorts      []scenarioShort
	displayName string
	handle      string
}

func insertCreatorScenario(
	ctx context.Context,
	pool *pgxpool.Pool,
	scenario creatorScenario,
	followViewer bool,
	viewerID uuid.UUID,
) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `INSERT INTO app.users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`, scenario.userID); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.user_profiles (user_id, display_name, handle, avatar_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET
			display_name = EXCLUDED.display_name,
			handle = EXCLUDED.handle,
			avatar_url = EXCLUDED.avatar_url,
			updated_at = CURRENT_TIMESTAMP
	`, scenario.userID, scenario.displayName, scenario.handle, "https://cdn.example.com/avatar/"+scenario.handle+".jpg"); err != nil {
		return fmt.Errorf("insert user profile: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.creator_capabilities (
			user_id,
			state,
			rejection_reason_code,
			is_resubmit_eligible,
			is_support_review_required,
			self_serve_resubmit_count,
			kyc_provider_case_ref,
			payout_provider_account_ref,
			submitted_at,
			approved_at,
			rejected_at,
			suspended_at
		) VALUES (
			$1,
			'approved',
			NULL,
			FALSE,
			FALSE,
			0,
			$2,
			$3,
			NULL,
			$4,
			NULL,
			NULL
		)
		ON CONFLICT (user_id) DO UPDATE
		SET
			state = EXCLUDED.state,
			kyc_provider_case_ref = EXCLUDED.kyc_provider_case_ref,
			payout_provider_account_ref = EXCLUDED.payout_provider_account_ref,
			approved_at = EXCLUDED.approved_at,
			updated_at = CURRENT_TIMESTAMP
	`, scenario.userID, "kyc-"+scenario.handle, "payout-"+scenario.handle, time.Now().UTC().Add(-24*time.Hour)); err != nil {
		return fmt.Errorf("insert creator capability: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.creator_profiles (user_id, display_name, handle, avatar_url, bio, published_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE
		SET
			display_name = EXCLUDED.display_name,
			handle = EXCLUDED.handle,
			avatar_url = EXCLUDED.avatar_url,
			bio = EXCLUDED.bio,
			published_at = EXCLUDED.published_at,
			updated_at = CURRENT_TIMESTAMP
	`, scenario.userID, scenario.displayName, scenario.handle, "https://cdn.example.com/avatar/"+scenario.handle+".jpg", "bio for "+scenario.handle, time.Now().UTC().Add(-23*time.Hour)); err != nil {
		return fmt.Errorf("insert creator profile: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.media_assets (
			id,
			creator_user_id,
			processing_state,
			storage_provider,
			storage_bucket,
			storage_key,
			playback_url,
			mime_type,
			duration_ms,
			external_upload_ref
		) VALUES (
			$1,
			$2,
			'ready',
			's3',
			'mock-media-bucket',
			$3,
			$4,
			'video/mp4',
			$5,
			NULL
		)
		ON CONFLICT (id) DO UPDATE
		SET
			playback_url = EXCLUDED.playback_url,
			duration_ms = EXCLUDED.duration_ms,
			updated_at = CURRENT_TIMESTAMP
	`, scenario.mainAssetID, scenario.userID, "mock/mains/"+scenario.handle+".mp4", "https://cdn.example.com/mains/"+scenario.handle+".m3u8", int64(182000)); err != nil {
		return fmt.Errorf("insert main asset: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.mains (
			id,
			creator_user_id,
			media_asset_id,
			state,
			review_reason_code,
			post_report_state,
			price_minor,
			currency_code,
			ownership_confirmed,
			consent_confirmed,
			approved_for_unlock_at
		) VALUES (
			$1,
			$2,
			$3,
			'approved_for_unlock',
			NULL,
			NULL,
			1800,
			'JPY',
			TRUE,
			TRUE,
			$4
		)
		ON CONFLICT (id) DO UPDATE
		SET
			state = EXCLUDED.state,
			price_minor = EXCLUDED.price_minor,
			approved_for_unlock_at = EXCLUDED.approved_for_unlock_at,
			updated_at = CURRENT_TIMESTAMP
	`, scenario.mainID, scenario.userID, scenario.mainAssetID, time.Now().UTC().Add(-22*time.Hour)); err != nil {
		return fmt.Errorf("insert main: %w", err)
	}
	for _, short := range scenario.shorts {
		if _, err := tx.Exec(ctx, `
			INSERT INTO app.media_assets (
				id,
				creator_user_id,
				processing_state,
				storage_provider,
				storage_bucket,
				storage_key,
				playback_url,
				mime_type,
				duration_ms,
				external_upload_ref
			) VALUES (
				$1,
				$2,
				'ready',
				's3',
				'mock-media-bucket',
				$3,
				$4,
				'video/mp4',
				18000,
				NULL
			)
			ON CONFLICT (id) DO UPDATE
			SET
				playback_url = EXCLUDED.playback_url,
				updated_at = CURRENT_TIMESTAMP
		`, short.assetID, scenario.userID, "mock/shorts/"+scenario.handle+"-"+short.id.String()+".mp4", "https://cdn.example.com/shorts/"+scenario.handle+"-"+short.id.String()+".m3u8"); err != nil {
			return fmt.Errorf("insert short asset short=%s: %w", short.id, err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO app.shorts (
				id,
				creator_user_id,
				canonical_main_id,
				media_asset_id,
				caption,
				state,
				review_reason_code,
				post_report_state,
				approved_for_publish_at,
				published_at
			) VALUES (
				$1,
				$2,
				$3,
				$4,
				$5,
				'approved_for_publish',
				NULL,
				NULL,
				$6,
				$7
			)
			ON CONFLICT (id) DO UPDATE
			SET
				caption = EXCLUDED.caption,
				approved_for_publish_at = EXCLUDED.approved_for_publish_at,
				published_at = EXCLUDED.published_at,
				updated_at = CURRENT_TIMESTAMP
		`, short.id, scenario.userID, scenario.mainID, short.assetID, short.caption, short.publishedAt.Add(-30*time.Minute), short.publishedAt); err != nil {
			return fmt.Errorf("insert short short=%s: %w", short.id, err)
		}
	}
	if followViewer {
		if _, err := tx.Exec(ctx, `
			INSERT INTO app.creator_follows (user_id, creator_user_id, followed_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, creator_user_id) DO UPDATE
			SET followed_at = EXCLUDED.followed_at
		`, viewerID, scenario.userID, time.Now().UTC().Add(-20*time.Hour)); err != nil {
			return fmt.Errorf("insert creator follow: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func upsertViewerCreatorFeatures(
	ctx context.Context,
	pool *pgxpool.Pool,
	viewerID uuid.UUID,
	creatorUserID uuid.UUID,
	mainClickCount int64,
	profileClickCount int64,
) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO app.recommendation_viewer_creator_features (
			viewer_user_id,
			creator_user_id,
			main_click_count,
			profile_click_count
		) VALUES ($1, $2, $3, $4)
		ON CONFLICT (viewer_user_id, creator_user_id) DO UPDATE
		SET
			main_click_count = EXCLUDED.main_click_count,
			profile_click_count = EXCLUDED.profile_click_count,
			updated_at = CURRENT_TIMESTAMP
	`, viewerID, creatorUserID, mainClickCount, profileClickCount)
	if err != nil {
		return fmt.Errorf("upsert recommendation_viewer_creator_features: %w", err)
	}

	return nil
}

func upsertViewerMainFeatures(
	ctx context.Context,
	pool *pgxpool.Pool,
	viewerID uuid.UUID,
	mainID uuid.UUID,
	creatorUserID uuid.UUID,
	mainClickCount int64,
) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO app.recommendation_viewer_main_features (
			viewer_user_id,
			canonical_main_id,
			creator_user_id,
			main_click_count
		) VALUES ($1, $2, $3, $4)
		ON CONFLICT (viewer_user_id, canonical_main_id) DO UPDATE
		SET
			main_click_count = EXCLUDED.main_click_count,
			updated_at = CURRENT_TIMESTAMP
		WHERE app.recommendation_viewer_main_features.creator_user_id = EXCLUDED.creator_user_id
	`, viewerID, mainID, creatorUserID, mainClickCount)
	if err != nil {
		return fmt.Errorf("upsert recommendation_viewer_main_features: %w", err)
	}

	return nil
}

func upsertShortGlobalFeatures(
	ctx context.Context,
	pool *pgxpool.Pool,
	shortID uuid.UUID,
	creatorUserID uuid.UUID,
	mainID uuid.UUID,
	unlockConversionCount int64,
	mainClickCount int64,
) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO app.recommendation_short_global_features (
			short_id,
			creator_user_id,
			canonical_main_id,
			unlock_conversion_count,
			main_click_count
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (short_id) DO UPDATE
		SET
			unlock_conversion_count = EXCLUDED.unlock_conversion_count,
			main_click_count = EXCLUDED.main_click_count,
			updated_at = CURRENT_TIMESTAMP
		WHERE app.recommendation_short_global_features.creator_user_id = EXCLUDED.creator_user_id
			AND app.recommendation_short_global_features.canonical_main_id = EXCLUDED.canonical_main_id
	`, shortID, creatorUserID, mainID, unlockConversionCount, mainClickCount)
	if err != nil {
		return fmt.Errorf("upsert recommendation_short_global_features: %w", err)
	}

	return nil
}

func insertFeedViewer(ctx context.Context, pool *pgxpool.Pool, viewerID uuid.UUID) error {
	if _, err := pool.Exec(ctx, `INSERT INTO app.users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`, viewerID); err != nil {
		return fmt.Errorf("insert viewer user: %w", err)
	}

	return nil
}

func newFeedTestDatabase(t *testing.T) (context.Context, *pgxpool.Pool, func()) {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv(feedIntegrationPostgresDSNEnv))
	if dsn == "" {
		t.Skipf("%s is required for feed integration tests", feedIntegrationPostgresDSNEnv)
	}

	ctx := context.Background()
	baseConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("pgx.ParseConfig() error = %v, want nil", err)
	}

	adminConn, err := connectFeedAdminDatabase(ctx, baseConfig)
	if err != nil {
		t.Fatalf("connectFeedAdminDatabase() error = %v, want nil", err)
	}

	tempDatabaseName := "feed_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := adminConn.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{tempDatabaseName}.Sanitize()); err != nil {
		adminConn.Close(ctx)
		t.Fatalf("CREATE DATABASE %q error = %v, want nil", tempDatabaseName, err)
	}

	tempConfig := baseConfig.Copy()
	tempConfig.Database = tempDatabaseName
	migrator := newFeedTestMigrator(t, tempConfig)
	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		closeFeedMigrator(t, migrator)
		dropFeedTempDatabase(t, ctx, adminConn, tempDatabaseName)
		adminConn.Close(ctx)
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		closeFeedMigrator(t, migrator)
		dropFeedTempDatabase(t, ctx, adminConn, tempDatabaseName)
		adminConn.Close(ctx)
		t.Fatalf("pgxpool.ParseConfig() error = %v, want nil", err)
	}
	poolConfig.ConnConfig.Database = tempDatabaseName

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		closeFeedMigrator(t, migrator)
		dropFeedTempDatabase(t, ctx, adminConn, tempDatabaseName)
		adminConn.Close(ctx)
		t.Fatalf("pgxpool.NewWithConfig() error = %v, want nil", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		closeFeedMigrator(t, migrator)
		dropFeedTempDatabase(t, ctx, adminConn, tempDatabaseName)
		adminConn.Close(ctx)
		t.Fatalf("pool.Ping() error = %v, want nil", err)
	}

	cleanup := func() {
		pool.Close()
		closeFeedMigrator(t, migrator)
		dropFeedTempDatabase(t, ctx, adminConn, tempDatabaseName)
		adminConn.Close(ctx)
	}

	return ctx, pool, cleanup
}

func connectFeedAdminDatabase(ctx context.Context, baseConfig *pgx.ConnConfig) (*pgx.Conn, error) {
	var lastErr error
	for _, databaseName := range uniqueNonEmptyFeedStrings("postgres", baseConfig.Database, "template1") {
		adminConfig := baseConfig.Copy()
		adminConfig.Database = databaseName

		conn, err := pgx.ConnectConfig(ctx, adminConfig)
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("connect admin database: %w", lastErr)
}

func newFeedTestMigrator(t *testing.T, config *pgx.ConnConfig) *migrate.Migrate {
	t.Helper()

	db := stdlib.OpenDB(*config)

	driver, err := pgmigrate.WithInstance(db, &pgmigrate.Config{})
	if err != nil {
		db.Close()
		t.Fatalf("postgres.WithInstance() error = %v, want nil", err)
	}

	sourceURL := (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(feedMigrationDir(t)),
	}).String()

	migrator, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		db.Close()
		t.Fatalf("migrate.NewWithDatabaseInstance() error = %v, want nil", err)
	}

	return migrator
}

func closeFeedMigrator(t *testing.T, migrator *migrate.Migrate) {
	t.Helper()

	if migrator == nil {
		return
	}

	sourceErr, databaseErr := migrator.Close()
	if sourceErr != nil {
		t.Fatalf("migrator.Close() source error = %v, want nil", sourceErr)
	}
	if databaseErr != nil {
		t.Fatalf("migrator.Close() database error = %v, want nil", databaseErr)
	}
}

func dropFeedTempDatabase(t *testing.T, ctx context.Context, adminConn *pgx.Conn, databaseName string) {
	t.Helper()

	if _, err := adminConn.Exec(
		ctx,
		`SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1
			AND pid <> pg_backend_pid()`,
		databaseName,
	); err != nil {
		t.Fatalf("pg_terminate_backend(%q) error = %v, want nil", databaseName, err)
	}

	if _, err := adminConn.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{databaseName}.Sanitize()); err != nil {
		t.Fatalf("DROP DATABASE %q error = %v, want nil", databaseName, err)
	}
}

func feedMigrationDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() ok = false, want true")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "db", "migrations"))
}

func uniqueNonEmptyFeedStrings(values ...string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}
