import {
  getCreatorProfileShortGrid,
} from "@/entities/creator";
import {
  fetchFanProfilePinnedShortsPage,
  type FanHubTab,
} from "@/entities/fan-profile";

export type ShortDetailReelState = {
  initialIndex: number;
  shortIds: readonly string[];
  sourceTab?: FanHubTab | undefined;
};

type LoadShortDetailReelStateOptions =
  | {
      creatorId: string;
      kind: "creator";
      sessionToken?: string | undefined;
      shortId: string;
    }
  | {
      kind: "fan";
      sessionToken?: string | undefined;
      shortId: string;
      tab: "pinned";
    };

async function loadShortIdsFromSource(
  options: LoadShortDetailReelStateOptions,
): Promise<readonly string[]> {
  if (options.kind === "creator") {
    const page = await getCreatorProfileShortGrid({
      creatorId: options.creatorId,
    });

    return page.items.map((item) => item.id);
  }

  const page = await fetchFanProfilePinnedShortsPage({
    sessionToken: options.sessionToken,
  });

  return page.items.map((item) => item.short.id);
}

/**
 * profile/list 起点の short reel state を取得する。
 */
export async function loadShortDetailReelState(
  options: LoadShortDetailReelStateOptions,
): Promise<ShortDetailReelState | null> {
  const shortIds = await loadShortIdsFromSource(options);
  const initialIndex = shortIds.indexOf(options.shortId);

  if (initialIndex < 0) {
    return null;
  }

  return {
    initialIndex,
    shortIds,
    ...(options.kind === "fan" ? { sourceTab: options.tab } : {}),
  };
}
