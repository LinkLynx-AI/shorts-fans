package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ViewerBootstrapReader は HTTP transport から current viewer bootstrap を読む境界です。
type ViewerBootstrapReader interface {
	ReadCurrentViewer(ctx context.Context, rawSessionToken string) (auth.Bootstrap, error)
}

func buildViewerBootstrapHandler(reader ViewerBootstrapReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawSessionToken, err := c.Cookie(auth.SessionCookieName)
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			c.JSON(http.StatusBadRequest, buildViewerBootstrapErrorEnvelope("invalid_cookie", "session cookie could not be read"))
			return
		}

		bootstrap, err := reader.ReadCurrentViewer(c.Request.Context(), rawSessionToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, buildViewerBootstrapErrorEnvelope("internal_error", "viewer bootstrap could not be loaded"))
			return
		}

		var currentViewer any
		if bootstrap.CurrentViewer != nil {
			currentViewer = gin.H{
				"id":                   bootstrap.CurrentViewer.ID.String(),
				"activeMode":           string(bootstrap.CurrentViewer.ActiveMode),
				"canAccessCreatorMode": bootstrap.CurrentViewer.CanAccessCreatorMode,
			}
		} else {
			currentViewer = nil
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"currentViewer": currentViewer,
			},
			"meta": gin.H{
				"page":      nil,
				"requestId": uuid.NewString(),
			},
			"error": nil,
		})
	}
}

func buildViewerBootstrapErrorEnvelope(code string, message string) gin.H {
	return gin.H{
		"data": nil,
		"meta": gin.H{
			"page":      nil,
			"requestId": uuid.NewString(),
		},
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}
}
