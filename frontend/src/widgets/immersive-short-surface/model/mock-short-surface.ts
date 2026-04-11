import { getCreatorById } from "@/entities/creator";
import {
  getFeedShortForTab,
  getShortById,
  type FeedTab,
  type ShortId,
} from "@/entities/short";
import { getUnlockSurfaceByShortId } from "@/features/unlock-entry";

import type {
  DetailShortSurface,
  DetailSurfaceViewerState,
  FeedShortSurface,
  FeedSurfaceViewerState,
  ShortSurfaceBase,
} from "./short-surface";

const feedViewerStateByShortId: Record<string, FeedSurfaceViewerState> = {
  afterrain: { isFollowingCreator: true, isPinned: true },
  balcony: { isFollowingCreator: true, isPinned: true },
  mirror: { isFollowingCreator: true, isPinned: false },
  poolcut: { isFollowingCreator: true, isPinned: false },
  rooftop: { isFollowingCreator: false, isPinned: true },
  softlight: { isFollowingCreator: true, isPinned: false },
};

const detailViewerStateByShortId: Record<string, DetailSurfaceViewerState> = {
  afterrain: { isFollowingCreator: true, isPinned: true },
  balcony: { isFollowingCreator: true, isPinned: true },
  mirror: { isFollowingCreator: true, isPinned: false },
  poolcut: { isFollowingCreator: true, isPinned: false },
  rooftop: { isFollowingCreator: true, isPinned: true },
  softlight: { isFollowingCreator: true, isPinned: false },
};

function buildShortSurfaceBase(shortId: ShortId): ShortSurfaceBase | undefined {
  const short = getShortById(shortId);

  if (!short) {
    return undefined;
  }

  const creator = getCreatorById(short.creatorId);

  if (!creator) {
    throw new Error(`Unknown creator for short surface: ${short.id}`);
  }

  const unlock = getUnlockSurfaceByShortId(short.id);

  if (!unlock) {
    throw new Error(`Unknown unlock surface state for short: ${short.id}`);
  }

  return {
    creator,
    mainEntryEnabled: true,
    short,
    unlock,
  };
}

/**
 * feed tab に対応する shorts surface 用 mock を返す。
 */
export function getFeedSurfaceByTab(tab: FeedTab): FeedShortSurface {
  const short = getFeedShortForTab(tab);
  const surface = buildShortSurfaceBase(short.id);

  if (!surface) {
    throw new Error(`Unknown feed short surface for tab: ${tab}`);
  }

  const viewer = feedViewerStateByShortId[short.id];

  if (!viewer) {
    throw new Error(`Unknown feed viewer state for short: ${short.id}`);
  }

  return {
    ...surface,
    viewer,
  };
}

/**
 * short detail 用の surface mock を返す。
 */
export function getShortSurfaceById(shortId: string): DetailShortSurface | undefined {
  const surface = buildShortSurfaceBase(shortId);

  if (!surface) {
    return undefined;
  }

  const viewer = detailViewerStateByShortId[surface.short.id];

  if (!viewer) {
    throw new Error(`Unknown detail viewer state for short: ${surface.short.id}`);
  }

  return {
    ...surface,
    viewer,
  };
}
