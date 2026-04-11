"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";

import {
  ShortPinApiError,
  updateShortPin,
} from "@/entities/short";
import { useHasViewerSession } from "@/entities/viewer";
import { buildFanLoginHref } from "@/features/fan-auth";
import { ApiError } from "@/shared/api";

type FeedPinSurface = {
  short: {
    id: string;
  };
  viewer: {
    isPinned: boolean;
  };
};

type FeedPinItemState = {
  errorMessage: string | null;
  isPending: boolean;
  isPinned: boolean;
};

type FeedPinInteraction = FeedPinItemState & {
  onToggle: () => void;
};

function buildFeedPinItemState(isPinned: boolean): FeedPinItemState {
  return {
    errorMessage: null,
    isPending: false,
    isPinned,
  };
}

function buildFeedPinStateByShortId(surfaces: readonly FeedPinSurface[]): Record<string, FeedPinItemState> {
  return Object.fromEntries(
    surfaces.map((surface) => [surface.short.id, buildFeedPinItemState(surface.viewer.isPinned)]),
  );
}

function mergeFeedPinState(
  currentStateByShortId: Record<string, FeedPinItemState>,
  surfaces: readonly FeedPinSurface[],
): Record<string, FeedPinItemState> {
  const nextStateByShortId = { ...currentStateByShortId };

  for (const surface of surfaces) {
    if (!(surface.short.id in nextStateByShortId)) {
      nextStateByShortId[surface.short.id] = buildFeedPinItemState(surface.viewer.isPinned);
    }
  }

  return nextStateByShortId;
}

function getFeedPinErrorMessage(error: unknown): string {
  if (error instanceof ShortPinApiError) {
    if (error.code === "not_found") {
      return "この short は現在利用できません。";
    }

    return "pin 状態を更新できませんでした。少し時間を置いてから再度お試しください。";
  }

  if (error instanceof ApiError) {
    return "pin 状態を更新できませんでした。通信状態を確認してから再度お試しください。";
  }

  return "pin 状態を更新できませんでした。少し時間を置いてから再度お試しください。";
}

/**
 * feed surface ごとの short pin pending / success / error state を管理する。
 */
export function useFeedPinState({ surfaces }: { surfaces: readonly FeedPinSurface[] }): {
  resolvePinState: (surface: FeedPinSurface) => FeedPinInteraction;
} {
  const hasViewerSession = useHasViewerSession();
  const pendingShortIdsRef = useRef<Set<string>>(new Set());
  const router = useRouter();
  const [pinStateByShortId, setPinStateByShortId] = useState<Record<string, FeedPinItemState>>(() =>
    buildFeedPinStateByShortId(surfaces),
  );

  useEffect(() => {
    setPinStateByShortId((currentStateByShortId) => mergeFeedPinState(currentStateByShortId, surfaces));
  }, [surfaces]);

  const togglePin = async (shortId: string, initialIsPinned: boolean) => {
    const currentState = pinStateByShortId[shortId] ?? buildFeedPinItemState(initialIsPinned);

    if (pendingShortIdsRef.current.has(shortId)) {
      return;
    }

    if (!hasViewerSession) {
      setPinStateByShortId((currentStateByShortId) => ({
        ...currentStateByShortId,
        [shortId]: {
          ...currentState,
          errorMessage: null,
        },
      }));
      router.push(buildFanLoginHref());
      return;
    }

    pendingShortIdsRef.current.add(shortId);
    setPinStateByShortId((currentStateByShortId) => ({
      ...currentStateByShortId,
      [shortId]: {
        ...(currentStateByShortId[shortId] ?? currentState),
        errorMessage: null,
        isPending: true,
      },
    }));

    try {
      const result = await updateShortPin({
        action: currentState.isPinned ? "unpin" : "pin",
        shortId,
      });

      setPinStateByShortId((currentStateByShortId) => ({
        ...currentStateByShortId,
        [shortId]: {
          errorMessage: null,
          isPending: false,
          isPinned: result.viewer.isPinned,
        },
      }));
    } catch (error) {
      if (error instanceof ShortPinApiError && error.code === "auth_required") {
        setPinStateByShortId((currentStateByShortId) => ({
          ...currentStateByShortId,
          [shortId]: {
            ...(currentStateByShortId[shortId] ?? currentState),
            errorMessage: null,
            isPending: false,
          },
        }));
        router.push(buildFanLoginHref());
        return;
      }

      setPinStateByShortId((currentStateByShortId) => ({
        ...currentStateByShortId,
        [shortId]: {
          ...(currentStateByShortId[shortId] ?? currentState),
          errorMessage: getFeedPinErrorMessage(error),
          isPending: false,
          isPinned: currentState.isPinned,
        },
      }));
    } finally {
      pendingShortIdsRef.current.delete(shortId);
    }
  };

  return {
    resolvePinState: (surface) => {
      const state = pinStateByShortId[surface.short.id] ?? buildFeedPinItemState(surface.viewer.isPinned);

      return {
        ...state,
        onToggle: () => {
          void togglePin(surface.short.id, surface.viewer.isPinned);
        },
      };
    },
  };
}
