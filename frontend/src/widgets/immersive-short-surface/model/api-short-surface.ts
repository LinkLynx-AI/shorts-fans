import type { CreatorSummary } from "@/entities/creator";
import type { FanFeedItem, PublicShortDetail } from "@/entities/short";
import {
  getMockMainAccessRoutePath,
  normalizeUnlockSurface,
  type MainAccessState,
  type RawUnlockSurfaceModel,
  type UnlockCtaState,
  type UnlockSurfaceModel,
} from "@/features/unlock-entry";

import type { DetailShortSurface, FeedShortSurface } from "./short-surface";

function buildShortPreviewMeta(item: FanFeedItem["short"]): UnlockSurfaceModel["short"] {
  return {
    caption: item.caption,
    canonicalMainId: item.canonicalMainId,
    creatorId: item.creatorId,
    id: item.id,
    media: item.media,
    previewDurationSeconds: item.previewDurationSeconds,
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
        reason: "purchased",
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
  short: UnlockSurfaceModel["short"];
  unlockCta: UnlockCtaState;
}): UnlockSurfaceModel {
  const access = buildMainAccessState(short.canonicalMainId, unlockCta);
  const purchaseState =
    unlockCta.state === "owner_preview"
      ? "owner_preview"
      : unlockCta.state === "continue_main"
        ? "already_purchased"
        : unlockCta.state === "setup_required"
          ? "setup_required"
          : "purchase_ready";

  return normalizeUnlockSurface({
    access,
    creator,
    entryContext: {
      accessEntryPath: getMockMainAccessRoutePath(short.canonicalMainId),
      purchasePath: `/api/fan/mains/${short.canonicalMainId}/purchase`,
      token: `disabled-${short.id}`,
    },
    main: {
      durationSeconds: unlockCta.mainDurationSeconds ?? short.previewDurationSeconds,
      id: short.canonicalMainId,
      priceJpy: unlockCta.priceJpy ?? 0,
    },
    purchase: {
      pendingReason: null,
      savedPaymentMethods: [],
      setup: {
        required: unlockCta.state === "setup_required",
        requiresAgeConfirmation: unlockCta.state === "setup_required",
        requiresCardSetup: unlockCta.state === "setup_required",
        requiresTermsAcceptance: unlockCta.state === "setup_required",
      },
      state: purchaseState,
      supportedCardBrands: ["visa", "mastercard", "jcb", "american_express"],
    },
    short,
    unlockCta,
  } satisfies RawUnlockSurfaceModel);
}

function buildShortSurfaceBase(item: Pick<FanFeedItem, "creator" | "short" | "unlockCta">) {
  const short = buildShortPreviewMeta(item.short);

  return {
    creator: item.creator,
    mainEntryEnabled: true,
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
    mainEntryEnabled: false,
    viewer: {
      isFollowingCreator: detail.viewer.isFollowingCreator,
      isPinned: detail.viewer.isPinned,
    },
  };
}
