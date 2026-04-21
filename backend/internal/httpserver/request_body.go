package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	jsonRequestBodyLimitBytes    int64 = 1 << 20
	webhookRequestBodyLimitBytes int64 = 256 << 10
)

var errRequestBodyTrailingData = errors.New("request body contains trailing data")

func newLimitedJSONDecoder(c *gin.Context) *json.Decoder {
	return json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, jsonRequestBodyLimitBytes))
}

func decodeLimitedJSONBody[T any](c *gin.Context, target *T, disallowUnknownFields bool) error {
	decoder := newLimitedJSONDecoder(c)
	if disallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(target); err != nil {
		return err
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	return errRequestBodyTrailingData
}

func readLimitedRequestBody(c *gin.Context, maxBytes int64) ([]byte, error) {
	return io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes))
}
