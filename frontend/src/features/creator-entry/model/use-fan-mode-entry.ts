"use client";

import { getFanModeEntryErrorMessage } from "./creator-entry";
import { useViewerModeEntry } from "./use-viewer-mode-entry";

type UseFanModeEntryResult = {
  clearError: () => void;
  enterFanMode: () => Promise<boolean>;
  errorMessage: string | null;
  isSubmitting: boolean;
};

/**
 * creator surface から fan mode へ戻る submit 状態を管理する。
 */
export function useFanModeEntry(): UseFanModeEntryResult {
  const {
    clearError,
    enterMode,
    errorMessage,
    isSubmitting,
  } = useViewerModeEntry({
    activeMode: "fan",
    destination: "/",
    getErrorMessage: getFanModeEntryErrorMessage,
  });

  return {
    clearError,
    enterFanMode: enterMode,
    errorMessage,
    isSubmitting,
  };
}
