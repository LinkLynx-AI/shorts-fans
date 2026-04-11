import { ApiError } from "@/shared/api";

import type {
  CreatorWorkspacePreviewMainListPage,
  CreatorWorkspacePreviewShortListPage,
} from "../api/get-creator-workspace-preview-collections";

export type CreatorWorkspacePreviewCollections = {
  mains: CreatorWorkspacePreviewMainListPage;
  shorts: CreatorWorkspacePreviewShortListPage;
};

export type CreatorWorkspacePreviewCollectionsState =
  | {
      kind: "error";
      message: string;
    }
  | {
      kind: "loading";
    }
  | {
      collections: CreatorWorkspacePreviewCollections;
      kind: "ready";
    };

const creatorWorkspacePreviewCollectionsErrorMessage =
  "動画一覧を読み込めませんでした。少し時間を置いてから再読み込みしてください。";
const creatorWorkspacePreviewCollectionsNetworkErrorMessage =
  "動画一覧を読み込めませんでした。通信環境を確認してから再読み込みしてください。";

/**
 * creator workspace preview list の loading state を組み立てる。
 */
export function buildLoadingCreatorWorkspacePreviewCollectionsState(): CreatorWorkspacePreviewCollectionsState {
  return {
    kind: "loading",
  };
}

/**
 * creator workspace preview list の error state を組み立てる。
 */
export function buildErrorCreatorWorkspacePreviewCollectionsState(
  error: unknown,
): CreatorWorkspacePreviewCollectionsState {
  if (error instanceof ApiError && error.code === "network") {
    return {
      kind: "error",
      message: creatorWorkspacePreviewCollectionsNetworkErrorMessage,
    };
  }

  return {
    kind: "error",
    message: creatorWorkspacePreviewCollectionsErrorMessage,
  };
}

/**
 * preview main item に保存済み price の override を反映する。
 */
export function applyCreatorWorkspaceMainPriceOverrides(
  state: CreatorWorkspacePreviewCollectionsState,
  priceByMainId: Readonly<Record<string, number>>,
): CreatorWorkspacePreviewCollectionsState {
  if (state.kind !== "ready") {
    return state;
  }

  return {
    collections: {
      mains: {
        ...state.collections.mains,
        items: state.collections.mains.items.map((item) => ({
          ...item,
          priceJpy: priceByMainId[item.id] ?? item.priceJpy,
        })),
      },
      shorts: state.collections.shorts,
    },
    kind: "ready",
  };
}
