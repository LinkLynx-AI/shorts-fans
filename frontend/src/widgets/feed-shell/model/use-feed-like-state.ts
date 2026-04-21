"use client";

import { useEffect, useRef, useState } from "react";

import {
  getShortLikeErrorMessage,
  ShortLikeApiError,
  updateShortLike,
} from "@/entities/short";
import { useHasViewerSession } from "@/entities/viewer";
import { useFanAuthDialogControls } from "@/features/fan-auth";

type FeedLikeSurface = {
  engagement: {
    likeCount: number;
  };
  short: {
    id: string;
  };
  viewer: {
    hasLiked: boolean;
  };
};

type FeedLikeItemState = {
  errorMessage: string | null;
  hasLiked: boolean;
  hasLocalOverride: boolean;
  isPending: boolean;
  likeCount: number;
};

type FeedLikeInteraction = FeedLikeItemState & {
  onToggle: () => void;
};

function buildFeedLikeItemState(surface: FeedLikeSurface): FeedLikeItemState {
  return {
    errorMessage: null,
    hasLiked: surface.viewer.hasLiked,
    hasLocalOverride: false,
    isPending: false,
    likeCount: surface.engagement.likeCount,
  };
}

function buildFeedLikeStateByShortId(surfaces: readonly FeedLikeSurface[]): Record<string, FeedLikeItemState> {
  return Object.fromEntries(surfaces.map((surface) => [surface.short.id, buildFeedLikeItemState(surface)]));
}

function buildTrackedShortIds(surfaces: readonly FeedLikeSurface[]): Set<string> {
  return new Set(surfaces.map((surface) => surface.short.id));
}

function matchesSurfaceState(state: FeedLikeItemState, surface: FeedLikeSurface): boolean {
  return state.hasLiked === surface.viewer.hasLiked && state.likeCount === surface.engagement.likeCount;
}

function mergeFeedLikeState(
  currentStateByShortId: Record<string, FeedLikeItemState>,
  surfaces: readonly FeedLikeSurface[],
): Record<string, FeedLikeItemState> {
  const nextStateByShortId: Record<string, FeedLikeItemState> = {};

  for (const surface of surfaces) {
    const currentState = currentStateByShortId[surface.short.id];

    if (!currentState) {
      nextStateByShortId[surface.short.id] = buildFeedLikeItemState(surface);
      continue;
    }

    if (currentState.isPending) {
      nextStateByShortId[surface.short.id] = currentState;
      continue;
    }

    if (currentState.hasLocalOverride) {
      nextStateByShortId[surface.short.id] = currentState.hasLiked === surface.viewer.hasLiked
        ? {
            ...currentState,
            errorMessage: null,
            hasLocalOverride: false,
            likeCount: surface.engagement.likeCount,
          }
        : currentState;
      continue;
    }

    nextStateByShortId[surface.short.id] = matchesSurfaceState(currentState, surface)
      ? currentState
      : {
          ...currentState,
          errorMessage: null,
          hasLiked: surface.viewer.hasLiked,
          likeCount: surface.engagement.likeCount,
        };
  }

  return nextStateByShortId;
}

/**
 * feed surface ごとの short like pending / success / error state を管理する。
 */
export function useFeedLikeState({ surfaces }: { surfaces: readonly FeedLikeSurface[] }): {
  resolveLikeState: (surface: FeedLikeSurface) => FeedLikeInteraction;
} {
  const hasViewerSession = useHasViewerSession();
  const { openFanAuthDialog } = useFanAuthDialogControls();
  const pendingShortIdsRef = useRef<Set<string>>(new Set());
  const trackedShortIdsRef = useRef<Set<string>>(buildTrackedShortIds(surfaces));
  const [likeStateByShortId, setLikeStateByShortId] = useState<Record<string, FeedLikeItemState>>(() =>
    buildFeedLikeStateByShortId(surfaces),
  );
  trackedShortIdsRef.current = buildTrackedShortIds(surfaces);

  useEffect(() => {
    setLikeStateByShortId((currentStateByShortId) => mergeFeedLikeState(currentStateByShortId, surfaces));
  }, [surfaces]);

  const toggleLike = async (surface: FeedLikeSurface) => {
    const shortId = surface.short.id;
    const currentState = likeStateByShortId[shortId] ?? buildFeedLikeItemState(surface);

    if (pendingShortIdsRef.current.has(shortId)) {
      return;
    }

    if (!hasViewerSession) {
      setLikeStateByShortId((currentStateByShortId) => {
        if (!trackedShortIdsRef.current.has(shortId)) {
          return currentStateByShortId;
        }

        return {
          ...currentStateByShortId,
          [shortId]: {
            ...currentState,
            errorMessage: null,
          },
        };
      });
      openFanAuthDialog({
        postAuthNavigation: "none",
      });
      return;
    }

    pendingShortIdsRef.current.add(shortId);
    setLikeStateByShortId((currentStateByShortId) => {
      if (!trackedShortIdsRef.current.has(shortId)) {
        return currentStateByShortId;
      }

      return {
        ...currentStateByShortId,
        [shortId]: {
          ...(currentStateByShortId[shortId] ?? currentState),
          errorMessage: null,
          isPending: true,
        },
      };
    });

    try {
      const result = await updateShortLike({
        action: currentState.hasLiked ? "unlike" : "like",
        shortId,
      });

      setLikeStateByShortId((currentStateByShortId) => {
        if (!trackedShortIdsRef.current.has(shortId)) {
          return currentStateByShortId;
        }

        return {
          ...currentStateByShortId,
          [shortId]: {
            errorMessage: null,
            hasLiked: result.viewer.hasLiked,
            hasLocalOverride: true,
            isPending: false,
            likeCount: result.engagement.likeCount,
          },
        };
      });
    } catch (error) {
      if (error instanceof ShortLikeApiError && error.code === "auth_required") {
        setLikeStateByShortId((currentStateByShortId) => {
          if (!trackedShortIdsRef.current.has(shortId)) {
            return currentStateByShortId;
          }

          return {
            ...currentStateByShortId,
            [shortId]: {
              ...(currentStateByShortId[shortId] ?? currentState),
              errorMessage: null,
              isPending: false,
            },
          };
        });
        openFanAuthDialog({
          postAuthNavigation: "none",
        });
        return;
      }

      setLikeStateByShortId((currentStateByShortId) => {
        if (!trackedShortIdsRef.current.has(shortId)) {
          return currentStateByShortId;
        }

        return {
          ...currentStateByShortId,
          [shortId]: {
            ...(currentStateByShortId[shortId] ?? currentState),
            errorMessage: getShortLikeErrorMessage(error),
            hasLiked: currentState.hasLiked,
            isPending: false,
            likeCount: currentState.likeCount,
          },
        };
      });
    } finally {
      pendingShortIdsRef.current.delete(shortId);
    }
  };

  return {
    resolveLikeState: (surface) => {
      const state = likeStateByShortId[surface.short.id] ?? buildFeedLikeItemState(surface);

      return {
        ...state,
        onToggle: () => {
          void toggleLike(surface);
        },
      };
    },
  };
}
