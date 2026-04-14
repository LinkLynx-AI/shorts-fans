package httpserver

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatoravatar"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/viewerprofile"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	viewerProfileAuthRequiredMessage          = "viewer profile requires authentication"
	viewerProfileAvatarUploadCreateScope      = "viewer_profile_avatar_upload_create"
	viewerProfileAvatarUploadCompleteScope    = "viewer_profile_avatar_upload_complete"
	viewerProfileGetRequestScope              = "viewer_profile_get"
	viewerProfileUpdateRequestScope           = "viewer_profile_update"
)

type viewerProfilePayload struct {
	Avatar      *mediaAsset `json:"avatar"`
	DisplayName string      `json:"displayName"`
	Handle      string      `json:"handle"`
}

type viewerProfileResponseData struct {
	Profile viewerProfilePayload `json:"profile"`
}

type viewerProfileUpdateRequest struct {
	AvatarUploadToken string `json:"avatarUploadToken"`
	DisplayName       string `json:"displayName"`
	Handle            string `json:"handle"`
}

func registerViewerProfileRoutes(
	router gin.IRouter,
	reader ViewerProfileReader,
	writer ViewerProfileWriter,
	avatarUploads ViewerCreatorAvatarUploadHandler,
	viewerBootstrap ViewerBootstrapReader,
) {
	if router == nil || viewerBootstrap == nil {
		return
	}

	if reader != nil {
		router.GET(
			"/api/viewer/profile",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerProfileGetRequestScope, viewerProfileAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerProfileGet(c, reader)
			},
		)
	}

	if avatarUploads != nil {
		router.POST(
			"/api/viewer/profile/avatar-uploads",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerProfileAvatarUploadCreateScope, viewerProfileAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerAvatarUploadCreate(c, avatarUploads, viewerProfileAvatarUploadCreateScope)
			},
		)
		router.POST(
			"/api/viewer/profile/avatar-uploads/complete",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerProfileAvatarUploadCompleteScope, viewerProfileAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerAvatarUploadComplete(c, avatarUploads, viewerProfileAvatarUploadCompleteScope)
			},
		)
	}

	if reader != nil && writer != nil {
		router.PUT(
			"/api/viewer/profile",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerProfileUpdateRequestScope, viewerProfileAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerProfileUpdate(c, reader, writer, avatarUploads)
			},
		)
	}
}

func handleViewerProfileGet(c *gin.Context, reader ViewerProfileReader) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeInternalServerError(c, viewerProfileGetRequestScope)
		return
	}

	profile, err := reader.GetProfile(c.Request.Context(), viewerUserID)
	if err != nil {
		if errors.Is(err, viewerprofile.ErrProfileNotFound) {
			writeViewerCreatorEntryError(c, http.StatusNotFound, "not_found", "viewer profile was not found", viewerProfileGetRequestScope)
			return
		}

		writeInternalServerError(c, viewerProfileGetRequestScope)
		return
	}

	payload, err := buildViewerProfilePayload(profile)
	if err != nil {
		writeInternalServerError(c, viewerProfileGetRequestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[viewerProfileResponseData]{
		Data: &viewerProfileResponseData{
			Profile: payload,
		},
		Meta: responseMeta{
			RequestID: newRequestID(viewerProfileGetRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleViewerProfileUpdate(
	c *gin.Context,
	reader ViewerProfileReader,
	writer ViewerProfileWriter,
	avatarUploads ViewerCreatorAvatarUploadHandler,
) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		writeInternalServerError(c, viewerProfileUpdateRequestScope)
		return
	}

	input, avatarUploadToken, ok := buildViewerProfileUpdateInput(c, reader, avatarUploads, viewerProfileUpdateRequestScope)
	if !ok {
		return
	}
	input.UserID = viewer.ID

	if _, err := writer.UpdateProfile(c.Request.Context(), input); err != nil {
		switch {
		case errors.Is(err, viewerprofile.ErrInvalidDisplayName):
			writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_display_name", "display name is invalid", viewerProfileUpdateRequestScope)
			return
		case errors.Is(err, viewerprofile.ErrInvalidHandle):
			writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_handle", "handle is invalid", viewerProfileUpdateRequestScope)
			return
		case errors.Is(err, viewerprofile.ErrHandleAlreadyTaken):
			writeViewerCreatorEntryError(c, http.StatusConflict, "handle_already_taken", "handle is already taken", viewerProfileUpdateRequestScope)
			return
		case errors.Is(err, viewerprofile.ErrProfileNotFound):
			writeViewerCreatorEntryError(c, http.StatusNotFound, "not_found", "viewer profile was not found", viewerProfileUpdateRequestScope)
			return
		default:
			writeInternalServerError(c, viewerProfileUpdateRequestScope)
			return
		}
	}

	if avatarUploadToken != "" {
		if avatarUploads == nil {
			writeInternalServerError(c, viewerProfileUpdateRequestScope)
			return
		}
		if err := avatarUploads.ConsumeCompletedUpload(c.Request.Context(), viewer.ID, avatarUploadToken); err != nil {
			writeInternalServerError(c, viewerProfileUpdateRequestScope)
			return
		}
	}

	c.Status(http.StatusNoContent)
}

func handleViewerAvatarUploadCreate(c *gin.Context, handler ViewerCreatorAvatarUploadHandler, requestScope string) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		writeInternalServerError(c, requestScope)
		return
	}

	var request viewerCreatorAvatarUploadCreateRequest
	if !decodeViewerCreatorEntryJSON(c, &request, "invalid_request", "viewer avatar upload request is invalid", requestScope) {
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
		if errors.As(err, &validationErr) {
			writeViewerCreatorEntryError(c, http.StatusBadRequest, validationErr.Code(), validationErr.Message(), requestScope)
			return
		}

		writeInternalServerError(c, requestScope)
		return
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
			RequestID: newRequestID(requestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleViewerAvatarUploadComplete(c *gin.Context, handler ViewerCreatorAvatarUploadHandler, requestScope string) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		writeInternalServerError(c, requestScope)
		return
	}

	var request viewerCreatorAvatarUploadCompleteRequest
	if !decodeViewerCreatorEntryJSON(c, &request, "invalid_request", "viewer avatar upload completion request is invalid", requestScope) {
		return
	}

	result, err := handler.CompleteUpload(c.Request.Context(), creatoravatar.CompleteUploadInput{
		AvatarUploadToken: request.AvatarUploadToken,
		ViewerUserID:      viewer.ID,
	})
	if err != nil {
		switch {
		case errors.Is(err, creatoravatar.ErrUploadNotFound):
			writeViewerCreatorEntryError(c, http.StatusNotFound, "avatar_upload_not_found", "avatar upload was not found", requestScope)
			return
		case errors.Is(err, creatoravatar.ErrUploadIncomplete):
			writeViewerCreatorEntryError(c, http.StatusConflict, "avatar_upload_incomplete", "avatar upload is not complete", requestScope)
			return
		case errors.Is(err, creatoravatar.ErrUploadExpired):
			writeViewerCreatorEntryError(c, http.StatusConflict, "avatar_upload_expired", "avatar upload has expired", requestScope)
			return
		default:
			writeInternalServerError(c, requestScope)
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
			RequestID: newRequestID(requestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func buildViewerProfileUpdateInput(
	c *gin.Context,
	reader ViewerProfileReader,
	avatarUploads ViewerCreatorAvatarUploadHandler,
	requestScope string,
) (viewerprofile.UpdateProfileInput, string, bool) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeInternalServerError(c, requestScope)
		return viewerprofile.UpdateProfileInput{}, "", false
	}

	var request viewerProfileUpdateRequest
	if !decodeViewerCreatorEntryJSON(c, &request, "invalid_request", "viewer profile request is invalid", requestScope) {
		return viewerprofile.UpdateProfileInput{}, "", false
	}

	currentProfile, err := reader.GetProfile(c.Request.Context(), viewerUserID)
	if err != nil {
		if errors.Is(err, viewerprofile.ErrProfileNotFound) {
			writeViewerCreatorEntryError(c, http.StatusNotFound, "not_found", "viewer profile was not found", requestScope)
			return viewerprofile.UpdateProfileInput{}, "", false
		}

		writeInternalServerError(c, requestScope)
		return viewerprofile.UpdateProfileInput{}, "", false
	}

	avatarURL, avatarUploadToken, ok := resolveViewerProfileAvatarURL(
		c,
		avatarUploads,
		viewerUserID,
		currentProfile.AvatarURL,
		request.AvatarUploadToken,
		requestScope,
	)
	if !ok {
		return viewerprofile.UpdateProfileInput{}, "", false
	}

	return viewerprofile.UpdateProfileInput{
		AvatarURL:   avatarURL,
		DisplayName: request.DisplayName,
		Handle:      request.Handle,
		UserID:      viewerUserID,
	}, avatarUploadToken, true
}

func resolveViewerProfileAvatarURL(
	c *gin.Context,
	avatarUploads ViewerCreatorAvatarUploadHandler,
	viewerUserID uuid.UUID,
	currentAvatarURL *string,
	rawAvatarUploadToken string,
	requestScope string,
) (*string, string, bool) {
	avatarUploadToken := strings.TrimSpace(rawAvatarUploadToken)
	if rawAvatarUploadToken != "" && avatarUploadToken == "" {
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_avatar_upload_token", "avatar upload token is invalid", requestScope)
		return nil, "", false
	}
	if avatarUploadToken == "" {
		return currentAvatarURL, "", true
	}
	if avatarUploads == nil {
		writeInternalServerError(c, requestScope)
		return nil, "", false
	}

	completedAvatar, err := avatarUploads.ResolveCompletedUpload(c.Request.Context(), viewerUserID, avatarUploadToken)
	if err != nil {
		switch {
		case errors.Is(err, creatoravatar.ErrUploadNotFound),
			errors.Is(err, creatoravatar.ErrUploadIncomplete),
			errors.Is(err, creatoravatar.ErrUploadExpired),
			errors.Is(err, creatoravatar.ErrUploadConsumed):
			writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_avatar_upload_token", "avatar upload token is invalid", requestScope)
			return nil, "", false
		default:
			writeInternalServerError(c, requestScope)
			return nil, "", false
		}
	}

	return &completedAvatar.AvatarURL, avatarUploadToken, true
}

func buildViewerProfilePayload(profile viewerprofile.Profile) (viewerProfilePayload, error) {
	payload := viewerProfilePayload{
		Avatar:      nil,
		DisplayName: profile.DisplayName,
		Handle:      "@" + profile.Handle,
	}
	if profile.AvatarURL == nil {
		return payload, nil
	}

	assetID := fmt.Sprintf("asset_viewer_profile_avatar_%s", strings.ReplaceAll(profile.UserID.String(), "-", ""))
	payload.Avatar = &mediaAsset{
		DurationSeconds: nil,
		ID:              assetID,
		Kind:            "image",
		PosterURL:       nil,
		URL:             *profile.AvatarURL,
	}

	return payload, nil
}
