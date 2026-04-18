package httpserver

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/payment"
	"github.com/gin-gonic/gin"
)

const ccbillWebhookRequestScope = "ccbill_webhook"

func registerPaymentWebhookRoutes(router gin.IRouter, handler PaymentWebhookHandler) {
	if router == nil || handler == nil {
		return
	}

	router.POST("/api/payments/ccbill/webhooks", func(c *gin.Context) {
		handleCCBillWebhook(c, handler)
	})
}

func handleCCBillWebhook(c *gin.Context, handler PaymentWebhookHandler) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		writePaymentWebhookError(c, http.StatusBadRequest, "invalid_request", "webhook payload was invalid")
		return
	}

	err = handler.HandleWebhook(
		c.Request.Context(),
		directRemoteIP(c.Request.RemoteAddr),
		c.Request.URL.Query(),
		c.GetHeader("Content-Type"),
		body,
	)
	if err != nil {
		switch {
		case errors.Is(err, payment.ErrCCBillWebhookOriginRejected):
			writePaymentWebhookError(c, http.StatusForbidden, "webhook_origin_rejected", "webhook origin was rejected")
		case errors.Is(err, payment.ErrCCBillWebhookInvalid):
			writePaymentWebhookError(c, http.StatusBadRequest, "invalid_request", "webhook payload was invalid")
		default:
			writeInternalServerError(c, ccbillWebhookRequestScope)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func writePaymentWebhookError(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, responseEnvelope[struct{}]{
		Data: nil,
		Meta: responseMeta{
			RequestID: newRequestID(ccbillWebhookRequestScope),
			Page:      nil,
		},
		Error: &responseError{
			Code:    code,
			Message: message,
		},
	})
}

func directRemoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		return strings.TrimSpace(remoteAddr)
	}

	return host
}
