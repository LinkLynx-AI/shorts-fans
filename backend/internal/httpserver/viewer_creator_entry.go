package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatoravatar"
	"github.com/gin-gonic/gin"
)

const (
	viewerActiveModeAuthRequiredMessage          = "viewer mode switch requires authentication"
	viewerActiveModeRequestScope                 = "viewer_active_mode"
	viewerCreatorRegistrationAuthRequiredMessage = "creator registration requires authentication"
	viewerCreatorRegistrationRequestScope        = "viewer_creator_registration"
)

type viewerCreatorRegistrationRequest struct {
	AvatarUploadToken string `json:"avatarUploadToken"`
	Bio               string `json:"bio"`
	DisplayName       string `json:"displayName"`
	Handle            string `json:"handle"`
}

type viewerActiveModeRequest struct {
	ActiveMode string `json:"activeMode"`
}

func registerViewerCreatorEntryRoutes(
	router gin.IRouter,
	registrationWriter ViewerCreatorRegistrationWriter,
	avatarUploadHandler ViewerCreatorAvatarUploadHandler,
	activeModeSwitcher ViewerActiveModeSwitcher,
	viewerBootstrap ViewerBootstrapReader,
) {
	if viewerBootstrap == nil {
		return
	}

	viewerGroup := router.Group("/api/viewer")

	if avatarUploadHandler != nil {
		viewerGroup.POST(
			"/creator-registration/avatar-uploads",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerCreatorAvatarUploadCreateRequestScope, viewerCreatorRegistrationAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerCreatorAvatarUploadCreate(c, avatarUploadHandler)
			},
		)
		viewerGroup.POST(
			"/creator-registration/avatar-uploads/complete",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerCreatorAvatarUploadCompleteRequestScope, viewerCreatorRegistrationAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerCreatorAvatarUploadComplete(c, avatarUploadHandler)
			},
		)
	}

	if registrationWriter != nil {
		viewerGroup.POST(
			"/creator-registration",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerCreatorRegistrationRequestScope, viewerCreatorRegistrationAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerCreatorRegistration(c, registrationWriter, avatarUploadHandler)
			},
		)
	}

	if activeModeSwitcher != nil {
		viewerGroup.PUT(
			"/active-mode",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerActiveModeRequestScope, viewerActiveModeAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerActiveModeSwitch(c, activeModeSwitcher)
			},
		)
	}
}

func handleViewerCreatorRegistration(
	c *gin.Context,
	writer ViewerCreatorRegistrationWriter,
	avatarUploads ViewerCreatorAvatarUploadHandler,
) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		writeInternalServerError(c, viewerCreatorRegistrationRequestScope)
		return
	}

	var request viewerCreatorRegistrationRequest
	if !decodeViewerCreatorEntryJSON(c, &request, "invalid_request", "creator registration request is invalid", viewerCreatorRegistrationRequestScope) {
		return
	}

	var avatarURL *string
	avatarUploadToken := strings.TrimSpace(request.AvatarUploadToken)
	if request.AvatarUploadToken != "" && avatarUploadToken == "" {
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_avatar_upload_token", "avatar upload token is invalid", viewerCreatorRegistrationRequestScope)
		return
	}
	if avatarUploadToken != "" {
		if avatarUploads == nil {
			writeInternalServerError(c, viewerCreatorRegistrationRequestScope)
			return
		}

		completedAvatar, err := avatarUploads.ResolveCompletedUpload(c.Request.Context(), viewer.ID, avatarUploadToken)
		if err != nil {
			switch {
			case errors.Is(err, creatoravatar.ErrUploadNotFound),
				errors.Is(err, creatoravatar.ErrUploadIncomplete),
				errors.Is(err, creatoravatar.ErrUploadExpired),
				errors.Is(err, creatoravatar.ErrUploadConsumed):
				writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_avatar_upload_token", "avatar upload token is invalid", viewerCreatorRegistrationRequestScope)
				return
			default:
				writeInternalServerError(c, viewerCreatorRegistrationRequestScope)
				return
			}
		}

		avatarURL = &completedAvatar.AvatarURL
	}

	_, err := writer.RegisterApprovedCreator(c.Request.Context(), creator.SelfServeRegistrationInput{
		AvatarURL:   avatarURL,
		UserID:      viewer.ID,
		DisplayName: request.DisplayName,
		Handle:      request.Handle,
		Bio:         request.Bio,
	})
	if err != nil {
		switch {
		case errors.Is(err, creator.ErrInvalidDisplayName):
			writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_display_name", "display name is invalid", viewerCreatorRegistrationRequestScope)
			return
		case errors.Is(err, creator.ErrInvalidHandle):
			writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_handle", "handle is invalid", viewerCreatorRegistrationRequestScope)
			return
		case errors.Is(err, creator.ErrHandleAlreadyTaken):
			writeViewerCreatorEntryError(c, http.StatusConflict, "handle_already_taken", "handle is already taken", viewerCreatorRegistrationRequestScope)
			return
		default:
			writeInternalServerError(c, viewerCreatorRegistrationRequestScope)
			return
		}
	}

	if avatarUploadToken != "" {
		if err := avatarUploads.ConsumeCompletedUpload(c.Request.Context(), viewer.ID, avatarUploadToken); err != nil {
			writeInternalServerError(c, viewerCreatorRegistrationRequestScope)
			return
		}
	}

	c.Status(http.StatusNoContent)
}

func handleViewerActiveModeSwitch(c *gin.Context, switcher ViewerActiveModeSwitcher) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		writeInternalServerError(c, viewerActiveModeRequestScope)
		return
	}

	var request viewerActiveModeRequest
	if !decodeViewerCreatorEntryJSON(c, &request, "invalid_request", "active mode request is invalid", viewerActiveModeRequestScope) {
		return
	}

	resolvedMode := strings.TrimSpace(request.ActiveMode)
	if resolvedMode != string(auth.ActiveModeFan) && resolvedMode != string(auth.ActiveModeCreator) {
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_active_mode", "active mode is invalid", viewerActiveModeRequestScope)
		return
	}

	if resolvedMode == string(auth.ActiveModeCreator) && !viewer.CanAccessCreatorMode {
		writeViewerCreatorEntryError(c, http.StatusForbidden, "creator_mode_unavailable", "creator mode is not available", viewerActiveModeRequestScope)
		return
	}

	rawSessionToken, err := c.Cookie(auth.SessionCookieName)
	if err != nil {
		writeInternalServerError(c, viewerActiveModeRequestScope)
		return
	}

	if err := switcher.SwitchActiveMode(c.Request.Context(), rawSessionToken, auth.ActiveMode(resolvedMode)); err != nil {
		if errors.Is(err, auth.ErrInvalidActiveMode) {
			writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_active_mode", "active mode is invalid", viewerActiveModeRequestScope)
			return
		}

		writeInternalServerError(c, viewerActiveModeRequestScope)
		return
	}

	c.Status(http.StatusNoContent)
}

func decodeViewerCreatorEntryJSON[T any](
	c *gin.Context,
	target *T,
	code string,
	message string,
	requestScope string,
) bool {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeViewerCreatorEntryError(c, http.StatusBadRequest, code, message, requestScope)
		return false
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != nil && !errors.Is(err, io.EOF) {
		writeViewerCreatorEntryError(c, http.StatusBadRequest, code, message, requestScope)
		return false
	}
	if len(extra) > 0 {
		writeViewerCreatorEntryError(c, http.StatusBadRequest, code, message, requestScope)
		return false
	}

	return true
}

func writeViewerCreatorEntryError(c *gin.Context, status int, code string, message string, requestScope string) {
	c.JSON(status, responseEnvelope[struct{}]{
		Data: nil,
		Meta: responseMeta{
			Page:      nil,
			RequestID: newRequestID(requestScope),
		},
		Error: &responseError{
			Code:    code,
			Message: message,
		},
	})
}
