"use client";

import { useEffect, useState } from "react";

import { getCreatorWorkspaceReviewSurface } from "../api/get-creator-workspace-review-surface";
import { resolveCreatorWorkspaceBlockedState } from "./creator-workspace-blocked-state";
import {
  buildErrorCreatorWorkspaceReviewSurfaceState,
  buildLoadingCreatorWorkspaceReviewSurfaceState,
  type CreatorWorkspaceReviewSurfaceState,
} from "./creator-workspace-review-surface";
import type { CreatorModeShellBlockedState } from "./creator-mode-shell";

type UseCreatorWorkspaceReviewSurfaceResult = {
  blockedState: CreatorModeShellBlockedState | null;
  retry: () => void;
  state: CreatorWorkspaceReviewSurfaceState;
};

/**
 * creator workspace dashboard の review surface を client で取得する。
 */
export function useCreatorWorkspaceReviewSurface(): UseCreatorWorkspaceReviewSurfaceResult {
  const [blockedState, setBlockedState] = useState<CreatorModeShellBlockedState | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const [state, setState] = useState<CreatorWorkspaceReviewSurfaceState>(buildLoadingCreatorWorkspaceReviewSurfaceState());

  useEffect(() => {
    const controller = new AbortController();

    void getCreatorWorkspaceReviewSurface({
      credentials: "include",
      signal: controller.signal,
    }).then((surface) => {
      if (controller.signal.aborted) {
        return;
      }

      setState({
        kind: "ready",
        surface,
      });
    }).catch((error: unknown) => {
      if (controller.signal.aborted) {
        return;
      }

      const nextBlockedState = resolveCreatorWorkspaceBlockedState(error);

      if (nextBlockedState) {
        setBlockedState(nextBlockedState);
        return;
      }

      setState(buildErrorCreatorWorkspaceReviewSurfaceState(error));
    });

    return () => {
      controller.abort();
    };
  }, [retryCount]);

  return {
    blockedState,
    retry: () => {
      setBlockedState(null);
      setState(buildLoadingCreatorWorkspaceReviewSurfaceState());
      setRetryCount((currentCount) => currentCount + 1);
    },
    state,
  };
}
