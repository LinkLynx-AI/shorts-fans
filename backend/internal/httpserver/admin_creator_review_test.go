package httpserver

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatorregistration"
	"github.com/google/uuid"
)

type adminCreatorReviewServiceStub struct {
	applyDecision func(context.Context, creatorregistration.ReviewDecisionInput) (creatorregistration.ReviewCase, error)
	getCase       func(context.Context, uuid.UUID) (creatorregistration.ReviewCase, error)
	listCases     func(context.Context, string) ([]creatorregistration.ReviewQueueItem, error)
}

func (s adminCreatorReviewServiceStub) ApplyDecision(
	ctx context.Context,
	input creatorregistration.ReviewDecisionInput,
) (creatorregistration.ReviewCase, error) {
	return s.applyDecision(ctx, input)
}

func (s adminCreatorReviewServiceStub) GetCase(ctx context.Context, userID uuid.UUID) (creatorregistration.ReviewCase, error) {
	return s.getCase(ctx, userID)
}

func (s adminCreatorReviewServiceStub) ListCases(ctx context.Context, state string) ([]creatorregistration.ReviewQueueItem, error) {
	return s.listCases(ctx, state)
}

func TestAdminCreatorReviewRoutesAreDisabledOutsideDevelopment(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		AppEnv: "production",
		AdminCreatorReview: adminCreatorReviewServiceStub{
			listCases: func(context.Context, string) ([]creatorregistration.ReviewQueueItem, error) {
				t.Fatal("ListCases() called, want no route registration")
				return nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/creator-reviews", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/admin/creator-reviews status got %d want %d", rec.Code, http.StatusNotFound)
	}
}

func TestAdminCreatorReviewRoutesRejectNonLoopbackRequests(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		AppEnv: developmentAppEnv,
		AdminCreatorReview: adminCreatorReviewServiceStub{
			listCases: func(context.Context, string) ([]creatorregistration.ReviewQueueItem, error) {
				t.Fatal("ListCases() called, want loopback guard to block request first")
				return nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/creator-reviews", nil)
	req.RemoteAddr = "203.0.113.10:4321"
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/admin/creator-reviews status got %d want %d", rec.Code, http.StatusNotFound)
	}
}

func TestAdminCreatorReviewQueueGetReturnsItems(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	submittedAt := time.Date(2026, 4, 18, 9, 0, 0, 0, time.UTC)

	router := NewHandler(HandlerConfig{
		AppEnv: developmentAppEnv,
		AdminCreatorReview: adminCreatorReviewServiceStub{
			listCases: func(_ context.Context, state string) ([]creatorregistration.ReviewQueueItem, error) {
				if state != creatorregistration.StateSubmitted {
					t.Fatalf("ListCases() state got %q want %q", state, creatorregistration.StateSubmitted)
				}
				return []creatorregistration.ReviewQueueItem{
					{
						CreatorBio: "quiet rooftop",
						LegalName:  "Creator Legal",
						Review: creatorregistration.ReviewTimeline{
							SubmittedAt: &submittedAt,
						},
						SharedProfile: creatorregistration.SharedProfilePreview{
							DisplayName: "Creator Display",
							Handle:      "creator.handle",
							UserID:      userID,
						},
						State:  creatorregistration.StateSubmitted,
						UserID: userID,
					},
				}, nil
			},
		},
	})

	req := newLoopbackAdminRequest(http.MethodGet, "/api/admin/creator-reviews", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/admin/creator-reviews status got %d want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"userId":"11111111-1111-1111-1111-111111111111"`) {
		t.Fatalf("GET /api/admin/creator-reviews body got %q want user id", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"state":"submitted"`) {
		t.Fatalf("GET /api/admin/creator-reviews body got %q want submitted state", rec.Body.String())
	}
}

func TestAdminCreatorReviewCaseGetMapsNotFound(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		AppEnv: developmentAppEnv,
		AdminCreatorReview: adminCreatorReviewServiceStub{
			getCase: func(context.Context, uuid.UUID) (creatorregistration.ReviewCase, error) {
				return creatorregistration.ReviewCase{}, creatorregistration.ErrReviewCaseNotFound
			},
		},
	})

	req := newLoopbackAdminRequest(http.MethodGet, "/api/admin/creator-reviews/11111111-1111-1111-1111-111111111111", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/admin/creator-reviews/:userId status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("GET /api/admin/creator-reviews/:userId body got %q want not_found", rec.Body.String())
	}
}

func TestAdminCreatorReviewDecisionPostReturnsUpdatedCase(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	submittedAt := time.Date(2026, 4, 18, 9, 0, 0, 0, time.UTC)
	approvedAt := time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC)
	var gotInput creatorregistration.ReviewDecisionInput

	router := NewHandler(HandlerConfig{
		AppEnv: developmentAppEnv,
		AdminCreatorReview: adminCreatorReviewServiceStub{
			applyDecision: func(_ context.Context, input creatorregistration.ReviewDecisionInput) (creatorregistration.ReviewCase, error) {
				gotInput = input
				return creatorregistration.ReviewCase{
					CreatorBio: "quiet rooftop",
					Evidences: []creatorregistration.ReviewEvidence{
						{
							Evidence: creatorregistration.Evidence{
								FileName:      "government-id.png",
								FileSizeBytes: 128,
								Kind:          creatorregistration.EvidenceKindGovernmentID,
								MimeType:      "image/png",
								UploadedAt:    submittedAt,
							},
							AccessURL: "https://signed.example.com/government-id",
						},
					},
					Intake: creatorregistration.ReviewCaseIntake{
						LegalName: "Creator Legal",
					},
					Review: creatorregistration.ReviewTimeline{
						ApprovedAt:  &approvedAt,
						SubmittedAt: &submittedAt,
					},
					SharedProfile: creatorregistration.SharedProfilePreview{
						DisplayName: "Creator Display",
						Handle:      "creator.handle",
						UserID:      userID,
					},
					State:  creatorregistration.StateApproved,
					UserID: userID,
				}, nil
			},
		},
	})

	req := newLoopbackAdminRequest(
		http.MethodPost,
		"/api/admin/creator-reviews/11111111-1111-1111-1111-111111111111/decision",
		bytes.NewBufferString(`{"decision":"approved","reasonCode":"","isResubmitEligible":false,"isSupportReviewRequired":false}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/admin/creator-reviews/:userId/decision status got %d want %d", rec.Code, http.StatusOK)
	}
	if gotInput.Decision != creatorregistration.StateApproved {
		t.Fatalf("ApplyDecision() decision got %q want %q", gotInput.Decision, creatorregistration.StateApproved)
	}
	if gotInput.UserID != userID {
		t.Fatalf("ApplyDecision() user id got %s want %s", gotInput.UserID, userID)
	}
	if !strings.Contains(rec.Body.String(), `"state":"approved"`) {
		t.Fatalf("POST /api/admin/creator-reviews/:userId/decision body got %q want approved state", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"accessUrl":"https://signed.example.com/government-id"`) {
		t.Fatalf("POST /api/admin/creator-reviews/:userId/decision body got %q want access url", rec.Body.String())
	}
}

func TestAdminCreatorReviewDecisionPostMapsValidationErrors(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		AppEnv: developmentAppEnv,
		AdminCreatorReview: adminCreatorReviewServiceStub{
			applyDecision: func(context.Context, creatorregistration.ReviewDecisionInput) (creatorregistration.ReviewCase, error) {
				return creatorregistration.ReviewCase{}, creatorregistration.ErrReviewDecisionReasonRequired
			},
		},
	})

	req := newLoopbackAdminRequest(
		http.MethodPost,
		"/api/admin/creator-reviews/11111111-1111-1111-1111-111111111111/decision",
		bytes.NewBufferString(`{"decision":"rejected","reasonCode":"","isResubmitEligible":false,"isSupportReviewRequired":false}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/admin/creator-reviews/:userId/decision status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"review_reason_required"`) {
		t.Fatalf("POST /api/admin/creator-reviews/:userId/decision body got %q want review_reason_required", rec.Body.String())
	}
}

func newLoopbackAdminRequest(method string, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.RemoteAddr = "127.0.0.1:4321"
	return req
}
