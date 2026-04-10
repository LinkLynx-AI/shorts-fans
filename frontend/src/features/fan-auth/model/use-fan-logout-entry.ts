"use client";

import { useRouter } from "next/navigation";
import {
  startTransition,
  useState,
} from "react";

import {
  useSetCurrentViewer,
  useSetViewerSession,
} from "@/entities/viewer";

import { logoutFanSession } from "../api/logout-fan-session";
import { getFanLogoutErrorMessage } from "./fan-auth";

type UseFanLogoutEntryResult = {
  clearError: () => void;
  errorMessage: string | null;
  isSubmitting: boolean;
  logout: () => Promise<boolean>;
};

/**
 * fan logout UI に必要な submit 状態を管理する。
 */
export function useFanLogoutEntry(): UseFanLogoutEntryResult {
  const router = useRouter();
  const setCurrentViewer = useSetCurrentViewer();
  const setViewerSession = useSetViewerSession();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const clearError = () => {
    if (errorMessage !== null) {
      setErrorMessage(null);
    }
  };

  const logout = async () => {
    if (isSubmitting) {
      return false;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      await logoutFanSession();
      setCurrentViewer(null);
      setViewerSession(false);

      startTransition(() => {
        router.push("/");
      });

      return true;
    } catch (error) {
      setErrorMessage(getFanLogoutErrorMessage(error));
      return false;
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    clearError,
    errorMessage,
    isSubmitting,
    logout,
  };
}
