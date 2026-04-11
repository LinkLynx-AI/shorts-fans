package media

import (
	"fmt"

	"github.com/google/uuid"
)

// DeliveryObjectKeys は 1 つの deliverable asset に紐づく object key の組です。
type DeliveryObjectKeys struct {
	Playback       string
	Poster         string
	PosterTempBase string
}

// BuildMainDeliveryObjectKeys は main asset 用の deterministic object key 群を返します。
func BuildMainDeliveryObjectKeys(mainID uuid.UUID) (DeliveryObjectKeys, error) {
	if mainID == uuid.Nil {
		return DeliveryObjectKeys{}, fmt.Errorf("main id is required")
	}

	return DeliveryObjectKeys{
		Playback:       fmt.Sprintf("mains/%s/playback.mp4", mainID),
		Poster:         fmt.Sprintf("mains/%s/poster.jpg", mainID),
		PosterTempBase: fmt.Sprintf("mains/%s/poster-temp", mainID),
	}, nil
}

// BuildShortDeliveryObjectKeys は short asset 用の deterministic object key 群を返します。
func BuildShortDeliveryObjectKeys(shortID uuid.UUID) (DeliveryObjectKeys, error) {
	if shortID == uuid.Nil {
		return DeliveryObjectKeys{}, fmt.Errorf("short id is required")
	}

	return DeliveryObjectKeys{
		Playback:       fmt.Sprintf("shorts/%s/playback.mp4", shortID),
		Poster:         fmt.Sprintf("shorts/%s/poster.jpg", shortID),
		PosterTempBase: fmt.Sprintf("shorts/%s/poster-temp", shortID),
	}, nil
}
