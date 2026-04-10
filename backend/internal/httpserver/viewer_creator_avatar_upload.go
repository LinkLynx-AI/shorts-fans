package httpserver

import (
	"errors"
	"net/http"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatoravatar"
	"github.com/gin-gonic/gin"
)

const (
	viewerCreatorAvatarUploadCreateRequestScope   = "viewer_creator_avatar_upload_create"
	viewerCreatorAvatarUploadCompleteRequestScope = "viewer_creator_avatar_upload_complete"
)

type viewerCreatorAvatarUploadCreateRequest struct {
	FileName      string `json:"fileName"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	MimeType      string `json:"mimeType"`
}

type viewerCreatorAvatarUploadCompleteRequest struct {
	AvatarUploadToken string `json:"avatarUploadToken"`
}

type viewerCreatorAvatarUploadCreateResponseData struct {
	AvatarUploadToken string                               `json:"avatarUploadToken"`
	ExpiresAt         string                               `json:"expiresAt"`
	UploadTarget      viewerCreatorAvatarUploadTargetShape `json:"uploadTarget"`
}

type viewerCreatorAvatarUploadTargetShape struct {
	FileName string                        `json:"fileName"`
	MimeType string                        `json:"mimeType"`
	Upload   viewerCreatorAvatarDirectLink `json:"upload"`
}

type viewerCreatorAvatarDirectLink struct {
	Headers map[string]string `json:"headers"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
}

type viewerCreatorAvatarUploadCompleteResponseData struct {
	Avatar            mediaAsset `json:"avatar"`
	AvatarUploadToken string     `json:"avatarUploadToken"`
}

func handleViewerCreatorAvatarUploadCreate(c *gin.Context, handler ViewerCreatorAvatarUploadHandler) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		writeInternalServerError(c, viewerCreatorAvatarUploadCreateRequestScope)
		return
	}

	var request viewerCreatorAvatarUploadCreateRequest
	if !decodeViewerCreatorEntryJSON(
		c,
		&request,
		"invalid_request",
		"creator avatar upload request is invalid",
		viewerCreatorAvatarUploadCreateRequestScope,
	) {
		return
	}

	result, err := handler.CreateUpload(c.Request.Context(), creatoravatar.CreateUploadInput{
		FileName:      request.FileName,
		FileSizeBytes: request.FileSizeBytes,
		MimeType:      request.MimeType,
		ViewerUserID:  viewer.ID,
	})
	if err != nil {
		var validationErr *creatoravatar.ValidationError
		switch {
		case errors.As(err, &validationErr):
			writeViewerCreatorEntryError(
				c,
				http.StatusBadRequest,
				validationErr.Code(),
				validationErr.Message(),
				viewerCreatorAvatarUploadCreateRequestScope,
			)
			return
		default:
			writeInternalServerError(c, viewerCreatorAvatarUploadCreateRequestScope)
			return
		}
	}

	c.JSON(http.StatusOK, responseEnvelope[viewerCreatorAvatarUploadCreateResponseData]{
		Data: &viewerCreatorAvatarUploadCreateResponseData{
			AvatarUploadToken: result.AvatarUploadToken,
			ExpiresAt:         result.ExpiresAt.Format(time.RFC3339),
			UploadTarget: viewerCreatorAvatarUploadTargetShape{
				FileName: result.UploadTarget.FileName,
				MimeType: result.UploadTarget.MimeType,
				Upload: viewerCreatorAvatarDirectLink{
					Headers: result.UploadTarget.Upload.Headers,
					Method:  result.UploadTarget.Upload.Method,
					URL:     result.UploadTarget.Upload.URL,
				},
			},
		},
		Meta: responseMeta{
			RequestID: newRequestID(viewerCreatorAvatarUploadCreateRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleViewerCreatorAvatarUploadComplete(c *gin.Context, handler ViewerCreatorAvatarUploadHandler) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		writeInternalServerError(c, viewerCreatorAvatarUploadCompleteRequestScope)
		return
	}

	var request viewerCreatorAvatarUploadCompleteRequest
	if !decodeViewerCreatorEntryJSON(
		c,
		&request,
		"invalid_request",
		"creator avatar upload completion request is invalid",
		viewerCreatorAvatarUploadCompleteRequestScope,
	) {
		return
	}

	result, err := handler.CompleteUpload(c.Request.Context(), creatoravatar.CompleteUploadInput{
		AvatarUploadToken: request.AvatarUploadToken,
		ViewerUserID:      viewer.ID,
	})
	if err != nil {
		switch {
		case errors.Is(err, creatoravatar.ErrUploadNotFound):
			writeViewerCreatorEntryError(c, http.StatusNotFound, "avatar_upload_not_found", "avatar upload was not found", viewerCreatorAvatarUploadCompleteRequestScope)
			return
		case errors.Is(err, creatoravatar.ErrUploadIncomplete):
			writeViewerCreatorEntryError(c, http.StatusConflict, "avatar_upload_incomplete", "avatar upload is not complete", viewerCreatorAvatarUploadCompleteRequestScope)
			return
		case errors.Is(err, creatoravatar.ErrUploadExpired):
			writeViewerCreatorEntryError(c, http.StatusConflict, "avatar_upload_expired", "avatar upload has expired", viewerCreatorAvatarUploadCompleteRequestScope)
			return
		default:
			writeInternalServerError(c, viewerCreatorAvatarUploadCompleteRequestScope)
			return
		}
	}

	c.JSON(http.StatusOK, responseEnvelope[viewerCreatorAvatarUploadCompleteResponseData]{
		Data: &viewerCreatorAvatarUploadCompleteResponseData{
			Avatar: mediaAsset{
				DurationSeconds: nil,
				ID:              result.Avatar.AvatarAssetID,
				Kind:            "image",
				PosterURL:       nil,
				URL:             result.Avatar.AvatarURL,
			},
			AvatarUploadToken: result.Avatar.AvatarUploadToken,
		},
		Meta: responseMeta{
			RequestID: newRequestID(viewerCreatorAvatarUploadCompleteRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}
