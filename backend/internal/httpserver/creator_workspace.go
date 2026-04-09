package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	creatorWorkspaceAuthRequiredMessage = "creator workspace requires authentication"
	creatorWorkspaceRequestScope        = "creator_workspace"
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

// registerCreatorWorkspaceRoutes は creator private workspace summary API を router に登録します。
func registerCreatorWorkspaceRoutes(
	router gin.IRouter,
	reader CreatorWorkspaceReader,
	viewerBootstrap ViewerBootstrapReader,
) {
	if router == nil || reader == nil || viewerBootstrap == nil {
		return
	}

	router.GET(
		"/api/creator/workspace",
		buildProtectedFanAuthGuard(viewerBootstrap, creatorWorkspaceRequestScope, creatorWorkspaceAuthRequiredMessage),
		func(c *gin.Context) {
			handleCreatorWorkspace(c, reader)
		},
	)
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

func writeCreatorWorkspaceError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, responseEnvelope[struct{}]{
		Data: nil,
		Meta: responseMeta{
			RequestID: newRequestID(creatorWorkspaceRequestScope),
			Page:      nil,
		},
		Error: &responseError{
			Code:    code,
			Message: message,
		},
	})
}

// CreatorWorkspaceReader は creator private workspace summary 用の read 操作を表します。
type CreatorWorkspaceReader interface {
	GetWorkspace(ctx context.Context, viewerUserID uuid.UUID) (creator.Workspace, error)
}
