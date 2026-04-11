"use client";

import { useState } from "react";

import { ApiError } from "@/shared/api";

import {
  CreatorFollowApiError,
  type CreatorFollowMutationResult,
  updateCreatorFollow,
} from "../api/update-creator-follow";

type UseCreatorFollowToggleOptions = {
  creatorId: string;
  hasViewerSession: boolean;
  initialIsFollowing: boolean;
  onAuthRequired: () => void;
  onSuccess?: ((result: CreatorFollowMutationResult) => void) | undefined;
  onUnauthenticated: () => void;
  updateFollow?: typeof updateCreatorFollow | undefined;
};

type UseCreatorFollowToggleResult = {
  errorMessage: string | null;
  isFollowing: boolean;
  isPending: boolean;
  toggleFollow: () => Promise<void>;
};

/**
 * creator follow mutation の失敗を surface 表示用の文言へ変換する。
 */
export function getCreatorFollowErrorMessage(error: unknown): string {
  if (error instanceof CreatorFollowApiError) {
    if (error.code === "not_found") {
      return "この creator profile は現在利用できません。";
    }

    return "フォロー状態を更新できませんでした。少し時間を置いてから再度お試しください。";
  }

  if (error instanceof ApiError) {
    if (error.code === "network") {
      return "フォロー状態を更新できませんでした。通信状態を確認してから再度お試しください。";
    }

    return "フォロー状態を更新できませんでした。少し時間を置いてから再度お試しください。";
  }

  return "フォロー状態を更新できませんでした。少し時間を置いてから再度お試しください。";
}

/**
 * 単一 creator の follow / unfollow CTA 状態を既存 mutation semantics で管理する。
 */
export function useCreatorFollowToggle({
  creatorId,
  hasViewerSession,
  initialIsFollowing,
  onAuthRequired,
  onSuccess,
  onUnauthenticated,
  updateFollow,
}: UseCreatorFollowToggleOptions): UseCreatorFollowToggleResult {
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isFollowing, setIsFollowing] = useState(initialIsFollowing);
  const [isPending, setIsPending] = useState(false);

  const toggleFollow = async () => {
    if (isPending) {
      return;
    }

    if (!hasViewerSession) {
      setErrorMessage(null);
      onUnauthenticated();
      return;
    }

    const action = isFollowing ? "unfollow" : "follow";

    setErrorMessage(null);
    setIsPending(true);

    try {
      const result = await (updateFollow ?? updateCreatorFollow)({
        action,
        creatorId,
      });

      setIsFollowing(result.viewer.isFollowing);
      onSuccess?.(result);
    } catch (error) {
      if (error instanceof CreatorFollowApiError && error.code === "auth_required") {
        onAuthRequired();
        return;
      }

      setErrorMessage(getCreatorFollowErrorMessage(error));
    } finally {
      setIsPending(false);
    }
  };

  return {
    errorMessage,
    isFollowing,
    isPending,
    toggleFollow,
  };
}
