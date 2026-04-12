package httpserver

import (
	"errors"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/google/uuid"
)

func buildPublicShortSummaryFields(
	shortID uuid.UUID,
	canonicalMainID uuid.UUID,
	creatorUserID uuid.UUID,
	caption string,
	mediaAssetID uuid.UUID,
	previewDurationSeconds int64,
	shortDisplayAssets ShortDisplayAssetResolver,
) (feedShortSummary, error) {
	if shortDisplayAssets == nil {
		return feedShortSummary{}, errors.New("short display asset resolver is required")
	}

	displayAsset, err := shortDisplayAssets.ResolveShortDisplayAsset(media.ShortDisplaySource{
		AssetID:    mediaAssetID,
		ShortID:    shortID,
		DurationMS: previewDurationSeconds * 1000,
	}, media.AccessBoundaryPublic)
	if err != nil {
		return feedShortSummary{}, err
	}

	return feedShortSummary{
		Caption:                caption,
		CanonicalMainID:        shorts.FormatPublicMainID(canonicalMainID),
		CreatorID:              creator.FormatPublicID(creatorUserID),
		ID:                     shorts.FormatPublicShortID(shortID),
		Media:                  buildVideoMediaAsset(displayAsset),
		PreviewDurationSeconds: previewDurationSeconds,
	}, nil
}
