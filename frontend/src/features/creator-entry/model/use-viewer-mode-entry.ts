"use client";

import { useRouter } from "next/navigation";
import {
  startTransition,
  useState,
} from "react";

import type { ViewerActiveMode } from "@/entities/viewer";

import { switchViewerActiveMode } from "../api/switch-viewer-active-mode";

type UseViewerModeEntryOptions = {
  activeMode: ViewerActiveMode;
  destination: string;
  getErrorMessage: (error: unknown) => string;
};

type UseViewerModeEntryResult = {
  clearError: () => void;
  enterMode: () => Promise<boolean>;
  errorMessage: string | null;
  isSubmitting: boolean;
};

/**
 * viewer mode 切り替え submit の共通状態を管理する。
 */
export function useViewerModeEntry({
  activeMode,
  destination,
  getErrorMessage,
}: UseViewerModeEntryOptions): UseViewerModeEntryResult {
  const router = useRouter();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const clearError = () => {
    if (errorMessage !== null) {
      setErrorMessage(null);
    }
  };

  const enterMode = async () => {
    if (isSubmitting) {
      return false;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      await switchViewerActiveMode(activeMode);

      startTransition(() => {
        router.push(destination);
      });

      return true;
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
      return false;
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    clearError,
    enterMode,
    errorMessage,
    isSubmitting,
  };
}
