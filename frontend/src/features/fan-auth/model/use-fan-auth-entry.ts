"use client";

import { useRouter } from "next/navigation";
import { startTransition, useState } from "react";

import { ApiError } from "@/shared/api";

import { authenticateFanWithEmail } from "../api/request-fan-auth";
import {
  FanAuthApiError,
  getFanAuthErrorMessage,
  type FanAuthMode,
} from "./fan-auth";

type UseFanAuthEntryOptions = {
  initialMode?: FanAuthMode;
  onAuthenticated?: () => Promise<string | null> | string | null;
};

type UseFanAuthEntryResult = {
  email: string;
  errorMessage: string | null;
  isSubmitting: boolean;
  mode: FanAuthMode;
  setEmail: (email: string) => void;
  submit: () => Promise<void>;
  switchMode: () => void;
};

/**
 * fan auth entry UI に必要な mode / email / submit 状態を管理する。
 */
export function useFanAuthEntry({
  initialMode = "sign-in",
  onAuthenticated,
}: UseFanAuthEntryOptions = {}): UseFanAuthEntryResult {
  const router = useRouter();
  const [email, setEmailState] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [mode, setMode] = useState<FanAuthMode>(initialMode);

  const setEmail = (nextEmail: string) => {
    setEmailState(nextEmail);
    if (errorMessage !== null) {
      setErrorMessage(null);
    }
  };

  const switchMode = () => {
    setMode((currentMode) => (currentMode === "sign-in" ? "sign-up" : "sign-in"));
    if (errorMessage !== null) {
      setErrorMessage(null);
    }
  };

  const submit = async () => {
    if (isSubmitting) {
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      await authenticateFanWithEmail(mode, email);

      if (onAuthenticated) {
        const postAuthErrorMessage = await onAuthenticated();

        if (postAuthErrorMessage) {
          setErrorMessage(postAuthErrorMessage);
        }

        return;
      }

      startTransition(() => {
        router.refresh();
      });
    } catch (error) {
      if (error instanceof FanAuthApiError) {
        setErrorMessage(getFanAuthErrorMessage(error.code));
        return;
      }

      if (error instanceof ApiError) {
        setErrorMessage("認証を完了できませんでした。通信状態を確認してから再度お試しください。");
        return;
      }

      setErrorMessage("認証を完了できませんでした。少し時間を置いてからやり直してください。");
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    email,
    errorMessage,
    isSubmitting,
    mode,
    setEmail,
    submit,
    switchMode,
  };
}
