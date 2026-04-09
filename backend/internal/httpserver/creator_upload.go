package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatorupload"
)

const (
	creatorUploadAuthRequiredMessage       = "creator upload requires authentication"
	creatorUploadCreateRequestScope        = "creator_upload_packages_create"
	creatorUploadCompleteRequestScope      = "creator_upload_packages_complete"
	creatorUploadCapabilityRequiredCode    = "capability_required"
	creatorUploadCapabilityRequiredMessage = "approved creator capability is required"
)

type creatorUploadCreateRequest struct {
	Main   *creatorUploadFileRequest  `json:"main"`
	Shorts []creatorUploadFileRequest `json:"shorts"`
}

type creatorUploadFileRequest struct {
	FileName      string `json:"fileName"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	MimeType      string `json:"mimeType"`
}

type creatorUploadCompleteRequest struct {
	Main         *creatorUploadEntryRequest  `json:"main"`
	PackageToken string                      `json:"packageToken"`
	Shorts       []creatorUploadEntryRequest `json:"shorts"`
}

type creatorUploadEntryRequest struct {
	UploadEntryID string `json:"uploadEntryId"`
}

type creatorUploadCreateResponseData struct {
	ExpiresAt     string                        `json:"expiresAt"`
	PackageToken  string                        `json:"packageToken"`
	UploadTargets creatorUploadTargetSetPayload `json:"uploadTargets"`
}

type creatorUploadTargetSetPayload struct {
	Main   creatorUploadTargetPayload   `json:"main"`
	Shorts []creatorUploadTargetPayload `json:"shorts"`
}

type creatorUploadTargetPayload struct {
	FileName      string                    `json:"fileName"`
	MimeType      string                    `json:"mimeType"`
	Role          string                    `json:"role"`
	Upload        creatorUploadDirectUpload `json:"upload"`
	UploadEntryID string                    `json:"uploadEntryId"`
}

type creatorUploadDirectUpload struct {
	Headers map[string]string `json:"headers"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
}

type creatorUploadCompleteResponseData struct {
	Main   creatorUploadCreatedMainPayload    `json:"main"`
	Shorts []creatorUploadCreatedShortPayload `json:"shorts"`
}

type creatorUploadCreatedMainPayload struct {
	ID         string                                `json:"id"`
	MediaAsset creatorUploadCreatedMediaAssetPayload `json:"mediaAsset"`
	State      string                                `json:"state"`
}

type creatorUploadCreatedShortPayload struct {
	CanonicalMainID string                                `json:"canonicalMainId"`
	ID              string                                `json:"id"`
	MediaAsset      creatorUploadCreatedMediaAssetPayload `json:"mediaAsset"`
	State           string                                `json:"state"`
}

type creatorUploadCreatedMediaAssetPayload struct {
	ID              string `json:"id"`
	MimeType        string `json:"mimeType"`
	ProcessingState string `json:"processingState"`
}

func registerCreatorUploadRoutes(router gin.IRouter, service CreatorUploadHandler, viewerBootstrap ViewerBootstrapReader) {
	if service == nil || viewerBootstrap == nil {
		return
	}

	creatorGroup := router.Group("/api/creator")
	creatorGroup.POST(
		"/upload-packages",
		buildProtectedFanAuthGuard(viewerBootstrap, creatorUploadCreateRequestScope, creatorUploadAuthRequiredMessage),
		func(c *gin.Context) {
			handleCreatorUploadCreate(c, service)
		},
	)
	creatorGroup.POST(
		"/upload-packages/complete",
		buildProtectedFanAuthGuard(viewerBootstrap, creatorUploadCompleteRequestScope, creatorUploadAuthRequiredMessage),
		func(c *gin.Context) {
			handleCreatorUploadComplete(c, service)
		},
	)
}

func handleCreatorUploadCreate(c *gin.Context, service CreatorUploadHandler) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		writeInternalServerError(c, creatorUploadCreateRequestScope)
		return
	}
	if !viewer.CanAccessCreatorMode {
		writeCreatorUploadError(c, http.StatusForbidden, creatorUploadCapabilityRequiredCode, creatorUploadCapabilityRequiredMessage, creatorUploadCreateRequestScope)
		return
	}

	var request creatorUploadCreateRequest
	if !decodeCreatorUploadJSON(c, &request, "creator upload request is invalid", creatorUploadCreateRequestScope) {
		return
	}

	input := creatorupload.CreatePackageInput{
		CreatorUserID: viewer.ID,
		Shorts:        make([]creatorupload.FileMetadata, 0, len(request.Shorts)),
	}
	if request.Main != nil {
		input.Main = &creatorupload.FileMetadata{
			FileName:      request.Main.FileName,
			FileSizeBytes: request.Main.FileSizeBytes,
			MimeType:      request.Main.MimeType,
		}
	}
	for _, short := range request.Shorts {
		input.Shorts = append(input.Shorts, creatorupload.FileMetadata{
			FileName:      short.FileName,
			FileSizeBytes: short.FileSizeBytes,
			MimeType:      short.MimeType,
		})
	}

	result, err := service.CreatePackage(c.Request.Context(), input)
	if err != nil {
		var validationErr *creatorupload.ValidationError
		switch {
		case errors.As(err, &validationErr):
			writeCreatorUploadError(c, http.StatusBadRequest, "validation_error", validationErr.Message(), creatorUploadCreateRequestScope)
			return
		case errors.Is(err, creatorupload.ErrStorageFailure):
			writeCreatorUploadError(c, http.StatusServiceUnavailable, "storage_failure", "creator upload target generation failed", creatorUploadCreateRequestScope)
			return
		default:
			writeInternalServerError(c, creatorUploadCreateRequestScope)
			return
		}
	}

	shortTargets := make([]creatorUploadTargetPayload, 0, len(result.UploadTargets.Shorts))
	for _, shortTarget := range result.UploadTargets.Shorts {
		shortTargets = append(shortTargets, buildCreatorUploadTargetPayload(shortTarget))
	}

	c.JSON(http.StatusOK, responseEnvelope[creatorUploadCreateResponseData]{
		Data: &creatorUploadCreateResponseData{
			ExpiresAt:    result.ExpiresAt.Format(time.RFC3339),
			PackageToken: result.PackageToken,
			UploadTargets: creatorUploadTargetSetPayload{
				Main:   buildCreatorUploadTargetPayload(result.UploadTargets.Main),
				Shorts: shortTargets,
			},
		},
		Meta: responseMeta{
			RequestID: newRequestID(creatorUploadCreateRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleCreatorUploadComplete(c *gin.Context, service CreatorUploadHandler) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		writeInternalServerError(c, creatorUploadCompleteRequestScope)
		return
	}
	if !viewer.CanAccessCreatorMode {
		writeCreatorUploadError(c, http.StatusForbidden, creatorUploadCapabilityRequiredCode, creatorUploadCapabilityRequiredMessage, creatorUploadCompleteRequestScope)
		return
	}

	var request creatorUploadCompleteRequest
	if !decodeCreatorUploadJSON(c, &request, "creator upload completion request is invalid", creatorUploadCompleteRequestScope) {
		return
	}

	input := creatorupload.CompletePackageInput{
		CreatorUserID: viewer.ID,
		PackageToken:  request.PackageToken,
		Shorts:        make([]creatorupload.UploadEntryReference, 0, len(request.Shorts)),
	}
	if request.Main != nil {
		input.Main = &creatorupload.UploadEntryReference{UploadEntryID: request.Main.UploadEntryID}
	}
	for _, short := range request.Shorts {
		input.Shorts = append(input.Shorts, creatorupload.UploadEntryReference{UploadEntryID: short.UploadEntryID})
	}

	result, err := service.CompletePackage(c.Request.Context(), input)
	if err != nil {
		var validationErr *creatorupload.ValidationError
		switch {
		case errors.As(err, &validationErr):
			writeCreatorUploadError(c, http.StatusBadRequest, "validation_error", validationErr.Message(), creatorUploadCompleteRequestScope)
			return
		case errors.Is(err, creatorupload.ErrPackageExpired):
			writeCreatorUploadError(c, http.StatusConflict, "upload_expired", "creator upload package has expired", creatorUploadCompleteRequestScope)
			return
		case errors.Is(err, creatorupload.ErrUploadFailure):
			writeCreatorUploadError(c, http.StatusUnprocessableEntity, "upload_failure", "creator upload package is incomplete", creatorUploadCompleteRequestScope)
			return
		case errors.Is(err, creatorupload.ErrStorageFailure):
			writeCreatorUploadError(c, http.StatusServiceUnavailable, "storage_failure", "creator upload storage verification failed", creatorUploadCompleteRequestScope)
			return
		default:
			writeInternalServerError(c, creatorUploadCompleteRequestScope)
			return
		}
	}

	shorts := make([]creatorUploadCreatedShortPayload, 0, len(result.Shorts))
	for _, short := range result.Shorts {
		shorts = append(shorts, creatorUploadCreatedShortPayload{
			CanonicalMainID: short.CanonicalMainID.String(),
			ID:              short.ID.String(),
			MediaAsset:      buildCreatorUploadCreatedMediaAsset(short.MediaAsset),
			State:           short.State,
		})
	}

	c.JSON(http.StatusOK, responseEnvelope[creatorUploadCompleteResponseData]{
		Data: &creatorUploadCompleteResponseData{
			Main: creatorUploadCreatedMainPayload{
				ID:         result.Main.ID.String(),
				MediaAsset: buildCreatorUploadCreatedMediaAsset(result.Main.MediaAsset),
				State:      result.Main.State,
			},
			Shorts: shorts,
		},
		Meta: responseMeta{
			RequestID: newRequestID(creatorUploadCompleteRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func decodeCreatorUploadJSON[T any](c *gin.Context, target *T, message string, requestScope string) bool {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeCreatorUploadError(c, http.StatusBadRequest, "invalid_request", message, requestScope)
		return false
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != nil && !errors.Is(err, io.EOF) {
		writeCreatorUploadError(c, http.StatusBadRequest, "invalid_request", message, requestScope)
		return false
	}
	if len(extra) > 0 {
		writeCreatorUploadError(c, http.StatusBadRequest, "invalid_request", message, requestScope)
		return false
	}

	return true
}

func writeCreatorUploadError(c *gin.Context, status int, code string, message string, requestScope string) {
	c.JSON(status, responseEnvelope[struct{}]{
		Data: nil,
		Meta: responseMeta{
			RequestID: newRequestID(requestScope),
			Page:      nil,
		},
		Error: &responseError{
			Code:    code,
			Message: message,
		},
	})
}

func buildCreatorUploadTargetPayload(target creatorupload.UploadTarget) creatorUploadTargetPayload {
	return creatorUploadTargetPayload{
		FileName: target.FileName,
		MimeType: target.MimeType,
		Role:     target.Role,
		Upload: creatorUploadDirectUpload{
			Headers: target.Upload.Headers,
			Method:  target.Upload.Method,
			URL:     target.Upload.URL,
		},
		UploadEntryID: target.UploadEntryID,
	}
}

func buildCreatorUploadCreatedMediaAsset(asset creatorupload.CreatedMediaAsset) creatorUploadCreatedMediaAssetPayload {
	return creatorUploadCreatedMediaAssetPayload{
		ID:              asset.ID.String(),
		MimeType:        asset.MimeType,
		ProcessingState: asset.ProcessingState,
	}
}
