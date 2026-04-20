package httpserver

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/submissionreview"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	adminSubmissionReviewCaseRequestScope     = "admin_submission_review_case_get"
	adminSubmissionReviewDecisionRequestScope = "admin_submission_review_decision_post"
	adminSubmissionReviewQueueRequestScope    = "admin_submission_review_queue_get"
)

type adminSubmissionReviewDecisionRequest struct {
	DecisionSource string                                      `json:"decisionSource"`
	MainDecision   *adminSubmissionReviewTargetDecisionRequest `json:"mainDecision"`
	ShortDecisions []adminSubmissionReviewShortDecisionRequest `json:"shortDecisions"`
}

type adminSubmissionReviewTargetDecisionRequest struct {
	Decision   string `json:"decision"`
	ReasonCode string `json:"reasonCode"`
	ReviewNote string `json:"reviewNote"`
}

type adminSubmissionReviewShortDecisionRequest struct {
	ShortID    string `json:"shortId"`
	Decision   string `json:"decision"`
	ReasonCode string `json:"reasonCode"`
	ReviewNote string `json:"reviewNote"`
}

type adminSubmissionReviewQueueItemPayload struct {
	Creator              creatorSummary `json:"creator"`
	IntakeID             string         `json:"intakeId"`
	MainDecisionRequired bool           `json:"mainDecisionRequired"`
	PendingShortCount    int64          `json:"pendingShortCount"`
	ShortCount           int64          `json:"shortCount"`
	SubmitKind           string         `json:"submitKind"`
	SubmittedAt          string         `json:"submittedAt"`
}

type adminSubmissionReviewQueueResponseData struct {
	Items []adminSubmissionReviewQueueItemPayload `json:"items"`
}

type adminSubmissionReviewIntakePayload struct {
	CanonicalMainID    string  `json:"canonicalMainId"`
	ConsentConfirmed   bool    `json:"consentConfirmed"`
	CreatorUserID      string  `json:"creatorUserId"`
	ID                 string  `json:"id"`
	MainMediaAssetID   string  `json:"mainMediaAssetId"`
	MainPriceJpy       int64   `json:"mainPriceJpy"`
	OwnershipConfirmed bool    `json:"ownershipConfirmed"`
	PreviousIntakeID   *string `json:"previousIntakeId"`
	Status             string  `json:"status"`
	SubmitKind         string  `json:"submitKind"`
	SubmittedAt        string  `json:"submittedAt"`
}

type adminSubmissionReviewProvenancePayload struct {
	DecisionSource *string `json:"decisionSource"`
	DecisionedAt   *string `json:"decisionedAt"`
	ReasonCode     *string `json:"reasonCode"`
	ReviewNote     *string `json:"reviewNote"`
}

type adminSubmissionReviewDecisionLogPayload struct {
	DecisionSource string  `json:"decisionSource"`
	DecisionedAt   string  `json:"decisionedAt"`
	ReasonCode     *string `json:"reasonCode"`
	ReviewNote     *string `json:"reviewNote"`
	TargetState    string  `json:"targetState"`
}

type adminSubmissionReviewMainPayload struct {
	CurrencyCode      string                                   `json:"currencyCode"`
	DecisionRequired  bool                                     `json:"decisionRequired"`
	ID                string                                   `json:"id"`
	IntakeDecisionLog *adminSubmissionReviewDecisionLogPayload `json:"intakeDecisionLog"`
	Media             mediaAsset                               `json:"media"`
	PriceJpy          int64                                    `json:"priceJpy"`
	Review            adminSubmissionReviewProvenancePayload   `json:"review"`
	State             string                                   `json:"state"`
}

type adminSubmissionReviewShortPayload struct {
	Caption           *string                                  `json:"caption"`
	DecisionRequired  bool                                     `json:"decisionRequired"`
	ID                string                                   `json:"id"`
	IntakeDecisionLog *adminSubmissionReviewDecisionLogPayload `json:"intakeDecisionLog"`
	Media             mediaAsset                               `json:"media"`
	Review            adminSubmissionReviewProvenancePayload   `json:"review"`
	State             string                                   `json:"state"`
}

type adminSubmissionReviewCasePayload struct {
	Creator creatorSummary                      `json:"creator"`
	Intake  adminSubmissionReviewIntakePayload  `json:"intake"`
	Main    adminSubmissionReviewMainPayload    `json:"main"`
	Shorts  []adminSubmissionReviewShortPayload `json:"shorts"`
}

type adminSubmissionReviewCaseResponseData struct {
	Case adminSubmissionReviewCasePayload `json:"case"`
}

func registerAdminSubmissionReviewRoutes(
	router gin.IRouter,
	appEnv string,
	service AdminSubmissionReviewService,
) {
	if appEnv != developmentAppEnv || service == nil {
		return
	}

	adminGroup := router.Group("/api/admin")
	adminGroup.Use(requireAdminLoopback())
	adminGroup.GET("/submission-reviews", func(c *gin.Context) {
		handleAdminSubmissionReviewQueueGet(c, service)
	})
	adminGroup.GET("/submission-reviews/:intakeId", func(c *gin.Context) {
		handleAdminSubmissionReviewCaseGet(c, service)
	})
	adminGroup.POST("/submission-reviews/:intakeId/decision", func(c *gin.Context) {
		handleAdminSubmissionReviewDecisionPost(c, service)
	})
}

func handleAdminSubmissionReviewQueueGet(c *gin.Context, service AdminSubmissionReviewService) {
	items, err := service.ListCases(c.Request.Context())
	if err != nil {
		if writeAdminSubmissionReviewError(c, err, adminSubmissionReviewQueueRequestScope) {
			return
		}
		writeInternalServerError(c, adminSubmissionReviewQueueRequestScope)
		return
	}

	payloadItems := make([]adminSubmissionReviewQueueItemPayload, 0, len(items))
	for _, item := range items {
		payload, buildErr := buildAdminSubmissionReviewQueueItemPayload(item)
		if buildErr != nil {
			writeInternalServerError(c, adminSubmissionReviewQueueRequestScope)
			return
		}
		payloadItems = append(payloadItems, payload)
	}

	c.JSON(http.StatusOK, responseEnvelope[adminSubmissionReviewQueueResponseData]{
		Data: &adminSubmissionReviewQueueResponseData{
			Items: payloadItems,
		},
		Meta: responseMeta{
			RequestID: newRequestID(adminSubmissionReviewQueueRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleAdminSubmissionReviewCaseGet(c *gin.Context, service AdminSubmissionReviewService) {
	intakeID, ok := parseAdminSubmissionReviewIntakeID(c, adminSubmissionReviewCaseRequestScope)
	if !ok {
		return
	}

	reviewCase, err := service.GetCase(c.Request.Context(), intakeID)
	if err != nil {
		if writeAdminSubmissionReviewError(c, err, adminSubmissionReviewCaseRequestScope) {
			return
		}
		writeInternalServerError(c, adminSubmissionReviewCaseRequestScope)
		return
	}

	payload, err := buildAdminSubmissionReviewCasePayload(reviewCase)
	if err != nil {
		writeInternalServerError(c, adminSubmissionReviewCaseRequestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[adminSubmissionReviewCaseResponseData]{
		Data: &adminSubmissionReviewCaseResponseData{Case: payload},
		Meta: responseMeta{
			RequestID: newRequestID(adminSubmissionReviewCaseRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleAdminSubmissionReviewDecisionPost(c *gin.Context, service AdminSubmissionReviewService) {
	intakeID, ok := parseAdminSubmissionReviewIntakeID(c, adminSubmissionReviewDecisionRequestScope)
	if !ok {
		return
	}

	var request adminSubmissionReviewDecisionRequest
	if !decodeViewerCreatorEntryJSON(
		c,
		&request,
		"invalid_request",
		"admin submission review decision request is invalid",
		adminSubmissionReviewDecisionRequestScope,
	) {
		return
	}

	mainDecision := buildAdminSubmissionReviewMainDecisionInput(request.MainDecision)
	shortDecisions, err := buildAdminSubmissionReviewShortDecisionInputs(request.ShortDecisions)
	if err != nil {
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_request", "review short id is invalid", adminSubmissionReviewDecisionRequestScope)
		return
	}

	err = service.ApplyDecision(c.Request.Context(), submissionreview.ReviewDecisionInput{
		IntakeID:       intakeID,
		DecisionSource: request.DecisionSource,
		MainDecision:   mainDecision,
		ShortDecisions: shortDecisions,
	})
	if err != nil {
		if writeAdminSubmissionReviewError(c, err, adminSubmissionReviewDecisionRequestScope) {
			return
		}
		writeInternalServerError(c, adminSubmissionReviewDecisionRequestScope)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseAdminSubmissionReviewIntakeID(c *gin.Context, requestScope string) (uuid.UUID, bool) {
	intakeID, err := uuid.Parse(strings.TrimSpace(c.Param("intakeId")))
	if err != nil {
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_request", "submission review intake id is invalid", requestScope)
		return uuid.Nil, false
	}

	return intakeID, true
}

func buildAdminSubmissionReviewMainDecisionInput(
	request *adminSubmissionReviewTargetDecisionRequest,
) *submissionreview.MainReviewDecisionInput {
	if request == nil {
		return nil
	}

	return &submissionreview.MainReviewDecisionInput{
		Decision:   request.Decision,
		ReasonCode: nullableTrimmedString(request.ReasonCode),
		ReviewNote: nullableTrimmedString(request.ReviewNote),
	}
}

func buildAdminSubmissionReviewShortDecisionInputs(
	requests []adminSubmissionReviewShortDecisionRequest,
) ([]submissionreview.ShortReviewDecisionInput, error) {
	inputs := make([]submissionreview.ShortReviewDecisionInput, 0, len(requests))
	for _, request := range requests {
		shortID, err := uuid.Parse(strings.TrimSpace(request.ShortID))
		if err != nil {
			return nil, err
		}

		inputs = append(inputs, submissionreview.ShortReviewDecisionInput{
			ShortID:    shortID,
			Decision:   request.Decision,
			ReasonCode: nullableTrimmedString(request.ReasonCode),
			ReviewNote: nullableTrimmedString(request.ReviewNote),
		})
	}

	return inputs, nil
}

func buildAdminSubmissionReviewQueueItemPayload(
	item submissionreview.AdminReviewQueueItem,
) (adminSubmissionReviewQueueItemPayload, error) {
	creator, err := buildCreatorSummaryFields(
		item.Creator.UserID,
		item.Creator.DisplayName,
		item.Creator.Handle,
		item.Creator.AvatarURL,
		item.Creator.Bio,
	)
	if err != nil {
		return adminSubmissionReviewQueueItemPayload{}, err
	}

	return adminSubmissionReviewQueueItemPayload{
		Creator:              creator,
		IntakeID:             item.IntakeID.String(),
		MainDecisionRequired: item.MainDecisionRequired,
		PendingShortCount:    item.PendingShortCount,
		ShortCount:           item.ShortCount,
		SubmitKind:           item.SubmitKind,
		SubmittedAt:          formatRequiredRFC3339(item.SubmittedAt),
	}, nil
}

func buildAdminSubmissionReviewCasePayload(
	reviewCase submissionreview.AdminReviewCase,
) (adminSubmissionReviewCasePayload, error) {
	creator, err := buildCreatorSummaryFields(
		reviewCase.Creator.UserID,
		reviewCase.Creator.DisplayName,
		reviewCase.Creator.Handle,
		reviewCase.Creator.AvatarURL,
		reviewCase.Creator.Bio,
	)
	if err != nil {
		return adminSubmissionReviewCasePayload{}, err
	}

	shorts := make([]adminSubmissionReviewShortPayload, 0, len(reviewCase.Shorts))
	for _, item := range reviewCase.Shorts {
		shorts = append(shorts, adminSubmissionReviewShortPayload{
			Caption:           item.Caption,
			DecisionRequired:  item.DecisionRequired,
			ID:                item.ID.String(),
			IntakeDecisionLog: buildAdminSubmissionReviewDecisionLogPayload(item.IntakeDecisionLog),
			Media:             buildVideoMediaAsset(item.Media),
			Review:            buildAdminSubmissionReviewProvenancePayload(item.Review),
			State:             item.State,
		})
	}

	return adminSubmissionReviewCasePayload{
		Creator: creator,
		Intake: adminSubmissionReviewIntakePayload{
			CanonicalMainID:    reviewCase.Intake.CanonicalMainID.String(),
			ConsentConfirmed:   reviewCase.Intake.ConsentConfirmed,
			CreatorUserID:      reviewCase.Intake.CreatorUserID.String(),
			ID:                 reviewCase.Intake.ID.String(),
			MainMediaAssetID:   reviewCase.Intake.MainMediaAssetID.String(),
			MainPriceJpy:       reviewCase.Intake.MainPriceJpy,
			OwnershipConfirmed: reviewCase.Intake.OwnershipConfirmed,
			PreviousIntakeID:   optionalUUIDString(reviewCase.Intake.PreviousIntakeID),
			Status:             reviewCase.Intake.Status,
			SubmitKind:         reviewCase.Intake.SubmitKind,
			SubmittedAt:        formatRequiredRFC3339(reviewCase.Intake.SubmittedAt),
		},
		Main: adminSubmissionReviewMainPayload{
			CurrencyCode:      reviewCase.Main.CurrencyCode,
			DecisionRequired:  reviewCase.Main.DecisionRequired,
			ID:                reviewCase.Main.ID.String(),
			IntakeDecisionLog: buildAdminSubmissionReviewDecisionLogPayload(reviewCase.Main.IntakeDecisionLog),
			Media:             buildVideoMediaAsset(reviewCase.Main.Media),
			PriceJpy:          reviewCase.Main.PriceJpy,
			Review:            buildAdminSubmissionReviewProvenancePayload(reviewCase.Main.Review),
			State:             reviewCase.Main.State,
		},
		Shorts: shorts,
	}, nil
}

func buildAdminSubmissionReviewProvenancePayload(
	provenance submissionreview.ReviewProvenance,
) adminSubmissionReviewProvenancePayload {
	return adminSubmissionReviewProvenancePayload{
		DecisionSource: provenance.DecisionSource,
		DecisionedAt:   formatOptionalRFC3339(provenance.DecisionedAt),
		ReasonCode:     provenance.ReasonCode,
		ReviewNote:     provenance.ReviewNote,
	}
}

func buildAdminSubmissionReviewDecisionLogPayload(
	log *submissionreview.IntakeDecisionLog,
) *adminSubmissionReviewDecisionLogPayload {
	if log == nil {
		return nil
	}

	return &adminSubmissionReviewDecisionLogPayload{
		DecisionSource: log.DecisionSource,
		DecisionedAt:   formatRequiredRFC3339(log.DecisionedAt),
		ReasonCode:     log.ReasonCode,
		ReviewNote:     log.ReviewNote,
		TargetState:    log.TargetState,
	}
}

func writeAdminSubmissionReviewError(c *gin.Context, err error, requestScope string) bool {
	switch {
	case errors.Is(err, submissionreview.ErrInvalidReviewDecision):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_review_decision", "review decision is invalid", requestScope)
	case errors.Is(err, submissionreview.ErrInvalidReviewDecisionSource):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "invalid_review_decision_source", "review decision source is invalid", requestScope)
	case errors.Is(err, submissionreview.ErrReviewDecisionReasonRequired):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "review_reason_required", "review reason is required", requestScope)
	case errors.Is(err, submissionreview.ErrReviewDecisionMetadataConflict):
		writeViewerCreatorEntryError(c, http.StatusBadRequest, "review_decision_metadata_conflict", "review decision metadata is invalid", requestScope)
	case errors.Is(err, submissionreview.ErrSubmissionReviewDecisionTargetsMismatch):
		writeViewerCreatorEntryError(c, http.StatusConflict, "review_targets_mismatch", "review decision targets do not match current intake", requestScope)
	case errors.Is(err, submissionreview.ErrReviewStateConflict):
		writeViewerCreatorEntryError(c, http.StatusConflict, "review_state_conflict", "review case is not in a valid state for this action", requestScope)
	case errors.Is(err, submissionreview.ErrSubmissionReviewIntakeNotFound), errors.Is(err, submissionreview.ErrAdminReviewCaseNotFound):
		writeViewerCreatorEntryError(c, http.StatusNotFound, "not_found", "review case was not found", requestScope)
	default:
		return false
	}

	return true
}

func formatRequiredRFC3339(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

func optionalUUIDString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}

	result := value.String()
	return &result
}
