"use client";

import { useEffect, useState } from "react";

import {
  getCreatorWorkspaceMainReviewSurface,
  getCreatorWorkspaceShortReviewSurface,
} from "../api/get-creator-workspace-review-surface";
import type { CreatorWorkspacePreviewDetailSelection } from "../ui/creator-mode-shell.types";
import { resolveCreatorWorkspaceBlockedState } from "./creator-workspace-blocked-state";
import {
  buildErrorCreatorWorkspaceItemReviewSurfaceState,
  type CreatorWorkspaceItemReviewSurfaceState,
} from "./creator-workspace-review-surface";
import type { CreatorModeShellBlockedState } from "./creator-mode-shell";

type InternalCreatorWorkspaceItemReviewSurfaceState =
  | {
      kind: "error";
      message: string;
      selectionKey: string;
    }
  | {
      kind: "idle";
      selectionKey: null;
    }
  | {
      kind: "loading";
      selectionKey: string;
    }
  | {
      kind: "ready";
      selectionKey: string;
      surface: Awaited<ReturnType<typeof getCreatorWorkspaceMainReviewSurface>>;
    };

type UseCreatorWorkspaceItemReviewSurfaceResult = {
  blockedState: CreatorModeShellBlockedState | null;
  retry: () => void;
  state: CreatorWorkspaceItemReviewSurfaceState;
};

function getCreatorWorkspaceItemReviewSelectionKey(
  selection: CreatorWorkspacePreviewDetailSelection | null,
): string | null {
  if (!selection) {
    return null;
  }

  return `${selection.kind}:${selection.item.id}`;
}

/**
 * creator workspace detail の item review surface を client で取得する。
 */
export function useCreatorWorkspaceItemReviewSurface(
  selection: CreatorWorkspacePreviewDetailSelection | null,
): UseCreatorWorkspaceItemReviewSurfaceResult {
  const [blockedState, setBlockedState] = useState<{
    selectionKey: string;
    state: CreatorModeShellBlockedState;
  } | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const [state, setState] = useState<InternalCreatorWorkspaceItemReviewSurfaceState>({
    kind: "idle",
    selectionKey: null,
  });
  const selectionKey = getCreatorWorkspaceItemReviewSelectionKey(selection);
  const mainId = selection?.kind === "preview-main" ? selection.item.id : null;
  const shortId = selection?.kind === "preview-short" ? selection.item.id : null;

  useEffect(() => {
    if (!selectionKey) {
      return;
    }

    const controller = new AbortController();

    const loadSurfacePromise = shortId
      ? getCreatorWorkspaceShortReviewSurface(shortId, {
          credentials: "include",
          signal: controller.signal,
        })
      : mainId
        ? getCreatorWorkspaceMainReviewSurface(mainId, {
            credentials: "include",
            signal: controller.signal,
          })
        : Promise.reject(new Error("review surface selection is missing a target id"));

    void loadSurfacePromise.then((surface) => {
      if (controller.signal.aborted) {
        return;
      }

      setState({
        kind: "ready",
        selectionKey,
        surface,
      });
    }).catch((error: unknown) => {
      if (controller.signal.aborted) {
        return;
      }

      const nextBlockedState = resolveCreatorWorkspaceBlockedState(error);

      if (nextBlockedState) {
        setBlockedState({
          selectionKey,
          state: nextBlockedState,
        });
        return;
      }

      setState({
        kind: "error",
        message: buildErrorCreatorWorkspaceItemReviewSurfaceState(error).message,
        selectionKey,
      });
    });

    return () => {
      controller.abort();
    };
  }, [mainId, retryCount, selectionKey, shortId]);

  if (!selectionKey) {
    return {
      blockedState: null,
      retry: () => {
        setRetryCount((currentCount) => currentCount + 1);
      },
      state: {
        kind: "idle",
      },
    };
  }

  return {
    blockedState: blockedState?.selectionKey === selectionKey ? blockedState.state : null,
    retry: () => {
      setBlockedState(null);
      setState({
        kind: "loading",
        selectionKey,
      });
      setRetryCount((currentCount) => currentCount + 1);
    },
    state:
      state.selectionKey === selectionKey
        ? state.kind === "ready"
          ? {
              kind: "ready",
              surface: state.surface,
            }
          : state.kind === "error"
            ? {
                kind: "error",
                message: state.message,
              }
            : {
                kind: state.kind,
              }
        : {
            kind: "loading",
          },
  };
}
