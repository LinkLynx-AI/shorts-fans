package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatorupload"
)

type creatorUploadServiceStub struct {
	createPackage   func(context.Context, creatorupload.CreatePackageInput) (creatorupload.CreatePackageResult, error)
	completePackage func(context.Context, creatorupload.CompletePackageInput) (creatorupload.CompletePackageResult, error)
}

func (s creatorUploadServiceStub) CreatePackage(ctx context.Context, input creatorupload.CreatePackageInput) (creatorupload.CreatePackageResult, error) {
	return s.createPackage(ctx, input)
}

func (s creatorUploadServiceStub) CompletePackage(ctx context.Context, input creatorupload.CompletePackageInput) (creatorupload.CompletePackageResult, error) {
	return s.completePackage(ctx, input)
}

func TestCreatorUploadCreateRequiresAuth(t *testing.T) {
	t.Parallel()

	called := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) { return auth.Bootstrap{}, nil },
		},
		CreatorUpload: creatorUploadServiceStub{
			createPackage: func(context.Context, creatorupload.CreatePackageInput) (creatorupload.CreatePackageResult, error) {
				called = true
				return creatorupload.CreatePackageResult{}, nil
			},
			completePackage: func(context.Context, creatorupload.CompletePackageInput) (creatorupload.CompletePackageResult, error) {
				return creatorupload.CompletePackageResult{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/creator/upload-packages", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST /api/creator/upload-packages status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if called {
		t.Fatal("POST /api/creator/upload-packages called = true, want false")
	}
}

func TestCreatorUploadCreateRejectsMissingCapability(t *testing.T) {
	t.Parallel()

	called := false
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
		CreatorUpload: creatorUploadServiceStub{
			createPackage: func(context.Context, creatorupload.CreatePackageInput) (creatorupload.CreatePackageResult, error) {
				called = true
				return creatorupload.CreatePackageResult{}, nil
			},
			completePackage: func(context.Context, creatorupload.CompletePackageInput) (creatorupload.CompletePackageResult, error) {
				return creatorupload.CompletePackageResult{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/creator/upload-packages", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST /api/creator/upload-packages status got %d want %d", rec.Code, http.StatusForbidden)
	}
	if called {
		t.Fatal("POST /api/creator/upload-packages called = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"capability_required"`) {
		t.Fatalf("POST /api/creator/upload-packages body got %q want capability_required", rec.Body.String())
	}
}

func TestCreatorUploadCreateSuccess(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var gotInput creatorupload.CreatePackageInput
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
		CreatorUpload: creatorUploadServiceStub{
			createPackage: func(_ context.Context, input creatorupload.CreatePackageInput) (creatorupload.CreatePackageResult, error) {
				gotInput = input
				return creatorupload.CreatePackageResult{
					ExpiresAt:    time.Unix(1710000000, 0).UTC(),
					PackageToken: "pkg-token",
					UploadTargets: creatorupload.UploadTargetSet{
						Main: creatorupload.UploadTarget{
							FileName:      "main.mp4",
							MimeType:      "video/mp4",
							Role:          "main",
							UploadEntryID: "main-entry",
							Upload: creatorupload.DirectUpload{
								Method: "PUT",
								URL:    "https://signed.example.com/main",
								Headers: map[string]string{
									"Content-Type": "video/mp4",
								},
							},
						},
						Shorts: []creatorupload.UploadTarget{{
							FileName:      "short.mp4",
							MimeType:      "video/mp4",
							Role:          "short",
							UploadEntryID: "short-entry",
							Upload: creatorupload.DirectUpload{
								Method: "PUT",
								URL:    "https://signed.example.com/short",
								Headers: map[string]string{
									"Content-Type": "video/mp4",
								},
							},
						}},
					},
				}, nil
			},
			completePackage: func(context.Context, creatorupload.CompletePackageInput) (creatorupload.CompletePackageResult, error) {
				return creatorupload.CompletePackageResult{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/creator/upload-packages", bytes.NewBufferString(`{"main":{"fileName":"main.mp4","mimeType":"video/mp4","fileSizeBytes":100},"shorts":[{"fileName":"short.mp4","mimeType":"video/mp4","fileSizeBytes":10}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/creator/upload-packages status got %d want %d", rec.Code, http.StatusOK)
	}
	if gotInput.CreatorUserID != viewerID {
		t.Fatalf("POST /api/creator/upload-packages creator id got %s want %s", gotInput.CreatorUserID, viewerID)
	}

	var response responseEnvelope[creatorUploadCreateResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.PackageToken != "pkg-token" {
		t.Fatalf("POST /api/creator/upload-packages response got %#v want package token", response.Data)
	}
}

func TestCreatorUploadCreateMapsValidationAndInvalidRequest(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		CreatorUpload: creatorUploadServiceStub{
			createPackage: func(context.Context, creatorupload.CreatePackageInput) (creatorupload.CreatePackageResult, error) {
				return creatorupload.CreatePackageResult{}, &creatorupload.ValidationError{}
			},
			completePackage: func(context.Context, creatorupload.CompletePackageInput) (creatorupload.CompletePackageResult, error) {
				return creatorupload.CompletePackageResult{}, nil
			},
		},
	})

	invalidReq := httptest.NewRequest(http.MethodPost, "/api/creator/upload-packages", bytes.NewBufferString(`{"main":{},"extra":true}`))
	invalidReq.Header.Set("Content-Type", "application/json")
	invalidReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	invalidRec := httptest.NewRecorder()
	router.ServeHTTP(invalidRec, invalidReq)
	if invalidRec.Code != http.StatusBadRequest {
		t.Fatalf("invalid request status got %d want %d", invalidRec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(invalidRec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("invalid request body got %q want invalid_request", invalidRec.Body.String())
	}

	serviceRouter := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		CreatorUpload: creatorUploadServiceStub{
			createPackage: func(context.Context, creatorupload.CreatePackageInput) (creatorupload.CreatePackageResult, error) {
				return creatorupload.CreatePackageResult{}, creatorupload.NewValidationError("main is required")
			},
			completePackage: func(context.Context, creatorupload.CompletePackageInput) (creatorupload.CompletePackageResult, error) {
				return creatorupload.CompletePackageResult{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/creator/upload-packages", bytes.NewBufferString(`{"shorts":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()
	serviceRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("validation request status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"validation_error"`) {
		t.Fatalf("validation request body got %q want validation_error", rec.Body.String())
	}
}

func TestCreatorUploadCompleteMapsContractErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		wantCode int
		wantBody string
	}{
		{
			name:     "upload expired",
			err:      creatorupload.ErrPackageExpired,
			wantCode: http.StatusConflict,
			wantBody: `"code":"upload_expired"`,
		},
		{
			name:     "upload failure",
			err:      creatorupload.ErrUploadFailure,
			wantCode: http.StatusUnprocessableEntity,
			wantBody: `"code":"upload_failure"`,
		},
		{
			name:     "storage failure",
			err:      creatorupload.ErrStorageFailure,
			wantCode: http.StatusServiceUnavailable,
			wantBody: `"code":"storage_failure"`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewHandler(HandlerConfig{
				ViewerBootstrap: viewerBootstrapReaderStub{
					readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
						return auth.Bootstrap{
							CurrentViewer: &auth.CurrentViewer{
								ID:                   uuid.New(),
								CanAccessCreatorMode: true,
							},
						}, nil
					},
				},
				CreatorUpload: creatorUploadServiceStub{
					createPackage: func(context.Context, creatorupload.CreatePackageInput) (creatorupload.CreatePackageResult, error) {
						return creatorupload.CreatePackageResult{}, nil
					},
					completePackage: func(context.Context, creatorupload.CompletePackageInput) (creatorupload.CompletePackageResult, error) {
						return creatorupload.CompletePackageResult{}, tt.err
					},
				},
			})

			req := httptest.NewRequest(http.MethodPost, "/api/creator/upload-packages/complete", bytes.NewBufferString(`{"packageToken":"pkg-token","main":{"uploadEntryId":"main-entry"},"shorts":[{"uploadEntryId":"short-entry"}]}`))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("POST /api/creator/upload-packages/complete status got %d want %d", rec.Code, tt.wantCode)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Fatalf("POST /api/creator/upload-packages/complete body got %q want %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestCreatorUploadCompleteSuccess(t *testing.T) {
	t.Parallel()

	mainID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	mainAssetID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	shortID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	shortAssetID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		CreatorUpload: creatorUploadServiceStub{
			createPackage: func(context.Context, creatorupload.CreatePackageInput) (creatorupload.CreatePackageResult, error) {
				return creatorupload.CreatePackageResult{}, nil
			},
			completePackage: func(context.Context, creatorupload.CompletePackageInput) (creatorupload.CompletePackageResult, error) {
				return creatorupload.CompletePackageResult{
					Main: creatorupload.CreatedMain{
						ID:    mainID,
						State: "draft",
						MediaAsset: creatorupload.CreatedMediaAsset{
							ID:              mainAssetID,
							MimeType:        "video/mp4",
							ProcessingState: "uploaded",
						},
					},
					Shorts: []creatorupload.CreatedShort{{
						ID:              shortID,
						CanonicalMainID: mainID,
						State:           "draft",
						MediaAsset: creatorupload.CreatedMediaAsset{
							ID:              shortAssetID,
							MimeType:        "video/mp4",
							ProcessingState: "uploaded",
						},
					}},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/creator/upload-packages/complete", bytes.NewBufferString(`{"packageToken":"pkg-token","main":{"uploadEntryId":"main-entry"},"shorts":[{"uploadEntryId":"short-entry"}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/creator/upload-packages/complete status got %d want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), mainID.String()) {
		t.Fatalf("POST /api/creator/upload-packages/complete body got %q want main id", rec.Body.String())
	}
}

func TestCreatorUploadCompleteInvalidRequest(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   uuid.New(),
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		CreatorUpload: creatorUploadServiceStub{
			createPackage: func(context.Context, creatorupload.CreatePackageInput) (creatorupload.CreatePackageResult, error) {
				return creatorupload.CreatePackageResult{}, nil
			},
			completePackage: func(context.Context, creatorupload.CompletePackageInput) (creatorupload.CompletePackageResult, error) {
				return creatorupload.CompletePackageResult{}, errors.New("should not be called")
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/creator/upload-packages/complete", bytes.NewBufferString(`{"packageToken":"pkg-token","unknown":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/creator/upload-packages/complete status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("POST /api/creator/upload-packages/complete body got %q want invalid_request", rec.Body.String())
	}
}
