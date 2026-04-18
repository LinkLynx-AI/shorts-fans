package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/fanmain"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	fanShortUnlockRequestScope     = "fan_short_unlock"
	fanMainPurchaseRequestScope    = "fan_main_purchase"
	fanMainAccessEntryRequestScope = "fan_main_access_entry"
	fanMainPlaybackRequestScope    = "fan_main_playback"
	fanMainAuthRequiredMessage     = "authentication required"
)

type unlockSurfaceResponseData struct {
	Access       mainAccessStatePayload     `json:"access"`
	Creator      creatorSummary             `json:"creator"`
	EntryContext entryContextPayload        `json:"entryContext"`
	Main         unlockMainSummaryPayload   `json:"main"`
	Purchase     unlockPurchaseStatePayload `json:"purchase"`
	Short        unlockShortSummary         `json:"short"`
	UnlockCta    unlockCtaStatePayload      `json:"unlockCta"`
}

type unlockMainSummaryPayload struct {
	DurationSeconds int64  `json:"durationSeconds"`
	ID              string `json:"id"`
	PriceJPY        int64  `json:"priceJpy"`
}

type unlockShortSummary struct {
	Caption                string     `json:"caption"`
	CanonicalMainID        string     `json:"canonicalMainId"`
	CreatorID              string     `json:"creatorId"`
	ID                     string     `json:"id"`
	Media                  mediaAsset `json:"media"`
	PreviewDurationSeconds int64      `json:"previewDurationSeconds"`
}

type entryContextPayload struct {
	AccessEntryPath string `json:"accessEntryPath"`
	PurchasePath    string `json:"purchasePath"`
	Token           string `json:"token"`
}

type purchaseSetupPayload struct {
	Required                bool `json:"required"`
	RequiresAgeConfirmation bool `json:"requiresAgeConfirmation"`
	RequiresCardSetup       bool `json:"requiresCardSetup"`
	RequiresTermsAcceptance bool `json:"requiresTermsAcceptance"`
}

type savedCardSummaryPayload struct {
	Brand           string `json:"brand"`
	Last4           string `json:"last4"`
	PaymentMethodID string `json:"paymentMethodId"`
}

type unlockPurchaseStatePayload struct {
	PendingReason       *string                   `json:"pendingReason"`
	SavedPaymentMethods []savedCardSummaryPayload `json:"savedPaymentMethods"`
	Setup               purchaseSetupPayload      `json:"setup"`
	State               string                    `json:"state"`
	SupportedCardBrands []string                  `json:"supportedCardBrands"`
}

type mainAccessEntryRequestPayload struct {
	EntryToken  string `json:"entryToken"`
	FromShortID string `json:"fromShortId"`
}

type mainAccessEntryResponseData struct {
	Href string `json:"href"`
}

type mainPurchasePaymentMethodPayload struct {
	CardSetupToken  string `json:"cardSetupToken"`
	Mode            string `json:"mode"`
	PaymentMethodID string `json:"paymentMethodId"`
}

type mainPurchaseRequestPayload struct {
	AcceptedAge   bool                             `json:"acceptedAge"`
	AcceptedTerms bool                             `json:"acceptedTerms"`
	EntryToken    string                           `json:"entryToken"`
	FromShortID   string                           `json:"fromShortId"`
	PaymentMethod mainPurchasePaymentMethodPayload `json:"paymentMethod"`
}

type purchaseOutcomePayload struct {
	CanRetry      bool    `json:"canRetry"`
	FailureReason *string `json:"failureReason"`
	Status        string  `json:"status"`
}

type mainPurchaseResponseData struct {
	Access       mainAccessStatePayload `json:"access"`
	EntryContext *entryContextPayload   `json:"entryContext"`
	Purchase     purchaseOutcomePayload `json:"purchase"`
}

type mainAccessStatePayload struct {
	MainID string `json:"mainId"`
	Reason string `json:"reason"`
	Status string `json:"status"`
}

type mainPlaybackResponseData struct {
	Access                mainAccessStatePayload `json:"access"`
	Creator               creatorSummary         `json:"creator"`
	EntryShort            *unlockShortSummary    `json:"entryShort"`
	Main                  mainPlaybackSummary    `json:"main"`
	ResumePositionSeconds *int64                 `json:"resumePositionSeconds"`
}

type mainPlaybackSummary struct {
	DurationSeconds int64      `json:"durationSeconds"`
	ID              string     `json:"id"`
	Media           mediaAsset `json:"media"`
}

func registerFanUnlockMainRoutes(
	router gin.IRouter,
	service FanUnlockMainService,
	shortDisplayAssets ShortDisplayAssetResolver,
	mainDisplayAssets MainDisplayAssetResolver,
	recommendationSignalExposure RecommendationSignalExposureStore,
	viewerBootstrap ViewerBootstrapReader,
) {
	if router == nil || service == nil || shortDisplayAssets == nil || mainDisplayAssets == nil || viewerBootstrap == nil {
		return
	}

	shortUnlockGroup := router.Group("/")
	shortUnlockGroup.Use(buildProtectedFanAuthGuard(viewerBootstrap, fanShortUnlockRequestScope, fanMainAuthRequiredMessage))
	shortUnlockGroup.GET("/api/fan/shorts/:shortId/unlock", func(c *gin.Context) {
		handleFanShortUnlock(c, service, shortDisplayAssets)
	})

	mainPurchaseGroup := router.Group("/")
	mainPurchaseGroup.Use(buildProtectedFanAuthGuard(viewerBootstrap, fanMainPurchaseRequestScope, fanMainAuthRequiredMessage))
	mainPurchaseGroup.POST("/api/fan/mains/:mainId/purchase", func(c *gin.Context) {
		handleFanMainPurchase(c, service)
	})

	mainAccessEntryGroup := router.Group("/")
	mainAccessEntryGroup.Use(buildProtectedFanAuthGuard(viewerBootstrap, fanMainAccessEntryRequestScope, fanMainAuthRequiredMessage))
	mainAccessEntryGroup.POST("/api/fan/mains/:mainId/access-entry", func(c *gin.Context) {
		handleFanMainAccessEntry(c, service)
	})

	mainPlaybackGroup := router.Group("/")
	mainPlaybackGroup.Use(buildProtectedFanAuthGuard(viewerBootstrap, fanMainPlaybackRequestScope, fanMainAuthRequiredMessage))
	mainPlaybackGroup.GET("/api/fan/mains/:mainId/playback", func(c *gin.Context) {
		handleFanMainPlayback(c, service, shortDisplayAssets, mainDisplayAssets, recommendationSignalExposure)
	})
}

func handleFanShortUnlock(
	c *gin.Context,
	service FanUnlockMainService,
	shortDisplayAssets ShortDisplayAssetResolver,
) {
	viewer, sessionBinding, ok := resolveAuthenticatedViewerRequest(c)
	if !ok {
		writeInternalServerError(c, fanShortUnlockRequestScope)
		return
	}

	shortID, err := shorts.ParsePublicShortID(c.Param("shortId"))
	if err != nil {
		writeNotFoundError(c, fanShortUnlockRequestScope, "short was not found")
		return
	}

	unlockSurface, err := service.GetUnlockSurface(c.Request.Context(), viewer.ID, sessionBinding, shortID)
	if err != nil {
		switch {
		case errors.Is(err, fanmain.ErrShortUnlockNotFound):
			writeNotFoundError(c, fanShortUnlockRequestScope, "short was not found")
		case errors.Is(err, fanmain.ErrMainLocked):
			writeFanMainLockedError(c, fanShortUnlockRequestScope, "main is not available for unlock")
		default:
			writeInternalServerError(c, fanShortUnlockRequestScope)
		}
		return
	}

	shortPayload, err := buildUnlockShortSummary(unlockSurface.Short, shortDisplayAssets)
	if err != nil {
		writeInternalServerError(c, fanShortUnlockRequestScope)
		return
	}

	creatorPayload, err := buildCreatorSummaryFields(
		unlockSurface.Creator.ID,
		unlockSurface.Creator.DisplayName,
		unlockSurface.Creator.Handle,
		unlockSurface.Creator.AvatarURL,
		unlockSurface.Creator.Bio,
	)
	if err != nil {
		writeInternalServerError(c, fanShortUnlockRequestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[unlockSurfaceResponseData]{
		Data: &unlockSurfaceResponseData{
			Access:       buildMainAccessStatePayload(unlockSurface.Access),
			Creator:      creatorPayload,
			EntryContext: buildEntryContextPayload(unlockSurface.Main.ID, unlockSurface.MainAccessToken),
			Main: unlockMainSummaryPayload{
				DurationSeconds: unlockSurface.Main.DurationSeconds,
				ID:              mainPublicID(unlockSurface.Main.ID),
				PriceJPY:        unlockSurface.Main.PriceJPY,
			},
			Purchase:  buildUnlockPurchaseStatePayload(unlockSurface.Purchase),
			Short:     shortPayload,
			UnlockCta: buildFanMainUnlockCtaStatePayload(unlockSurface.UnlockCta),
		},
		Meta: responseMeta{
			RequestID: newRequestID(fanShortUnlockRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleFanMainPurchase(c *gin.Context, service FanUnlockMainService) {
	viewer, sessionBinding, ok := resolveAuthenticatedViewerRequest(c)
	if !ok {
		writeInternalServerError(c, fanMainPurchaseRequestScope)
		return
	}

	mainID, err := shorts.ParsePublicMainID(c.Param("mainId"))
	if err != nil {
		writeNotFoundError(c, fanMainPurchaseRequestScope, "main or short was not found")
		return
	}

	var request mainPurchaseRequestPayload
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		writeInvalidFanMainRequest(c, fanMainPurchaseRequestScope, "main purchase request was invalid")
		return
	}

	fromShortID, err := shorts.ParsePublicShortID(request.FromShortID)
	if err != nil {
		writeNotFoundError(c, fanMainPurchaseRequestScope, "main or short was not found")
		return
	}

	result, err := service.PurchaseMain(c.Request.Context(), sessionBinding, fanmain.PurchaseInput{
		AcceptedAge:   request.AcceptedAge,
		AcceptedTerms: request.AcceptedTerms,
		EntryToken:    request.EntryToken,
		FromShortID:   fromShortID,
		MainID:        mainID,
		PaymentMethod: fanmain.PurchasePaymentMethodInput{
			CardSetupToken:  request.PaymentMethod.CardSetupToken,
			Mode:            request.PaymentMethod.Mode,
			PaymentMethodID: request.PaymentMethod.PaymentMethodID,
		},
		ViewerID: viewer.ID,
		ViewerIP: c.ClientIP(),
	})
	if err != nil {
		switch {
		case errors.Is(err, fanmain.ErrInvalidPurchaseRequest):
			writeInvalidFanMainRequest(c, fanMainPurchaseRequestScope, "main purchase request was invalid")
		case errors.Is(err, fanmain.ErrPurchaseNotFound):
			writeNotFoundError(c, fanMainPurchaseRequestScope, "main or short was not found")
		case errors.Is(err, fanmain.ErrMainLocked):
			writeFanMainLockedError(c, fanMainPurchaseRequestScope, "main purchase could not be started")
		default:
			writeInternalServerError(c, fanMainPurchaseRequestScope)
		}
		return
	}

	var entryContext *entryContextPayload
	if result.EntryToken != nil {
		payload := buildEntryContextPayload(mainID, *result.EntryToken)
		entryContext = &payload
	}

	statusCode := http.StatusOK
	if result.Purchase.Status == "pending" {
		statusCode = http.StatusAccepted
	}

	c.JSON(statusCode, responseEnvelope[mainPurchaseResponseData]{
		Data: &mainPurchaseResponseData{
			Access:       buildMainAccessStatePayload(result.Access),
			EntryContext: entryContext,
			Purchase: purchaseOutcomePayload{
				CanRetry:      result.Purchase.CanRetry,
				FailureReason: result.Purchase.FailureReason,
				Status:        result.Purchase.Status,
			},
		},
		Meta: responseMeta{
			RequestID: newRequestID(fanMainPurchaseRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleFanMainAccessEntry(c *gin.Context, service FanUnlockMainService) {
	viewer, sessionBinding, ok := resolveAuthenticatedViewerRequest(c)
	if !ok {
		writeInternalServerError(c, fanMainAccessEntryRequestScope)
		return
	}

	mainID, err := shorts.ParsePublicMainID(c.Param("mainId"))
	if err != nil {
		writeNotFoundError(c, fanMainAccessEntryRequestScope, "main or short was not found")
		return
	}

	var request mainAccessEntryRequestPayload
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		writeInvalidFanMainRequest(c, fanMainAccessEntryRequestScope, "main access entry request was invalid")
		return
	}

	fromShortID, err := shorts.ParsePublicShortID(request.FromShortID)
	if err != nil {
		writeNotFoundError(c, fanMainAccessEntryRequestScope, "main or short was not found")
		return
	}

	issued, err := service.IssueAccessEntry(c.Request.Context(), sessionBinding, fanmain.AccessEntryInput{
		EntryToken:  request.EntryToken,
		FromShortID: fromShortID,
		MainID:      mainID,
		ViewerID:    viewer.ID,
	})
	if err != nil {
		switch {
		case errors.Is(err, fanmain.ErrAccessEntryNotFound):
			writeNotFoundError(c, fanMainAccessEntryRequestScope, "main or short was not found")
		case errors.Is(err, fanmain.ErrMainLocked):
			writeFanMainLockedError(c, fanMainAccessEntryRequestScope, "main access entry could not be issued")
		default:
			writeInternalServerError(c, fanMainAccessEntryRequestScope)
		}
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[mainAccessEntryResponseData]{
		Data: &mainAccessEntryResponseData{
			Href: buildMainPlaybackHref(mainID, fromShortID, issued.GrantToken),
		},
		Meta: responseMeta{
			RequestID: newRequestID(fanMainAccessEntryRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleFanMainPlayback(
	c *gin.Context,
	service FanUnlockMainService,
	shortDisplayAssets ShortDisplayAssetResolver,
	mainDisplayAssets MainDisplayAssetResolver,
	recommendationSignalExposure RecommendationSignalExposureStore,
) {
	viewer, sessionBinding, ok := resolveAuthenticatedViewerRequest(c)
	if !ok {
		writeInternalServerError(c, fanMainPlaybackRequestScope)
		return
	}

	mainID, err := shorts.ParsePublicMainID(c.Param("mainId"))
	if err != nil {
		writeNotFoundError(c, fanMainPlaybackRequestScope, "main was not found")
		return
	}

	fromShortID, err := shorts.ParsePublicShortID(c.Query("fromShortId"))
	if err != nil {
		writeNotFoundError(c, fanMainPlaybackRequestScope, "main was not found")
		return
	}

	grantToken := c.Query("grant")
	if grantToken == "" {
		writeFanMainLockedError(c, fanMainPlaybackRequestScope, "main is not available for playback")
		return
	}

	playback, err := service.GetPlaybackSurface(c.Request.Context(), viewer.ID, sessionBinding, mainID, fromShortID, grantToken)
	if err != nil {
		switch {
		case errors.Is(err, fanmain.ErrPlaybackNotFound):
			writeNotFoundError(c, fanMainPlaybackRequestScope, "main was not found")
		case errors.Is(err, fanmain.ErrMainLocked):
			writeFanMainLockedError(c, fanMainPlaybackRequestScope, "main is not available for playback")
		default:
			writeInternalServerError(c, fanMainPlaybackRequestScope)
		}
		return
	}

	creatorPayload, err := buildCreatorSummaryFields(
		playback.Creator.ID,
		playback.Creator.DisplayName,
		playback.Creator.Handle,
		playback.Creator.AvatarURL,
		playback.Creator.Bio,
	)
	if err != nil {
		writeInternalServerError(c, fanMainPlaybackRequestScope)
		return
	}

	entryShort, err := buildUnlockShortSummary(playback.EntryShort, shortDisplayAssets)
	if err != nil {
		writeInternalServerError(c, fanMainPlaybackRequestScope)
		return
	}

	mainMedia, err := mainDisplayAssets.ResolveMainDisplayAsset(c.Request.Context(), media.MainDisplaySource{
		AssetID:    playback.Main.MediaAssetID,
		MainID:     playback.Main.ID,
		DurationMS: playback.Main.DurationSeconds * 1000,
	}, resolveMainPlaybackBoundary(playback.Access.Status), 0)
	if err != nil {
		writeInternalServerError(c, fanMainPlaybackRequestScope)
		return
	}
	if playback.Access.Status != "owner" && recommendationSignalExposure != nil {
		_ = recommendationSignalExposure.RememberCreatorExposure(c.Request.Context(), viewer.ID, playback.Creator.ID)
	}

	c.JSON(http.StatusOK, responseEnvelope[mainPlaybackResponseData]{
		Data: &mainPlaybackResponseData{
			Access:     buildMainAccessStatePayload(playback.Access),
			Creator:    creatorPayload,
			EntryShort: &entryShort,
			Main: mainPlaybackSummary{
				DurationSeconds: playback.Main.DurationSeconds,
				ID:              mainPublicID(playback.Main.ID),
				Media:           buildVideoMediaAsset(mainMedia),
			},
			ResumePositionSeconds: playback.ResumePositionSeconds,
		},
		Meta: responseMeta{
			RequestID: newRequestID(fanMainPlaybackRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func buildEntryContextPayload(mainID uuid.UUID, token string) entryContextPayload {
	return entryContextPayload{
		AccessEntryPath: buildMainAccessEntryPath(mainID),
		PurchasePath:    buildMainPurchasePath(mainID),
		Token:           token,
	}
}

func buildMainPurchasePath(mainID uuid.UUID) string {
	return fmt.Sprintf("/api/fan/mains/%s/purchase", mainPublicID(mainID))
}

func buildMainAccessEntryPath(mainID uuid.UUID) string {
	return fmt.Sprintf("/api/fan/mains/%s/access-entry", mainPublicID(mainID))
}

func buildMainAccessStatePayload(state fanmain.MainAccessState) mainAccessStatePayload {
	return mainAccessStatePayload{
		MainID: mainPublicID(state.MainID),
		Reason: state.Reason,
		Status: state.Status,
	}
}

func buildFanMainUnlockCtaStatePayload(state fanmain.UnlockCtaState) unlockCtaStatePayload {
	return unlockCtaStatePayload{
		MainDurationSeconds:   state.MainDurationSeconds,
		PriceJPY:              state.PriceJPY,
		ResumePositionSeconds: state.ResumePositionSeconds,
		State:                 state.State,
	}
}

func buildUnlockPurchaseStatePayload(state fanmain.UnlockPurchaseState) unlockPurchaseStatePayload {
	savedMethods := make([]savedCardSummaryPayload, 0, len(state.SavedPaymentMethods))
	for _, method := range state.SavedPaymentMethods {
		savedMethods = append(savedMethods, savedCardSummaryPayload{
			Brand:           method.Brand,
			Last4:           method.Last4,
			PaymentMethodID: method.PaymentMethodID,
		})
	}

	return unlockPurchaseStatePayload{
		PendingReason:       state.PendingReason,
		SavedPaymentMethods: savedMethods,
		Setup: purchaseSetupPayload{
			Required:                state.Setup.Required,
			RequiresAgeConfirmation: state.Setup.RequiresAgeConfirmation,
			RequiresCardSetup:       state.Setup.RequiresCardSetup,
			RequiresTermsAcceptance: state.Setup.RequiresTermsAcceptance,
		},
		State:               state.State,
		SupportedCardBrands: append([]string(nil), state.SupportedCardBrands...),
	}
}

func buildMainPlaybackHref(mainID uuid.UUID, fromShortID uuid.UUID, grantToken string) string {
	searchParams := url.Values{}
	searchParams.Set("fromShortId", shortPublicID(fromShortID))
	searchParams.Set("grant", grantToken)

	return fmt.Sprintf("/mains/%s?%s", mainPublicID(mainID), searchParams.Encode())
}

func buildUnlockShortSummary(source fanmain.ShortSummary, shortDisplayAssets ShortDisplayAssetResolver) (unlockShortSummary, error) {
	displayAsset, err := shortDisplayAssets.ResolveShortDisplayAsset(media.ShortDisplaySource{
		AssetID:    source.MediaAssetID,
		ShortID:    source.ID,
		DurationMS: source.PreviewDurationSeconds * 1000,
	}, media.AccessBoundaryPublic)
	if err != nil {
		return unlockShortSummary{}, err
	}

	return unlockShortSummary{
		Caption:                source.Caption,
		CanonicalMainID:        mainPublicID(source.CanonicalMainID),
		CreatorID:              creatorPublicID(source.CreatorUserID),
		ID:                     shortPublicID(source.ID),
		Media:                  buildVideoMediaAsset(displayAsset),
		PreviewDurationSeconds: source.PreviewDurationSeconds,
	}, nil
}

func creatorPublicID(creatorID uuid.UUID) string {
	return creator.FormatPublicID(creatorID)
}

func resolveAuthenticatedViewerRequest(c *gin.Context) (auth.CurrentViewer, string, bool) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		return auth.CurrentViewer{}, "", false
	}

	rawSessionToken, err := c.Cookie(auth.SessionCookieName)
	if err != nil {
		return auth.CurrentViewer{}, "", false
	}

	return viewer, auth.HashSessionToken(rawSessionToken), true
}

func resolveMainPlaybackBoundary(status string) media.AccessBoundary {
	if status == "owner" {
		return media.AccessBoundaryOwner
	}

	return media.AccessBoundaryPrivate
}

func writeFanMainLockedError(c *gin.Context, requestScope string, message string) {
	c.JSON(http.StatusForbidden, responseEnvelope[struct{}]{
		Data: nil,
		Meta: responseMeta{
			RequestID: newRequestID(requestScope),
			Page:      nil,
		},
		Error: &responseError{
			Code:    "main_locked",
			Message: message,
		},
	})
}

func writeInvalidFanMainRequest(c *gin.Context, requestScope string, message string) {
	c.JSON(http.StatusBadRequest, responseEnvelope[struct{}]{
		Data: nil,
		Meta: responseMeta{
			RequestID: newRequestID(requestScope),
			Page:      nil,
		},
		Error: &responseError{
			Code:    "invalid_request",
			Message: message,
		},
	})
}
