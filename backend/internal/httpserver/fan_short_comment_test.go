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

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shortcomment"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/google/uuid"
)

type stubFanShortCommentReader struct {
	listComments func(context.Context, uuid.UUID, *shortcomment.Cursor, int) ([]shortcomment.Comment, *shortcomment.Cursor, error)
}

func (s stubFanShortCommentReader) ListComments(ctx context.Context, shortID uuid.UUID, cursor *shortcomment.Cursor, limit int) ([]shortcomment.Comment, *shortcomment.Cursor, error) {
	return s.listComments(ctx, shortID, cursor, limit)
}

type stubFanShortCommentWriter struct {
	createComment func(context.Context, shortcomment.CreateInput) (shortcomment.Comment, error)
}

func (s stubFanShortCommentWriter) CreateComment(ctx context.Context, input shortcomment.CreateInput) (shortcomment.Comment, error) {
	return s.createComment(ctx, input)
}

func TestFanShortCommentListRoute(t *testing.T) {
	t.Parallel()

	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	commentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	nextCommentID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	authorID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	createdAt := time.Date(2026, 1, 2, 14, 5, 0, 0, time.UTC)
	nextCreatedAt := createdAt.Add(-time.Minute)
	avatarURL := "https://cdn.example.com/avatar.jpg"

	router := NewHandler(HandlerConfig{
		FanShortComments: stubFanShortCommentReader{
			listComments: func(_ context.Context, gotShortID uuid.UUID, cursor *shortcomment.Cursor, limit int) ([]shortcomment.Comment, *shortcomment.Cursor, error) {
				if gotShortID != shortID {
					t.Fatalf("ListComments() short id got %s want %s", gotShortID, shortID)
				}
				if cursor != nil {
					t.Fatalf("ListComments() cursor got %#v want nil", cursor)
				}
				if limit != shortcomment.DefaultPageSize {
					t.Fatalf("ListComments() limit got %d want %d", limit, shortcomment.DefaultPageSize)
				}

				return []shortcomment.Comment{
					{
						Author: shortcomment.Author{
							AvatarURL:   &avatarURL,
							DisplayName: "Kana Mori",
							Handle:      "kanamori",
							ID:          authorID,
						},
						Body:      "hello",
						CreatedAt: createdAt,
						ID:        commentID,
						ShortID:   shortID,
					},
				}, &shortcomment.Cursor{CommentID: nextCommentID, CreatedAt: nextCreatedAt}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/shorts/"+shorts.FormatPublicShortID(shortID)+"/comments", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/shorts/{shortId}/comments status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[shortCommentListResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || len(response.Data.Items) != 1 {
		t.Fatalf("response.Data got %#v want one comment", response.Data)
	}
	if response.Data.Items[0].ID != shortCommentPublicID(commentID) {
		t.Fatalf("comment id got %q want %q", response.Data.Items[0].ID, shortCommentPublicID(commentID))
	}
	if response.Data.Items[0].ShortID != shortPublicID(shortID) {
		t.Fatalf("comment short id got %q want %q", response.Data.Items[0].ShortID, shortPublicID(shortID))
	}
	if response.Data.Items[0].Author.Handle != "@kanamori" {
		t.Fatalf("comment author handle got %q want @kanamori", response.Data.Items[0].Author.Handle)
	}
	if response.Data.Items[0].Author.Avatar == nil || response.Data.Items[0].Author.Avatar.Kind != "image" {
		t.Fatalf("comment author avatar got %#v want image asset", response.Data.Items[0].Author.Avatar)
	}
	if strings.Contains(rec.Body.String(), authorID.String()) || strings.Contains(rec.Body.String(), strings.ReplaceAll(authorID.String(), "-", "")) {
		t.Fatalf("comment response leaked raw author id in body %q", rec.Body.String())
	}
	if response.Meta.Page == nil || !response.Meta.Page.HasNext || response.Meta.Page.NextCursor == nil {
		t.Fatalf("response.Meta.Page got %#v want next cursor", response.Meta.Page)
	}
	decodedCursor, ok := decodeShortCommentCursor(*response.Meta.Page.NextCursor, shortID)
	if !ok {
		t.Fatalf("decodeShortCommentCursor(%q) ok = false, want true", *response.Meta.Page.NextCursor)
	}
	if decodedCursor.CommentID != nextCommentID {
		t.Fatalf("next cursor comment id got %s want %s", decodedCursor.CommentID, nextCommentID)
	}
}

func TestFanShortCommentListRouteForwardsValidCursor(t *testing.T) {
	t.Parallel()

	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cursorCommentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	cursorCreatedAt := time.Date(2026, 1, 2, 14, 5, 0, 0, time.UTC)
	cursor := encodeShortCommentCursor(shortID, &shortcomment.Cursor{
		CommentID: cursorCommentID,
		CreatedAt: cursorCreatedAt,
	})
	if cursor == nil {
		t.Fatal("encodeShortCommentCursor() = nil, want cursor")
	}
	readerCalled := false
	router := NewHandler(HandlerConfig{
		FanShortComments: stubFanShortCommentReader{
			listComments: func(_ context.Context, gotShortID uuid.UUID, gotCursor *shortcomment.Cursor, limit int) ([]shortcomment.Comment, *shortcomment.Cursor, error) {
				readerCalled = true
				if gotShortID != shortID {
					t.Fatalf("ListComments() short id got %s want %s", gotShortID, shortID)
				}
				if gotCursor == nil {
					t.Fatal("ListComments() cursor = nil, want decoded cursor")
				}
				if gotCursor.CommentID != cursorCommentID {
					t.Fatalf("ListComments() cursor comment id got %s want %s", gotCursor.CommentID, cursorCommentID)
				}
				if !gotCursor.CreatedAt.Equal(cursorCreatedAt) {
					t.Fatalf("ListComments() cursor created_at got %s want %s", gotCursor.CreatedAt, cursorCreatedAt)
				}
				if limit != shortcomment.DefaultPageSize {
					t.Fatalf("ListComments() limit got %d want %d", limit, shortcomment.DefaultPageSize)
				}

				return []shortcomment.Comment{}, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/shorts/"+shorts.FormatPublicShortID(shortID)+"/comments?cursor="+*cursor, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/shorts/{shortId}/comments status got %d want %d", rec.Code, http.StatusOK)
	}
	if !readerCalled {
		t.Fatal("GET /api/fan/shorts/{shortId}/comments readerCalled = false, want true")
	}
}

func TestFanShortCommentListRouteRejectsInvalidCursor(t *testing.T) {
	t.Parallel()

	readerCalled := false
	router := NewHandler(HandlerConfig{
		FanShortComments: stubFanShortCommentReader{
			listComments: func(context.Context, uuid.UUID, *shortcomment.Cursor, int) ([]shortcomment.Comment, *shortcomment.Cursor, error) {
				readerCalled = true
				return nil, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/shorts/short_11111111111111111111111111111111/comments?cursor=***", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("GET /api/fan/shorts/{shortId}/comments status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if readerCalled {
		t.Fatal("GET /api/fan/shorts/{shortId}/comments readerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("GET /api/fan/shorts/{shortId}/comments body got %q want invalid_request", rec.Body.String())
	}
}

func TestFanShortCommentListRouteRejectsOversizedCursor(t *testing.T) {
	t.Parallel()

	readerCalled := false
	router := NewHandler(HandlerConfig{
		FanShortComments: stubFanShortCommentReader{
			listComments: func(context.Context, uuid.UUID, *shortcomment.Cursor, int) ([]shortcomment.Comment, *shortcomment.Cursor, error) {
				readerCalled = true
				return nil, nil, nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/fan/shorts/short_11111111111111111111111111111111/comments?cursor="+strings.Repeat("a", maxFanShortCommentCursorLength+1),
		nil,
	)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("GET /api/fan/shorts/{shortId}/comments status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if readerCalled {
		t.Fatal("GET /api/fan/shorts/{shortId}/comments readerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("GET /api/fan/shorts/{shortId}/comments body got %q want invalid_request", rec.Body.String())
	}
}

func TestFanShortCommentListRouteRejectsCursorForDifferentShort(t *testing.T) {
	t.Parallel()

	routeShortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cursorShortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	cursorCommentID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	cursorCreatedAt := time.Date(2026, 1, 2, 14, 5, 0, 0, time.UTC)
	readerCalled := false
	cursor := encodeShortCommentCursor(cursorShortID, &shortcomment.Cursor{
		CommentID: cursorCommentID,
		CreatedAt: cursorCreatedAt,
	})
	if cursor == nil {
		t.Fatal("encodeShortCommentCursor() = nil, want cursor")
	}
	router := NewHandler(HandlerConfig{
		FanShortComments: stubFanShortCommentReader{
			listComments: func(context.Context, uuid.UUID, *shortcomment.Cursor, int) ([]shortcomment.Comment, *shortcomment.Cursor, error) {
				readerCalled = true
				return nil, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/shorts/"+shorts.FormatPublicShortID(routeShortID)+"/comments?cursor="+*cursor, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("GET /api/fan/shorts/{shortId}/comments status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if readerCalled {
		t.Fatal("GET /api/fan/shorts/{shortId}/comments readerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("GET /api/fan/shorts/{shortId}/comments body got %q want invalid_request", rec.Body.String())
	}
}

func TestFanShortCommentCreateRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	commentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	createdAt := time.Date(2026, 1, 2, 14, 5, 0, 0, time.UTC)

	router := NewHandler(HandlerConfig{
		FanShortCommentWriter: stubFanShortCommentWriter{
			createComment: func(_ context.Context, input shortcomment.CreateInput) (shortcomment.Comment, error) {
				if input.AuthorUserID != viewerID {
					t.Fatalf("CreateComment() viewer id got %s want %s", input.AuthorUserID, viewerID)
				}
				if input.ShortID != shortID {
					t.Fatalf("CreateComment() short id got %s want %s", input.ShortID, shortID)
				}
				if input.Body != "hello" {
					t.Fatalf("CreateComment() body got %q want hello", input.Body)
				}

				return shortcomment.Comment{
					Author: shortcomment.Author{
						DisplayName: "Kana Mori",
						Handle:      "kanamori",
						ID:          viewerID,
					},
					Body:      "hello",
					CreatedAt: createdAt,
					ID:        commentID,
					ShortID:   shortID,
				}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/shorts/"+shorts.FormatPublicShortID(shortID)+"/comments", bytes.NewBufferString(`{"body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /api/fan/shorts/{shortId}/comments status got %d want %d", rec.Code, http.StatusCreated)
	}

	var response responseEnvelope[shortCommentCreateResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Comment.ID != shortCommentPublicID(commentID) {
		t.Fatalf("response.Data got %#v want created comment", response.Data)
	}
	if response.Meta.Page != nil {
		t.Fatalf("response.Meta.Page got %#v want nil", response.Meta.Page)
	}
}

func TestFanShortCommentCreateRouteRejectsUnauthenticatedRequest(t *testing.T) {
	t.Parallel()

	writerCalled := false
	router := NewHandler(HandlerConfig{
		FanShortCommentWriter: stubFanShortCommentWriter{
			createComment: func(context.Context, shortcomment.CreateInput) (shortcomment.Comment, error) {
				writerCalled = true
				return shortcomment.Comment{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/shorts/short_11111111111111111111111111111111/comments", bytes.NewBufferString(`{"body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST /api/fan/shorts/{shortId}/comments status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if writerCalled {
		t.Fatal("POST /api/fan/shorts/{shortId}/comments writerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"auth_required"`) {
		t.Fatalf("POST /api/fan/shorts/{shortId}/comments body got %q want auth_required", rec.Body.String())
	}
}

func TestFanShortCommentCreateRouteReturnsNotFound(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	writerCalled := false
	router := NewHandler(HandlerConfig{
		FanShortCommentWriter: stubFanShortCommentWriter{
			createComment: func(context.Context, shortcomment.CreateInput) (shortcomment.Comment, error) {
				writerCalled = true
				return shortcomment.Comment{}, shortcomment.ErrShortNotFound
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/shorts/short_11111111111111111111111111111111/comments", bytes.NewBufferString(`{"body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("POST /api/fan/shorts/{shortId}/comments status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !writerCalled {
		t.Fatal("POST /api/fan/shorts/{shortId}/comments writerCalled = false, want true")
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("POST /api/fan/shorts/{shortId}/comments body got %q want not_found", rec.Body.String())
	}
}

func TestFanShortCommentCreateRouteRejectsMalformedJSONBeforeWriter(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	writerCalled := false
	router := NewHandler(HandlerConfig{
		FanShortCommentWriter: stubFanShortCommentWriter{
			createComment: func(context.Context, shortcomment.CreateInput) (shortcomment.Comment, error) {
				writerCalled = true
				return shortcomment.Comment{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fan/shorts/short_11111111111111111111111111111111/comments", bytes.NewBufferString(`{"body":`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/fan/shorts/{shortId}/comments status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if writerCalled {
		t.Fatal("POST /api/fan/shorts/{shortId}/comments writerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("POST /api/fan/shorts/{shortId}/comments body got %q want invalid_request", rec.Body.String())
	}
}

func TestFanShortCommentCreateRouteClassifiesInvalidBodyAsValidationError(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	tests := []struct {
		name             string
		body             string
		wantWriterCalled bool
	}{
		{
			name:             "missing body",
			body:             `{}`,
			wantWriterCalled: true,
		},
		{
			name:             "non string body",
			body:             `{"body":123}`,
			wantWriterCalled: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writerCalled := false
			router := NewHandler(HandlerConfig{
				FanShortCommentWriter: stubFanShortCommentWriter{
					createComment: func(context.Context, shortcomment.CreateInput) (shortcomment.Comment, error) {
						writerCalled = true
						return shortcomment.Comment{}, shortcomment.ErrInvalidCommentBody
					},
				},
				ViewerBootstrap: viewerBootstrapReaderStub{
					readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
						return auth.Bootstrap{
							CurrentViewer: &auth.CurrentViewer{
								ID:                   viewerID,
								ActiveMode:           auth.ActiveModeFan,
								CanAccessCreatorMode: false,
							},
						}, nil
					},
				},
			})

			req := httptest.NewRequest(http.MethodPost, "/api/fan/shorts/short_11111111111111111111111111111111/comments", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("POST /api/fan/shorts/{shortId}/comments status got %d want %d", rec.Code, http.StatusBadRequest)
			}
			if writerCalled != tt.wantWriterCalled {
				t.Fatalf("POST /api/fan/shorts/{shortId}/comments writerCalled got %t want %t", writerCalled, tt.wantWriterCalled)
			}
			if !strings.Contains(rec.Body.String(), `"code":"validation_error"`) {
				t.Fatalf("POST /api/fan/shorts/{shortId}/comments body got %q want validation_error", rec.Body.String())
			}
		})
	}
}

func TestFanShortCommentRoutesDoNotRegisterMainCommentEndpoints(t *testing.T) {
	t.Parallel()

	mainID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	readerCalled := false
	writerCalled := false
	router := NewHandler(HandlerConfig{
		FanShortComments: stubFanShortCommentReader{
			listComments: func(context.Context, uuid.UUID, *shortcomment.Cursor, int) ([]shortcomment.Comment, *shortcomment.Cursor, error) {
				readerCalled = true
				return nil, nil, nil
			},
		},
		FanShortCommentWriter: stubFanShortCommentWriter{
			createComment: func(context.Context, shortcomment.CreateInput) (shortcomment.Comment, error) {
				writerCalled = true
				return shortcomment.Comment{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         uuid.MustParse("99999999-9999-9999-9999-999999999999"),
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
	})

	listReq := httptest.NewRequest(http.MethodGet, "/api/fan/mains/"+shorts.FormatPublicMainID(mainID)+"/comments", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/fan/mains/{mainId}/comments status got %d want %d", listRec.Code, http.StatusNotFound)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/fan/mains/"+shorts.FormatPublicMainID(mainID)+"/comments", bytes.NewBufferString(`{"body":"hello"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusNotFound {
		t.Fatalf("POST /api/fan/mains/{mainId}/comments status got %d want %d", createRec.Code, http.StatusNotFound)
	}

	if readerCalled {
		t.Fatal("main comment GET called short comment reader, want false")
	}
	if writerCalled {
		t.Fatal("main comment POST called short comment writer, want false")
	}
}

func TestFanShortCommentRoutesDoNotRegisterUnsupportedCommentCapabilities(t *testing.T) {
	t.Parallel()

	shortID := shorts.FormatPublicShortID(uuid.MustParse("11111111-1111-1111-1111-111111111111"))
	commentID := shortCommentPublicID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))
	readerCalled := false
	writerCalled := false
	router := NewHandler(HandlerConfig{
		FanShortComments: stubFanShortCommentReader{
			listComments: func(context.Context, uuid.UUID, *shortcomment.Cursor, int) ([]shortcomment.Comment, *shortcomment.Cursor, error) {
				readerCalled = true
				return nil, nil, nil
			},
		},
		FanShortCommentWriter: stubFanShortCommentWriter{
			createComment: func(context.Context, shortcomment.CreateInput) (shortcomment.Comment, error) {
				writerCalled = true
				return shortcomment.Comment{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         uuid.MustParse("99999999-9999-9999-9999-999999999999"),
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
	})

	tests := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodDelete,
			path:   "/api/fan/shorts/" + shortID + "/comments/" + commentID,
		},
		{
			method: http.MethodPost,
			path:   "/api/fan/shorts/" + shortID + "/comments/" + commentID + "/likes",
		},
		{
			method: http.MethodPost,
			path:   "/api/fan/shorts/" + shortID + "/comments/" + commentID + "/replies",
		},
		{
			method: http.MethodPost,
			path:   "/api/fan/shorts/" + shortID + "/comments/" + commentID + "/report",
		},
		{
			method: http.MethodPost,
			path:   "/api/fan/shorts/" + shortID + "/comments/" + commentID + "/moderation",
		},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(`{"body":"hello"}`))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s %s status got %d want 404 or 405", tt.method, tt.path, rec.Code)
		}
	}

	if readerCalled {
		t.Fatal("unsupported comment capability called short comment reader, want false")
	}
	if writerCalled {
		t.Fatal("unsupported comment capability called short comment writer, want false")
	}
}

func TestFanShortCommentRoutesReturnExpectedErrors(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	router := NewHandler(HandlerConfig{
		FanShortComments: stubFanShortCommentReader{
			listComments: func(context.Context, uuid.UUID, *shortcomment.Cursor, int) ([]shortcomment.Comment, *shortcomment.Cursor, error) {
				return nil, nil, shortcomment.ErrShortNotFound
			},
		},
		FanShortCommentWriter: stubFanShortCommentWriter{
			createComment: func(context.Context, shortcomment.CreateInput) (shortcomment.Comment, error) {
				return shortcomment.Comment{}, shortcomment.ErrInvalidCommentBody
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
	})

	listReq := httptest.NewRequest(http.MethodGet, "/api/fan/shorts/short_11111111111111111111111111111111/comments", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/fan/shorts/{shortId}/comments status got %d want %d", listRec.Code, http.StatusNotFound)
	}
	if !strings.Contains(listRec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("GET /api/fan/shorts/{shortId}/comments body got %q want not_found", listRec.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/fan/shorts/short_11111111111111111111111111111111/comments", bytes.NewBufferString(`{"body":" "}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/fan/shorts/{shortId}/comments status got %d want %d", createRec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(createRec.Body.String(), `"code":"validation_error"`) {
		t.Fatalf("POST /api/fan/shorts/{shortId}/comments body got %q want validation_error", createRec.Body.String())
	}

	router = NewHandler(HandlerConfig{
		FanShortCommentWriter: stubFanShortCommentWriter{
			createComment: func(context.Context, shortcomment.CreateInput) (shortcomment.Comment, error) {
				return shortcomment.Comment{}, errors.New("unexpected")
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{ID: viewerID},
				}, nil
			},
		},
	})
	internalReq := httptest.NewRequest(http.MethodPost, "/api/fan/shorts/short_11111111111111111111111111111111/comments", bytes.NewBufferString(`{"body":"hello"}`))
	internalReq.Header.Set("Content-Type", "application/json")
	internalReq.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	internalRec := httptest.NewRecorder()
	router.ServeHTTP(internalRec, internalReq)
	if internalRec.Code != http.StatusInternalServerError {
		t.Fatalf("POST /api/fan/shorts/{shortId}/comments status got %d want %d", internalRec.Code, http.StatusInternalServerError)
	}
}
