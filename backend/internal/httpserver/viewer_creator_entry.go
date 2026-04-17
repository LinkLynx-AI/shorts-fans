package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatorregistration"
	"github.com/gin-gonic/gin"
)

const (
	viewerActiveModeAuthRequiredMessage                   = "viewer mode switch requires authentication"
	viewerActiveModeRequestScope                          = "viewer_active_mode"
	viewerCreatorRegistrationAuthRequiredMessage          = "creator registration requires authentication"
	viewerCreatorRegistrationEvidenceCreateRequestScope   = "viewer_creator_registration_evidence_upload_create"
	viewerCreatorRegistrationEvidenceCompleteRequestScope = "viewer_creator_registration_evidence_upload_complete"
	viewerCreatorRegistrationGetRequestScope              = "viewer_creator_registration_get"
	viewerCreatorRegistrationIntakeGetRequestScope        = "viewer_creator_registration_intake_get"
	viewerCreatorRegistrationIntakePutRequestScope        = "viewer_creator_registration_intake_put"
	viewerCreatorRegistrationSubmitRequestScope           = "viewer_creator_registration_submit"
)

type viewerCreatorRegistrationRequest struct{}

type viewerCreatorRegistrationIntakeRequest struct {
	AcceptsConsentResponsibility bool   `json:"acceptsConsentResponsibility"`
	BirthDate                    string `json:"birthDate"`
	CreatorBio                   string `json:"creatorBio"`
	DeclaresNoProhibitedCategory bool   `json:"declaresNoProhibitedCategory"`
	LegalName                    string `json:"legalName"`
	PayoutRecipientName          string `json:"payoutRecipientName"`
	PayoutRecipientType          string `json:"payoutRecipientType"`
}

type viewerCreatorRegistrationEvidenceUploadCreateRequest struct {
	FileName      string `json:"fileName"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	Kind          string `json:"kind"`
	MimeType      string `json:"mimeType"`
}

type viewerCreatorRegistrationEvidenceUploadCompleteRequest struct {
	EvidenceUploadToken string `json:"evidenceUploadToken"`
}

type viewerCreatorRegistrationActionsPayload struct {
	CanEnterCreatorMode bool `json:"canEnterCreatorMode"`
	CanResubmit         bool `json:"canResubmit"`
	CanSubmit           bool `json:"canSubmit"`
}

type viewerCreatorRegistrationCreatorDraftPayload struct {
	Bio string `json:"bio"`
}

type viewerCreatorRegistrationEvidencePayload struct {
	FileName      string `json:"fileName"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	Kind          string `json:"kind"`
	MimeType      string `json:"mimeType"`
	UploadedAt    string `json:"uploadedAt"`
}

type viewerCreatorRegistrationEvidenceUploadCreateResponseData struct {
	EvidenceKind        string                                `json:"evidenceKind"`
	EvidenceUploadToken string                                `json:"evidenceUploadToken"`
	ExpiresAt           string                                `json:"expiresAt"`
	UploadTarget        viewerCreatorRegistrationUploadTarget `json:"uploadTarget"`
}

type viewerCreatorRegistrationEvidenceUploadCompleteResponseData struct {
	Evidence            viewerCreatorRegistrationEvidencePayload `json:"evidence"`
	EvidenceKind        string                                   `json:"evidenceKind"`
	EvidenceUploadToken string                                   `json:"evidenceUploadToken"`
}

type viewerCreatorRegistrationGetResponseData struct {
	Registration *viewerCreatorRegistrationPayload `json:"registration"`
}

type viewerCreatorRegistrationIntakePayload struct {
	AcceptsConsentResponsibility bool                                       `json:"acceptsConsentResponsibility"`
	BirthDate                    *string                                    `json:"birthDate"`
	CanSubmit                    bool                                       `json:"canSubmit"`
	CreatorBio                   string                                     `json:"creatorBio"`
	DeclaresNoProhibitedCategory bool                                       `json:"declaresNoProhibitedCategory"`
	Evidences                    []viewerCreatorRegistrationEvidencePayload `json:"evidences"`
	IsReadOnly                   bool                                       `json:"isReadOnly"`
	LegalName                    string                                     `json:"legalName"`
	PayoutRecipientName          string                                     `json:"payoutRecipientName"`
	PayoutRecipientType          *string                                    `json:"payoutRecipientType"`
	RegistrationState            *string                                    `json:"registrationState"`
	SharedProfile                viewerProfilePayload                       `json:"sharedProfile"`
}

type viewerCreatorRegistrationIntakeResponseData struct {
	Intake viewerCreatorRegistrationIntakePayload `json:"intake"`
}

type viewerCreatorRegistrationPayload struct {
	Actions       viewerCreatorRegistrationActionsPayload      `json:"actions"`
	CreatorDraft  viewerCreatorRegistrationCreatorDraftPayload `json:"creatorDraft"`
	Rejection     *viewerCreatorRegistrationRejectionPayload   `json:"rejection"`
	Review        viewerCreatorRegistrationReviewPayload       `json:"review"`
	SharedProfile viewerProfilePayload                         `json:"sharedProfile"`
	State         string                                       `json:"state"`
	Surface       viewerCreatorRegistrationSurfacePayload      `json:"surface"`
}

type viewerCreatorRegistrationRejectionPayload struct {
	IsResubmitEligible      bool    `json:"isResubmitEligible"`
	IsSupportReviewRequired bool    `json:"isSupportReviewRequired"`
	ReasonCode              *string `json:"reasonCode"`
	SelfServeResubmitCount  int32   `json:"selfServeResubmitCount"`
	SelfServeResubmitRemain int32   `json:"selfServeResubmitRemaining"`
}

type viewerCreatorRegistrationReviewPayload struct {
	ApprovedAt  *string `json:"approvedAt"`
	RejectedAt  *string `json:"rejectedAt"`
	SubmittedAt *string `json:"submittedAt"`
	SuspendedAt *string `json:"suspendedAt"`
}

type viewerCreatorRegistrationSurfacePayload struct {
	Kind             string  `json:"kind"`
	WorkspacePreview *string `json:"workspacePreview"`
}

type viewerCreatorRegistrationUploadTarget struct {
	FileName string                        `json:"fileName"`
	MimeType string                        `json:"mimeType"`
	Upload   viewerCreatorAvatarDirectLink `json:"upload"`
}

type viewerActiveModeRequest struct {
	ActiveMode string `json:"activeMode"`
}

func registerViewerCreatorEntryRoutes(
	router gin.IRouter,
	registrationService ViewerCreatorRegistrationService,
	evidenceUploadHandler ViewerCreatorRegistrationEvidenceUploadHandler,
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

	if evidenceUploadHandler != nil {
		viewerGroup.POST(
			"/creator-registration/evidence-uploads",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerCreatorRegistrationEvidenceCreateRequestScope, viewerCreatorRegistrationAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerCreatorRegistrationEvidenceUploadCreate(c, evidenceUploadHandler)
			},
		)
		viewerGroup.POST(
			"/creator-registration/evidence-uploads/complete",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerCreatorRegistrationEvidenceCompleteRequestScope, viewerCreatorRegistrationAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerCreatorRegistrationEvidenceUploadComplete(c, evidenceUploadHandler)
			},
		)
	}

	if registrationService != nil {
		viewerGroup.GET(
			"/creator-registration",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerCreatorRegistrationGetRequestScope, viewerCreatorRegistrationAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerCreatorRegistrationGet(c, registrationService)
			},
		)
		viewerGroup.GET(
			"/creator-registration/intake",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerCreatorRegistrationIntakeGetRequestScope, viewerCreatorRegistrationAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerCreatorRegistrationIntakeGet(c, registrationService)
			},
		)
		viewerGroup.PUT(
			"/creator-registration/intake",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerCreatorRegistrationIntakePutRequestScope, viewerCreatorRegistrationAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerCreatorRegistrationIntakePut(c, registrationService)
			},
		)
		viewerGroup.POST(
			"/creator-registration",
			buildProtectedFanAuthGuard(viewerBootstrap, viewerCreatorRegistrationSubmitRequestScope, viewerCreatorRegistrationAuthRequiredMessage),
			func(c *gin.Context) {
				handleViewerCreatorRegistrationSubmit(c, registrationService)
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

func handleViewerCreatorRegistrationGet(c *gin.Context, service ViewerCreatorRegistrationService) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeInternalServerError(c, viewerCreatorRegistrationGetRequestScope)
		return
	}

	registration, err := service.GetRegistration(c.Request.Context(), viewerUserID)
	if err != nil {
		if errors.Is(err, creatorregistration.ErrSharedProfileNotFound) {
			writeViewerCreatorEntryError(c, http.StatusNotFound, "not_found", "viewer profile was not found", viewerCreatorRegistrationGetRequestScope)
			return
		}
		writeInternalServerError(c, viewerCreatorRegistrationGetRequestScope)
		return
	}

	var payload *viewerCreatorRegistrationPayload
	if registration != nil {
		built, err := buildViewerCreatorRegistrationPayload(*registration)
		if err != nil {
			writeInternalServerError(c, viewerCreatorRegistrationGetRequestScope)
			return
		}
		payload = &built
	}

	c.JSON(http.StatusOK, responseEnvelope[viewerCreatorRegistrationGetResponseData]{
		Data: &viewerCreatorRegistrationGetResponseData{
			Registration: payload,
		},
		Meta: responseMeta{
			RequestID: newRequestID(viewerCreatorRegistrationGetRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleViewerCreatorRegistrationIntakeGet(c *gin.Context, service ViewerCreatorRegistrationService) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeInternalServerError(c, viewerCreatorRegistrationIntakeGetRequestScope)
		return
	}

	intake, err := service.GetIntake(c.Request.Context(), viewerUserID)
	if err != nil {
		if errors.Is(err, creatorregistration.ErrSharedProfileNotFound) {
			writeViewerCreatorEntryError(c, http.StatusNotFound, "not_found", "viewer profile was not found", viewerCreatorRegistrationIntakeGetRequestScope)
			return
		}
		writeInternalServerError(c, viewerCreatorRegistrationIntakeGetRequestScope)
		return
	}

	payload, err := buildViewerCreatorRegistrationIntakePayload(intake)
	if err != nil {
		writeInternalServerError(c, viewerCreatorRegistrationIntakeGetRequestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[viewerCreatorRegistrationIntakeResponseData]{
		Data: &viewerCreatorRegistrationIntakeResponseData{
			Intake: payload,
		},
		Meta: responseMeta{
			RequestID: newRequestID(viewerCreatorRegistrationIntakeGetRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleViewerCreatorRegistrationIntakePut(c *gin.Context, service ViewerCreatorRegistrationService) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeInternalServerError(c, viewerCreatorRegistrationIntakePutRequestScope)
		return
	}

	var request viewerCreatorRegistrationIntakeRequest
	if !decodeViewerCreatorEntryJSON(
		c,
		&request,
		"invalid_request",
		"creator registration intake request is invalid",
		viewerCreatorRegistrationIntakePutRequestScope,
	) {
		return
	}

	intake, err := service.SaveIntake(c.Request.Context(), creatorregistration.SaveIntakeInput{
		AcceptsConsentResponsibility: request.AcceptsConsentResponsibility,
		BirthDate:                    request.BirthDate,
		CreatorBio:                   request.CreatorBio,
		DeclaresNoProhibitedCategory: request.DeclaresNoProhibitedCategory,
		LegalName:                    request.LegalName,
		PayoutRecipientName:          request.PayoutRecipientName,
		PayoutRecipientType:          request.PayoutRecipientType,
		UserID:                       viewerUserID,
	})
	if err != nil {
		if writeViewerCreatorRegistrationError(c, err, viewerCreatorRegistrationIntakePutRequestScope) {
			return
		}
		writeInternalServerError(c, viewerCreatorRegistrationIntakePutRequestScope)
		return
	}

	payload, err := buildViewerCreatorRegistrationIntakePayload(intake)
	if err != nil {
		writeInternalServerError(c, viewerCreatorRegistrationIntakePutRequestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[viewerCreatorRegistrationIntakeResponseData]{
		Data: &viewerCreatorRegistrationIntakeResponseData{
			Intake: payload,
		},
		Meta: responseMeta{
			RequestID: newRequestID(viewerCreatorRegistrationIntakePutRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleViewerCreatorRegistrationSubmit(c *gin.Context, service ViewerCreatorRegistrationService) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeInternalServerError(c, viewerCreatorRegistrationSubmitRequestScope)
		return
	}

	var request viewerCreatorRegistrationRequest
	if !decodeViewerCreatorEntryJSON(
		c,
		&request,
		"invalid_request",
		"creator registration request is invalid",
		viewerCreatorRegistrationSubmitRequestScope,
	) {
		return
	}

	if _, err := service.Submit(c.Request.Context(), viewerUserID); err != nil {
		if writeViewerCreatorRegistrationError(c, err, viewerCreatorRegistrationSubmitRequestScope) {
			return
		}
		writeInternalServerError(c, viewerCreatorRegistrationSubmitRequestScope)
		return
	}

	c.Status(http.StatusNoContent)
}

func handleViewerCreatorRegistrationEvidenceUploadCreate(
	c *gin.Context,
	handler ViewerCreatorRegistrationEvidenceUploadHandler,
) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeInternalServerError(c, viewerCreatorRegistrationEvidenceCreateRequestScope)
		return
	}

	var request viewerCreatorRegistrationEvidenceUploadCreateRequest
	if !decodeViewerCreatorEntryJSON(
		c,
		&request,
		"invalid_request",
		"creator registration evidence upload request is invalid",
		viewerCreatorRegistrationEvidenceCreateRequestScope,
	) {
		return
	}

	result, err := handler.CreateUpload(c.Request.Context(), creatorregistration.CreateEvidenceUploadInput{
		FileName:      request.FileName,
		FileSizeBytes: request.FileSizeBytes,
		Kind:          request.Kind,
		MimeType:      request.MimeType,
		ViewerUserID:  viewerUserID,
	})
	if err != nil {
		var validationErr *creatorregistration.ValidationError
		if errors.As(err, &validationErr) {
			writeViewerCreatorEntryError(c, http.StatusBadRequest, validationErr.Code(), validationErr.Message(), viewerCreatorRegistrationEvidenceCreateRequestScope)
			return
		}
		if writeViewerCreatorRegistrationError(c, err, viewerCreatorRegistrationEvidenceCreateRequestScope) {
			return
		}
		writeInternalServerError(c, viewerCreatorRegistrationEvidenceCreateRequestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[viewerCreatorRegistrationEvidenceUploadCreateResponseData]{
		Data: &viewerCreatorRegistrationEvidenceUploadCreateResponseData{
			EvidenceKind:        result.EvidenceKind,
			EvidenceUploadToken: result.EvidenceUploadToken,
			ExpiresAt:           result.ExpiresAt.Format(time.RFC3339),
			UploadTarget: viewerCreatorRegistrationUploadTarget{
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
			RequestID: newRequestID(viewerCreatorRegistrationEvidenceCreateRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleViewerCreatorRegistrationEvidenceUploadComplete(
	c *gin.Context,
	handler ViewerCreatorRegistrationEvidenceUploadHandler,
) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeInternalServerError(c, viewerCreatorRegistrationEvidenceCompleteRequestScope)
		return
	}

	var request viewerCreatorRegistrationEvidenceUploadCompleteRequest
	if !decodeViewerCreatorEntryJSON(
		c,
		&request,
		"invalid_request",
		"creator registration evidence upload completion request is invalid",
		viewerCreatorRegistrationEvidenceCompleteRequestScope,
	) {
		return
	}

	result, err := handler.CompleteUpload(c.Request.Context(), creatorregistration.CompleteEvidenceUploadInput{
		EvidenceUploadToken: request.EvidenceUploadToken,
		ViewerUserID:        viewerUserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, creatorregistration.ErrEvidenceUploadNotFound):
			writeViewerCreatorEntryError(c, http.StatusNotFound, "evidence_upload_not_found", "evidence upload was not found", viewerCreatorRegistrationEvidenceCompleteRequestScope)
			return
		case errors.Is(err, creatorregistration.ErrEvidenceUploadIncomplete):
			writeViewerCreatorEntryError(c, http.StatusConflict, "evidence_upload_incomplete", "evidence upload is not complete", viewerCreatorRegistrationEvidenceCompleteRequestScope)
			return
		case errors.Is(err, creatorregistration.ErrEvidenceUploadExpired):
			writeViewerCreatorEntryError(c, http.StatusConflict, "evidence_upload_expired", "evidence upload has expired", viewerCreatorRegistrationEvidenceCompleteRequestScope)
			return
		default:
			if writeViewerCreatorRegistrationError(c, err, viewerCreatorRegistrationEvidenceCompleteRequestScope) {
				return
			}
			writeInternalServerError(c, viewerCreatorRegistrationEvidenceCompleteRequestScope)
			return
		}
	}

	payload := buildViewerCreatorRegistrationEvidencePayload(result.Evidence)
	c.JSON(http.StatusOK, responseEnvelope[viewerCreatorRegistrationEvidenceUploadCompleteResponseData]{
		Data: &viewerCreatorRegistrationEvidenceUploadCompleteResponseData{
			Evidence:            payload,
			EvidenceKind:        result.EvidenceKind,
			EvidenceUploadToken: result.EvidenceUploadToken,
		},
		Meta: responseMeta{
			RequestID: newRequestID(viewerCreatorRegistrationEvidenceCompleteRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
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

func buildViewerCreatorRegistrationEvidencePayload(evidence creatorregistration.Evidence) viewerCreatorRegistrationEvidencePayload {
	return viewerCreatorRegistrationEvidencePayload{
		FileName:      evidence.FileName,
		FileSizeBytes: evidence.FileSizeBytes,
		Kind:          evidence.Kind,
		MimeType:      evidence.MimeType,
		UploadedAt:    evidence.UploadedAt.UTC().Format(time.RFC3339),
	}
}

func buildViewerCreatorRegistrationIntakePayload(intake creatorregistration.Intake) (viewerCreatorRegistrationIntakePayload, error) {
	sharedProfile, err := buildViewerCreatorRegistrationSharedProfilePayload(intake.SharedProfile)
	if err != nil {
		return viewerCreatorRegistrationIntakePayload{}, err
	}

	evidences := make([]viewerCreatorRegistrationEvidencePayload, 0, len(intake.Evidences))
	for _, evidence := range intake.Evidences {
		evidences = append(evidences, buildViewerCreatorRegistrationEvidencePayload(evidence))
	}

	return viewerCreatorRegistrationIntakePayload{
		AcceptsConsentResponsibility: intake.AcceptsConsentResponsibility,
		BirthDate:                    nullableTrimmedString(intake.BirthDate),
		CanSubmit:                    intake.CanSubmit,
		CreatorBio:                   intake.CreatorBio,
		DeclaresNoProhibitedCategory: intake.DeclaresNoProhibitedCategory,
		Evidences:                    evidences,
		IsReadOnly:                   intake.IsReadOnly,
		LegalName:                    intake.LegalName,
		PayoutRecipientName:          intake.PayoutRecipientName,
		PayoutRecipientType:          nullableTrimmedString(intake.PayoutRecipientType),
		RegistrationState:            intake.RegistrationState,
		SharedProfile:                sharedProfile,
	}, nil
}

func buildViewerCreatorRegistrationPayload(registration creatorregistration.Registration) (viewerCreatorRegistrationPayload, error) {
	sharedProfile, err := buildViewerCreatorRegistrationSharedProfilePayload(registration.SharedProfile)
	if err != nil {
		return viewerCreatorRegistrationPayload{}, err
	}

	return viewerCreatorRegistrationPayload{
		Actions: viewerCreatorRegistrationActionsPayload{
			CanEnterCreatorMode: registration.Actions.CanEnterCreatorMode,
			CanResubmit:         registration.Actions.CanResubmit,
			CanSubmit:           registration.Actions.CanSubmit,
		},
		CreatorDraft: viewerCreatorRegistrationCreatorDraftPayload{
			Bio: registration.CreatorDraft.Bio,
		},
		Rejection: buildViewerCreatorRegistrationRejectionPayload(registration.Rejection),
		Review: viewerCreatorRegistrationReviewPayload{
			ApprovedAt:  formatOptionalRFC3339(registration.Review.ApprovedAt),
			RejectedAt:  formatOptionalRFC3339(registration.Review.RejectedAt),
			SubmittedAt: formatOptionalRFC3339(registration.Review.SubmittedAt),
			SuspendedAt: formatOptionalRFC3339(registration.Review.SuspendedAt),
		},
		SharedProfile: sharedProfile,
		State:         registration.State,
		Surface: viewerCreatorRegistrationSurfacePayload{
			Kind:             registration.Surface.Kind,
			WorkspacePreview: registration.Surface.WorkspacePreview,
		},
	}, nil
}

func buildViewerCreatorRegistrationRejectionPayload(rejection *creatorregistration.Rejection) *viewerCreatorRegistrationRejectionPayload {
	if rejection == nil {
		return nil
	}

	return &viewerCreatorRegistrationRejectionPayload{
		IsResubmitEligible:      rejection.IsResubmitEligible,
		IsSupportReviewRequired: rejection.IsSupportReviewRequired,
		ReasonCode:              rejection.ReasonCode,
		SelfServeResubmitCount:  rejection.SelfServeResubmitCount,
		SelfServeResubmitRemain: rejection.SelfServeResubmitRemain,
	}
}

func buildViewerCreatorRegistrationSharedProfilePayload(profile creatorregistration.SharedProfilePreview) (viewerProfilePayload, error) {
	payload := viewerProfilePayload{
		Avatar:      nil,
		DisplayName: profile.DisplayName,
		Handle:      "@" + strings.TrimPrefix(profile.Handle, "@"),
	}
	if profile.AvatarURL == nil {
		return payload, nil
	}

	payload.Avatar = &mediaAsset{
		DurationSeconds: nil,
		ID:              fmt.Sprintf("asset_viewer_profile_avatar_%s", strings.ReplaceAll(profile.UserID.String(), "-", "")),
		Kind:            "image",
		PosterURL:       nil,
		URL:             *profile.AvatarURL,
	}

	return payload, nil
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

func formatOptionalRFC3339(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}

func nullableTrimmedString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
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

func writeViewerCreatorRegistrationError(c *gin.Context, err error, requestScope string) bool {
	switch {
	case errors.Is(err, creatorregistration.ErrInvalidBirthDate):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_birth_date", "birth date is invalid", requestScope)
	case errors.Is(err, creatorregistration.ErrInvalidDisplayName):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_display_name", "display name is invalid", requestScope)
	case errors.Is(err, creatorregistration.ErrInvalidHandle):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_handle", "handle is invalid", requestScope)
	case errors.Is(err, creatorregistration.ErrInvalidLegalName):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_legal_name", "legal name is invalid", requestScope)
	case errors.Is(err, creatorregistration.ErrInvalidPayoutRecipient):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_payout_recipient_name", "payout recipient name is invalid", requestScope)
	case errors.Is(err, creatorregistration.ErrInvalidPayoutRecipientTyp):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_payout_recipient_type", "payout recipient type is invalid", requestScope)
	case errors.Is(err, creatorregistration.ErrHandleAlreadyTaken):
		writeViewerCreatorEntryError(c, http.StatusConflict, "handle_already_taken", "handle is already taken", requestScope)
	case errors.Is(err, creatorregistration.ErrRegistrationIncomplete):
		writeViewerCreatorEntryError(c, http.StatusConflict, "registration_incomplete", "registration intake is incomplete", requestScope)
	case errors.Is(err, creatorregistration.ErrRegistrationStateConflict):
		writeViewerCreatorEntryError(c, http.StatusConflict, "registration_state_conflict", "creator registration state does not allow this action", requestScope)
	case errors.Is(err, creatorregistration.ErrSharedProfileNotFound):
		writeViewerCreatorEntryError(c, http.StatusNotFound, "not_found", "viewer profile was not found", requestScope)
	default:
		return false
	}

	return true
}
