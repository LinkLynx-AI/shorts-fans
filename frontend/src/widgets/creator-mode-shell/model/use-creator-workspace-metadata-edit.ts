"use client";

import { useState } from "react";

import { ApiError } from "@/shared/api";

import {
  updateCreatorWorkspaceMainPrice,
  updateCreatorWorkspaceShortCaption,
} from "../api/update-creator-workspace-metadata";

type UseCreatorWorkspaceMetadataEditResult = {
  clearError: () => void;
  errorMessage: string | null;
  isSubmitting: boolean;
  saveMainPrice: (mainId: string, priceJpy: number) => Promise<boolean>;
  saveShortCaption: (shortId: string, caption: string | null) => Promise<boolean>;
};

function resolveCreatorWorkspaceMainPriceErrorMessage(error: unknown): string {
  if (
    error instanceof ApiError &&
    error.code === "http" &&
    error.status === 400 &&
    error.details?.includes(`"validation_error"`)
  ) {
    return "価格は1円以上で入力してください。";
  }

  if (error instanceof ApiError && error.code === "network") {
    return "価格を更新できませんでした。通信環境を確認してからやり直してください。";
  }

  return "価格を更新できませんでした。少し時間を置いてからやり直してください。";
}

function resolveCreatorWorkspaceShortCaptionErrorMessage(error: unknown): string {
  if (error instanceof ApiError && error.code === "network") {
    return "caption を更新できませんでした。通信環境を確認してからやり直してください。";
  }

  return "caption を更新できませんでした。少し時間を置いてからやり直してください。";
}

/**
 * creator workspace の metadata edit submit 状態を管理する。
 */
export function useCreatorWorkspaceMetadataEdit(): UseCreatorWorkspaceMetadataEditResult {
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const clearError = () => {
    setErrorMessage(null);
  };

  const saveMainPrice = async (mainId: string, priceJpy: number): Promise<boolean> => {
    if (isSubmitting) {
      return false;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      await updateCreatorWorkspaceMainPrice(mainId, priceJpy);
      return true;
    } catch (error) {
      setErrorMessage(resolveCreatorWorkspaceMainPriceErrorMessage(error));
      return false;
    } finally {
      setIsSubmitting(false);
    }
  };

  const saveShortCaption = async (shortId: string, caption: string | null): Promise<boolean> => {
    if (isSubmitting) {
      return false;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      await updateCreatorWorkspaceShortCaption(shortId, caption);
      return true;
    } catch (error) {
      setErrorMessage(resolveCreatorWorkspaceShortCaptionErrorMessage(error));
      return false;
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    clearError,
    errorMessage,
    isSubmitting,
    saveMainPrice,
    saveShortCaption,
  };
}
