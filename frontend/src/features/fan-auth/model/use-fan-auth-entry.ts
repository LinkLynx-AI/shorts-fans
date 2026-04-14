"use client";

import { useRouter } from "next/navigation";
import { startTransition, useState } from "react";

import { ApiError } from "@/shared/api";
import {
  updateViewerProfile,
  useViewerProfileDraft,
} from "@/features/viewer-profile";

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
  avatar: ReturnType<typeof useViewerProfileDraft>["avatar"];
  avatarInputKey: number;
  clearAvatarSelection: () => void;
  displayName: string;
  email: string;
  errorMessage: string | null;
  handle: string;
  isSubmitting: boolean;
  mode: FanAuthMode;
  selectAvatarFile: (file: File | null) => void;
  setDisplayName: (displayName: string) => void;
  setEmail: (email: string) => void;
  setHandle: (handle: string) => void;
  submit: () => Promise<void>;
  switchMode: () => void;
};

const signUpAvatarSaveFailureMessage =
  "アカウントは作成されましたが、avatar の保存に失敗しました。fan settings から再度設定してください。";

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
  const draft = useViewerProfileDraft({
    mode: "sign-up",
    onDirty: () => {
      if (errorMessage !== null) {
        setErrorMessage(null);
      }
    },
  });

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
      if (mode === "sign-up") {
        const profileValidationError = draft.getProfileValidationError();
        if (profileValidationError !== null) {
          setErrorMessage(profileValidationError);
          return;
        }

        const avatarSubmissionError = draft.getAvatarSubmissionError();
        if (avatarSubmissionError !== null) {
          setErrorMessage(avatarSubmissionError);
          return;
        }
      }

      await authenticateFanWithEmail({
        ...(mode === "sign-up"
          ? {
              displayName: draft.displayName,
              handle: draft.handle,
            }
          : {}),
        email,
        mode,
      });

      if (mode === "sign-up") {
        try {
          const avatarUploadToken = await draft.uploadAvatarIfNeeded();

          if (avatarUploadToken) {
            await updateViewerProfile({
              avatarUploadToken,
              displayName: draft.displayName,
              handle: draft.handle,
            });
          }
        } catch {
          setErrorMessage(signUpAvatarSaveFailureMessage);
          return;
        }
      }

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
    avatar: draft.avatar,
    avatarInputKey: draft.avatarInputKey,
    clearAvatarSelection: draft.clearAvatarSelection,
    displayName: draft.displayName,
    email,
    errorMessage,
    handle: draft.handle,
    isSubmitting,
    mode,
    selectAvatarFile: draft.selectAvatarFile,
    setDisplayName: draft.setDisplayName,
    setEmail,
    setHandle: draft.setHandle,
    submit,
    switchMode,
  };
}
