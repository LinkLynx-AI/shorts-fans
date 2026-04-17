package httpserver

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatorregistration"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	adminCreatorReviewCaseRequestScope     = "admin_creator_review_case_get"
	adminCreatorReviewDecisionRequestScope = "admin_creator_review_decision_post"
	adminCreatorReviewQueueRequestScope    = "admin_creator_review_queue_get"
)

type adminCreatorReviewDecisionRequest struct {
	Decision                string `json:"decision"`
	IsResubmitEligible      bool   `json:"isResubmitEligible"`
	IsSupportReviewRequired bool   `json:"isSupportReviewRequired"`
	ReasonCode              string `json:"reasonCode"`
}

type adminCreatorReviewEvidencePayload struct {
	AccessURL     string `json:"accessUrl"`
	FileName      string `json:"fileName"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	Kind          string `json:"kind"`
	MimeType      string `json:"mimeType"`
	UploadedAt    string `json:"uploadedAt"`
}

type adminCreatorReviewIntakePayload struct {
	AcceptsConsentResponsibility bool    `json:"acceptsConsentResponsibility"`
	BirthDate                    *string `json:"birthDate"`
	DeclaresNoProhibitedCategory bool    `json:"declaresNoProhibitedCategory"`
	LegalName                    string  `json:"legalName"`
	PayoutRecipientName          string  `json:"payoutRecipientName"`
	PayoutRecipientType          *string `json:"payoutRecipientType"`
}

type adminCreatorReviewQueueItemPayload struct {
	CreatorBio    string                                 `json:"creatorBio"`
	LegalName     string                                 `json:"legalName"`
	Review        viewerCreatorRegistrationReviewPayload `json:"review"`
	SharedProfile viewerProfilePayload                   `json:"sharedProfile"`
	State         string                                 `json:"state"`
	UserID        string                                 `json:"userId"`
}

type adminCreatorReviewQueueResponseData struct {
	Items []adminCreatorReviewQueueItemPayload `json:"items"`
	State string                               `json:"state"`
}

type adminCreatorReviewCasePayload struct {
	CreatorBio    string                                     `json:"creatorBio"`
	Evidences     []adminCreatorReviewEvidencePayload        `json:"evidences"`
	Intake        adminCreatorReviewIntakePayload            `json:"intake"`
	Rejection     *viewerCreatorRegistrationRejectionPayload `json:"rejection"`
	Review        viewerCreatorRegistrationReviewPayload     `json:"review"`
	SharedProfile viewerProfilePayload                       `json:"sharedProfile"`
	State         string                                     `json:"state"`
	UserID        string                                     `json:"userId"`
}

type adminCreatorReviewCaseResponseData struct {
	Case adminCreatorReviewCasePayload `json:"case"`
}

func registerAdminCreatorReviewRoutes(
	router gin.IRouter,
	appEnv string,
	service AdminCreatorReviewService,
) {
	if appEnv != developmentAppEnv || service == nil {
		return
	}

	adminGroup := router.Group("/api/admin")
	adminGroup.Use(requireAdminLoopback())
	adminGroup.GET("/creator-reviews", func(c *gin.Context) {
		handleAdminCreatorReviewQueueGet(c, service)
	})
	adminGroup.GET("/creator-reviews/:userId", func(c *gin.Context) {
		handleAdminCreatorReviewCaseGet(c, service)
	})
	adminGroup.POST("/creator-reviews/:userId/decision", func(c *gin.Context) {
		handleAdminCreatorReviewDecisionPost(c, service)
	})
}

func requireAdminLoopback() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isLoopbackRemoteAddr(c.Request.RemoteAddr) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Next()
	}
}

func isLoopbackRemoteAddr(remoteAddr string) bool {
	candidate := strings.TrimSpace(remoteAddr)
	if candidate == "" {
		return false
	}

	if host, _, err := net.SplitHostPort(candidate); err == nil {
		candidate = host
	}

	candidate = strings.Trim(candidate, "[]")
	ip := net.ParseIP(candidate)
	return ip != nil && ip.IsLoopback()
}

func handleAdminCreatorReviewQueueGet(c *gin.Context, service AdminCreatorReviewService) {
	state := strings.TrimSpace(c.DefaultQuery("state", creatorregistration.StateSubmitted))
	items, err := service.ListCases(c.Request.Context(), state)
	if err != nil {
		if writeAdminCreatorReviewError(c, err, adminCreatorReviewQueueRequestScope) {
			return
		}
		writeInternalServerError(c, adminCreatorReviewQueueRequestScope)
		return
	}

	payloadItems := make([]adminCreatorReviewQueueItemPayload, 0, len(items))
	for _, item := range items {
		payload, buildErr := buildAdminCreatorReviewQueueItemPayload(item)
		if buildErr != nil {
			writeInternalServerError(c, adminCreatorReviewQueueRequestScope)
			return
		}
		payloadItems = append(payloadItems, payload)
	}

	c.JSON(http.StatusOK, responseEnvelope[adminCreatorReviewQueueResponseData]{
		Data: &adminCreatorReviewQueueResponseData{
			Items: payloadItems,
			State: state,
		},
		Meta: responseMeta{
			RequestID: newRequestID(adminCreatorReviewQueueRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleAdminCreatorReviewCaseGet(c *gin.Context, service AdminCreatorReviewService) {
	userID, ok := parseAdminCreatorReviewUserID(c, adminCreatorReviewCaseRequestScope)
	if !ok {
		return
	}

	reviewCase, err := service.GetCase(c.Request.Context(), userID)
	if err != nil {
		if writeAdminCreatorReviewError(c, err, adminCreatorReviewCaseRequestScope) {
			return
		}
		writeInternalServerError(c, adminCreatorReviewCaseRequestScope)
		return
	}

	payload, err := buildAdminCreatorReviewCasePayload(reviewCase)
	if err != nil {
		writeInternalServerError(c, adminCreatorReviewCaseRequestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[adminCreatorReviewCaseResponseData]{
		Data: &adminCreatorReviewCaseResponseData{
			Case: payload,
		},
		Meta: responseMeta{
			RequestID: newRequestID(adminCreatorReviewCaseRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleAdminCreatorReviewDecisionPost(c *gin.Context, service AdminCreatorReviewService) {
	userID, ok := parseAdminCreatorReviewUserID(c, adminCreatorReviewDecisionRequestScope)
	if !ok {
		return
	}

	var request adminCreatorReviewDecisionRequest
	if !decodeViewerCreatorEntryJSON(
		c,
		&request,
		"invalid_request",
		"admin creator review decision request is invalid",
		adminCreatorReviewDecisionRequestScope,
	) {
		return
	}

	reasonCode := nullableTrimmedString(request.ReasonCode)
	reviewCase, err := service.ApplyDecision(c.Request.Context(), creatorregistration.ReviewDecisionInput{
		Decision:                request.Decision,
		IsResubmitEligible:      request.IsResubmitEligible,
		IsSupportReviewRequired: request.IsSupportReviewRequired,
		ReasonCode:              reasonCode,
		UserID:                  userID,
	})
	if err != nil {
		if writeAdminCreatorReviewError(c, err, adminCreatorReviewDecisionRequestScope) {
			return
		}
		writeInternalServerError(c, adminCreatorReviewDecisionRequestScope)
		return
	}

	payload, err := buildAdminCreatorReviewCasePayload(reviewCase)
	if err != nil {
		writeInternalServerError(c, adminCreatorReviewDecisionRequestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[adminCreatorReviewCaseResponseData]{
		Data: &adminCreatorReviewCaseResponseData{
			Case: payload,
		},
		Meta: responseMeta{
			RequestID: newRequestID(adminCreatorReviewDecisionRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func parseAdminCreatorReviewUserID(c *gin.Context, requestScope string) (uuid.UUID, bool) {
	userID, err := uuid.Parse(strings.TrimSpace(c.Param("userId")))
	if err != nil {
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_request", "review case user id is invalid", requestScope)
		return uuid.Nil, false
	}

	return userID, true
}

func buildAdminCreatorReviewQueueItemPayload(item creatorregistration.ReviewQueueItem) (adminCreatorReviewQueueItemPayload, error) {
	sharedProfile, err := buildViewerCreatorRegistrationSharedProfilePayload(item.SharedProfile)
	if err != nil {
		return adminCreatorReviewQueueItemPayload{}, err
	}

	return adminCreatorReviewQueueItemPayload{
		CreatorBio: item.CreatorBio,
		LegalName:  item.LegalName,
		Review: viewerCreatorRegistrationReviewPayload{
			ApprovedAt:  formatOptionalRFC3339(item.Review.ApprovedAt),
			RejectedAt:  formatOptionalRFC3339(item.Review.RejectedAt),
			SubmittedAt: formatOptionalRFC3339(item.Review.SubmittedAt),
			SuspendedAt: formatOptionalRFC3339(item.Review.SuspendedAt),
		},
		SharedProfile: sharedProfile,
		State:         item.State,
		UserID:        item.UserID.String(),
	}, nil
}

func buildAdminCreatorReviewCasePayload(reviewCase creatorregistration.ReviewCase) (adminCreatorReviewCasePayload, error) {
	sharedProfile, err := buildViewerCreatorRegistrationSharedProfilePayload(reviewCase.SharedProfile)
	if err != nil {
		return adminCreatorReviewCasePayload{}, err
	}

	evidences := make([]adminCreatorReviewEvidencePayload, 0, len(reviewCase.Evidences))
	for _, evidence := range reviewCase.Evidences {
		evidences = append(evidences, adminCreatorReviewEvidencePayload{
			AccessURL:     evidence.AccessURL,
			FileName:      evidence.FileName,
			FileSizeBytes: evidence.FileSizeBytes,
			Kind:          evidence.Kind,
			MimeType:      evidence.MimeType,
			UploadedAt:    evidence.UploadedAt.UTC().Format(time.RFC3339),
		})
	}

	return adminCreatorReviewCasePayload{
		CreatorBio: reviewCase.CreatorBio,
		Evidences:  evidences,
		Intake: adminCreatorReviewIntakePayload{
			AcceptsConsentResponsibility: reviewCase.Intake.AcceptsConsentResponsibility,
			BirthDate:                    nullableTrimmedString(reviewCase.Intake.BirthDate),
			DeclaresNoProhibitedCategory: reviewCase.Intake.DeclaresNoProhibitedCategory,
			LegalName:                    reviewCase.Intake.LegalName,
			PayoutRecipientName:          reviewCase.Intake.PayoutRecipientName,
			PayoutRecipientType:          nullableTrimmedString(reviewCase.Intake.PayoutRecipientType),
		},
		Rejection: buildViewerCreatorRegistrationRejectionPayload(reviewCase.Rejection),
		Review: viewerCreatorRegistrationReviewPayload{
			ApprovedAt:  formatOptionalRFC3339(reviewCase.Review.ApprovedAt),
			RejectedAt:  formatOptionalRFC3339(reviewCase.Review.RejectedAt),
			SubmittedAt: formatOptionalRFC3339(reviewCase.Review.SubmittedAt),
			SuspendedAt: formatOptionalRFC3339(reviewCase.Review.SuspendedAt),
		},
		SharedProfile: sharedProfile,
		State:         reviewCase.State,
		UserID:        reviewCase.UserID.String(),
	}, nil
}

func writeAdminCreatorReviewError(c *gin.Context, err error, requestScope string) bool {
	switch {
	case errors.Is(err, creatorregistration.ErrInvalidReviewState):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_review_state", "review state is invalid", requestScope)
	case errors.Is(err, creatorregistration.ErrInvalidReviewDecision):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_review_decision", "review decision is invalid", requestScope)
	case errors.Is(err, creatorregistration.ErrReviewDecisionReasonRequired):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "review_reason_required", "review rejection reason is required", requestScope)
	case errors.Is(err, creatorregistration.ErrReviewDecisionMetadataConflict):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "review_decision_metadata_conflict", "review decision metadata is invalid", requestScope)
	case errors.Is(err, creatorregistration.ErrRegistrationStateConflict):
		writeViewerCreatorEntryError(c, http.StatusConflict, "review_state_conflict", "review case is not in a valid state for this action", requestScope)
	case errors.Is(err, creatorregistration.ErrReviewCaseNotFound), errors.Is(err, creatorregistration.ErrSharedProfileNotFound):
		writeViewerCreatorEntryError(c, http.StatusNotFound, "not_found", "review case was not found", requestScope)
	default:
		return false
	}

	return true
}
