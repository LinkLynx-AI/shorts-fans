"use client";

import { getCreatorModeEntryErrorMessage } from "./creator-entry";
import { useViewerModeEntry } from "./use-viewer-mode-entry";

type UseCreatorModeEntryResult = {
  clearError: () => void;
  enterCreatorMode: () => Promise<boolean>;
  errorMessage: string | null;
  isSubmitting: boolean;
};

/**
 * fan surface から creator mode へ遷移する submit 状態を管理する。
 */
export function useCreatorModeEntry(): UseCreatorModeEntryResult {
  const {
    clearError,
    enterMode,
    errorMessage,
    isSubmitting,
  } = useViewerModeEntry({
    activeMode: "creator",
    destination: "/creator",
    getErrorMessage: getCreatorModeEntryErrorMessage,
  });

  return {
    clearError,
    enterCreatorMode: enterMode,
    errorMessage,
    isSubmitting,
  };
}
