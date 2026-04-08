"use client";

import { useRouter } from "next/navigation";
import {
  startTransition,
  useState,
} from "react";

import { switchViewerActiveMode } from "../api/switch-viewer-active-mode";
import { getCreatorModeEntryErrorMessage } from "./creator-entry";

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
  const router = useRouter();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const clearError = () => {
    if (errorMessage !== null) {
      setErrorMessage(null);
    }
  };

  const enterCreatorMode = async () => {
    if (isSubmitting) {
      return false;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      await switchViewerActiveMode("creator");

      startTransition(() => {
        router.push("/creator");
      });

      return true;
    } catch (error) {
      setErrorMessage(getCreatorModeEntryErrorMessage(error));
      return false;
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    clearError,
    enterCreatorMode,
    errorMessage,
    isSubmitting,
  };
}
