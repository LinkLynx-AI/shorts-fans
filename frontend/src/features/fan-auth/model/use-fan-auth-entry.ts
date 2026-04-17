"use client";

import { useRouter } from "next/navigation";
import { startTransition, useState } from "react";

import { ApiError } from "@/shared/api";
import {
  updateViewerProfile,
  useViewerProfileDraft,
} from "@/features/viewer-profile";

import {
  confirmFanPasswordReset,
  confirmFanSignUp,
  reAuthenticateFan,
  signInFan,
  signUpFan,
  startFanPasswordReset,
} from "../api/request-fan-auth";
import {
  FanAuthApiError,
  getFanAuthErrorMessage,
  mapFanAuthNextStepToMode,
  type FanAuthMode,
} from "./fan-auth";

type UseFanAuthEntryAuthenticatedOptions = {
  afterViewerSynced?: () => Promise<string | null> | string | null;
  authenticatedMode?: FanAuthMode;
};

type UseFanAuthEntryOptions = {
  initialMode?: FanAuthMode;
  onFallbackToSignIn?: (() => void) | undefined;
  onAuthenticated?: (
    options?: UseFanAuthEntryAuthenticatedOptions,
  ) => Promise<string | null> | string | null;
};

type PendingSignUpDraftSnapshot = {
  email: string;
  password: string;
};

type UseFanAuthEntryResult = {
  avatar: ReturnType<typeof useViewerProfileDraft>["avatar"];
  avatarInputKey: number;
  canResend: boolean;
  clearAvatarSelection: () => void;
  confirmationCode: string;
  deliveryDestinationHint: string | null;
  displayName: string;
  email: string;
  errorMessage: string | null;
  handle: string;
  hasConfirmedSignUp: boolean;
  infoMessage: string | null;
  isSubmitting: boolean;
  mode: FanAuthMode;
  newPassword: string;
  password: string;
  resend: () => Promise<void>;
  selectAvatarFile: (file: File | null) => void;
  setConfirmationCode: (confirmationCode: string) => void;
  setDisplayName: (displayName: string) => void;
  setEmail: (email: string) => void;
  setHandle: (handle: string) => void;
  setMode: (mode: FanAuthMode) => void;
  setNewPassword: (newPassword: string) => void;
  setPassword: (password: string) => void;
  submit: () => Promise<void>;
};

function getAvatarInitializationErrorMessage(error: unknown): string {
  if (error instanceof ApiError && error.code === "network") {
    return "avatar の初期化に失敗しました。通信状態を確認してから再度お試しください。";
  }
  if (error instanceof ApiError) {
    return "avatar の初期化に失敗しました。もう一度お試しください。";
  }

  return "avatar の初期化に失敗しました。少し時間を置いてから再度お試しください。";
}

function getGenericFanAuthErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return "認証を完了できませんでした。通信状態を確認してから再度お試しください。";
  }

  return "認証を完了できませんでした。少し時間を置いてからやり直してください。";
}

function normalizeEmailForComparison(value: string): string | null {
  const normalizedEmail = value.trim().toLowerCase();

  return normalizedEmail === "" ? null : normalizedEmail;
}

/**
 * fan auth entry UI に必要な multi-step auth state を管理する。
 */
export function useFanAuthEntry({
  initialMode = "sign-in",
  onFallbackToSignIn,
  onAuthenticated,
}: UseFanAuthEntryOptions = {}): UseFanAuthEntryResult {
  const router = useRouter();
  const [confirmationCode, setConfirmationCodeState] = useState("");
  const [deliveryDestinationHint, setDeliveryDestinationHint] = useState<string | null>(null);
  const [email, setEmailState] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [hasConfirmedSignUp, setHasConfirmedSignUp] = useState(false);
  const [infoMessage, setInfoMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [mode, setModeState] = useState<FanAuthMode>(initialMode);
  const [newPassword, setNewPasswordState] = useState("");
  const [password, setPasswordState] = useState("");
  const [pendingSignUpDraft, setPendingSignUpDraft] = useState<PendingSignUpDraftSnapshot | null>(null);
  const draft = useViewerProfileDraft({
    mode: "sign-up",
    onDirty: () => {
      if (errorMessage !== null) {
        setErrorMessage(null);
      }
    },
  });

  const clearFeedback = () => {
    if (errorMessage !== null) {
      setErrorMessage(null);
    }
    if (infoMessage !== null) {
      setInfoMessage(null);
    }
  };

  // Resend is safe only while this modal session still holds the same sign-up draft
  // including the password that was accepted for the current confirmation flow.
  const normalizedEmail = normalizeEmailForComparison(email);
  const hasPendingSignUpDraftForCurrentEmail =
    pendingSignUpDraft !== null &&
    pendingSignUpDraft.email === normalizedEmail &&
    pendingSignUpDraft.password === password;

  const invalidatePendingSignUpDraft = () => {
    if (pendingSignUpDraft !== null) {
      setPendingSignUpDraft(null);
    }
  };

  const setMode = (nextMode: FanAuthMode) => {
    setModeState(nextMode);
    setErrorMessage(null);

    const shouldRequirePasswordReentry =
      nextMode === "sign-up" &&
      mode === "confirm-sign-up" &&
      !hasPendingSignUpDraftForCurrentEmail;

    if (shouldRequirePasswordReentry) {
      setPasswordState("");
    }

    if (nextMode !== "confirm-sign-up") {
      setConfirmationCodeState("");
      setDeliveryDestinationHint(null);
      setHasConfirmedSignUp(false);
    }

    if (nextMode !== "confirm-password-reset") {
      setConfirmationCodeState((currentCode) =>
        nextMode === "confirm-sign-up" ? currentCode : "",
      );
      setNewPasswordState("");
    }

    if (nextMode !== "sign-up") {
      setInfoMessage(null);
    }
  };

  const setEmail = (nextEmail: string) => {
    setEmailState(nextEmail);
    clearFeedback();
  };

  const setPassword = (nextPassword: string) => {
    if (
      mode === "sign-up" ||
      (pendingSignUpDraft !== null && nextPassword !== pendingSignUpDraft.password)
    ) {
      invalidatePendingSignUpDraft();
    }
    setPasswordState(nextPassword);
    clearFeedback();
  };

  const setConfirmationCode = (nextConfirmationCode: string) => {
    setConfirmationCodeState(nextConfirmationCode);
    clearFeedback();
  };

  const setNewPassword = (nextNewPassword: string) => {
    setNewPasswordState(nextNewPassword);
    clearFeedback();
  };

  const returnToSignUpForConfirmationRecovery = () => {
    invalidatePendingSignUpDraft();
    setPasswordState("");
    setMode("sign-up");
    setInfoMessage("確認コードを再送するには登録情報を入力し直してください。");
  };

  const completeAvatarInitialization = async (): Promise<string | null> => {
    const avatarSubmissionError = draft.getAvatarSubmissionError();
    if (avatarSubmissionError !== null) {
      return avatarSubmissionError;
    }

    try {
      const avatarUploadToken = await draft.uploadAvatarIfNeeded();

      if (!avatarUploadToken) {
        return null;
      }

      await updateViewerProfile({
        avatarUploadToken,
        displayName: draft.displayName,
        handle: draft.handle,
      });

      return null;
    } catch (error) {
      return getAvatarInitializationErrorMessage(error);
    }
  };

  const finishAuthenticated = async (
    options: UseFanAuthEntryAuthenticatedOptions = {},
  ): Promise<void> => {
    const authenticatedOptions = {
      ...options,
      authenticatedMode: mode,
    } satisfies UseFanAuthEntryAuthenticatedOptions;

    if (onAuthenticated) {
      const postAuthErrorMessage = await onAuthenticated(authenticatedOptions);

      if (postAuthErrorMessage) {
        setErrorMessage(postAuthErrorMessage);
      }

      return;
    }

    const postSyncErrorMessage = await authenticatedOptions.afterViewerSynced?.();

    if (postSyncErrorMessage) {
      setErrorMessage(postSyncErrorMessage);
      return;
    }

    startTransition(() => {
      router.refresh();
    });
  };

  const handleFanAuthError = (error: unknown, fallbackMode: FanAuthMode) => {
    if (error instanceof FanAuthApiError) {
      if (fallbackMode === "sign-in" && error.code === "confirmation_required") {
        const keepPendingSignUpDraft =
          pendingSignUpDraft !== null &&
          pendingSignUpDraft.email === normalizedEmail &&
          pendingSignUpDraft.password === password;

        if (!keepPendingSignUpDraft) {
          setPendingSignUpDraft(null);
        }

        setConfirmationCodeState("");
        setDeliveryDestinationHint(null);
        setModeState("confirm-sign-up");
        setErrorMessage(null);
        setInfoMessage("確認コードを入力して登録を完了してください。");
        setHasConfirmedSignUp(false);
        return;
      }

      if (fallbackMode === "re-auth" && error.code === "auth_required") {
        onFallbackToSignIn?.();
        setModeState("sign-in");
        setErrorMessage(getFanAuthErrorMessage(error.code));
        return;
      }

      setErrorMessage(getFanAuthErrorMessage(error.code));
      return;
    }

    setErrorMessage(getGenericFanAuthErrorMessage(error));
  };

  const submit = async () => {
    if (isSubmitting) {
      return;
    }

    setIsSubmitting(true);
    clearFeedback();

    try {
      switch (mode) {
        case "sign-in": {
          await signInFan({
            email,
            password,
          });
          await finishAuthenticated();
          return;
        }
        case "sign-up": {
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

          const acceptedStep = await signUpFan({
            displayName: draft.displayName,
            email,
            handle: draft.handle,
            password,
          });

          setPendingSignUpDraft(
            normalizedEmail === null
              ? null
              : {
                  email: normalizedEmail,
                  password,
                },
          );
          setDeliveryDestinationHint(acceptedStep.deliveryDestinationHint);
          setModeState(mapFanAuthNextStepToMode(acceptedStep.nextStep));
          setInfoMessage("確認コードを送信しました。メールを確認してください。");
          setHasConfirmedSignUp(false);
          return;
        }
        case "confirm-sign-up": {
          if (!hasConfirmedSignUp) {
            await confirmFanSignUp({
              confirmationCode,
              email,
            });
            setHasConfirmedSignUp(true);
          }

          await finishAuthenticated({
            afterViewerSynced: completeAvatarInitialization,
          });
          return;
        }
        case "password-reset-request": {
          const acceptedStep = await startFanPasswordReset({ email });

          setDeliveryDestinationHint(acceptedStep.deliveryDestinationHint);
          setModeState(mapFanAuthNextStepToMode(acceptedStep.nextStep));
          setInfoMessage("確認コードを送信しました。メールを確認してください。");
          return;
        }
        case "confirm-password-reset": {
          await confirmFanPasswordReset({
            confirmationCode,
            email,
            newPassword,
          });

          setModeState("sign-in");
          setNewPasswordState("");
          setPasswordState("");
          setConfirmationCodeState("");
          setDeliveryDestinationHint(null);
          setInfoMessage("パスワードを更新しました。サインインを続けてください。");
          return;
        }
        case "re-auth": {
          await reAuthenticateFan({
            password,
          });
          await finishAuthenticated();
          return;
        }
      }
    } catch (error) {
      handleFanAuthError(error, mode);
    } finally {
      setIsSubmitting(false);
    }
  };

  const resend = async () => {
    if (isSubmitting) {
      return;
    }

    if (mode !== "confirm-sign-up" && mode !== "confirm-password-reset") {
      return;
    }

    setIsSubmitting(true);
    clearFeedback();

    try {
      if (mode === "confirm-sign-up") {
        if (!hasPendingSignUpDraftForCurrentEmail) {
          returnToSignUpForConfirmationRecovery();
          return;
        }

        if (password.trim() === "") {
          returnToSignUpForConfirmationRecovery();
          return;
        }

        const profileValidationError = draft.getProfileValidationError();
        if (profileValidationError !== null) {
          returnToSignUpForConfirmationRecovery();
          return;
        }

        const avatarSubmissionError = draft.getAvatarSubmissionError();
        if (avatarSubmissionError !== null) {
          setErrorMessage(avatarSubmissionError);
          return;
        }

        const acceptedStep = await signUpFan({
          displayName: draft.displayName,
          email,
          handle: draft.handle,
          password,
        });

        setDeliveryDestinationHint(acceptedStep.deliveryDestinationHint);
        setInfoMessage("確認コードを再送しました。メールを確認してください。");
        return;
      }

      const acceptedStep = await startFanPasswordReset({ email });

      setDeliveryDestinationHint(acceptedStep.deliveryDestinationHint);
      setInfoMessage("確認コードを再送しました。メールを確認してください。");
    } catch (error) {
      handleFanAuthError(error, mode);
    } finally {
      setIsSubmitting(false);
    }
  };

  const hasPendingPasswordForConfirmation = password.trim() !== "";

  return {
    avatar: draft.avatar,
    avatarInputKey: draft.avatarInputKey,
    canResend:
      mode === "confirm-password-reset" ||
      (mode === "confirm-sign-up" &&
        !hasConfirmedSignUp &&
        hasPendingSignUpDraftForCurrentEmail &&
        hasPendingPasswordForConfirmation),
    clearAvatarSelection: () => {
      invalidatePendingSignUpDraft();
      draft.clearAvatarSelection();
    },
    confirmationCode,
    deliveryDestinationHint,
    displayName: draft.displayName,
    email,
    errorMessage,
    handle: draft.handle,
    hasConfirmedSignUp,
    infoMessage,
    isSubmitting,
    mode,
    newPassword,
    password,
    resend,
    selectAvatarFile: (file) => {
      invalidatePendingSignUpDraft();
      draft.selectAvatarFile(file);
    },
    setConfirmationCode,
    setDisplayName: (nextDisplayName) => {
      invalidatePendingSignUpDraft();
      draft.setDisplayName(nextDisplayName);
    },
    setEmail,
    setHandle: (nextHandle) => {
      invalidatePendingSignUpDraft();
      draft.setHandle(nextHandle);
    },
    setMode,
    setNewPassword,
    setPassword,
    submit,
  };
}
