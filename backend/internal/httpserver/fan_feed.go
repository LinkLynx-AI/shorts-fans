package httpserver

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FanFeedService は fan feed read surface を提供します。
type FanFeedService interface {
	ListRecommended(ctx context.Context, input feed.ListRecommendedInput) (feed.RecommendedFeed, error)
}

type responseEnvelope struct {
	Data  any            `json:"data"`
	Meta  responseMeta   `json:"meta"`
	Error *responseError `json:"error"`
}

type responseMeta struct {
	RequestID string        `json:"requestId"`
	Page      *responsePage `json:"page"`
}

type responsePage struct {
	NextCursor *string `json:"nextCursor"`
	HasNext    bool    `json:"hasNext"`
}

type responseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func registerFanRoutes(router *gin.Engine, fanServices FanServices) {
	fanGroup := router.Group("/api/fan")
	fanGroup.GET("/feed", func(c *gin.Context) {
		handleFanFeed(c, fanServices.Feed)
	})
}

func handleFanFeed(c *gin.Context, service FanFeedService) {
	requestID := newRequestID("req_feed")
	tab := strings.TrimSpace(c.Query("tab"))
	if tab == "" {
		tab = "recommended"
	}

	switch tab {
	case "recommended":
		if service == nil {
			c.JSON(http.StatusServiceUnavailable, responseEnvelope{
				Data: nil,
				Meta: responseMeta{
					RequestID: requestID,
					Page:      nil,
				},
				Error: &responseError{
					Code:    "service_unavailable",
					Message: "recommended feed service is unavailable",
				},
			})
			return
		}

		result, err := service.ListRecommended(c.Request.Context(), feed.ListRecommendedInput{
			Cursor: c.Query("cursor"),
		})
		if err != nil {
			status := http.StatusInternalServerError
			errorCode := "internal_error"
			errorMessage := "recommended feed request failed"
			if errors.Is(err, feed.ErrInvalidCursor) {
				status = http.StatusBadRequest
				errorCode = "invalid_request"
				errorMessage = "cursor is invalid"
			}

			c.JSON(status, responseEnvelope{
				Data: nil,
				Meta: responseMeta{
					RequestID: requestID,
					Page:      nil,
				},
				Error: &responseError{
					Code:    errorCode,
					Message: errorMessage,
				},
			})
			return
		}

		c.JSON(http.StatusOK, responseEnvelope{
			Data: result,
			Meta: responseMeta{
				RequestID: requestID,
				Page: &responsePage{
					NextCursor: result.NextCursor,
					HasNext:    result.HasNext,
				},
			},
			Error: nil,
		})
	case "following":
		c.JSON(http.StatusUnauthorized, responseEnvelope{
			Data: nil,
			Meta: responseMeta{
				RequestID: requestID,
				Page:      nil,
			},
			Error: &responseError{
				Code:    "auth_required",
				Message: "following feed requires authentication",
			},
		})
	default:
		c.JSON(http.StatusBadRequest, responseEnvelope{
			Data: nil,
			Meta: responseMeta{
				RequestID: requestID,
				Page:      nil,
			},
			Error: &responseError{
				Code:    "invalid_request",
				Message: "tab is invalid",
			},
		})
	}
}

func newRequestID(prefix string) string {
	return prefix + "_" + strings.ReplaceAll(uuid.NewString(), "-", "")
}
