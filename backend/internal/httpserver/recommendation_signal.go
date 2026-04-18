package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/recommendation"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	fanRecommendationEventRequestScope = "fan_recommendation_event"
)

var errRecommendationSignalExposureRequired = errors.New("recommendation signal target exposure is required")

type recommendationSignalRequestPayload struct {
	CreatorID      string `json:"creatorId"`
	EventKind      string `json:"eventKind"`
	IdempotencyKey string `json:"idempotencyKey"`
	ShortID        string `json:"shortId"`
}

type recommendationSignalResponseData struct {
	Recorded bool `json:"recorded"`
}

// RecommendationSignalWriter は fan recommendation signal mutation を表します。
type RecommendationSignalWriter interface {
	RecordProfileClick(ctx context.Context, viewerID uuid.UUID, creatorUserID uuid.UUID, idempotencyKey string) (recommendation.RecordEventResult, error)
	RecordShortSignal(ctx context.Context, viewerID uuid.UUID, shortID uuid.UUID, eventKind recommendation.EventKind, idempotencyKey string) (recommendation.RecordEventResult, error)
}

func registerRecommendationSignalRoutes(
	router gin.IRouter,
	writer RecommendationSignalWriter,
	recommendationSignalExposure RecommendationSignalExposureStore,
	viewerBootstrap ViewerBootstrapReader,
) {
	if router == nil || writer == nil || recommendationSignalExposure == nil || viewerBootstrap == nil {
		return
	}

	recommendationGroup := router.Group("/")
	recommendationGroup.Use(buildProtectedFanAuthGuard(viewerBootstrap, fanRecommendationEventRequestScope, fanMainAuthRequiredMessage))
	recommendationGroup.POST("/api/fan/recommendation/events", func(c *gin.Context) {
		handleRecommendationSignal(c, writer, recommendationSignalExposure)
	})
}

func handleRecommendationSignal(c *gin.Context, writer RecommendationSignalWriter, recommendationSignalExposure RecommendationSignalExposureStore) {
	viewer, _, ok := resolveAuthenticatedViewerRequest(c)
	if !ok {
		writeInternalServerError(c, fanRecommendationEventRequestScope)
		return
	}

	var request recommendationSignalRequestPayload
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		writeRecommendationSignalInvalidRequest(c, "recommendation signal request was invalid")
		return
	}

	result, err := recordRecommendationSignal(
		c.Request.Context(),
		writer,
		recommendationSignalExposure,
		viewer.ID,
		request,
	)
	if err != nil {
		switch {
		case errors.Is(err, recommendation.ErrSignalTargetNotFound):
			writeNotFoundError(c, fanRecommendationEventRequestScope, "recommendation signal target was not found")
		case errors.Is(err, errRecommendationSignalExposureRequired):
			writeRecommendationSignalInvalidRequest(c, "recommendation signal target was not recently surfaced")
		case errors.Is(err, recommendation.ErrIdempotencyConflict):
			writeRecommendationSignalError(c, http.StatusConflict, "idempotency_conflict", "recommendation signal idempotency key conflicted")
		case errors.Is(err, recommendation.ErrEventKindInvalid),
			errors.Is(err, recommendation.ErrIdempotencyKeyRequired),
			errors.Is(err, recommendation.ErrCreatorUserIDRequired),
			errors.Is(err, recommendation.ErrCanonicalMainIDRequired),
			errors.Is(err, recommendation.ErrCanonicalMainIDForbidden),
			errors.Is(err, recommendation.ErrShortIDRequired),
			errors.Is(err, recommendation.ErrShortIDForbidden),
			errors.Is(err, recommendation.ErrShortIDInvalid):
			writeRecommendationSignalInvalidRequest(c, "recommendation signal request was invalid")
		default:
			writeInternalServerError(c, fanRecommendationEventRequestScope)
		}
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[recommendationSignalResponseData]{
		Data: &recommendationSignalResponseData{
			Recorded: result.Recorded,
		},
		Meta: responseMeta{
			RequestID: newRequestID(fanRecommendationEventRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func recordRecommendationSignal(
	ctx context.Context,
	writer RecommendationSignalWriter,
	recommendationSignalExposure RecommendationSignalExposureStore,
	viewerID uuid.UUID,
	request recommendationSignalRequestPayload,
) (recommendation.RecordEventResult, error) {
	if writer == nil || recommendationSignalExposure == nil {
		return recommendation.RecordEventResult{}, errors.New("recommendation signal dependency が初期化されていません")
	}

	eventKind := recommendation.EventKind(strings.TrimSpace(request.EventKind))
	idempotencyKey := strings.TrimSpace(request.IdempotencyKey)

	switch eventKind {
	case recommendation.EventKindImpression,
		recommendation.EventKindViewStart,
		recommendation.EventKindViewCompletion,
		recommendation.EventKindRewatchLoop,
		recommendation.EventKindMainClick:
		shortIDText := strings.TrimSpace(request.ShortID)
		if shortIDText == "" {
			return recommendation.RecordEventResult{}, recommendation.ErrShortIDRequired
		}

		shortID, err := shorts.ParsePublicShortID(shortIDText)
		if err != nil {
			return recommendation.RecordEventResult{}, recommendation.ErrShortIDInvalid
		}
		hasExposure, err := recommendationSignalExposure.HasShortExposure(ctx, viewerID, shortID)
		if err != nil {
			return recommendation.RecordEventResult{}, err
		}
		if !hasExposure {
			return recommendation.RecordEventResult{}, errRecommendationSignalExposureRequired
		}

		return writer.RecordShortSignal(ctx, viewerID, shortID, eventKind, idempotencyKey)
	case recommendation.EventKindProfileClick:
		creatorUserID, err := creator.ParsePublicID(request.CreatorID)
		if err != nil {
			return recommendation.RecordEventResult{}, recommendation.ErrCreatorUserIDRequired
		}
		hasExposure, err := recommendationSignalExposure.HasCreatorExposure(ctx, viewerID, creatorUserID)
		if err != nil {
			return recommendation.RecordEventResult{}, err
		}
		if !hasExposure {
			return recommendation.RecordEventResult{}, errRecommendationSignalExposureRequired
		}

		return writer.RecordProfileClick(ctx, viewerID, creatorUserID, idempotencyKey)
	default:
		return recommendation.RecordEventResult{}, recommendation.ErrEventKindInvalid
	}
}

func writeRecommendationSignalInvalidRequest(c *gin.Context, message string) {
	writeRecommendationSignalError(c, http.StatusBadRequest, "invalid_request", message)
}

func writeRecommendationSignalError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, responseEnvelope[struct{}]{
		Data: nil,
		Meta: responseMeta{
			RequestID: newRequestID(fanRecommendationEventRequestScope),
			Page:      nil,
		},
		Error: &responseError{
			Code:    code,
			Message: message,
		},
	})
}
