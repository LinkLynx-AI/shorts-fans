"use client";

import { useState } from "react";

import type { FanFollowingItem } from "@/entities/fan-profile";

type FollowingCreatorRelationAction = "follow" | "unfollow";

export type UpdateFollowingCreatorRelation = (input: {
  action: FollowingCreatorRelationAction;
  creatorId: string;
}) => Promise<void>;

type FollowingCreatorRow = {
  creator: FanFollowingItem["creator"];
  isFollowing: boolean;
  isPending: boolean;
};

type UseFollowingCreatorRowsOptions = {
  items: readonly FanFollowingItem[];
  updateFollowingCreatorRelation?: UpdateFollowingCreatorRelation | undefined;
};

type UseFollowingCreatorRowsResult = {
  rows: readonly FollowingCreatorRow[];
  toggleFollowing: (creatorId: string) => Promise<void>;
};

const noopUpdateFollowingCreatorRelation: UpdateFollowingCreatorRelation = async () => undefined;

function buildFollowingState(items: readonly FanFollowingItem[]): Record<string, boolean> {
  return Object.fromEntries(items.map((item) => [item.creator.id, item.viewer.isFollowing]));
}

function getInitialFollowingState(items: readonly FanFollowingItem[], creatorId: string): boolean | undefined {
  return items.find((item) => item.creator.id === creatorId)?.viewer.isFollowing;
}

function removePendingState(
  pendingByCreatorId: Record<string, boolean>,
  creatorId: string,
): Record<string, boolean> {
  if (!(creatorId in pendingByCreatorId)) {
    return pendingByCreatorId;
  }

  const nextPendingByCreatorId = { ...pendingByCreatorId };

  delete nextPendingByCreatorId[creatorId];

  return nextPendingByCreatorId;
}

/**
 * following 一覧の row ごとの follow state と pending state を管理する。
 */
export function useFollowingCreatorRows({
  items,
  updateFollowingCreatorRelation = noopUpdateFollowingCreatorRelation,
}: UseFollowingCreatorRowsOptions): UseFollowingCreatorRowsResult {
  const [followingByCreatorId, setFollowingByCreatorId] = useState<Record<string, boolean>>(() => buildFollowingState(items));
  const [pendingByCreatorId, setPendingByCreatorId] = useState<Record<string, boolean>>({});

  const rows = items.map((item) => ({
    creator: item.creator,
    isFollowing: followingByCreatorId[item.creator.id] ?? item.viewer.isFollowing,
    isPending: Boolean(pendingByCreatorId[item.creator.id]),
  }));

  const toggleFollowing = async (creatorId: string) => {
    if (pendingByCreatorId[creatorId]) {
      return;
    }

    const isFollowing = followingByCreatorId[creatorId] ?? getInitialFollowingState(items, creatorId);

    if (isFollowing === undefined) {
      return;
    }

    const action = isFollowing ? "unfollow" : "follow";

    setPendingByCreatorId((currentPendingByCreatorId) => ({
      ...currentPendingByCreatorId,
      [creatorId]: true,
    }));

    try {
      await updateFollowingCreatorRelation({ action, creatorId });

      setFollowingByCreatorId((currentFollowingByCreatorId) => ({
        ...currentFollowingByCreatorId,
        [creatorId]: !isFollowing,
      }));
    } catch (error) {
      throw error;
    } finally {
      setPendingByCreatorId((currentPendingByCreatorId) =>
        removePendingState(currentPendingByCreatorId, creatorId),
      );
    }
  };

  return {
    rows,
    toggleFollowing,
  };
}
