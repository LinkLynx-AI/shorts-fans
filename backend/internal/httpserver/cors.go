package httpserver

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

const (
	developmentAllowedHeaders = "Accept, Content-Type"
	developmentAllowedMethods = "GET, POST, PUT, DELETE, OPTIONS"
	developmentAppEnv         = "development"
	productionAppEnv          = "production"
)

// devLoopbackCORS は local frontend からの cross-origin request を許可します。
func devLoopbackCORS(appEnv string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if appEnv != developmentAppEnv {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")
		if !isAllowedDevelopmentOrigin(origin) {
			c.Next()
			return
		}

		headers := c.Writer.Header()
		headers.Set("Access-Control-Allow-Origin", origin)
		headers.Set("Access-Control-Allow-Credentials", "true")
		headers.Set("Access-Control-Allow-Methods", developmentAllowedMethods)
		headers.Set("Access-Control-Allow-Headers", developmentAllowedHeaders)
		headers.Add("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isAllowedDevelopmentOrigin(origin string) bool {
	if origin == "" {
		return false
	}

	parsedOrigin, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if parsedOrigin.Scheme != "http" && parsedOrigin.Scheme != "https" {
		return false
	}

	switch parsedOrigin.Hostname() {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}
