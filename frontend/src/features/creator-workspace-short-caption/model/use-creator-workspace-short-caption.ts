"use client";

import { startTransition, useEffect, useState } from "react";

import { ApiError } from "@/shared/api";

import { updateCreatorWorkspaceShortCaption } from "../api/update-creator-workspace-short-caption";

type UseCreatorWorkspaceShortCaptionOptions = {
  initialCaption: string;
  onSaved: () => void;
  open: boolean;
  shortId: string;
};

type UseCreatorWorkspaceShortCaptionResult = {
  caption: string;
  errorMessage: string | null;
  isSubmitting: boolean;
  setCaption: (caption: string) => void;
  submit: () => Promise<void>;
};

const creatorWorkspaceShortCaptionErrorMessage =
  "captionを更新できませんでした。少し時間を置いてからやり直してください。";
const creatorWorkspaceShortCaptionNetworkErrorMessage =
  "captionを更新できませんでした。通信環境を確認してからやり直してください。";
const creatorWorkspaceShortCaptionNotFoundErrorMessage =
  "対象のショートが見つからないため、captionを更新できませんでした。詳細を開き直してください。";

function buildCreatorWorkspaceShortCaptionErrorMessage(error: unknown): string {
  if (error instanceof ApiError && error.code === "network") {
    return creatorWorkspaceShortCaptionNetworkErrorMessage;
  }

  if (error instanceof ApiError && error.code === "http" && error.status === 404) {
    return creatorWorkspaceShortCaptionNotFoundErrorMessage;
  }

  return creatorWorkspaceShortCaptionErrorMessage;
}

/**
 * creator workspace short caption edit dialog の入力状態と保存状態を管理する。
 */
export function useCreatorWorkspaceShortCaption({
  initialCaption,
  onSaved,
  open,
  shortId,
}: UseCreatorWorkspaceShortCaptionOptions): UseCreatorWorkspaceShortCaptionResult {
  const [caption, setCaptionState] = useState(initialCaption);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (!open) {
      return;
    }

    setCaptionState(initialCaption);
    setErrorMessage(null);
  }, [initialCaption, open, shortId]);

  const setCaption = (nextCaption: string) => {
    setCaptionState(nextCaption);

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
      await updateCreatorWorkspaceShortCaption(shortId, caption);

      startTransition(() => {
        onSaved();
      });
    } catch (error) {
      setErrorMessage(buildCreatorWorkspaceShortCaptionErrorMessage(error));
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    caption,
    errorMessage,
    isSubmitting,
    setCaption,
    submit,
  };
}
