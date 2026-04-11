package httpserver

import (
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/google/uuid"
)

func mainPublicID(mainID uuid.UUID) string {
	return shorts.FormatPublicMainID(mainID)
}

func shortPublicID(shortID uuid.UUID) string {
	return shorts.FormatPublicShortID(shortID)
}
