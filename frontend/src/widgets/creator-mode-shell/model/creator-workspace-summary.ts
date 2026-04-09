import type { CreatorSummary } from "@/entities/creator";
import { ApiError } from "@/shared/api";

import type {
  ApprovedCreatorWorkspaceOverviewMetrics,
  ApprovedCreatorWorkspaceRevisionRequestedSummary,
} from "./approved-creator-workspace";

export type CreatorWorkspaceSummary = {
  creator: CreatorSummary;
  overviewMetrics: ApprovedCreatorWorkspaceOverviewMetrics;
  revisionRequestedSummary: ApprovedCreatorWorkspaceRevisionRequestedSummary | null;
};

export type CreatorWorkspaceSummaryState =
  | {
      kind: "error";
      message: string;
    }
  | {
      kind: "loading";
    }
  | {
      kind: "ready";
      summary: CreatorWorkspaceSummary;
    };

const creatorWorkspaceSummaryErrorMessage =
  "creator workspace summary を読み込めませんでした。少し時間を置いてから再読み込みしてください。";
const creatorWorkspaceSummaryNetworkErrorMessage =
  "creator workspace summary を読み込めませんでした。通信環境を確認してから再読み込みしてください。";

/**
 * creator workspace summary の loading state を組み立てる。
 */
export function buildLoadingCreatorWorkspaceSummaryState(): CreatorWorkspaceSummaryState {
  return {
    kind: "loading",
  };
}

/**
 * creator workspace summary の error state を組み立てる。
 */
export function buildErrorCreatorWorkspaceSummaryState(error: unknown): CreatorWorkspaceSummaryState {
  if (error instanceof ApiError && error.code === "network") {
    return {
      kind: "error",
      message: creatorWorkspaceSummaryNetworkErrorMessage,
    };
  }

  return {
    kind: "error",
    message: creatorWorkspaceSummaryErrorMessage,
  };
}
