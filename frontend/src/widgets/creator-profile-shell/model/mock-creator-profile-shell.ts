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

const creatorProfileShortIdsById: Record<string, readonly ShortId[]> = {
  aoi: ["softlight", "balcony"],
  mina: ["rooftop", "mirror"],
  sora: [],
};

const creatorProfileViewerStateById: Record<string, { isFollowing: boolean }> = {
  aoi: { isFollowing: true },
  mina: { isFollowing: true },
  sora: { isFollowing: false },
};

/**
 * creator profile 用の shell state を返す。
 */
export function getCreatorProfileShellState(
  creatorId: string,
): CreatorProfileShellState | undefined {
  const creator = getCreatorById(creatorId);
  const stats = getCreatorProfileStatsById(creatorId);
  const viewer = creatorProfileViewerStateById[creatorId];

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
