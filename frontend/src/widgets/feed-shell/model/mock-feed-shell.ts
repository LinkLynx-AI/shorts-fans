import type { FeedTab } from "@/entities/short";
import { getFeedSurfaceByTab, type FeedShortSurface } from "@/widgets/immersive-short-surface";

export type FeedShellPageInfo = {
  hasNext: boolean;
  nextCursor: string | null;
};

export type FeedShellState =
  | {
      kind: "auth_required";
      tab: "following";
    }
  | {
      kind: "empty";
      tab: FeedTab;
    }
  | {
      detailHref?: string;
      kind: "ready";
      page: FeedShellPageInfo | null;
      surface: FeedShortSurface;
      tab: FeedTab;
    };

/**
 * ready な feed shell state を組み立てる。
 */
export function createReadyFeedShellState({
  detailHref,
  page = null,
  surface,
  tab,
}: {
  detailHref?: string;
  page?: FeedShellPageInfo | null;
  surface: FeedShortSurface;
  tab: FeedTab;
}): FeedShellState {
  return {
    kind: "ready",
    page,
    surface,
    tab,
    ...(detailHref ? { detailHref } : {}),
  };
}

/**
 * empty な feed shell state を組み立てる。
 */
export function getEmptyFeedShellState(tab: FeedTab): FeedShellState {
  return {
    kind: "empty",
    tab,
  };
}

/**
 * mock feed state を返す。
 * `following` の状態分岐は shell 側で表現できるように残す。
 */
export function getMockFeedShellState(tab: FeedTab): FeedShellState {
  return createReadyFeedShellState({
    detailHref: `/shorts/${getFeedSurfaceByTab(tab).short.id}`,
    surface: getFeedSurfaceByTab(tab),
    tab,
  });
}

/**
 * following feed の非 ready state を組み立てる。
 */
export function getFollowingFeedShellState(kind: "auth_required" | "empty"): FeedShellState {
  if (kind === "empty") {
    return getEmptyFeedShellState("following");
  }

  return {
    kind,
    tab: "following",
  };
}
