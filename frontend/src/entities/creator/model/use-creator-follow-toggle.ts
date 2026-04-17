"use client";

import { useState } from "react";

import { useCurrentViewer } from "@/entities/viewer";

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
  const currentViewer = useCurrentViewer();
  const viewerIdentityKey = currentViewer?.id ?? null;
  const viewerScopeKey = `${creatorId}::${viewerIdentityKey ?? "anonymous"}`;
  const [followState, setFollowState] = useState<{
    errorMessage: string | null;
    isFollowing: boolean;
    isPending: boolean;
    viewerScopeKey: string;
  }>(() => ({
    errorMessage: null,
    isFollowing: initialIsFollowing,
    isPending: false,
    viewerScopeKey,
  }));
  const isCurrentViewerScope = followState.viewerScopeKey === viewerScopeKey;
  const errorMessage = isCurrentViewerScope ? followState.errorMessage : null;
  const isFollowing = isCurrentViewerScope ? followState.isFollowing : initialIsFollowing;
  const isPending = isCurrentViewerScope ? followState.isPending : false;

  const toggleFollow = async () => {
    if (isPending) {
      return;
    }

    if (!hasViewerSession) {
      setFollowState({
        errorMessage: null,
        isFollowing,
        isPending: false,
        viewerScopeKey,
      });
      onUnauthenticated();
      return;
    }

    const action = isFollowing ? "unfollow" : "follow";

    setFollowState({
      errorMessage: null,
      isFollowing,
      isPending: true,
      viewerScopeKey,
    });

    try {
      const result = await (updateFollow ?? updateCreatorFollow)({
        action,
        creatorId,
      });

      setFollowState({
        errorMessage: null,
        isFollowing: result.viewer.isFollowing,
        isPending: false,
        viewerScopeKey,
      });
      onSuccess?.(result);
    } catch (error) {
      if (error instanceof CreatorFollowApiError && error.code === "auth_required") {
        setFollowState({
          errorMessage: null,
          isFollowing,
          isPending: false,
          viewerScopeKey,
        });
        onAuthRequired();
        return;
      }

      setFollowState({
        errorMessage: getCreatorFollowErrorMessage(error),
        isFollowing,
        isPending: false,
        viewerScopeKey,
      });
    }
  };

  return {
    errorMessage,
    isFollowing,
    isPending,
    toggleFollow,
  };
}
