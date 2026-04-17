package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatoravatar"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatorregistration"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/viewerprofile"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type viewerCreatorRegistrationServiceStub struct {
	getIntake       func(context.Context, uuid.UUID) (creatorregistration.Intake, error)
	getRegistration func(context.Context, uuid.UUID) (*creatorregistration.Registration, error)
	saveIntake      func(context.Context, creatorregistration.SaveIntakeInput) (creatorregistration.Intake, error)
	submit          func(context.Context, uuid.UUID) (creatorregistration.Registration, error)
}

func (s viewerCreatorRegistrationServiceStub) GetIntake(ctx context.Context, userID uuid.UUID) (creatorregistration.Intake, error) {
	if s.getIntake == nil {
		return creatorregistration.Intake{}, nil
	}

	return s.getIntake(ctx, userID)
}

func (s viewerCreatorRegistrationServiceStub) GetRegistration(ctx context.Context, userID uuid.UUID) (*creatorregistration.Registration, error) {
	if s.getRegistration == nil {
		return nil, nil
	}

	return s.getRegistration(ctx, userID)
}

func (s viewerCreatorRegistrationServiceStub) SaveIntake(ctx context.Context, input creatorregistration.SaveIntakeInput) (creatorregistration.Intake, error) {
	if s.saveIntake == nil {
		return creatorregistration.Intake{}, nil
	}

	return s.saveIntake(ctx, input)
}

func (s viewerCreatorRegistrationServiceStub) Submit(ctx context.Context, userID uuid.UUID) (creatorregistration.Registration, error) {
	if s.submit == nil {
		return creatorregistration.Registration{}, nil
	}

	return s.submit(ctx, userID)
}

type viewerCreatorRegistrationEvidenceUploadHandlerStub struct {
	completeUpload func(context.Context, creatorregistration.CompleteEvidenceUploadInput) (creatorregistration.CompleteEvidenceUploadResult, error)
	createUpload   func(context.Context, creatorregistration.CreateEvidenceUploadInput) (creatorregistration.CreateEvidenceUploadResult, error)
}

func (s viewerCreatorRegistrationEvidenceUploadHandlerStub) CompleteUpload(
	ctx context.Context,
	input creatorregistration.CompleteEvidenceUploadInput,
) (creatorregistration.CompleteEvidenceUploadResult, error) {
	return s.completeUpload(ctx, input)
}

func (s viewerCreatorRegistrationEvidenceUploadHandlerStub) CreateUpload(
	ctx context.Context,
	input creatorregistration.CreateEvidenceUploadInput,
) (creatorregistration.CreateEvidenceUploadResult, error) {
	return s.createUpload(ctx, input)
}

type viewerCreatorAvatarUploadHandlerStub struct {
	completeUpload         func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error)
	consumeCompletedUpload func(context.Context, uuid.UUID, string) error
	createUpload           func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error)
	resolveCompletedUpload func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error)
}

func (s viewerCreatorAvatarUploadHandlerStub) CompleteUpload(
	ctx context.Context,
	input creatoravatar.CompleteUploadInput,
) (creatoravatar.CompleteUploadResult, error) {
	return s.completeUpload(ctx, input)
}

func (s viewerCreatorAvatarUploadHandlerStub) ConsumeCompletedUpload(
	ctx context.Context,
	viewerUserID uuid.UUID,
	avatarUploadToken string,
) error {
	return s.consumeCompletedUpload(ctx, viewerUserID, avatarUploadToken)
}

func (s viewerCreatorAvatarUploadHandlerStub) CreateUpload(
	ctx context.Context,
	input creatoravatar.CreateUploadInput,
) (creatoravatar.CreateUploadResult, error) {
	return s.createUpload(ctx, input)
}

func (s viewerCreatorAvatarUploadHandlerStub) ResolveCompletedUpload(
	ctx context.Context,
	viewerUserID uuid.UUID,
	avatarUploadToken string,
) (creatoravatar.CompletedUpload, error) {
	return s.resolveCompletedUpload(ctx, viewerUserID, avatarUploadToken)
}

type viewerActiveModeSwitcherStub struct {
	switchActiveMode func(context.Context, string, auth.ActiveMode) error
}

func (s viewerActiveModeSwitcherStub) SwitchActiveMode(
	ctx context.Context,
	rawSessionToken string,
	activeMode auth.ActiveMode,
) error {
	return s.switchActiveMode(ctx, rawSessionToken, activeMode)
}

type viewerProfileReaderStub struct {
	getProfile func(context.Context, uuid.UUID) (viewerprofile.Profile, error)
}

func (s viewerProfileReaderStub) GetProfile(ctx context.Context, userID uuid.UUID) (viewerprofile.Profile, error) {
	return s.getProfile(ctx, userID)
}

type viewerProfileWriterStub struct {
	updateProfile func(context.Context, viewerprofile.UpdateProfileInput) (viewerprofile.Profile, error)
}

func (s viewerProfileWriterStub) UpdateProfile(
	ctx context.Context,
	input viewerprofile.UpdateProfileInput,
) (viewerprofile.Profile, error) {
	return s.updateProfile(ctx, input)
}

func TestViewerCreatorRegistrationRequiresAuth(t *testing.T) {
	t.Parallel()

	submitCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationServiceStub{
			submit: func(context.Context, uuid.UUID) (creatorregistration.Registration, error) {
				submitCalled = true
				return creatorregistration.Registration{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/viewer/creator-registration", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if submitCalled {
		t.Fatal("POST /api/viewer/creator-registration submitCalled = true, want false")
	}
}

func TestViewerCreatorRegistrationSubmitUsesCurrentViewer(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var gotUserID uuid.UUID

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         viewerID,
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationServiceStub{
			submit: func(_ context.Context, userID uuid.UUID) (creatorregistration.Registration, error) {
				gotUserID = userID
				return creatorregistration.Registration{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/viewer/creator-registration", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusNoContent)
	}
	if gotUserID != viewerID {
		t.Fatalf("POST /api/viewer/creator-registration user id got %s want %s", gotUserID, viewerID)
	}
}

func TestViewerCreatorRegistrationRejectsUnexpectedInput(t *testing.T) {
	t.Parallel()

	submitCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         uuid.New(),
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationServiceStub{
			submit: func(context.Context, uuid.UUID) (creatorregistration.Registration, error) {
				submitCalled = true
				return creatorregistration.Registration{}, nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/creator-registration",
		bytes.NewBufferString(`{"displayName":"Mina"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if submitCalled {
		t.Fatal("POST /api/viewer/creator-registration submitCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("POST /api/viewer/creator-registration body got %q want invalid_request", rec.Body.String())
	}
}

func TestViewerCreatorRegistrationMapsIncompleteConflict(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         uuid.New(),
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationServiceStub{
			submit: func(context.Context, uuid.UUID) (creatorregistration.Registration, error) {
				return creatorregistration.Registration{}, creatorregistration.ErrRegistrationIncomplete
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/viewer/creator-registration", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("POST /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusConflict)
	}
	if !strings.Contains(rec.Body.String(), `"code":"registration_incomplete"`) {
		t.Fatalf("POST /api/viewer/creator-registration body got %q want registration_incomplete", rec.Body.String())
	}
}

func TestViewerCreatorRegistrationIntakePutPassesNormalizedPayload(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var gotInput creatorregistration.SaveIntakeInput

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         viewerID,
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationServiceStub{
			saveIntake: func(_ context.Context, input creatorregistration.SaveIntakeInput) (creatorregistration.Intake, error) {
				gotInput = input
				return creatorregistration.Intake{
					AcceptsConsentResponsibility: true,
					BirthDate:                    "1999-04-02",
					CanSubmit:                    false,
					CreatorBio:                   input.CreatorBio,
					DeclaresNoProhibitedCategory: true,
					Evidences:                    []creatorregistration.Evidence{},
					IsReadOnly:                   false,
					LegalName:                    input.LegalName,
					PayoutRecipientName:          input.PayoutRecipientName,
					PayoutRecipientType:          input.PayoutRecipientType,
					SharedProfile: creatorregistration.SharedProfilePreview{
						DisplayName: "Mina",
						Handle:      "mina",
						UserID:      viewerID,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/creator-registration/intake",
		bytes.NewBufferString(`{"creatorBio":"quiet rooftop","legalName":"Mina Rei","birthDate":"1999-04-02","payoutRecipientType":"self","payoutRecipientName":"Mina Rei","declaresNoProhibitedCategory":true,"acceptsConsentResponsibility":true}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /api/viewer/creator-registration/intake status got %d want %d", rec.Code, http.StatusOK)
	}
	if gotInput.UserID != viewerID {
		t.Fatalf("SaveIntake() user id got %s want %s", gotInput.UserID, viewerID)
	}
	if gotInput.CreatorBio != "quiet rooftop" {
		t.Fatalf("SaveIntake() creator bio got %q want %q", gotInput.CreatorBio, "quiet rooftop")
	}
	if !strings.Contains(rec.Body.String(), `"creatorBio":"quiet rooftop"`) {
		t.Fatalf("PUT /api/viewer/creator-registration/intake body got %q want creatorBio", rec.Body.String())
	}
}

func TestViewerCreatorRegistrationGetReturnsCurrentStatus(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	submittedAt := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	avatarURL := "https://cdn.example.com/avatars/mina.jpg"

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         viewerID,
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistration: viewerCreatorRegistrationServiceStub{
			getRegistration: func(context.Context, uuid.UUID) (*creatorregistration.Registration, error) {
				return &creatorregistration.Registration{
					Actions: creatorregistration.Actions{
						CanEnterCreatorMode: false,
						CanResubmit:         false,
						CanSubmit:           false,
					},
					CreatorDraft: creatorregistration.CreatorDraft{Bio: "quiet rooftop"},
					Review:       creatorregistration.ReviewTimeline{SubmittedAt: &submittedAt},
					SharedProfile: creatorregistration.SharedProfilePreview{
						AvatarURL:   &avatarURL,
						DisplayName: "Mina",
						Handle:      "mina",
						UserID:      viewerID,
					},
					State: "submitted",
					Surface: creatorregistration.Surface{
						Kind: "read_only_onboarding",
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/viewer/creator-registration", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/viewer/creator-registration status got %d want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"state":"submitted"`) {
		t.Fatalf("GET /api/viewer/creator-registration body got %q want submitted state", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"id":"asset_viewer_profile_avatar_11111111111111111111111111111111"`) {
		t.Fatalf("GET /api/viewer/creator-registration body got %q want stable avatar id", rec.Body.String())
	}
}

func TestViewerCreatorRegistrationEvidenceUploadCreateRejectsReadOnlyState(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         viewerID,
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistrationEvidence: viewerCreatorRegistrationEvidenceUploadHandlerStub{
			createUpload: func(context.Context, creatorregistration.CreateEvidenceUploadInput) (creatorregistration.CreateEvidenceUploadResult, error) {
				return creatorregistration.CreateEvidenceUploadResult{}, creatorregistration.ErrRegistrationStateConflict
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/creator-registration/evidence-uploads",
		bytes.NewBufferString(`{"kind":"government_id","fileName":"government-id.png","mimeType":"image/png","fileSizeBytes":128}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("POST /api/viewer/creator-registration/evidence-uploads status got %d want %d", rec.Code, http.StatusConflict)
	}
	if !strings.Contains(rec.Body.String(), `"code":"registration_state_conflict"`) {
		t.Fatalf("POST /api/viewer/creator-registration/evidence-uploads body got %q want registration_state_conflict", rec.Body.String())
	}
}

func TestViewerCreatorRegistrationEvidenceUploadCompleteReturnsEvidence(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	uploadedAt := time.Date(2026, 4, 17, 10, 30, 0, 0, time.UTC)

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         viewerID,
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorRegistrationEvidence: viewerCreatorRegistrationEvidenceUploadHandlerStub{
			completeUpload: func(_ context.Context, input creatorregistration.CompleteEvidenceUploadInput) (creatorregistration.CompleteEvidenceUploadResult, error) {
				if input.ViewerUserID != viewerID {
					t.Fatalf("CompleteUpload() viewer id got %s want %s", input.ViewerUserID, viewerID)
				}
				return creatorregistration.CompleteEvidenceUploadResult{
					Evidence: creatorregistration.Evidence{
						FileName:      "government-id.png",
						FileSizeBytes: 128,
						Kind:          creatorregistration.EvidenceKindGovernmentID,
						MimeType:      "image/png",
						UploadedAt:    uploadedAt,
					},
					EvidenceKind:        creatorregistration.EvidenceKindGovernmentID,
					EvidenceUploadToken: input.EvidenceUploadToken,
				}, nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/creator-registration/evidence-uploads/complete",
		bytes.NewBufferString(`{"evidenceUploadToken":"vcevd_123"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/viewer/creator-registration/evidence-uploads/complete status got %d want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"evidenceKind":"government_id"`) {
		t.Fatalf("POST /api/viewer/creator-registration/evidence-uploads/complete body got %q want government_id", rec.Body.String())
	}
}

func TestViewerActiveModeSwitchSuccess(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var gotRawToken string
	var gotActiveMode auth.ActiveMode

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		ViewerActiveMode: viewerActiveModeSwitcherStub{
			switchActiveMode: func(_ context.Context, rawSessionToken string, activeMode auth.ActiveMode) error {
				gotRawToken = rawSessionToken
				gotActiveMode = activeMode
				return nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/viewer/active-mode", bytes.NewBufferString(`{"activeMode":"creator"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT /api/viewer/active-mode status got %d want %d", rec.Code, http.StatusNoContent)
	}
	if gotRawToken != "raw-session-token" {
		t.Fatalf("PUT /api/viewer/active-mode raw token got %q want %q", gotRawToken, "raw-session-token")
	}
	if gotActiveMode != auth.ActiveModeCreator {
		t.Fatalf("PUT /api/viewer/active-mode active mode got %q want %q", gotActiveMode, auth.ActiveModeCreator)
	}
}

func TestViewerActiveModeSwitchRejectsUnavailableCreatorMode(t *testing.T) {
	t.Parallel()

	switcherCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
		ViewerActiveMode: viewerActiveModeSwitcherStub{
			switchActiveMode: func(context.Context, string, auth.ActiveMode) error {
				switcherCalled = true
				return nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/viewer/active-mode", bytes.NewBufferString(`{"activeMode":"creator"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("PUT /api/viewer/active-mode status got %d want %d", rec.Code, http.StatusForbidden)
	}
	if switcherCalled {
		t.Fatal("PUT /api/viewer/active-mode switcherCalled = true, want false")
	}

	var response responseEnvelope[json.RawMessage]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "creator_mode_unavailable" {
		t.Fatalf("PUT /api/viewer/active-mode error got %#v want creator_mode_unavailable", response.Error)
	}
}

func TestViewerActiveModeSwitchRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	switcherCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		ViewerActiveMode: viewerActiveModeSwitcherStub{
			switchActiveMode: func(context.Context, string, auth.ActiveMode) error {
				switcherCalled = true
				return nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/active-mode",
		bytes.NewBufferString(`{"activeMode":"creator"}{"extra":true}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /api/viewer/active-mode status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if switcherCalled {
		t.Fatal("PUT /api/viewer/active-mode switcherCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("PUT /api/viewer/active-mode body got %q want invalid_request", rec.Body.String())
	}
}

func TestViewerActiveModeSwitchRejectsUnknownField(t *testing.T) {
	t.Parallel()

	switcherCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		ViewerActiveMode: viewerActiveModeSwitcherStub{
			switchActiveMode: func(context.Context, string, auth.ActiveMode) error {
				switcherCalled = true
				return nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/active-mode",
		bytes.NewBufferString(`{"activeMode":"fan","unexpected":true}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /api/viewer/active-mode status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if switcherCalled {
		t.Fatal("PUT /api/viewer/active-mode switcherCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("PUT /api/viewer/active-mode body got %q want invalid_request", rec.Body.String())
	}
}

func TestViewerActiveModeSwitchRejectsInvalidMode(t *testing.T) {
	t.Parallel()

	switcherCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		ViewerActiveMode: viewerActiveModeSwitcherStub{
			switchActiveMode: func(context.Context, string, auth.ActiveMode) error {
				switcherCalled = true
				return nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/viewer/active-mode", bytes.NewBufferString(`{"activeMode":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /api/viewer/active-mode status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if switcherCalled {
		t.Fatal("PUT /api/viewer/active-mode switcherCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_active_mode"`) {
		t.Fatalf("PUT /api/viewer/active-mode body got %q want invalid_active_mode", rec.Body.String())
	}
}

func TestViewerActiveModeSwitchMapsInvalidModeFromSwitcher(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						ActiveMode:           auth.ActiveModeCreator,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		ViewerActiveMode: viewerActiveModeSwitcherStub{
			switchActiveMode: func(context.Context, string, auth.ActiveMode) error {
				return auth.ErrInvalidActiveMode
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/viewer/active-mode", bytes.NewBufferString(`{"activeMode":"fan"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /api/viewer/active-mode status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_active_mode"`) {
		t.Fatalf("PUT /api/viewer/active-mode body got %q want invalid_active_mode", rec.Body.String())
	}
}

func TestViewerActiveModeSwitchReturnsInternalErrorWhenCookieMissing(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/viewer/active-mode", bytes.NewBufferString(`{"activeMode":"fan"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set(authenticatedViewerContextKey, auth.CurrentViewer{
		ID:                   uuid.New(),
		ActiveMode:           auth.ActiveModeCreator,
		CanAccessCreatorMode: true,
	})

	handleViewerActiveModeSwitch(ctx, viewerActiveModeSwitcherStub{
		switchActiveMode: func(context.Context, string, auth.ActiveMode) error {
			t.Fatal("SwitchActiveMode() should not be called")
			return nil
		},
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("handleViewerActiveModeSwitch() status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}
