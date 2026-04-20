package httpserver

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/submissionreview"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type adminSubmissionReviewServiceStub struct {
	applyDecision func(context.Context, submissionreview.ReviewDecisionInput) error
	getCase       func(context.Context, uuid.UUID) (submissionreview.AdminReviewCase, error)
	listCases     func(context.Context) ([]submissionreview.AdminReviewQueueItem, error)
}

func (s adminSubmissionReviewServiceStub) ApplyDecision(
	ctx context.Context,
	input submissionreview.ReviewDecisionInput,
) error {
	return s.applyDecision(ctx, input)
}

func (s adminSubmissionReviewServiceStub) GetCase(ctx context.Context, intakeID uuid.UUID) (submissionreview.AdminReviewCase, error) {
	return s.getCase(ctx, intakeID)
}

func (s adminSubmissionReviewServiceStub) ListCases(ctx context.Context) ([]submissionreview.AdminReviewQueueItem, error) {
	return s.listCases(ctx)
}

func TestAdminSubmissionReviewRoutesAreDisabledOutsideDevelopment(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		AppEnv: "production",
		AdminSubmissionReview: adminSubmissionReviewServiceStub{
			listCases: func(context.Context) ([]submissionreview.AdminReviewQueueItem, error) {
				t.Fatal("ListCases() called, want no route registration")
				return nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/submission-reviews", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/admin/submission-reviews status got %d want %d", rec.Code, http.StatusNotFound)
	}
}

func TestAdminSubmissionReviewRoutesRejectNonLoopbackRequests(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		AppEnv: developmentAppEnv,
		AdminSubmissionReview: adminSubmissionReviewServiceStub{
			listCases: func(context.Context) ([]submissionreview.AdminReviewQueueItem, error) {
				t.Fatal("ListCases() called, want loopback guard to block request first")
				return nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/submission-reviews", nil)
	req.RemoteAddr = "203.0.113.10:4321"
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/admin/submission-reviews status got %d want %d", rec.Code, http.StatusNotFound)
	}
}

func TestAdminSubmissionReviewQueueGetReturnsItems(t *testing.T) {
	t.Parallel()

	intakeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorUserID := uuid.MustParse("12111111-1111-1111-1111-111111111111")
	submittedAt := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)

	router := NewHandler(HandlerConfig{
		AppEnv: developmentAppEnv,
		AdminSubmissionReview: adminSubmissionReviewServiceStub{
			listCases: func(context.Context) ([]submissionreview.AdminReviewQueueItem, error) {
				return []submissionreview.AdminReviewQueueItem{
					{
						Creator: submissionreview.AdminReviewCreator{
							Bio:         "quiet rooftop",
							DisplayName: "Mina Rei",
							Handle:      "minarei",
							UserID:      creatorUserID,
						},
						IntakeID:             intakeID,
						MainDecisionRequired: true,
						PendingShortCount:    2,
						ShortCount:           3,
						SubmitKind:           "initial_submit",
						SubmittedAt:          submittedAt,
					},
				}, nil
			},
		},
	})

	req := newLoopbackAdminRequest(http.MethodGet, "/api/admin/submission-reviews", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/admin/submission-reviews status got %d want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"intakeId":"11111111-1111-1111-1111-111111111111"`) {
		t.Fatalf("GET /api/admin/submission-reviews body got %q want intake id", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"mainDecisionRequired":true`) {
		t.Fatalf("GET /api/admin/submission-reviews body got %q want main decision required", rec.Body.String())
	}
}

func TestAdminSubmissionReviewCaseGetMapsNotFound(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		AppEnv: developmentAppEnv,
		AdminSubmissionReview: adminSubmissionReviewServiceStub{
			getCase: func(context.Context, uuid.UUID) (submissionreview.AdminReviewCase, error) {
				return submissionreview.AdminReviewCase{}, submissionreview.ErrAdminReviewCaseNotFound
			},
		},
	})

	req := newLoopbackAdminRequest(http.MethodGet, "/api/admin/submission-reviews/11111111-1111-1111-1111-111111111111", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/admin/submission-reviews/:intakeId status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("GET /api/admin/submission-reviews/:intakeId body got %q want not_found", rec.Body.String())
	}
}

func TestAdminSubmissionReviewCaseGetRejectsInvalidIntakeID(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		AppEnv: developmentAppEnv,
		AdminSubmissionReview: adminSubmissionReviewServiceStub{
			getCase: func(context.Context, uuid.UUID) (submissionreview.AdminReviewCase, error) {
				t.Fatal("GetCase() called, want invalid request to fail first")
				return submissionreview.AdminReviewCase{}, nil
			},
		},
	})

	req := newLoopbackAdminRequest(http.MethodGet, "/api/admin/submission-reviews/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("GET /api/admin/submission-reviews/:intakeId status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("GET /api/admin/submission-reviews/:intakeId body got %q want invalid_request", rec.Body.String())
	}
}

func TestAdminSubmissionReviewDecisionPostReturnsUpdatedCase(t *testing.T) {
	t.Parallel()

	intakeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	var gotInput submissionreview.ReviewDecisionInput

	router := NewHandler(HandlerConfig{
		AppEnv: developmentAppEnv,
		AdminSubmissionReview: adminSubmissionReviewServiceStub{
			applyDecision: func(_ context.Context, input submissionreview.ReviewDecisionInput) error {
				gotInput = input
				return nil
			},
		},
	})

	req := newLoopbackAdminRequest(
		http.MethodPost,
		"/api/admin/submission-reviews/11111111-1111-1111-1111-111111111111/decision",
		bytes.NewBufferString(`{
			"mainDecision": {
				"decision": "approved",
				"reviewNote": "unlock ready"
			},
			"shortDecisions": [
				{
					"shortId": "33333333-3333-3333-3333-333333333333",
					"decision": "revision_requested",
					"reasonCode": "quality_issue",
					"reviewNote": "reframe intro"
				}
			]
		}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST /api/admin/submission-reviews/:intakeId/decision status got %d want %d", rec.Code, http.StatusNoContent)
	}
	if gotInput.IntakeID != intakeID {
		t.Fatalf("ApplyDecision() intake id got %s want %s", gotInput.IntakeID, intakeID)
	}
	if gotInput.MainDecision == nil || gotInput.MainDecision.Decision != "approved" {
		t.Fatalf("ApplyDecision() main decision got %#v want approved", gotInput.MainDecision)
	}
	if gotInput.MainDecision.ReviewNote == nil || *gotInput.MainDecision.ReviewNote != "unlock ready" {
		t.Fatalf("ApplyDecision() main review note got %#v want unlock ready", gotInput.MainDecision.ReviewNote)
	}
	if len(gotInput.ShortDecisions) != 1 {
		t.Fatalf("ApplyDecision() short decisions len got %d want 1", len(gotInput.ShortDecisions))
	}
	if gotInput.ShortDecisions[0].ShortID != shortID {
		t.Fatalf("ApplyDecision() short id got %s want %s", gotInput.ShortDecisions[0].ShortID, shortID)
	}
	if gotInput.ShortDecisions[0].ReviewNote == nil || *gotInput.ShortDecisions[0].ReviewNote != "reframe intro" {
		t.Fatalf("ApplyDecision() short review note got %#v want reframe intro", gotInput.ShortDecisions[0].ReviewNote)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "" {
		t.Fatalf("POST /api/admin/submission-reviews/:intakeId/decision body got %q want empty", rec.Body.String())
	}
}

func TestAdminSubmissionReviewDecisionPostRejectsInvalidShortID(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		AppEnv: developmentAppEnv,
		AdminSubmissionReview: adminSubmissionReviewServiceStub{
			applyDecision: func(context.Context, submissionreview.ReviewDecisionInput) error {
				t.Fatal("ApplyDecision() called, want invalid request to fail first")
				return nil
			},
		},
	})

	req := newLoopbackAdminRequest(
		http.MethodPost,
		"/api/admin/submission-reviews/11111111-1111-1111-1111-111111111111/decision",
		bytes.NewBufferString(`{
			"shortDecisions": [
				{
					"shortId": "not-a-uuid",
					"decision": "revision_requested",
					"reasonCode": "quality_issue"
				}
			]
		}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/admin/submission-reviews/:intakeId/decision status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("POST /api/admin/submission-reviews/:intakeId/decision body got %q want invalid_request", rec.Body.String())
	}
}

func TestWriteAdminSubmissionReviewError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err        error
		name       string
		wantCode   string
		wantMapped bool
		wantStatus int
	}{
		{
			err:        submissionreview.ErrInvalidReviewDecision,
			name:       "invalid decision",
			wantCode:   "invalid_review_decision",
			wantMapped: true,
			wantStatus: http.StatusBadRequest,
		},
		{
			err:        submissionreview.ErrReviewDecisionReasonRequired,
			name:       "reason required",
			wantCode:   "review_reason_required",
			wantMapped: true,
			wantStatus: http.StatusBadRequest,
		},
		{
			err:        submissionreview.ErrSubmissionReviewDecisionTargetsMismatch,
			name:       "targets mismatch",
			wantCode:   "review_targets_mismatch",
			wantMapped: true,
			wantStatus: http.StatusConflict,
		},
		{
			err:        submissionreview.ErrReviewStateConflict,
			name:       "state conflict",
			wantCode:   "review_state_conflict",
			wantMapped: true,
			wantStatus: http.StatusConflict,
		},
		{
			err:        errors.New("boom"),
			name:       "unmapped",
			wantMapped: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = newLoopbackAdminRequest(http.MethodGet, "/api/admin/submission-reviews", nil)

			gotMapped := writeAdminSubmissionReviewError(c, tt.err, "admin_submission_review_test")
			if gotMapped != tt.wantMapped {
				t.Fatalf("writeAdminSubmissionReviewError() mapped got %t want %t", gotMapped, tt.wantMapped)
			}
			if !tt.wantMapped {
				return
			}
			if rec.Code != tt.wantStatus {
				t.Fatalf("writeAdminSubmissionReviewError() status got %d want %d", rec.Code, tt.wantStatus)
			}
			if !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
				t.Fatalf("writeAdminSubmissionReviewError() body got %q want code %q", rec.Body.String(), tt.wantCode)
			}
		})
	}
}

func sampleAdminSubmissionReviewCase() submissionreview.AdminReviewCase {
	intakeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorUserID := uuid.MustParse("12111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainAssetID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	shortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortAssetID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	submittedAt := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)
	decisionedAt := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	reviewSource := "manual"
	mainNote := "unlock ready"
	shortReason := "quality_issue"
	shortNote := "reframe intro"

	return submissionreview.AdminReviewCase{
		Creator: submissionreview.AdminReviewCreator{
			Bio:         "quiet rooftop",
			DisplayName: "Mina Rei",
			Handle:      "minarei",
			UserID:      creatorUserID,
		},
		Intake: submissionreview.AdminReviewIntake{
			CanonicalMainID:    mainID,
			ConsentConfirmed:   true,
			CreatorUserID:      creatorUserID,
			ID:                 intakeID,
			MainMediaAssetID:   mainAssetID,
			MainPriceJpy:       1800,
			OwnershipConfirmed: true,
			Status:             "decision_applied",
			SubmitKind:         "initial_submit",
			SubmittedAt:        submittedAt,
		},
		Main: submissionreview.AdminReviewMain{
			CurrencyCode:     "JPY",
			DecisionRequired: false,
			ID:               mainID,
			IntakeDecisionLog: &submissionreview.IntakeDecisionLog{
				DecisionSource: reviewSource,
				DecisionedAt:   decisionedAt,
				ReviewNote:     &mainNote,
				TargetState:    "approved_for_unlock",
			},
			Media: media.VideoDisplayAsset{
				DurationSeconds: 540,
				ID:              mainAssetID,
				Kind:            "video",
				PosterURL:       "https://cdn.example.com/mains/poster.jpg",
				URL:             "https://cdn.example.com/mains/video.mp4",
			},
			PriceJpy: 1800,
			Review: submissionreview.ReviewProvenance{
				DecisionSource: &reviewSource,
				DecisionedAt:   &decisionedAt,
				ReviewNote:     &mainNote,
			},
			State: "approved_for_unlock",
		},
		Shorts: []submissionreview.AdminReviewShort{
			{
				Caption:          ptrString("quiet rooftop cut"),
				DecisionRequired: false,
				ID:               shortID,
				IntakeDecisionLog: &submissionreview.IntakeDecisionLog{
					DecisionSource: reviewSource,
					DecisionedAt:   decisionedAt,
					ReasonCode:     &shortReason,
					ReviewNote:     &shortNote,
					TargetState:    "revision_requested",
				},
				Media: media.VideoDisplayAsset{
					DurationSeconds: 18,
					ID:              shortAssetID,
					Kind:            "video",
					PosterURL:       "https://cdn.example.com/shorts/poster.jpg",
					URL:             "https://cdn.example.com/shorts/video.mp4",
				},
				Review: submissionreview.ReviewProvenance{
					DecisionSource: &reviewSource,
					DecisionedAt:   &decisionedAt,
					ReasonCode:     &shortReason,
					ReviewNote:     &shortNote,
				},
				State: "revision_requested",
			},
		},
	}
}

func ptrString(value string) *string {
	return &value
}
