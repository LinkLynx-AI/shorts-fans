import { getFanFeedPage, type FeedTab } from "@/entities/short";
import { isAuthRequiredApiError } from "@/features/fan-auth";
import { buildFeedSurfaceFromApiItem } from "@/widgets/immersive-short-surface";

import type { FeedShellState } from "./mock-feed-shell";

type LoadFeedShellStateOptions = {
  sessionToken?: string | undefined;
};

/**
 * fan feed shell 用の初期 state を API から取得する。
 */
export async function loadFeedShellState(
  tab: FeedTab,
  options: LoadFeedShellStateOptions = {},
): Promise<FeedShellState> {
  try {
    const response = await getFanFeedPage({
      sessionToken: options.sessionToken,
      tab,
    });

    if (response.items.length === 0) {
      return {
        kind: "empty",
        tab,
      };
    }

    return {
      kind: "ready",
      surfaces: response.items.map(buildFeedSurfaceFromApiItem),
      tab: response.tab,
    };
  } catch (error) {
    if (tab === "following" && isAuthRequiredApiError(error)) {
      return {
        kind: "auth_required",
        tab: "following",
      };
    }

    throw error;
  }
}
