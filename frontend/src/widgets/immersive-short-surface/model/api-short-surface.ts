import type { CreatorSummary } from "@/entities/creator";
import type { FanFeedItem, PublicShortDetail, ShortPreviewMeta } from "@/entities/short";
import { getMockMainAccessRoutePath, type MainAccessState, type UnlockCtaState, type UnlockSurfaceModel } from "@/features/unlock-entry";

import type { DetailShortSurface, FeedShortSurface } from "./short-surface";

function buildShortPreviewMeta(item: FanFeedItem["short"]): ShortPreviewMeta {
  return {
    caption: item.caption,
    canonicalMainId: item.canonicalMainId,
    creatorId: item.creatorId,
    id: item.id,
    media: item.media,
    previewDurationSeconds: item.previewDurationSeconds,
    title: item.caption,
  };
}

function buildMainAccessState(mainId: string, unlockCta: UnlockCtaState): MainAccessState {
  switch (unlockCta.state) {
    case "owner_preview":
      return {
        mainId,
        reason: "owner_preview",
        status: "owner",
      };
    case "continue_main":
      return {
        mainId,
        reason: "session_unlocked",
        status: "unlocked",
      };
    default:
      return {
        mainId,
        reason: "unlock_required",
        status: "locked",
      };
  }
}

function buildUnlockSurfaceModel({
  creator,
  short,
  unlockCta,
}: {
  creator: CreatorSummary;
  short: ShortPreviewMeta;
  unlockCta: UnlockCtaState;
}): UnlockSurfaceModel {
  return {
    access: buildMainAccessState(short.canonicalMainId, unlockCta),
    creator,
    main: {
      durationSeconds: unlockCta.mainDurationSeconds ?? short.previewDurationSeconds,
      id: short.canonicalMainId,
      priceJpy: unlockCta.priceJpy ?? 0,
      title: short.caption,
    },
    mainAccessEntry: {
      routePath: getMockMainAccessRoutePath(short.canonicalMainId),
      token: `disabled-${short.id}`,
    },
    setup: {
      required: unlockCta.state === "setup_required",
      requiresAgeConfirmation: unlockCta.state === "setup_required",
      requiresTermsAcceptance: unlockCta.state === "setup_required",
    },
    short,
    unlockCta,
  };
}

function buildShortSurfaceBase(item: Pick<FanFeedItem, "creator" | "short" | "unlockCta">) {
  const short = buildShortPreviewMeta(item.short);

  return {
    creator: item.creator,
    mainEntryEnabled: false,
    short,
    unlock: buildUnlockSurfaceModel({
      creator: item.creator,
      short,
      unlockCta: item.unlockCta,
    }),
  };
}

/**
 * feed API item を immersive feed surface へ変換する。
 */
export function buildFeedSurfaceFromApiItem(item: FanFeedItem): FeedShortSurface {
  return {
    ...buildShortSurfaceBase(item),
    viewer: {
      isFollowingCreator: item.viewer.isFollowingCreator,
      isPinned: item.viewer.isPinned,
    },
  };
}

/**
 * short detail API payload を immersive detail surface へ変換する。
 */
export function buildDetailSurfaceFromApi(detail: PublicShortDetail): DetailShortSurface {
  return {
    ...buildShortSurfaceBase(detail),
    viewer: {
      isFollowingCreator: detail.viewer.isFollowingCreator,
      isPinned: detail.viewer.isPinned,
    },
  };
}
