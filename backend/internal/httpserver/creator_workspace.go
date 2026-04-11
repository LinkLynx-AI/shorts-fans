package httpserver

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	creatorWorkspaceAuthRequiredMessage      = "creator workspace requires authentication"
	creatorWorkspaceMainDetailRequestScope   = "creator_workspace_main_detail"
	creatorWorkspaceMainListRequestScope     = "creator_workspace_mains"
	creatorWorkspaceMainPriceRequestScope    = "creator_workspace_main_price"
	creatorWorkspaceRequestScope             = "creator_workspace"
	creatorWorkspaceShortCaptionRequestScope = "creator_workspace_short_caption"
	creatorWorkspaceShortDetailRequestScope  = "creator_workspace_short_detail"
	creatorWorkspaceShortListRequestScope    = "creator_workspace_shorts"
	creatorWorkspaceTopPerformersScope       = "creator_workspace_top_performers"
)

type creatorWorkspaceResponseData struct {
	Workspace creatorWorkspacePayload `json:"workspace"`
}

type creatorWorkspacePayload struct {
	Creator                  creatorSummary                          `json:"creator"`
	OverviewMetrics          creatorWorkspaceOverviewMetrics         `json:"overviewMetrics"`
	RevisionRequestedSummary *creatorWorkspaceRevisionRequestedCount `json:"revisionRequestedSummary"`
}

type creatorWorkspaceOverviewMetrics struct {
	GrossUnlockRevenueJpy int64 `json:"grossUnlockRevenueJpy"`
	UnlockCount           int64 `json:"unlockCount"`
	UniquePurchaserCount  int64 `json:"uniquePurchaserCount"`
}

type creatorWorkspaceRevisionRequestedCount struct {
	MainCount  int64 `json:"mainCount"`
	ShortCount int64 `json:"shortCount"`
	TotalCount int64 `json:"totalCount"`
}

type creatorWorkspacePreviewShortListResponseData struct {
	Items []creatorWorkspacePreviewShortItem `json:"items"`
}

type creatorWorkspacePreviewMainListResponseData struct {
	Items []creatorWorkspacePreviewMainItem `json:"items"`
}

type creatorWorkspacePreviewShortDetailResponseData struct {
	Preview creatorWorkspacePreviewShortDetailPayload `json:"preview"`
}

type creatorWorkspacePreviewMainDetailResponseData struct {
	Preview creatorWorkspacePreviewMainDetailPayload `json:"preview"`
}

type creatorWorkspaceTopPerformersResponseData struct {
	TopPerformers creatorWorkspaceTopPerformersPayload `json:"topPerformers"`
}

type creatorWorkspaceTopPerformersPayload struct {
	TopMain  *creatorWorkspaceTopMainPerformer  `json:"topMain"`
	TopShort *creatorWorkspaceTopShortPerformer `json:"topShort"`
}

type creatorWorkspaceTopMainPerformer struct {
	ID          string                            `json:"id"`
	Media       creatorWorkspacePreviewMediaAsset `json:"media"`
	UnlockCount int64                             `json:"unlockCount"`
}

type creatorWorkspaceTopShortPerformer struct {
	AttributedUnlockCount int64                             `json:"attributedUnlockCount"`
	ID                    string                            `json:"id"`
	Media                 creatorWorkspacePreviewMediaAsset `json:"media"`
}

type creatorWorkspacePreviewShortItem struct {
	CanonicalMainID        string                            `json:"canonicalMainId"`
	ID                     string                            `json:"id"`
	Media                  creatorWorkspacePreviewMediaAsset `json:"media"`
	PreviewDurationSeconds int64                             `json:"previewDurationSeconds"`
}

type creatorWorkspacePreviewMainItem struct {
	DurationSeconds int64                             `json:"durationSeconds"`
	ID              string                            `json:"id"`
	LeadShortID     string                            `json:"leadShortId"`
	Media           creatorWorkspacePreviewMediaAsset `json:"media"`
	PriceJpy        int64                             `json:"priceJpy"`
}

type creatorWorkspaceOwnerPreviewAccessPayload struct {
	MainID string `json:"mainId"`
	Reason string `json:"reason"`
	Status string `json:"status"`
}

type creatorWorkspacePreviewShortDetailPayload struct {
	Access  creatorWorkspaceOwnerPreviewAccessPayload `json:"access"`
	Creator creatorSummary                            `json:"creator"`
	Short   creatorWorkspacePreviewShortDetailItem    `json:"short"`
}

type creatorWorkspacePreviewMainDetailPayload struct {
	Access     creatorWorkspaceOwnerPreviewAccessPayload `json:"access"`
	Creator    creatorSummary                            `json:"creator"`
	EntryShort creatorWorkspacePreviewShortDetailItem    `json:"entryShort"`
	Main       creatorWorkspacePreviewMainDetailItem     `json:"main"`
}

type creatorWorkspacePreviewShortDetailItem struct {
	Caption                string     `json:"caption"`
	CanonicalMainID        string     `json:"canonicalMainId"`
	CreatorID              string     `json:"creatorId"`
	ID                     string     `json:"id"`
	Media                  mediaAsset `json:"media"`
	PreviewDurationSeconds int64      `json:"previewDurationSeconds"`
}

type creatorWorkspacePreviewMainDetailItem struct {
	DurationSeconds int64      `json:"durationSeconds"`
	ID              string     `json:"id"`
	Media           mediaAsset `json:"media"`
	PriceJpy        int64      `json:"priceJpy"`
}

type creatorWorkspacePreviewMediaAsset struct {
	DurationSeconds int64  `json:"durationSeconds"`
	ID              string `json:"id"`
	Kind            string `json:"kind"`
	PosterURL       string `json:"posterUrl"`
}

type creatorWorkspacePreviewCursorPayload struct {
	CreatedAt string `json:"createdAt"`
	ID        string `json:"id"`
}

type creatorWorkspaceUpdateMainPriceRequest struct {
	PriceJpy int64 `json:"priceJpy"`
}

type creatorWorkspaceUpdateShortCaptionRequest struct {
	Caption creatorWorkspaceNullableStringRequestField `json:"caption"`
}

type creatorWorkspaceNullableStringRequestField struct {
	Present bool
	Value   *string
}

func (f *creatorWorkspaceNullableStringRequestField) UnmarshalJSON(data []byte) error {
	f.Present = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		f.Value = nil
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	f.Value = &value
	return nil
}

// registerCreatorWorkspaceRoutes は creator private workspace summary API を router に登録します。
func registerCreatorWorkspaceRoutes(
	router gin.IRouter,
	reader CreatorWorkspaceReader,
	writer CreatorWorkspaceWriter,
	viewerBootstrap ViewerBootstrapReader,
) {
	if router == nil || viewerBootstrap == nil {
		return
	}

	if reader != nil {
		router.GET(
			"/api/creator/workspace",
			buildProtectedFanAuthGuard(viewerBootstrap, creatorWorkspaceRequestScope, creatorWorkspaceAuthRequiredMessage),
			func(c *gin.Context) {
				handleCreatorWorkspace(c, reader)
			},
		)
		router.GET(
			"/api/creator/workspace/shorts",
			buildProtectedFanAuthGuard(viewerBootstrap, creatorWorkspaceShortListRequestScope, creatorWorkspaceAuthRequiredMessage),
			func(c *gin.Context) {
				handleCreatorWorkspacePreviewShorts(c, reader)
			},
		)
		router.GET(
			"/api/creator/workspace/shorts/:shortId/preview",
			buildProtectedFanAuthGuard(viewerBootstrap, creatorWorkspaceShortDetailRequestScope, creatorWorkspaceAuthRequiredMessage),
			func(c *gin.Context) {
				handleCreatorWorkspacePreviewShortDetail(c, reader)
			},
		)
		router.GET(
			"/api/creator/workspace/mains",
			buildProtectedFanAuthGuard(viewerBootstrap, creatorWorkspaceMainListRequestScope, creatorWorkspaceAuthRequiredMessage),
			func(c *gin.Context) {
				handleCreatorWorkspacePreviewMains(c, reader)
			},
		)
		router.GET(
			"/api/creator/workspace/mains/:mainId/preview",
			buildProtectedFanAuthGuard(viewerBootstrap, creatorWorkspaceMainDetailRequestScope, creatorWorkspaceAuthRequiredMessage),
			func(c *gin.Context) {
				handleCreatorWorkspacePreviewMainDetail(c, reader)
			},
		)
		router.GET(
			"/api/creator/workspace/top-performers",
			buildProtectedFanAuthGuard(viewerBootstrap, creatorWorkspaceTopPerformersScope, creatorWorkspaceAuthRequiredMessage),
			func(c *gin.Context) {
				handleCreatorWorkspaceTopPerformers(c, reader)
			},
		)
	}

	if writer != nil {
		router.PUT(
			"/api/creator/workspace/mains/:mainId/price",
			buildProtectedFanAuthGuard(viewerBootstrap, creatorWorkspaceMainPriceRequestScope, creatorWorkspaceAuthRequiredMessage),
			func(c *gin.Context) {
				handleCreatorWorkspaceMainPriceUpdate(c, writer)
			},
		)
		router.PUT(
			"/api/creator/workspace/shorts/:shortId/caption",
			buildProtectedFanAuthGuard(viewerBootstrap, creatorWorkspaceShortCaptionRequestScope, creatorWorkspaceAuthRequiredMessage),
			func(c *gin.Context) {
				handleCreatorWorkspaceShortCaptionUpdate(c, writer)
			},
		)
	}
}

func handleCreatorWorkspace(c *gin.Context, reader CreatorWorkspaceReader) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeCreatorWorkspaceError(c, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
		return
	}

	workspace, err := reader.GetWorkspace(c.Request.Context(), viewerUserID)
	if err != nil {
		switch {
		case errors.Is(err, creator.ErrCreatorModeUnavailable):
			writeCreatorWorkspaceError(c, http.StatusForbidden, "creator_mode_unavailable", "creator mode is not available")
			return
		case errors.Is(err, creator.ErrProfileNotFound):
			writeCreatorWorkspaceError(c, http.StatusNotFound, "not_found", "creator workspace was not found")
			return
		default:
			writeCreatorWorkspaceError(c, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
			return
		}
	}

	creatorSummary, err := buildCreatorSummary(workspace.Creator)
	if err != nil {
		writeCreatorWorkspaceError(c, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[creatorWorkspaceResponseData]{
		Data: &creatorWorkspaceResponseData{
			Workspace: creatorWorkspacePayload{
				Creator: creatorSummary,
				OverviewMetrics: creatorWorkspaceOverviewMetrics{
					GrossUnlockRevenueJpy: workspace.OverviewMetrics.GrossUnlockRevenueJpy,
					UnlockCount:           workspace.OverviewMetrics.UnlockCount,
					UniquePurchaserCount:  workspace.OverviewMetrics.UniquePurchaserCount,
				},
				RevisionRequestedSummary: buildCreatorWorkspaceRevisionSummary(workspace.RevisionRequestedSummary),
			},
		},
		Meta: responseMeta{
			RequestID: newRequestID(creatorWorkspaceRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func buildCreatorWorkspaceRevisionSummary(summary *creator.RevisionRequestedSummary) *creatorWorkspaceRevisionRequestedCount {
	if summary == nil {
		return nil
	}

	return &creatorWorkspaceRevisionRequestedCount{
		MainCount:  summary.MainCount,
		ShortCount: summary.ShortCount,
		TotalCount: summary.TotalCount,
	}
}

func handleCreatorWorkspacePreviewShorts(c *gin.Context, reader CreatorWorkspaceReader) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeCreatorWorkspaceListError(c, creatorWorkspaceShortListRequestScope, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
		return
	}

	cursor := decodeCreatorWorkspacePreviewCursor(strings.TrimSpace(c.Query("cursor")))
	items, nextCursor, err := reader.ListWorkspacePreviewShorts(
		c.Request.Context(),
		viewerUserID,
		cursor,
		creator.DefaultWorkspacePreviewPageSize,
	)
	if err != nil {
		writeCreatorWorkspaceReadError(c, creatorWorkspaceShortListRequestScope, err)
		return
	}

	responseItems := make([]creatorWorkspacePreviewShortItem, 0, len(items))
	for _, item := range items {
		responseItem, buildErr := buildCreatorWorkspacePreviewShortItem(item)
		if buildErr != nil {
			writeCreatorWorkspaceListError(c, creatorWorkspaceShortListRequestScope, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
			return
		}
		responseItems = append(responseItems, responseItem)
	}

	c.JSON(http.StatusOK, responseEnvelope[creatorWorkspacePreviewShortListResponseData]{
		Data: &creatorWorkspacePreviewShortListResponseData{
			Items: responseItems,
		},
		Meta: responseMeta{
			RequestID: newRequestID(creatorWorkspaceShortListRequestScope),
			Page: &cursorPageInfo{
				HasNext:    nextCursor != nil,
				NextCursor: encodeCreatorWorkspacePreviewCursor(nextCursor),
			},
		},
		Error: nil,
	})
}

func handleCreatorWorkspacePreviewMains(c *gin.Context, reader CreatorWorkspaceReader) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeCreatorWorkspaceListError(c, creatorWorkspaceMainListRequestScope, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
		return
	}

	cursor := decodeCreatorWorkspacePreviewCursor(strings.TrimSpace(c.Query("cursor")))
	items, nextCursor, err := reader.ListWorkspacePreviewMains(
		c.Request.Context(),
		viewerUserID,
		cursor,
		creator.DefaultWorkspacePreviewPageSize,
	)
	if err != nil {
		writeCreatorWorkspaceReadError(c, creatorWorkspaceMainListRequestScope, err)
		return
	}

	responseItems := make([]creatorWorkspacePreviewMainItem, 0, len(items))
	for _, item := range items {
		responseItem, buildErr := buildCreatorWorkspacePreviewMainItem(item)
		if buildErr != nil {
			writeCreatorWorkspaceListError(c, creatorWorkspaceMainListRequestScope, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
			return
		}
		responseItems = append(responseItems, responseItem)
	}

	c.JSON(http.StatusOK, responseEnvelope[creatorWorkspacePreviewMainListResponseData]{
		Data: &creatorWorkspacePreviewMainListResponseData{
			Items: responseItems,
		},
		Meta: responseMeta{
			RequestID: newRequestID(creatorWorkspaceMainListRequestScope),
			Page: &cursorPageInfo{
				HasNext:    nextCursor != nil,
				NextCursor: encodeCreatorWorkspacePreviewCursor(nextCursor),
			},
		},
		Error: nil,
	})
}

func handleCreatorWorkspacePreviewShortDetail(c *gin.Context, reader CreatorWorkspaceReader) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeCreatorWorkspaceScopedError(c, creatorWorkspaceShortDetailRequestScope, http.StatusInternalServerError, "internal_error", "creator workspace preview could not be loaded")
		return
	}

	shortID, err := shorts.ParsePublicShortID(c.Param("shortId"))
	if err != nil {
		writeCreatorWorkspaceScopedError(c, creatorWorkspaceShortDetailRequestScope, http.StatusNotFound, "not_found", "creator workspace preview was not found")
		return
	}

	detail, err := reader.GetWorkspacePreviewShortDetail(c.Request.Context(), viewerUserID, shortID)
	if err != nil {
		writeCreatorWorkspacePreviewReadError(c, creatorWorkspaceShortDetailRequestScope, err)
		return
	}

	responsePayload, err := buildCreatorWorkspacePreviewShortDetailPayload(detail)
	if err != nil {
		writeCreatorWorkspaceScopedError(c, creatorWorkspaceShortDetailRequestScope, http.StatusInternalServerError, "internal_error", "creator workspace preview could not be loaded")
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[creatorWorkspacePreviewShortDetailResponseData]{
		Data: &creatorWorkspacePreviewShortDetailResponseData{
			Preview: responsePayload,
		},
		Meta: responseMeta{
			RequestID: newRequestID(creatorWorkspaceShortDetailRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleCreatorWorkspacePreviewMainDetail(c *gin.Context, reader CreatorWorkspaceReader) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeCreatorWorkspaceScopedError(c, creatorWorkspaceMainDetailRequestScope, http.StatusInternalServerError, "internal_error", "creator workspace preview could not be loaded")
		return
	}

	mainID, err := shorts.ParsePublicMainID(c.Param("mainId"))
	if err != nil {
		writeCreatorWorkspaceScopedError(c, creatorWorkspaceMainDetailRequestScope, http.StatusNotFound, "not_found", "creator workspace preview was not found")
		return
	}

	detail, err := reader.GetWorkspacePreviewMainDetail(c.Request.Context(), viewerUserID, mainID)
	if err != nil {
		writeCreatorWorkspacePreviewReadError(c, creatorWorkspaceMainDetailRequestScope, err)
		return
	}

	responsePayload, err := buildCreatorWorkspacePreviewMainDetailPayload(detail)
	if err != nil {
		writeCreatorWorkspaceScopedError(c, creatorWorkspaceMainDetailRequestScope, http.StatusInternalServerError, "internal_error", "creator workspace preview could not be loaded")
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[creatorWorkspacePreviewMainDetailResponseData]{
		Data: &creatorWorkspacePreviewMainDetailResponseData{
			Preview: responsePayload,
		},
		Meta: responseMeta{
			RequestID: newRequestID(creatorWorkspaceMainDetailRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleCreatorWorkspaceMainPriceUpdate(c *gin.Context, writer CreatorWorkspaceWriter) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeCreatorWorkspaceMutationError(c, creatorWorkspaceMainPriceRequestScope, http.StatusInternalServerError, "internal_error", "creator workspace metadata could not be updated")
		return
	}

	mainID, err := shorts.ParsePublicMainID(c.Param("mainId"))
	if err != nil {
		writeCreatorWorkspaceMutationError(c, creatorWorkspaceMainPriceRequestScope, http.StatusNotFound, "not_found", "creator workspace preview was not found")
		return
	}

	var request creatorWorkspaceUpdateMainPriceRequest
	if !decodeCreatorWorkspaceJSON(c, &request, "creator workspace metadata request is invalid", creatorWorkspaceMainPriceRequestScope) {
		return
	}
	if request.PriceJpy <= 0 {
		writeCreatorWorkspaceMutationError(c, creatorWorkspaceMainPriceRequestScope, http.StatusBadRequest, "validation_error", "priceJpy must be positive")
		return
	}

	if err := writer.UpdateWorkspaceMainPrice(c.Request.Context(), viewerUserID, mainID, request.PriceJpy); err != nil {
		writeCreatorWorkspaceWriteError(c, creatorWorkspaceMainPriceRequestScope, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func handleCreatorWorkspaceShortCaptionUpdate(c *gin.Context, writer CreatorWorkspaceWriter) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeCreatorWorkspaceMutationError(c, creatorWorkspaceShortCaptionRequestScope, http.StatusInternalServerError, "internal_error", "creator workspace metadata could not be updated")
		return
	}

	shortID, err := shorts.ParsePublicShortID(c.Param("shortId"))
	if err != nil {
		writeCreatorWorkspaceMutationError(c, creatorWorkspaceShortCaptionRequestScope, http.StatusNotFound, "not_found", "creator workspace preview was not found")
		return
	}

	var request creatorWorkspaceUpdateShortCaptionRequest
	if !decodeCreatorWorkspaceJSON(c, &request, "creator workspace metadata request is invalid", creatorWorkspaceShortCaptionRequestScope) {
		return
	}
	if !request.Caption.Present {
		writeCreatorWorkspaceMutationError(c, creatorWorkspaceShortCaptionRequestScope, http.StatusBadRequest, "invalid_request", "creator workspace metadata request is invalid")
		return
	}

	if err := writer.UpdateWorkspaceShortCaption(c.Request.Context(), viewerUserID, shortID, request.Caption.Value); err != nil {
		writeCreatorWorkspaceWriteError(c, creatorWorkspaceShortCaptionRequestScope, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func handleCreatorWorkspaceTopPerformers(c *gin.Context, reader CreatorWorkspaceReader) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeCreatorWorkspaceListError(c, creatorWorkspaceTopPerformersScope, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
		return
	}

	topPerformers, err := reader.GetWorkspaceTopPerformers(c.Request.Context(), viewerUserID)
	if err != nil {
		writeCreatorWorkspaceReadError(c, creatorWorkspaceTopPerformersScope, err)
		return
	}

	responsePayload := buildCreatorWorkspaceTopPerformersPayload(topPerformers)

	c.JSON(http.StatusOK, responseEnvelope[creatorWorkspaceTopPerformersResponseData]{
		Data: &creatorWorkspaceTopPerformersResponseData{
			TopPerformers: responsePayload,
		},
		Meta: responseMeta{
			RequestID: newRequestID(creatorWorkspaceTopPerformersScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func buildCreatorWorkspacePreviewShortItem(item creator.WorkspacePreviewShortItem) (creatorWorkspacePreviewShortItem, error) {
	mediaPayload := buildCreatorWorkspacePreviewMediaAssetPayload(item.Media)

	return creatorWorkspacePreviewShortItem{
		CanonicalMainID:        mainPublicID(item.CanonicalMainID),
		ID:                     shortPublicID(item.ID),
		Media:                  mediaPayload,
		PreviewDurationSeconds: item.PreviewDurationSeconds,
	}, nil
}

func buildCreatorWorkspacePreviewMainItem(item creator.WorkspacePreviewMainItem) (creatorWorkspacePreviewMainItem, error) {
	mediaPayload := buildCreatorWorkspacePreviewMediaAssetPayload(item.Media)

	return creatorWorkspacePreviewMainItem{
		DurationSeconds: item.DurationSeconds,
		ID:              mainPublicID(item.ID),
		LeadShortID:     shortPublicID(item.LeadShortID),
		Media:           mediaPayload,
		PriceJpy:        item.PriceJpy,
	}, nil
}

func buildCreatorWorkspacePreviewShortDetailPayload(
	detail creator.WorkspacePreviewShortDetail,
) (creatorWorkspacePreviewShortDetailPayload, error) {
	creatorSummary, err := buildCreatorSummary(detail.Creator)
	if err != nil {
		return creatorWorkspacePreviewShortDetailPayload{}, err
	}

	return creatorWorkspacePreviewShortDetailPayload{
		Access:  buildCreatorWorkspaceOwnerPreviewAccessPayload(detail.Short.CanonicalMainID),
		Creator: creatorSummary,
		Short:   buildCreatorWorkspacePreviewShortDetailItemPayload(detail.Short),
	}, nil
}

func buildCreatorWorkspacePreviewMainDetailPayload(
	detail creator.WorkspacePreviewMainDetail,
) (creatorWorkspacePreviewMainDetailPayload, error) {
	creatorSummary, err := buildCreatorSummary(detail.Creator)
	if err != nil {
		return creatorWorkspacePreviewMainDetailPayload{}, err
	}

	return creatorWorkspacePreviewMainDetailPayload{
		Access:     buildCreatorWorkspaceOwnerPreviewAccessPayload(detail.Main.ID),
		Creator:    creatorSummary,
		EntryShort: buildCreatorWorkspacePreviewShortDetailItemPayload(detail.EntryShort),
		Main: creatorWorkspacePreviewMainDetailItem{
			DurationSeconds: detail.Main.DurationSeconds,
			ID:              mainPublicID(detail.Main.ID),
			Media:           buildVideoMediaAsset(detail.Main.Media),
			PriceJpy:        detail.Main.PriceJpy,
		},
	}, nil
}

func buildCreatorWorkspacePreviewShortDetailItemPayload(
	item creator.WorkspacePreviewShortDetailItem,
) creatorWorkspacePreviewShortDetailItem {
	return creatorWorkspacePreviewShortDetailItem{
		Caption:                item.Caption,
		CanonicalMainID:        mainPublicID(item.CanonicalMainID),
		CreatorID:              creator.FormatPublicID(item.CreatorUserID),
		ID:                     shortPublicID(item.ID),
		Media:                  buildVideoMediaAsset(item.Media),
		PreviewDurationSeconds: item.PreviewDurationSeconds,
	}
}

func buildCreatorWorkspaceOwnerPreviewAccessPayload(mainID uuid.UUID) creatorWorkspaceOwnerPreviewAccessPayload {
	return creatorWorkspaceOwnerPreviewAccessPayload{
		MainID: mainPublicID(mainID),
		Reason: "owner_preview",
		Status: "owner",
	}
}

func buildCreatorWorkspacePreviewMediaAssetPayload(asset media.VideoPreviewCardAsset) creatorWorkspacePreviewMediaAsset {
	return creatorWorkspacePreviewMediaAsset{
		DurationSeconds: asset.DurationSeconds,
		ID:              mediaAssetPublicID(asset.ID),
		Kind:            asset.Kind,
		PosterURL:       asset.PosterURL,
	}
}

func buildCreatorWorkspaceTopPerformersPayload(
	topPerformers creator.WorkspaceTopPerformers,
) creatorWorkspaceTopPerformersPayload {
	var topMainPayload *creatorWorkspaceTopMainPerformer
	if topPerformers.TopMain != nil {
		topMainPayload = &creatorWorkspaceTopMainPerformer{
			ID:          mainPublicID(topPerformers.TopMain.ID),
			Media:       buildCreatorWorkspacePreviewMediaAssetPayload(topPerformers.TopMain.Media),
			UnlockCount: topPerformers.TopMain.UnlockCount,
		}
	}

	var topShortPayload *creatorWorkspaceTopShortPerformer
	if topPerformers.TopShort != nil {
		topShortPayload = &creatorWorkspaceTopShortPerformer{
			AttributedUnlockCount: topPerformers.TopShort.AttributedUnlockCount,
			ID:                    shortPublicID(topPerformers.TopShort.ID),
			Media:                 buildCreatorWorkspacePreviewMediaAssetPayload(topPerformers.TopShort.Media),
		}
	}

	return creatorWorkspaceTopPerformersPayload{
		TopMain:  topMainPayload,
		TopShort: topShortPayload,
	}
}

func writeCreatorWorkspaceReadError(c *gin.Context, requestScope string, err error) {
	switch {
	case errors.Is(err, creator.ErrCreatorModeUnavailable):
		writeCreatorWorkspaceScopedError(c, requestScope, http.StatusForbidden, "creator_mode_unavailable", "creator mode is not available")
	case errors.Is(err, creator.ErrProfileNotFound):
		writeCreatorWorkspaceScopedError(c, requestScope, http.StatusNotFound, "not_found", "creator workspace was not found")
	default:
		writeCreatorWorkspaceScopedError(c, requestScope, http.StatusInternalServerError, "internal_error", "creator workspace could not be loaded")
	}
}

func writeCreatorWorkspaceListError(c *gin.Context, requestScope string, status int, code string, message string) {
	writeCreatorWorkspaceScopedError(c, requestScope, status, code, message)
}

func writeCreatorWorkspacePreviewReadError(c *gin.Context, requestScope string, err error) {
	switch {
	case errors.Is(err, creator.ErrCreatorModeUnavailable):
		writeCreatorWorkspaceScopedError(c, requestScope, http.StatusForbidden, "creator_mode_unavailable", "creator mode is not available")
	case errors.Is(err, creator.ErrProfileNotFound),
		errors.Is(err, creator.ErrWorkspacePreviewMainNotFound),
		errors.Is(err, creator.ErrWorkspacePreviewShortNotFound):
		writeCreatorWorkspaceScopedError(c, requestScope, http.StatusNotFound, "not_found", "creator workspace preview was not found")
	default:
		writeCreatorWorkspaceScopedError(c, requestScope, http.StatusInternalServerError, "internal_error", "creator workspace preview could not be loaded")
	}
}

func writeCreatorWorkspaceWriteError(c *gin.Context, requestScope string, err error) {
	switch {
	case errors.Is(err, creator.ErrCreatorModeUnavailable):
		writeCreatorWorkspaceMutationError(c, requestScope, http.StatusForbidden, "creator_mode_unavailable", "creator mode is not available")
	case errors.Is(err, creator.ErrProfileNotFound),
		errors.Is(err, creator.ErrWorkspacePreviewMainNotFound),
		errors.Is(err, creator.ErrWorkspacePreviewShortNotFound):
		writeCreatorWorkspaceMutationError(c, requestScope, http.StatusNotFound, "not_found", "creator workspace preview was not found")
	default:
		writeCreatorWorkspaceMutationError(c, requestScope, http.StatusInternalServerError, "internal_error", "creator workspace metadata could not be updated")
	}
}

func writeCreatorWorkspaceMutationError(c *gin.Context, requestScope string, status int, code string, message string) {
	writeCreatorWorkspaceScopedError(c, requestScope, status, code, message)
}

func writeCreatorWorkspaceScopedError(c *gin.Context, requestScope string, status int, code string, message string) {
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

func writeCreatorWorkspaceError(c *gin.Context, status int, code string, message string) {
	writeCreatorWorkspaceScopedError(c, creatorWorkspaceRequestScope, status, code, message)
}

// CreatorWorkspaceReader は creator private workspace summary 用の read 操作を表します。
type CreatorWorkspaceReader interface {
	GetWorkspace(ctx context.Context, viewerUserID uuid.UUID) (creator.Workspace, error)
	GetWorkspacePreviewMainDetail(ctx context.Context, viewerUserID uuid.UUID, mainID uuid.UUID) (creator.WorkspacePreviewMainDetail, error)
	GetWorkspacePreviewShortDetail(ctx context.Context, viewerUserID uuid.UUID, shortID uuid.UUID) (creator.WorkspacePreviewShortDetail, error)
	GetWorkspaceTopPerformers(ctx context.Context, viewerUserID uuid.UUID) (creator.WorkspaceTopPerformers, error)
	ListWorkspacePreviewMains(ctx context.Context, viewerUserID uuid.UUID, cursor *creator.WorkspacePreviewCursor, limit int) ([]creator.WorkspacePreviewMainItem, *creator.WorkspacePreviewCursor, error)
	ListWorkspacePreviewShorts(ctx context.Context, viewerUserID uuid.UUID, cursor *creator.WorkspacePreviewCursor, limit int) ([]creator.WorkspacePreviewShortItem, *creator.WorkspacePreviewCursor, error)
}

// CreatorWorkspaceWriter は creator workspace metadata mutation を表します。
type CreatorWorkspaceWriter interface {
	UpdateWorkspaceMainPrice(ctx context.Context, viewerUserID uuid.UUID, mainID uuid.UUID, priceJpy int64) error
	UpdateWorkspaceShortCaption(ctx context.Context, viewerUserID uuid.UUID, shortID uuid.UUID, caption *string) error
}

func decodeCreatorWorkspaceJSON[T any](c *gin.Context, target *T, message string, requestScope string) bool {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeCreatorWorkspaceMutationError(c, requestScope, http.StatusBadRequest, "invalid_request", message)
		return false
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != nil && !errors.Is(err, io.EOF) {
		writeCreatorWorkspaceMutationError(c, requestScope, http.StatusBadRequest, "invalid_request", message)
		return false
	}
	if len(extra) > 0 {
		writeCreatorWorkspaceMutationError(c, requestScope, http.StatusBadRequest, "invalid_request", message)
		return false
	}

	return true
}

func decodeCreatorWorkspacePreviewCursor(encoded string) *creator.WorkspacePreviewCursor {
	if encoded == "" {
		return nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil
	}

	var payload creatorWorkspacePreviewCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil
	}

	createdAt, err := time.Parse(time.RFC3339Nano, payload.CreatedAt)
	if err != nil {
		return nil
	}
	itemID, err := uuid.Parse(strings.TrimSpace(payload.ID))
	if err != nil {
		return nil
	}

	return &creator.WorkspacePreviewCursor{
		CreatedAt: createdAt,
		ID:        itemID,
	}
}

func encodeCreatorWorkspacePreviewCursor(cursor *creator.WorkspacePreviewCursor) *string {
	if cursor == nil {
		return nil
	}

	payload, err := json.Marshal(creatorWorkspacePreviewCursorPayload{
		CreatedAt: cursor.CreatedAt.Format(time.RFC3339Nano),
		ID:        cursor.ID.String(),
	})
	if err != nil {
		return nil
	}

	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return &encoded
}
