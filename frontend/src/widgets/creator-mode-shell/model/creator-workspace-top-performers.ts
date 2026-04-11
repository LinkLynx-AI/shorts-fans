import { ApiError } from "@/shared/api";

import type { CreatorWorkspaceTopPerformersResponse } from "../api/get-creator-workspace-top-performers";

export type CreatorWorkspaceTopPerformersState =
  | {
      kind: "error";
      message: string;
    }
  | {
      kind: "loading";
    }
  | {
      kind: "ready";
      topPerformers: CreatorWorkspaceTopPerformersResponse;
    };

const creatorWorkspaceTopPerformersErrorMessage =
  "top performers を読み込めませんでした。少し時間を置いてから再読み込みしてください。";
const creatorWorkspaceTopPerformersNetworkErrorMessage =
  "top performers を読み込めませんでした。通信環境を確認してから再読み込みしてください。";

/**
 * creator workspace top performers の loading state を組み立てる。
 */
export function buildLoadingCreatorWorkspaceTopPerformersState(): CreatorWorkspaceTopPerformersState {
  return {
    kind: "loading",
  };
}

/**
 * creator workspace top performers の error state を組み立てる。
 */
export function buildErrorCreatorWorkspaceTopPerformersState(error: unknown): CreatorWorkspaceTopPerformersState {
  if (error instanceof ApiError && error.code === "network") {
    return {
      kind: "error",
      message: creatorWorkspaceTopPerformersNetworkErrorMessage,
    };
  }

  return {
    kind: "error",
    message: creatorWorkspaceTopPerformersErrorMessage,
  };
}
