import type { CreatorId } from "@/entities/creator";
import type { ShortId } from "@/entities/short";

const followedCreatorIds = new Set<CreatorId>(["aoi", "mina", "sora"]);
const pinnedShortIds = new Set<ShortId>(["afterrain", "balcony", "rooftop"]);

/**
 * mock follow state を返す。
 */
export function isCreatorFollowed(creatorId: CreatorId): boolean {
  return followedCreatorIds.has(creatorId);
}

/**
 * mock pin state を返す。
 */
export function isShortPinned(shortId: ShortId): boolean {
  return pinnedShortIds.has(shortId);
}
