import {
  getCreatorById,
  getCreatorProfileStatsById,
  type CreatorProfileStats,
  type CreatorSummary,
} from "@/entities/creator";
import { getShortById, type ShortId, type ShortPreviewMeta } from "@/entities/short";

export type CreatorProfileShellState =
  | {
      creator: CreatorSummary;
      kind: "empty";
      shorts: readonly [];
      stats: CreatorProfileStats;
      viewer: {
        isFollowing: boolean;
      };
    }
  | {
      creator: CreatorSummary;
      kind: "ready";
      shorts: readonly ShortPreviewMeta[];
      stats: CreatorProfileStats;
      viewer: {
        isFollowing: boolean;
      };
  };

type SearchPrefillCreator = {
  displayName?: string | undefined;
  handle?: string | undefined;
};

const creatorProfileShortIdsById: Record<string, readonly ShortId[]> = {
  aoi: ["softlight", "balcony"],
  creator_11111111111111111111111111111111: [],
  creator_aoi_n: ["softlight", "balcony"],
  creator_mina_rei: ["rooftop", "mirror"],
  creator_sora_vale: [],
  mina: ["rooftop", "mirror"],
  sora: [],
};

const creatorProfileViewerStateById: Record<string, { isFollowing: boolean }> = {
  aoi: { isFollowing: true },
  creator_11111111111111111111111111111111: { isFollowing: false },
  creator_aoi_n: { isFollowing: true },
  creator_mina_rei: { isFollowing: true },
  creator_sora_vale: { isFollowing: false },
  mina: { isFollowing: true },
  sora: { isFollowing: false },
};

function buildProvisionalCreatorSummary(
  creatorId: string,
  prefilledCreator?: SearchPrefillCreator,
): CreatorSummary | undefined {
  const displayName = prefilledCreator?.displayName?.trim();
  const handle = prefilledCreator?.handle?.trim();

  if (!displayName || !handle || !handle.startsWith("@")) {
    return undefined;
  }

  return {
    avatar: null,
    bio: "",
    displayName,
    handle: handle as `@${string}`,
    id: creatorId,
  };
}

/**
 * creator profile 用の shell state を返す。
 */
export function getCreatorProfileShellState(
  creatorId: string,
  prefilledCreator?: SearchPrefillCreator,
): CreatorProfileShellState | undefined {
  const creator = getCreatorById(creatorId) ?? buildProvisionalCreatorSummary(creatorId, prefilledCreator);
  const stats = getCreatorProfileStatsById(creatorId) ?? (creator
    ? { fanCount: 0, shortCount: 0, viewCount: 0 }
    : undefined);
  const viewer = creatorProfileViewerStateById[creatorId] ?? (creator
    ? { isFollowing: false }
    : undefined);

  if (!creator || !stats || !viewer) {
    return undefined;
  }

  const shorts = (creatorProfileShortIdsById[creatorId] ?? []).flatMap((shortId) => {
    const short = getShortById(shortId);

    return short ? [short] : [];
  });

  if (shorts.length === 0) {
    return {
      creator,
      kind: "empty",
      shorts: [],
      stats,
      viewer,
    };
  }

  return {
    creator,
    kind: "ready",
    shorts,
    stats,
    viewer,
  };
}
