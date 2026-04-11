"use client";

import { useEffect, useState } from "react";

import { ApiError } from "@/shared/api";

import {
  getCreatorWorkspacePreviewMainDetail,
  getCreatorWorkspacePreviewShortDetail,
} from "../api/get-creator-workspace-preview-detail";
import type {
  CreatorWorkspacePreviewDetailData,
  CreatorWorkspacePreviewDetailSelection,
} from "../ui/creator-mode-shell.types";

type CreatorWorkspacePreviewDetailState =
  | {
      kind: "error";
      message: string;
    }
  | {
      kind: "loading";
    }
  | {
      detail: CreatorWorkspacePreviewDetailData;
      kind: "ready";
    };

type CreatorWorkspacePreviewDetailResolvedState =
  | {
      detail: CreatorWorkspacePreviewDetailData;
      kind: "ready";
      requestKey: string;
    }
  | {
      kind: "error";
      message: string;
      requestKey: string;
    };

type UseCreatorWorkspacePreviewDetailResult = {
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

/**
 * creator workspace owner preview detail の取得状態を client で管理する。
 */
export function useCreatorWorkspacePreviewDetail(
  selection: CreatorWorkspacePreviewDetailSelection,
): UseCreatorWorkspacePreviewDetailResult {
  const [retryCount, setRetryCount] = useState(0);
  const [resolvedState, setResolvedState] = useState<CreatorWorkspacePreviewDetailResolvedState | null>(null);
  const requestKey = `${selection.kind}:${selection.id}:${retryCount}`;

  useEffect(() => {
    const controller = new AbortController();

    const loadDetail = async (): Promise<CreatorWorkspacePreviewDetailData> => {
      if (selection.kind === "preview-main") {
        return {
          detail: await getCreatorWorkspacePreviewMainDetail(selection.id, {
            signal: controller.signal,
          }),
          kind: "preview-main",
        };
      }

      return {
        detail: await getCreatorWorkspacePreviewShortDetail(selection.id, {
          signal: controller.signal,
        }),
        kind: "preview-short",
      };
    };

    void loadDetail().then((detail) => {
      if (controller.signal.aborted) {
        return;
      }

      setResolvedState({
        detail,
        kind: "ready",
        requestKey,
      });
    }).catch((error: unknown) => {
      if (controller.signal.aborted) {
        return;
      }

      setResolvedState({
        kind: "error",
        message: buildCreatorWorkspacePreviewDetailErrorMessage(error),
        requestKey,
      });
    });

    return () => {
      controller.abort();
    };
  }, [requestKey, selection.id, selection.kind]);

  const state: CreatorWorkspacePreviewDetailState =
    resolvedState !== null && resolvedState.requestKey === requestKey
      ? (resolvedState.kind === "ready"
        ? {
            detail: resolvedState.detail,
            kind: "ready",
          }
        : {
            kind: "error",
            message: resolvedState.message,
          })
      : {
          kind: "loading",
        };

  return {
    retry: () => {
      setRetryCount((currentCount) => currentCount + 1);
    },
    state,
  };
}
