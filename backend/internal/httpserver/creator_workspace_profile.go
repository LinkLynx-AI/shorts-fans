package httpserver

import (
	"errors"
	"net/http"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/viewerprofile"
	"github.com/gin-gonic/gin"
)

func handleCreatorWorkspaceProfileUpdate(
	c *gin.Context,
	writer CreatorWorkspaceProfileWriter,
	avatarUploads ViewerCreatorAvatarUploadHandler,
) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeCreatorWorkspaceError(c, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
		return
	}

	var request creatorWorkspaceProfileUpdateRequest
	if !decodeViewerCreatorEntryJSON(
		c,
		&request,
		"invalid_request",
		"creator workspace profile request is invalid",
		creatorWorkspaceProfileUpdateRequestScope,
	) {
		return
	}

	currentProfile, err := writer.GetProfile(c.Request.Context(), viewerUserID)
	if err != nil {
		if errors.Is(err, viewerprofile.ErrProfileNotFound) {
			writeCreatorWorkspaceError(c, http.StatusNotFound, "not_found", "creator workspace was not found")
			return
		}

		writeCreatorWorkspaceError(c, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
		return
	}

	avatarURL, avatarUploadToken, ok := resolveViewerProfileAvatarURL(
		c,
		avatarUploads,
		viewerUserID,
		currentProfile.AvatarURL,
		request.AvatarUploadToken,
		creatorWorkspaceProfileUpdateRequestScope,
	)
	if !ok {
		return
	}

	if _, err := writer.UpdateCreatorProfileSync(c.Request.Context(), viewerprofile.UpdateProfileInput{
		AvatarURL:   avatarURL,
		DisplayName: request.DisplayName,
		Handle:      request.Handle,
		UserID:      viewerUserID,
	}, request.Bio); err != nil {
		switch {
		case errors.Is(err, viewerprofile.ErrInvalidDisplayName):
			writeCreatorWorkspaceError(c, http.StatusBadRequest, "invalid_display_name", "display name is invalid")
			return
		case errors.Is(err, viewerprofile.ErrInvalidHandle):
			writeCreatorWorkspaceError(c, http.StatusBadRequest, "invalid_handle", "handle is invalid")
			return
		case errors.Is(err, viewerprofile.ErrHandleAlreadyTaken):
			writeCreatorWorkspaceError(c, http.StatusConflict, "handle_already_taken", "handle is already taken")
			return
		case errors.Is(err, viewerprofile.ErrCreatorProfileNotFound):
			writeCreatorWorkspaceError(c, http.StatusForbidden, "creator_mode_unavailable", "creator mode is not available")
			return
		case errors.Is(err, viewerprofile.ErrProfileNotFound):
			writeCreatorWorkspaceError(c, http.StatusNotFound, "not_found", "creator workspace was not found")
			return
		default:
			writeCreatorWorkspaceError(c, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
			return
		}
	}

	if avatarUploadToken != "" {
		if avatarUploads == nil {
			writeCreatorWorkspaceError(c, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
			return
		}
		if err := avatarUploads.ConsumeCompletedUpload(c.Request.Context(), viewerUserID, avatarUploadToken); err != nil {
			writeCreatorWorkspaceError(c, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
			return
		}
	}

	c.Status(http.StatusNoContent)
}
