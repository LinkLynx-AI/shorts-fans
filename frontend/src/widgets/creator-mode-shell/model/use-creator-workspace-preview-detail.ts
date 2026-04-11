"use client";

import { useEffect, useState } from "react";

import { ApiError } from "@/shared/api";

import type {
  CreatorWorkspacePreviewDetail,
} from "../api/get-creator-workspace-preview-detail";
import {
  getCreatorWorkspacePreviewMainDetail,
  getCreatorWorkspacePreviewShortDetail,
} from "../api/get-creator-workspace-preview-detail";
import type { CreatorWorkspacePreviewDetailSelection } from "../ui/creator-mode-shell.types";
import { resolveCreatorWorkspaceBlockedState } from "./creator-workspace-blocked-state";
import type { CreatorModeShellBlockedState } from "./creator-mode-shell";

export type CreatorWorkspacePreviewDetailState =
  | {
      kind: "error";
      message: string;
    }
  | {
      kind: "idle";
    }
  | {
      kind: "loading";
    }
  | {
      detail: CreatorWorkspacePreviewDetail;
      kind: "ready";
    };

type InternalCreatorWorkspacePreviewDetailState =
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
      detail: CreatorWorkspacePreviewDetail;
      kind: "ready";
      selectionKey: string;
    };

type UseCreatorWorkspacePreviewDetailResult = {
  blockedState: CreatorModeShellBlockedState | null;
  retry: () => void;
  state: CreatorWorkspacePreviewDetailState;
};

const creatorWorkspacePreviewDetailErrorMessage =
  "動画詳細を読み込めませんでした。少し時間を置いてから再読み込みしてください。";
const creatorWorkspacePreviewDetailNetworkErrorMessage =
  "動画詳細を読み込めませんでした。通信環境を確認してから再読み込みしてください。";

function buildCreatorWorkspacePreviewDetailErrorMessage(error: unknown): string {
  if (error instanceof ApiError && error.code === "network") {
    return creatorWorkspacePreviewDetailNetworkErrorMessage;
  }

  return creatorWorkspacePreviewDetailErrorMessage;
}

function getCreatorWorkspacePreviewSelectionKey(
  selection: CreatorWorkspacePreviewDetailSelection | null,
): string | null {
  if (!selection) {
    return null;
  }

  return `${selection.kind}:${selection.item.id}`;
}

/**
 * creator workspace preview detail の取得状態を client で管理する。
 */
export function useCreatorWorkspacePreviewDetail(
  selection: CreatorWorkspacePreviewDetailSelection | null,
): UseCreatorWorkspacePreviewDetailResult {
  const [blockedState, setBlockedState] = useState<{
    selectionKey: string;
    state: CreatorModeShellBlockedState;
  } | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const [state, setState] = useState<InternalCreatorWorkspacePreviewDetailState>({
    kind: "idle",
    selectionKey: null,
  });
  const selectionKey = getCreatorWorkspacePreviewSelectionKey(selection);
  const mainId = selection?.kind === "preview-main" ? selection.item.id : null;
  const shortId = selection?.kind === "preview-short" ? selection.item.id : null;

  useEffect(() => {
    if (!selectionKey) {
      return;
    }

    const controller = new AbortController();

    const loadDetailPromise = shortId
      ? getCreatorWorkspacePreviewShortDetail(shortId, {
          credentials: "include",
          signal: controller.signal,
        })
      : mainId
        ? getCreatorWorkspacePreviewMainDetail(mainId, {
            credentials: "include",
            signal: controller.signal,
          })
        : Promise.reject(new Error("preview detail selection is missing a target id"));

    void loadDetailPromise.then((detail) => {
      if (controller.signal.aborted) {
        return;
      }

      setState({
        detail,
        kind: "ready",
        selectionKey,
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
        message: buildCreatorWorkspacePreviewDetailErrorMessage(error),
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
              detail: state.detail,
              kind: "ready",
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
