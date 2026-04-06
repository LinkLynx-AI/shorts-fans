package httpserver

import (
	"net/http"
	"strings"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

const authenticatedViewerContextKey = "authenticatedViewer"

// buildProtectedFanAuthGuard は protected fan surface 用の認証 middleware を返します。
func buildProtectedFanAuthGuard(
	reader ViewerBootstrapReader,
	requestScope string,
	authRequiredMessage string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		if reader == nil {
			writeInternalServerError(c, requestScope)
			c.Abort()
			return
		}

		rawSessionToken, err := c.Cookie(auth.SessionCookieName)
		if err != nil {
			writeAuthRequiredError(c, requestScope, authRequiredMessage)
			c.Abort()
			return
		}

		bootstrap, err := reader.ReadCurrentViewer(c.Request.Context(), rawSessionToken)
		if err != nil {
			writeInternalServerError(c, requestScope)
			c.Abort()
			return
		}

		if bootstrap.CurrentViewer == nil {
			writeAuthRequiredError(c, requestScope, authRequiredMessage)
			c.Abort()
			return
		}

		c.Set(authenticatedViewerContextKey, *bootstrap.CurrentViewer)
		c.Next()
	}
}

func writeAuthRequiredError(c *gin.Context, requestScope string, message string) {
	resolvedMessage := strings.TrimSpace(message)
	if resolvedMessage == "" {
		resolvedMessage = "authentication is required"
	}

	c.JSON(http.StatusUnauthorized, responseEnvelope[struct{}]{
		Data: nil,
		Meta: responseMeta{
			RequestID: newRequestID(requestScope),
			Page:      nil,
		},
		Error: &responseError{
			Code:    "auth_required",
			Message: resolvedMessage,
		},
	})
}

func authenticatedViewerFromContext(c *gin.Context) (auth.CurrentViewer, bool) {
	if c == nil {
		return auth.CurrentViewer{}, false
	}

	value, ok := c.Get(authenticatedViewerContextKey)
	if !ok {
		return auth.CurrentViewer{}, false
	}

	viewer, ok := value.(auth.CurrentViewer)
	if !ok {
		return auth.CurrentViewer{}, false
	}

	return viewer, true
}
