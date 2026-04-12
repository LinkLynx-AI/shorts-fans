package httpserver

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/fanprofile"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const fanProfileLibraryRequestScope = "fan_profile_library"

type fanProfileLibraryResponseData struct {
	Items []fanProfileLibraryItem `json:"items"`
}

type fanProfileLibraryItem struct {
	Access     mainAccessStatePayload    `json:"access"`
	Creator    creatorSummary            `json:"creator"`
	EntryShort feedShortSummary          `json:"entryShort"`
	Main       fanProfileLibraryMainInfo `json:"main"`
}

type fanProfileLibraryMainInfo struct {
	DurationSeconds int64  `json:"durationSeconds"`
	ID              string `json:"id"`
}

type fanProfileLibraryCursorPayload struct {
	CreatedAt   string `json:"createdAt"`
	MainID      string `json:"mainId"`
	PurchasedAt string `json:"purchasedAt"`
}

func handleFanProfileLibrary(
	c *gin.Context,
	reader FanProfileLibraryReader,
	shortDisplayAssets ShortDisplayAssetResolver,
) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		writeInternalServerError(c, fanProfileLibraryRequestScope)
		return
	}

	cursor := decodeFanProfileLibraryCursor(strings.TrimSpace(c.Query("cursor")))
	items, nextCursor, err := reader.ListLibrary(c.Request.Context(), viewer.ID, cursor, fanprofile.DefaultLibraryPageSize)
	if err != nil {
		writeInternalServerError(c, fanProfileLibraryRequestScope)
		return
	}

	responseItems := make([]fanProfileLibraryItem, 0, len(items))
	for _, item := range items {
		responseItem, buildErr := buildFanProfileLibraryItem(item, shortDisplayAssets)
		if buildErr != nil {
			writeInternalServerError(c, fanProfileLibraryRequestScope)
			return
		}

		responseItems = append(responseItems, responseItem)
	}

	c.JSON(http.StatusOK, responseEnvelope[fanProfileLibraryResponseData]{
		Data: &fanProfileLibraryResponseData{
			Items: responseItems,
		},
		Meta: responseMeta{
			RequestID: newRequestID(fanProfileLibraryRequestScope),
			Page: &cursorPageInfo{
				HasNext:    nextCursor != nil,
				NextCursor: encodeFanProfileLibraryCursor(nextCursor),
			},
		},
		Error: nil,
	})
}

func buildFanProfileLibraryItem(
	item fanprofile.LibraryItem,
	shortDisplayAssets ShortDisplayAssetResolver,
) (fanProfileLibraryItem, error) {
	creator, err := buildCreatorSummaryFields(
		item.CreatorUserID,
		item.CreatorDisplayName,
		item.CreatorHandle,
		item.CreatorAvatarURL,
		item.CreatorBio,
	)
	if err != nil {
		return fanProfileLibraryItem{}, err
	}

	entryShort, err := buildPublicShortSummaryFields(
		item.EntryShortID,
		item.EntryShortCanonicalMainID,
		item.CreatorUserID,
		item.EntryShortCaption,
		item.EntryShortMediaAssetID,
		item.EntryShortPreviewDurationSeconds,
		shortDisplayAssets,
	)
	if err != nil {
		return fanProfileLibraryItem{}, err
	}

	return fanProfileLibraryItem{
		Access: mainAccessStatePayload{
			MainID: mainPublicID(item.MainID),
			Reason: "session_unlocked",
			Status: "unlocked",
		},
		Creator:    creator,
		EntryShort: entryShort,
		Main: fanProfileLibraryMainInfo{
			DurationSeconds: item.MainDurationSeconds,
			ID:              mainPublicID(item.MainID),
		},
	}, nil
}

func decodeFanProfileLibraryCursor(encoded string) *fanprofile.LibraryCursor {
	if encoded == "" {
		return nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil
	}

	var payload fanProfileLibraryCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil
	}

	purchasedAt, err := time.Parse(time.RFC3339Nano, payload.PurchasedAt)
	if err != nil {
		return nil
	}
	createdAt, err := time.Parse(time.RFC3339Nano, payload.CreatedAt)
	if err != nil {
		return nil
	}
	mainID, err := uuid.Parse(strings.TrimSpace(payload.MainID))
	if err != nil {
		return nil
	}

	return &fanprofile.LibraryCursor{
		MainID:          mainID,
		PurchasedAt:     purchasedAt,
		UnlockCreatedAt: createdAt,
	}
}

func encodeFanProfileLibraryCursor(cursor *fanprofile.LibraryCursor) *string {
	if cursor == nil {
		return nil
	}

	payload, err := json.Marshal(fanProfileLibraryCursorPayload{
		CreatedAt:   cursor.UnlockCreatedAt.Format(time.RFC3339Nano),
		MainID:      cursor.MainID.String(),
		PurchasedAt: cursor.PurchasedAt.Format(time.RFC3339Nano),
	})
	if err != nil {
		return nil
	}

	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return &encoded
}
