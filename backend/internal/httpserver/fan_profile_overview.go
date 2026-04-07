package httpserver

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/fanprofile"
)

const (
	fanProfileOverviewRequestScope = "fan_profile_overview"
	fanProfileAuthRequiredMessage  = "fan profile requires authentication"
)

type fanProfileOverviewResponseData struct {
	FanProfile fanProfileOverviewPayload `json:"fanProfile"`
}

type fanProfileOverviewPayload struct {
	Counts fanProfileOverviewCounts `json:"counts"`
	Title  string                   `json:"title"`
}

type fanProfileOverviewCounts struct {
	Following    int64 `json:"following"`
	PinnedShorts int64 `json:"pinnedShorts"`
	Library      int64 `json:"library"`
}

// registerFanProfileRoutes は fan profile private hub API を router に登録します。
func registerFanProfileRoutes(
	router gin.IRouter,
	overviewReader FanProfileOverviewReader,
	followingReader FanProfileFollowingReader,
	viewerBootstrap ViewerBootstrapReader,
) {
	if router == nil || viewerBootstrap == nil {
		return
	}

	if overviewReader != nil {
		router.GET(
			"/api/fan/profile",
			buildProtectedFanAuthGuard(viewerBootstrap, fanProfileOverviewRequestScope, fanProfileAuthRequiredMessage),
			func(c *gin.Context) {
				handleFanProfileOverview(c, overviewReader)
			},
		)
	}

	if followingReader != nil {
		router.GET(
			"/api/fan/profile/following",
			buildProtectedFanAuthGuard(viewerBootstrap, fanProfileFollowingRequestScope, fanProfileAuthRequiredMessage),
			func(c *gin.Context) {
				handleFanProfileFollowing(c, followingReader)
			},
		)
	}
}

// handleFanProfileOverview は private hub overview を返します。
func handleFanProfileOverview(c *gin.Context, reader FanProfileOverviewReader) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeInternalServerError(c, fanProfileOverviewRequestScope)
		return
	}

	overview, err := reader.GetOverview(c.Request.Context(), viewerUserID)
	if err != nil {
		if errors.Is(err, fanprofile.ErrProfileNotFound) {
			writeNotFoundError(c, fanProfileOverviewRequestScope, "fan profile was not found")
			return
		}

		writeInternalServerError(c, fanProfileOverviewRequestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[fanProfileOverviewResponseData]{
		Data: &fanProfileOverviewResponseData{
			FanProfile: fanProfileOverviewPayload{
				Counts: fanProfileOverviewCounts{
					Following:    overview.Counts.Following,
					PinnedShorts: overview.Counts.PinnedShorts,
					Library:      overview.Counts.Library,
				},
				Title: overview.Title,
			},
		},
		Meta: responseMeta{
			RequestID: newRequestID(fanProfileOverviewRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func authenticatedViewerIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		return uuid.Nil, false
	}

	return viewer.ID, true
}
