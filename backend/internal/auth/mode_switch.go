package auth

import (
	"context"
	"fmt"
	"strings"
)

// ErrInvalidActiveMode は要求された active mode が不正なことを表します。
var ErrInvalidActiveMode = fmt.Errorf("active mode が不正です")

type activeModeRepository interface {
	UpdateActiveModeByTokenHash(ctx context.Context, sessionTokenHash string, activeMode ActiveMode) (SessionRecord, error)
}

// ModeSwitcher は active mode の切替を扱います。
type ModeSwitcher struct {
	repository activeModeRepository
}

// NewModeSwitcher は active mode switcher を構築します。
func NewModeSwitcher(repository activeModeRepository) *ModeSwitcher {
	return &ModeSwitcher{repository: repository}
}

// SwitchActiveMode は raw session token に紐づく active mode を更新します。
func (s *ModeSwitcher) SwitchActiveMode(ctx context.Context, rawSessionToken string, activeMode ActiveMode) error {
	if s == nil || s.repository == nil {
		return fmt.Errorf("mode switcher が初期化されていません")
	}

	switch activeMode {
	case ActiveModeFan, ActiveModeCreator:
	default:
		return ErrInvalidActiveMode
	}

	trimmedToken := strings.TrimSpace(rawSessionToken)
	if trimmedToken == "" {
		return ErrSessionNotFound
	}

	if _, err := s.repository.UpdateActiveModeByTokenHash(ctx, HashSessionToken(trimmedToken), activeMode); err != nil {
		return fmt.Errorf("active mode 切替 token=%s mode=%s: %w", trimmedToken, activeMode, err)
	}

	return nil
}
