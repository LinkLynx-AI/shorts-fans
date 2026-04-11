import type { FeedTab } from "@/entities/short";
import { getFeedSurfaceByTab, type FeedShortSurface } from "@/widgets/immersive-short-surface";

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
      kind: "ready";
      surfaces: readonly FeedShortSurface[];
      tab: FeedTab;
    };

/**
 * mock feed state を返す。
 * `following` の状態分岐は shell 側で表現できるように残す。
 */
export function getMockFeedShellState(tab: FeedTab): FeedShellState {
  return {
    kind: "ready",
    surfaces: [getFeedSurfaceByTab(tab)],
    tab,
  };
}

/**
 * following feed の非 ready state を組み立てる。
 */
export function getFollowingFeedShellState(kind: "auth_required" | "empty"): FeedShellState {
  return {
    kind,
    tab: "following",
  };
}
