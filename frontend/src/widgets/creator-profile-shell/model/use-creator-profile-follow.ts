"use client";

import { useState } from "react";

import {
  updateCreatorFollow,
  useCreatorFollowToggle,
} from "@/entities/creator";

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
  const { errorMessage, isFollowing, isPending, toggleFollow } = useCreatorFollowToggle({
    creatorId,
    hasViewerSession,
    initialIsFollowing,
    onAuthRequired,
    onSuccess: (result) => {
      setFanCount(result.stats.fanCount);
    },
    onUnauthenticated,
    updateFollow: updateCreatorFollow,
  });

  return {
    errorMessage,
    fanCount,
    isFollowing,
    isPending,
    toggleFollow,
  };
}
