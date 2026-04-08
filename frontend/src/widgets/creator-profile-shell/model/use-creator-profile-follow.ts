"use client";

import { useState } from "react";

import {
  CreatorFollowApiError,
  updateCreatorFollow,
} from "@/entities/creator";
import { ApiError } from "@/shared/api";

type UseCreatorProfileFollowOptions = {
  creatorId: string;
  hasViewerSession: boolean;
  initialFanCount: number;
  initialIsFollowing: boolean;
  onAuthRequired: () => void;
  onUnauthenticated: () => void;
};

type UseCreatorProfileFollowResult = {
  errorMessage: string | null;
  fanCount: number;
  isFollowing: boolean;
  isPending: boolean;
  toggleFollow: () => Promise<void>;
};

function getCreatorFollowErrorMessage(error: unknown): string {
  if (error instanceof CreatorFollowApiError) {
    if (error.code === "not_found") {
      return "この creator profile は現在利用できません。";
    }

    return "フォロー状態を更新できませんでした。少し時間を置いてから再度お試しください。";
  }

  if (error instanceof ApiError) {
    return "フォロー状態を更新できませんでした。通信状態を確認してから再度お試しください。";
  }

  return "フォロー状態を更新できませんでした。少し時間を置いてから再度お試しください。";
}

/**
 * creator profile follow CTA の pending / success / error state を管理する。
 */
export function useCreatorProfileFollow({
  creatorId,
  hasViewerSession,
  initialFanCount,
  initialIsFollowing,
  onAuthRequired,
  onUnauthenticated,
}: UseCreatorProfileFollowOptions): UseCreatorProfileFollowResult {
  const [fanCount, setFanCount] = useState(initialFanCount);
  const [isFollowing, setIsFollowing] = useState(initialIsFollowing);
  const [isPending, setIsPending] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

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
      const result = await updateCreatorFollow({
        action,
        creatorId,
      });

      setFanCount(result.stats.fanCount);
      setIsFollowing(result.viewer.isFollowing);
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
    fanCount,
    isFollowing,
    isPending,
    toggleFollow,
  };
}
