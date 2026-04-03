import { getCreatorById, type CreatorSummary } from "@/entities/creator";
import {
  getFeedShortForTab,
  getShortById,
  type FeedTab,
  type ShortId,
  type ShortPreviewMeta,
} from "@/entities/short";
import type { UnlockCtaState } from "@/features/unlock-entry";

export type FeedSurfaceViewerState = {
  isPinned: boolean;
};

export type DetailSurfaceViewerState = FeedSurfaceViewerState & {
  isFollowingCreator: boolean;
};

type ShortSurfaceBase = {
  creator: CreatorSummary;
  short: ShortPreviewMeta;
  unlockCta: UnlockCtaState;
};

export type FeedShortSurface = ShortSurfaceBase & {
  viewer: FeedSurfaceViewerState;
};

export type DetailShortSurface = ShortSurfaceBase & {
  viewer: DetailSurfaceViewerState;
};

const feedViewerStateByShortId: Record<string, FeedSurfaceViewerState> = {
  afterrain: { isPinned: true },
  balcony: { isPinned: true },
  mirror: { isPinned: false },
  poolcut: { isPinned: false },
  rooftop: { isPinned: true },
  softlight: { isPinned: false },
};

const detailViewerStateByShortId: Record<string, DetailSurfaceViewerState> = {
  afterrain: { isFollowingCreator: true, isPinned: true },
  balcony: { isFollowingCreator: true, isPinned: true },
  mirror: { isFollowingCreator: true, isPinned: false },
  poolcut: { isFollowingCreator: true, isPinned: false },
  rooftop: { isFollowingCreator: true, isPinned: true },
  softlight: { isFollowingCreator: true, isPinned: false },
};

const unlockCtaByShortId: Record<string, UnlockCtaState> = {
  afterrain: {
    mainDurationSeconds: 540,
    priceJpy: 2100,
    resumePositionSeconds: null,
    state: "unlock_available",
  },
  balcony: {
    mainDurationSeconds: 720,
    priceJpy: null,
    resumePositionSeconds: null,
    state: "owner_preview",
  },
  mirror: {
    mainDurationSeconds: 660,
    priceJpy: 2400,
    resumePositionSeconds: null,
    state: "setup_required",
  },
  poolcut: {
    mainDurationSeconds: 480,
    priceJpy: 1900,
    resumePositionSeconds: null,
    state: "unlock_available",
  },
  rooftop: {
    mainDurationSeconds: 480,
    priceJpy: 1800,
    resumePositionSeconds: null,
    state: "unlock_available",
  },
  softlight: {
    mainDurationSeconds: null,
    priceJpy: null,
    resumePositionSeconds: 198,
    state: "continue_main",
  },
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

  const unlockCta = unlockCtaByShortId[short.id];

  if (!unlockCta) {
    throw new Error(`Unknown unlock CTA state for short: ${short.id}`);
  }

  return {
    creator,
    short,
    unlockCta,
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
