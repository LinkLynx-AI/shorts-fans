package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/fanmain"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubFanUnlockMainService struct {
	createCardSetupSession func(context.Context, string, fanmain.CardSetupSessionInput) (fanmain.CardSetupSessionResult, error)
	getPlaybackSurface     func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, string) (fanmain.PlaybackSurface, error)
	getUnlockSurface       func(context.Context, uuid.UUID, string, uuid.UUID) (fanmain.UnlockSurface, error)
	issueAccessEntry       func(context.Context, string, fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error)
	issueCardSetupToken    func(context.Context, string, fanmain.CardSetupTokenInput) (fanmain.CardSetupTokenResult, error)
	purchaseMain           func(context.Context, string, fanmain.PurchaseInput) (fanmain.PurchaseResult, error)
}

const validSignedEntryToken = "signed-entry-token.payload"
const validCardSetupSessionToken = "card-setup-session-token.payload"

func (s stubFanUnlockMainService) CreateCardSetupSession(ctx context.Context, sessionBinding string, input fanmain.CardSetupSessionInput) (fanmain.CardSetupSessionResult, error) {
	return s.createCardSetupSession(ctx, sessionBinding, input)
}

func (s stubFanUnlockMainService) GetPlaybackSurface(ctx context.Context, viewerID uuid.UUID, sessionBinding string, mainID uuid.UUID, fromShortID uuid.UUID, grantToken string) (fanmain.PlaybackSurface, error) {
	return s.getPlaybackSurface(ctx, viewerID, sessionBinding, mainID, fromShortID, grantToken)
}

func (s stubFanUnlockMainService) GetUnlockSurface(ctx context.Context, viewerID uuid.UUID, sessionBinding string, shortID uuid.UUID) (fanmain.UnlockSurface, error) {
	return s.getUnlockSurface(ctx, viewerID, sessionBinding, shortID)
}

func (s stubFanUnlockMainService) IssueAccessEntry(ctx context.Context, sessionBinding string, input fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error) {
	return s.issueAccessEntry(ctx, sessionBinding, input)
}

func (s stubFanUnlockMainService) IssueCardSetupToken(ctx context.Context, sessionBinding string, input fanmain.CardSetupTokenInput) (fanmain.CardSetupTokenResult, error) {
	return s.issueCardSetupToken(ctx, sessionBinding, input)
}

func (s stubFanUnlockMainService) PurchaseMain(ctx context.Context, sessionBinding string, input fanmain.PurchaseInput) (fanmain.PurchaseResult, error) {
	return s.purchaseMain(ctx, sessionBinding, input)
}

type stubMainDisplayAssetResolver struct {
	resolve func(context.Context, media.MainDisplaySource, media.AccessBoundary, time.Duration) (media.VideoDisplayAsset, error)
}

func (s stubMainDisplayAssetResolver) ResolveMainDisplayAsset(ctx context.Context, source media.MainDisplaySource, boundary media.AccessBoundary, ttl time.Duration) (media.VideoDisplayAsset, error) {
	return s.resolve(ctx, source, boundary, ttl)
}

func assertEnvelopeErrorCode(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("response status got %d want %d", rec.Code, wantStatus)
	}
	if !strings.Contains(rec.Body.String(), `"code":"`+wantCode+`"`) {
		t.Fatalf("response body got %q want code %q", rec.Body.String(), wantCode)
	}
}

func TestFanShortUnlockRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	mainDurationSeconds := int64(480)
	priceJPY := int64(1800)

	router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
		getUnlockSurface: func(_ context.Context, gotViewerID uuid.UUID, sessionBinding string, gotShortID uuid.UUID) (fanmain.UnlockSurface, error) {
			if gotViewerID != viewerID || gotShortID != shortID || sessionBinding == "" {
				t.Fatalf("GetUnlockSurface() got viewer=%s short=%s session=%q", gotViewerID, gotShortID, sessionBinding)
			}

			return fanmain.UnlockSurface{
				Access: fanmain.MainAccessState{
					MainID: mainID,
					Reason: "unlock_required",
					Status: "locked",
				},
				Creator: fanmain.CreatorSummary{
					Bio:         "quiet rooftop specialist",
					DisplayName: "Mina Rei",
					Handle:      "minarei",
					ID:          viewerID,
				},
				Main: fanmain.MainSummary{
					DurationSeconds: 480,
					ID:              mainID,
					MediaAssetID:    uuid.MustParse("55555555-5555-5555-5555-555555555555"),
					PriceJPY:        1800,
				},
				MainAccessToken: "signed-entry-token",
				Purchase: fanmain.UnlockPurchaseState{
					SavedPaymentMethods: []fanmain.SavedPaymentMethodSummary{
						{
							Brand:           "visa",
							Last4:           "4242",
							PaymentMethodID: "paymeth_66666666666666666666666666666666",
						},
					},
					Setup: fanmain.PurchaseSetupState{},
					State: "purchase_ready",
					SupportedCardBrands: []string{
						"visa",
						"mastercard",
						"jcb",
						"american_express",
					},
				},
				Short: fanmain.ShortSummary{
					Caption:                "quiet rooftop preview",
					CanonicalMainID:        mainID,
					CreatorUserID:          viewerID,
					ID:                     shortID,
					MediaAssetID:           shortAssetID,
					PreviewDurationSeconds: 16,
				},
				UnlockCta: fanmain.UnlockCtaState{
					MainDurationSeconds: &mainDurationSeconds,
					PriceJPY:            &priceJPY,
					State:               "unlock_available",
				},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/shorts/short_22222222222222222222222222222222/unlock", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/shorts/:shortId/unlock status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[unlockSurfaceResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil {
		t.Fatal("response.Data = nil, want value")
	}
	if response.Data.EntryContext.Token != "signed-entry-token" {
		t.Fatalf("response.Data.EntryContext.Token got %q want %q", response.Data.EntryContext.Token, "signed-entry-token")
	}
	if response.Data.EntryContext.PurchasePath != "/api/fan/mains/main_33333333333333333333333333333333/purchase" {
		t.Fatalf("response.Data.EntryContext.PurchasePath got %q", response.Data.EntryContext.PurchasePath)
	}
	if response.Data.Purchase.State != "purchase_ready" || len(response.Data.Purchase.SavedPaymentMethods) != 1 {
		t.Fatalf("response.Data.Purchase got %#v want purchase_ready with saved card", response.Data.Purchase)
	}
}

func TestFanMainPurchaseRouteReturnsAcceptedForPending(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
		purchaseMain: func(_ context.Context, sessionBinding string, input fanmain.PurchaseInput) (fanmain.PurchaseResult, error) {
			if sessionBinding == "" || input.ViewerID != viewerID || input.MainID != mainID || input.FromShortID != shortID {
				t.Fatalf("PurchaseMain() input got %+v session=%q", input, sessionBinding)
			}
			if input.PaymentMethod.Mode != "saved_card" || input.PaymentMethod.PaymentMethodID != "paymeth_44444444444444444444444444444444" {
				t.Fatalf("PurchaseMain() payment method got %+v", input.PaymentMethod)
			}
			if input.ViewerIP != "192.0.2.1" {
				t.Fatalf("PurchaseMain() viewerIP got %q want %q", input.ViewerIP, "192.0.2.1")
			}

			return fanmain.PurchaseResult{
				Access: fanmain.MainAccessState{
					MainID: mainID,
					Reason: "unlock_required",
					Status: "locked",
				},
				Purchase: fanmain.PurchaseOutcome{
					CanRetry: false,
					Status:   "pending",
				},
			}, nil
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/fan/mains/main_33333333333333333333333333333333/purchase",
		strings.NewReader(`{"acceptedAge":true,"acceptedTerms":true,"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222","paymentMethod":{"mode":"saved_card","paymentMethodId":"paymeth_44444444444444444444444444444444"}}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.0.2.1:1234"
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("POST /api/fan/mains/:mainId/purchase status got %d want %d", rec.Code, http.StatusAccepted)
	}

	var response responseEnvelope[mainPurchaseResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Purchase.Status != "pending" {
		t.Fatalf("response.Data got %#v want pending", response.Data)
	}
	if response.Data.EntryContext != nil {
		t.Fatalf("response.Data.EntryContext got %#v want nil", response.Data.EntryContext)
	}
}

func TestFanMainCardSetupSessionRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
		createCardSetupSession: func(_ context.Context, sessionBinding string, input fanmain.CardSetupSessionInput) (fanmain.CardSetupSessionResult, error) {
			if sessionBinding == "" || input.ViewerID != viewerID || input.MainID != mainID || input.FromShortID != shortID {
				t.Fatalf("CreateCardSetupSession() input got %+v session=%q", input, sessionBinding)
			}
			if input.EntryToken != validSignedEntryToken {
				t.Fatalf("CreateCardSetupSession() entry token got %q want %q", input.EntryToken, validSignedEntryToken)
			}

			return fanmain.CardSetupSessionResult{
				APIBaseURL:    "https://api.ccbill.com",
				APIKey:        "frontend-token",
				ClientAccount: "900100",
				Currency:      "JPY",
				InitialPeriod: "30",
				InitialPrice:  "1800.00",
				SessionToken:  validCardSetupSessionToken,
				SubAccount:    "1",
			}, nil
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/fan/mains/main_33333333333333333333333333333333/card-setup-session",
		strings.NewReader(`{"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/fan/mains/:mainId/card-setup-session status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[cardSetupSessionResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.APIKey != "frontend-token" {
		t.Fatalf("response.Data got %#v want widget session", response.Data)
	}
	if response.Data.SessionToken != validCardSetupSessionToken {
		t.Fatalf("response.Data.SessionToken got %q want %q", response.Data.SessionToken, validCardSetupSessionToken)
	}
}

func TestFanMainCardSetupTokenRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
		issueCardSetupToken: func(_ context.Context, sessionBinding string, input fanmain.CardSetupTokenInput) (fanmain.CardSetupTokenResult, error) {
			if sessionBinding == "" || input.ViewerID != viewerID || input.MainID != mainID || input.FromShortID != shortID {
				t.Fatalf("IssueCardSetupToken() input got %+v session=%q", input, sessionBinding)
			}
			if input.CardSetupSessionToken != validCardSetupSessionToken {
				t.Fatalf("IssueCardSetupToken() session token got %q want %q", input.CardSetupSessionToken, validCardSetupSessionToken)
			}
			if input.PaymentTokenRef != "payment-token-1" {
				t.Fatalf("IssueCardSetupToken() payment token got %q want %q", input.PaymentTokenRef, "payment-token-1")
			}

			return fanmain.CardSetupTokenResult{CardSetupToken: "opaque-card-setup-token"}, nil
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/fan/mains/main_33333333333333333333333333333333/card-setup-token",
		strings.NewReader(`{"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222","paymentTokenId":"payment-token-1","sessionToken":"card-setup-session-token.payload"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/fan/mains/:mainId/card-setup-token status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[cardSetupTokenResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.CardSetupToken != "opaque-card-setup-token" {
		t.Fatalf("response.Data got %#v want card setup token", response.Data)
	}
}

func TestFanMainCardSetupSessionRouteErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		path        string
		body        string
		service     stubFanUnlockMainService
		wantCode    string
		wantStatus  int
	}{
		{
			name:       "invalid request body",
			path:       "/api/fan/mains/main_33333333333333333333333333333333/card-setup-session",
			body:       `{`,
			wantCode:   "invalid_request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid short id",
			path:       "/api/fan/mains/main_33333333333333333333333333333333/card-setup-session",
			body:       `{"entryToken":"signed-entry-token.payload","fromShortId":"invalid-short"}`,
			wantCode:   "invalid_request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing entry token",
			path:       "/api/fan/mains/main_33333333333333333333333333333333/card-setup-session",
			body:       `{"entryToken":"","fromShortId":"short_22222222222222222222222222222222"}`,
			wantCode:   "invalid_request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "main locked",
			path: "/api/fan/mains/main_33333333333333333333333333333333/card-setup-session",
			body: `{"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222"}`,
			service: stubFanUnlockMainService{
				createCardSetupSession: func(context.Context, string, fanmain.CardSetupSessionInput) (fanmain.CardSetupSessionResult, error) {
					return fanmain.CardSetupSessionResult{}, fanmain.ErrMainLocked
				},
			},
			wantCode:   "main_locked",
			wantStatus: http.StatusForbidden,
		},
		{
			name: "not found",
			path: "/api/fan/mains/main_33333333333333333333333333333333/card-setup-session",
			body: `{"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222"}`,
			service: stubFanUnlockMainService{
				createCardSetupSession: func(context.Context, string, fanmain.CardSetupSessionInput) (fanmain.CardSetupSessionResult, error) {
					return fanmain.CardSetupSessionResult{}, fanmain.ErrPurchaseNotFound
				},
			},
			wantCode:   "not_found",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			router := newFanUnlockMainRouter(t, tc.service)
			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assertEnvelopeErrorCode(t, rec, tc.wantStatus, tc.wantCode)
		})
	}
}

func TestFanMainCardSetupTokenRouteErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		path        string
		body        string
		service     stubFanUnlockMainService
		wantCode    string
		wantStatus  int
	}{
		{
			name:       "invalid request body",
			path:       "/api/fan/mains/main_33333333333333333333333333333333/card-setup-token",
			body:       `{`,
			wantCode:   "invalid_request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid short id",
			path:       "/api/fan/mains/main_33333333333333333333333333333333/card-setup-token",
			body:       `{"entryToken":"signed-entry-token.payload","fromShortId":"invalid-short","paymentTokenId":"payment-token-1","sessionToken":"card-setup-session-token.payload"}`,
			wantCode:   "invalid_request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing payment token",
			path:       "/api/fan/mains/main_33333333333333333333333333333333/card-setup-token",
			body:       `{"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222","paymentTokenId":"   ","sessionToken":"card-setup-session-token.payload"}`,
			wantCode:   "invalid_request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing session token",
			path:       "/api/fan/mains/main_33333333333333333333333333333333/card-setup-token",
			body:       `{"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222","paymentTokenId":"payment-token-1","sessionToken":""}`,
			wantCode:   "invalid_request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid payment token request",
			path: "/api/fan/mains/main_33333333333333333333333333333333/card-setup-token",
			body: `{"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222","paymentTokenId":"payment-token-1","sessionToken":"card-setup-session-token.payload"}`,
			service: stubFanUnlockMainService{
				issueCardSetupToken: func(context.Context, string, fanmain.CardSetupTokenInput) (fanmain.CardSetupTokenResult, error) {
					return fanmain.CardSetupTokenResult{}, fanmain.ErrInvalidCardSetupRequest
				},
			},
			wantCode:   "invalid_request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "main locked",
			path: "/api/fan/mains/main_33333333333333333333333333333333/card-setup-token",
			body: `{"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222","paymentTokenId":"payment-token-1","sessionToken":"card-setup-session-token.payload"}`,
			service: stubFanUnlockMainService{
				issueCardSetupToken: func(context.Context, string, fanmain.CardSetupTokenInput) (fanmain.CardSetupTokenResult, error) {
					return fanmain.CardSetupTokenResult{}, fanmain.ErrMainLocked
				},
			},
			wantCode:   "main_locked",
			wantStatus: http.StatusForbidden,
		},
		{
			name: "not found",
			path: "/api/fan/mains/main_33333333333333333333333333333333/card-setup-token",
			body: `{"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222","paymentTokenId":"payment-token-1","sessionToken":"card-setup-session-token.payload"}`,
			service: stubFanUnlockMainService{
				issueCardSetupToken: func(context.Context, string, fanmain.CardSetupTokenInput) (fanmain.CardSetupTokenResult, error) {
					return fanmain.CardSetupTokenResult{}, fanmain.ErrPurchaseNotFound
				},
			},
			wantCode:   "not_found",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			router := newFanUnlockMainRouter(t, tc.service)
			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assertEnvelopeErrorCode(t, rec, tc.wantStatus, tc.wantCode)
		})
	}
}

func TestFanMainPurchaseRouteReturnsEntryContextForSuccess(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
		purchaseMain: func(_ context.Context, sessionBinding string, input fanmain.PurchaseInput) (fanmain.PurchaseResult, error) {
			if sessionBinding == "" || input.ViewerID != viewerID || input.MainID != mainID || input.FromShortID != shortID {
				t.Fatalf("PurchaseMain() input got %+v session=%q", input, sessionBinding)
			}

			entryToken := validSignedEntryToken
			return fanmain.PurchaseResult{
				Access: fanmain.MainAccessState{
					MainID: mainID,
					Reason: "purchased",
					Status: "unlocked",
				},
				EntryToken: &entryToken,
				Purchase: fanmain.PurchaseOutcome{
					CanRetry: false,
					Status:   "succeeded",
				},
			}, nil
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/fan/mains/main_33333333333333333333333333333333/purchase",
		strings.NewReader(`{"acceptedAge":true,"acceptedTerms":true,"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222","paymentMethod":{"mode":"saved_card","paymentMethodId":"paymeth_44444444444444444444444444444444"}}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/fan/mains/:mainId/purchase status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[mainPurchaseResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.EntryContext == nil {
		t.Fatalf("response.Data got %#v want entry context", response.Data)
	}
	if response.Data.EntryContext.AccessEntryPath != "/api/fan/mains/main_33333333333333333333333333333333/access-entry" {
		t.Fatalf("response.Data.EntryContext got %#v", response.Data.EntryContext)
	}
}

func TestFanMainAccessEntryRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
		issueAccessEntry: func(_ context.Context, sessionBinding string, input fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error) {
			if sessionBinding == "" || input.ViewerID != viewerID || input.MainID != mainID || input.FromShortID != shortID {
				t.Fatalf("IssueAccessEntry() input got %+v session=%q", input, sessionBinding)
			}
			if input.EntryToken != validSignedEntryToken {
				t.Fatalf("IssueAccessEntry() entry token got %q want %q", input.EntryToken, validSignedEntryToken)
			}

			return fanmain.AccessEntryResult{
				GrantKind:  fanmain.MainPlaybackGrantKindPurchased,
				GrantToken: "signed-grant-token",
			}, nil
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/fan/mains/main_33333333333333333333333333333333/access-entry",
		strings.NewReader(`{"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/fan/mains/:mainId/access-entry status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[mainAccessEntryResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || !strings.Contains(response.Data.Href, "signed-grant-token") {
		t.Fatalf("response.Data.Href got %#v want grant token", response.Data)
	}
}

func TestFanMainPlaybackRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	mainAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")

	router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
		getPlaybackSurface: func(_ context.Context, gotViewerID uuid.UUID, sessionBinding string, gotMainID uuid.UUID, gotFromShortID uuid.UUID, gotGrant string) (fanmain.PlaybackSurface, error) {
			if gotViewerID != viewerID || gotMainID != mainID || gotFromShortID != shortID || gotGrant != "signed-grant-token" || sessionBinding == "" {
				t.Fatalf("GetPlaybackSurface() got viewer=%s main=%s fromShort=%s grant=%q session=%q", gotViewerID, gotMainID, gotFromShortID, gotGrant, sessionBinding)
			}

			return fanmain.PlaybackSurface{
				Access: fanmain.MainAccessState{
					MainID: mainID,
					Reason: "purchased",
					Status: "unlocked",
				},
				Creator: fanmain.CreatorSummary{
					Bio:         "quiet rooftop specialist",
					DisplayName: "Mina Rei",
					Handle:      "minarei",
					ID:          viewerID,
				},
				EntryShort: fanmain.ShortSummary{
					Caption:                "quiet rooftop preview",
					CanonicalMainID:        mainID,
					CreatorUserID:          viewerID,
					ID:                     shortID,
					MediaAssetID:           shortAssetID,
					PreviewDurationSeconds: 16,
				},
				Main: fanmain.MainSummary{
					DurationSeconds: 480,
					ID:              mainID,
					MediaAssetID:    mainAssetID,
					PriceJPY:        1800,
				},
			}, nil
		},
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/fan/mains/main_33333333333333333333333333333333/playback?fromShortId=short_22222222222222222222222222222222&grant=signed-grant-token",
		nil,
	)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/mains/:mainId/playback status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[mainPlaybackResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Access.Reason != "purchased" {
		t.Fatalf("response.Data.Access got %#v want purchased", response.Data)
	}
}

func TestFanUnlockMainErrorRoutes(t *testing.T) {
	t.Parallel()

	t.Run("unlock locked", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
			getUnlockSurface: func(context.Context, uuid.UUID, string, uuid.UUID) (fanmain.UnlockSurface, error) {
				return fanmain.UnlockSurface{}, fanmain.ErrMainLocked
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/api/fan/shorts/short_22222222222222222222222222222222/unlock", nil)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("GET /unlock status got %d want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("purchase invalid request", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
			purchaseMain: func(context.Context, string, fanmain.PurchaseInput) (fanmain.PurchaseResult, error) {
				return fanmain.PurchaseResult{}, fanmain.ErrInvalidPurchaseRequest
			},
		})

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/fan/mains/main_33333333333333333333333333333333/purchase",
			strings.NewReader(`{"acceptedAge":true,"acceptedTerms":true,"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222","paymentMethod":{"mode":"saved_card","paymentMethodId":"paymeth_44444444444444444444444444444444"}}`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("POST /purchase status got %d want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("purchase invalid entry context payload", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{})

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/fan/mains/main_33333333333333333333333333333333/purchase",
			strings.NewReader(`{"acceptedAge":true,"acceptedTerms":true,"entryToken":"signed-entry-token.payload","fromShortId":"invalid-short","paymentMethod":{"mode":"saved_card","paymentMethodId":"paymeth_44444444444444444444444444444444"}}`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertEnvelopeErrorCode(t, rec, http.StatusBadRequest, "invalid_request")
	})

	t.Run("access entry locked", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
			issueAccessEntry: func(context.Context, string, fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error) {
				return fanmain.AccessEntryResult{}, fanmain.ErrMainLocked
			},
		})

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/fan/mains/main_33333333333333333333333333333333/access-entry",
			strings.NewReader(`{"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222"}`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("POST /access-entry status got %d want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("access entry invalid entry context payload", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{})

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/fan/mains/main_33333333333333333333333333333333/access-entry",
			strings.NewReader(`{"entryToken":"invalid","fromShortId":"short_22222222222222222222222222222222"}`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assertEnvelopeErrorCode(t, rec, http.StatusBadRequest, "invalid_request")
	})

	t.Run("playback missing grant", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{})

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/fan/mains/main_33333333333333333333333333333333/playback?fromShortId=short_22222222222222222222222222222222",
			nil,
		)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("GET /playback status got %d want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("playback not found", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
			getPlaybackSurface: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, string) (fanmain.PlaybackSurface, error) {
				return fanmain.PlaybackSurface{}, fanmain.ErrPlaybackNotFound
			},
		})

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/fan/mains/main_33333333333333333333333333333333/playback?fromShortId=short_22222222222222222222222222222222&grant=signed-grant-token",
			nil,
		)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("GET /playback status got %d want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("purchase not found", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
			purchaseMain: func(context.Context, string, fanmain.PurchaseInput) (fanmain.PurchaseResult, error) {
				return fanmain.PurchaseResult{}, fanmain.ErrPurchaseNotFound
			},
		})

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/fan/mains/main_33333333333333333333333333333333/purchase",
			strings.NewReader(`{"acceptedAge":true,"acceptedTerms":true,"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222","paymentMethod":{"mode":"saved_card","paymentMethodId":"paymeth_44444444444444444444444444444444"}}`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("POST /purchase status got %d want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("purchase locked", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
			purchaseMain: func(context.Context, string, fanmain.PurchaseInput) (fanmain.PurchaseResult, error) {
				return fanmain.PurchaseResult{}, fanmain.ErrMainLocked
			},
		})

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/fan/mains/main_33333333333333333333333333333333/purchase",
			strings.NewReader(`{"acceptedAge":true,"acceptedTerms":true,"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222","paymentMethod":{"mode":"saved_card","paymentMethodId":"paymeth_44444444444444444444444444444444"}}`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("POST /purchase status got %d want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("access entry not found", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
			issueAccessEntry: func(context.Context, string, fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error) {
				return fanmain.AccessEntryResult{}, fanmain.ErrAccessEntryNotFound
			},
		})

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/fan/mains/main_33333333333333333333333333333333/access-entry",
			strings.NewReader(`{"entryToken":"signed-entry-token.payload","fromShortId":"short_22222222222222222222222222222222"}`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("POST /access-entry status got %d want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("access entry invalid request body", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{})

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/fan/mains/main_33333333333333333333333333333333/access-entry",
			strings.NewReader(`{`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("POST /access-entry status got %d want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("playback locked", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
			getPlaybackSurface: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, string) (fanmain.PlaybackSurface, error) {
				return fanmain.PlaybackSurface{}, fanmain.ErrMainLocked
			},
		})

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/fan/mains/main_33333333333333333333333333333333/playback?fromShortId=short_22222222222222222222222222222222&grant=signed-grant-token",
			nil,
		)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("GET /playback status got %d want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("playback internal error", func(t *testing.T) {
		t.Parallel()

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
			getPlaybackSurface: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, string) (fanmain.PlaybackSurface, error) {
				return fanmain.PlaybackSurface{}, errors.New("unexpected playback failure")
			},
		})

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/fan/mains/main_33333333333333333333333333333333/playback?fromShortId=short_22222222222222222222222222222222&grant=signed-grant-token",
			nil,
		)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("GET /playback status got %d want %d", rec.Code, http.StatusInternalServerError)
		}
	})

	t.Run("unlock creator payload invalid", func(t *testing.T) {
		t.Parallel()

		viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

		router := newFanUnlockMainRouter(t, stubFanUnlockMainService{
			getUnlockSurface: func(context.Context, uuid.UUID, string, uuid.UUID) (fanmain.UnlockSurface, error) {
				return fanmain.UnlockSurface{
					Access: fanmain.MainAccessState{
						MainID: mainID,
						Reason: "unlock_required",
						Status: "locked",
					},
					Creator: fanmain.CreatorSummary{
						DisplayName: "",
						Handle:      "",
						ID:          viewerID,
					},
					Main: fanmain.MainSummary{
						DurationSeconds: 480,
						ID:              mainID,
						PriceJPY:        1800,
					},
					MainAccessToken: "signed-entry-token",
					Purchase:        fanmain.UnlockPurchaseState{State: "purchase_ready"},
					Short: fanmain.ShortSummary{
						CanonicalMainID:        mainID,
						CreatorUserID:          viewerID,
						ID:                     shortID,
						MediaAssetID:           uuid.MustParse("44444444-4444-4444-4444-444444444444"),
						PreviewDurationSeconds: 16,
					},
					UnlockCta: fanmain.UnlockCtaState{State: "unlock_available"},
				}, nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/api/fan/shorts/short_22222222222222222222222222222222/unlock", nil)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("GET /unlock status got %d want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

func TestFanUnlockMainHelpers(t *testing.T) {
	t.Parallel()

	t.Run("resolve authenticated viewer request", func(t *testing.T) {
		t.Parallel()

		viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		ctx.Request = req
		ctx.Set(authenticatedViewerContextKey, auth.CurrentViewer{ID: viewerID})

		viewer, sessionBinding, ok := resolveAuthenticatedViewerRequest(ctx)
		if !ok || viewer.ID != viewerID || sessionBinding != auth.HashSessionToken("raw-session-token") {
			t.Fatalf("resolveAuthenticatedViewerRequest() got viewer=%#v session=%q ok=%t", viewer, sessionBinding, ok)
		}

		ctxNoCookie, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctxNoCookie.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		ctxNoCookie.Set(authenticatedViewerContextKey, auth.CurrentViewer{ID: viewerID})

		if _, _, ok := resolveAuthenticatedViewerRequest(ctxNoCookie); ok {
			t.Fatal("resolveAuthenticatedViewerRequest() ok = true, want false without cookie")
		}
	})

	t.Run("build unlock short summary returns resolver error", func(t *testing.T) {
		t.Parallel()

		_, err := buildUnlockShortSummary(
			fanmain.ShortSummary{
				CanonicalMainID:        uuid.MustParse("33333333-3333-3333-3333-333333333333"),
				CreatorUserID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				ID:                     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				MediaAssetID:           uuid.MustParse("44444444-4444-4444-4444-444444444444"),
				PreviewDurationSeconds: 16,
			},
			stubShortDisplayAssetResolver{
				resolve: func(media.ShortDisplaySource, media.AccessBoundary) (media.VideoDisplayAsset, error) {
					return media.VideoDisplayAsset{}, errors.New("asset lookup failed")
				},
			},
		)
		if err == nil {
			t.Fatal("buildUnlockShortSummary() error = nil, want resolver error")
		}
	})

	t.Run("resolve main playback boundary", func(t *testing.T) {
		t.Parallel()

		if got := resolveMainPlaybackBoundary("owner"); got != media.AccessBoundaryOwner {
			t.Fatalf("resolveMainPlaybackBoundary(owner) got %s want %s", got, media.AccessBoundaryOwner)
		}
		if got := resolveMainPlaybackBoundary("unlocked"); got != media.AccessBoundaryPrivate {
			t.Fatalf("resolveMainPlaybackBoundary(unlocked) got %s want %s", got, media.AccessBoundaryPrivate)
		}
	})
}

func newFanUnlockMainRouter(t *testing.T, service stubFanUnlockMainService) *gin.Engine {
	t.Helper()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	shortAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	mainAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")

	if service.getUnlockSurface == nil {
		service.getUnlockSurface = func(context.Context, uuid.UUID, string, uuid.UUID) (fanmain.UnlockSurface, error) {
			t.Fatal("GetUnlockSurface() was called unexpectedly")
			return fanmain.UnlockSurface{}, nil
		}
	}
	if service.purchaseMain == nil {
		service.purchaseMain = func(context.Context, string, fanmain.PurchaseInput) (fanmain.PurchaseResult, error) {
			t.Fatal("PurchaseMain() was called unexpectedly")
			return fanmain.PurchaseResult{}, nil
		}
	}
	if service.createCardSetupSession == nil {
		service.createCardSetupSession = func(context.Context, string, fanmain.CardSetupSessionInput) (fanmain.CardSetupSessionResult, error) {
			t.Fatal("CreateCardSetupSession() was called unexpectedly")
			return fanmain.CardSetupSessionResult{}, nil
		}
	}
	if service.issueAccessEntry == nil {
		service.issueAccessEntry = func(context.Context, string, fanmain.AccessEntryInput) (fanmain.AccessEntryResult, error) {
			t.Fatal("IssueAccessEntry() was called unexpectedly")
			return fanmain.AccessEntryResult{}, nil
		}
	}
	if service.issueCardSetupToken == nil {
		service.issueCardSetupToken = func(context.Context, string, fanmain.CardSetupTokenInput) (fanmain.CardSetupTokenResult, error) {
			t.Fatal("IssueCardSetupToken() was called unexpectedly")
			return fanmain.CardSetupTokenResult{}, nil
		}
	}
	if service.getPlaybackSurface == nil {
		service.getPlaybackSurface = func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, string) (fanmain.PlaybackSurface, error) {
			t.Fatal("GetPlaybackSurface() was called unexpectedly")
			return fanmain.PlaybackSurface{}, nil
		}
	}

	return NewHandler(HandlerConfig{
		FanUnlockMain: service,
		MainDisplayAssets: stubMainDisplayAssetResolver{
			resolve: func(_ context.Context, source media.MainDisplaySource, boundary media.AccessBoundary, ttl time.Duration) (media.VideoDisplayAsset, error) {
				if source.MainID != mainID || boundary != media.AccessBoundaryPrivate || ttl != 0 {
					t.Fatalf("ResolveMainDisplayAsset() got source=%+v boundary=%s ttl=%s", source, boundary, ttl)
				}

				return media.VideoDisplayAsset{
					DurationSeconds: 480,
					ID:              mainAssetID,
					Kind:            "video",
					PosterURL:       "https://cdn.example.com/mains/poster.jpg",
					URL:             "https://cdn.example.com/mains/playback.mp4",
				}, nil
			},
		},
		ShortDisplayAssets: stubShortDisplayAssetResolver{
			resolve: func(source media.ShortDisplaySource, boundary media.AccessBoundary) (media.VideoDisplayAsset, error) {
				if source.ShortID != shortID || boundary != media.AccessBoundaryPublic {
					t.Fatalf("ResolveShortDisplayAsset() got source=%+v boundary=%s", source, boundary)
				}

				return media.VideoDisplayAsset{
					DurationSeconds: 16,
					ID:              shortAssetID,
					Kind:            "video",
					PosterURL:       "https://cdn.example.com/shorts/poster.jpg",
					URL:             "https://cdn.example.com/shorts/playback.mp4",
				}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID: viewerID,
					},
				}, nil
			},
		},
	})
}
