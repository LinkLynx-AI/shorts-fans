import { z } from "zod";

import { getCreatorById } from "@/entities/creator";
import { getMainById } from "@/entities/main";
import { getShortById } from "@/entities/short";

import type { UnlockSurfaceModel } from "../model/unlock-entry";

const purchaseSchema = z.object({
  mainId: z.string().min(1),
  status: z.enum(["not_purchased", "purchased"]),
});

const accessSchema = z.object({
  mainId: z.string().min(1),
  reason: z.enum(["owner_preview", "purchase_required", "purchased_access"]),
  status: z.enum(["locked", "owner", "purchased"]),
});

const unlockCtaSchema = z.object({
  mainDurationSeconds: z.number().int().positive().nullable(),
  priceJpy: z.number().int().positive().nullable(),
  resumePositionSeconds: z.number().int().nonnegative().nullable(),
  state: z.enum(["continue_main", "owner_preview", "setup_required", "unavailable", "unlock_available"]),
});

const setupSchema = z.object({
  required: z.boolean(),
  requiresAgeConfirmation: z.boolean(),
  requiresTermsAcceptance: z.boolean(),
});

const mediaAssetSchema = z.object({
  durationSeconds: z.number().int().nonnegative().nullable(),
  id: z.string().min(1),
  kind: z.literal("video"),
  posterUrl: z.string().nullable(),
  url: z.string().url(),
});

const shortSchema = z.object({
  canonicalMainId: z.string().min(1),
  caption: z.string().min(1),
  creatorId: z.string().min(1),
  id: z.string().min(1),
  media: mediaAssetSchema,
  previewDurationSeconds: z.number().int().nonnegative(),
  title: z.string().min(1),
});

const creatorAssetSchema = z.object({
  durationSeconds: z.null(),
  id: z.string().min(1),
  kind: z.literal("image"),
  posterUrl: z.null(),
  url: z.string().min(1),
});

const creatorSchema = z.object({
  avatar: creatorAssetSchema,
  bio: z.string().min(1),
  displayName: z.string().min(1),
  handle: z.custom<`@${string}`>((value) => typeof value === "string" && value.startsWith("@")),
  id: z.string().min(1),
});

const mainSummarySchema = z.object({
  durationSeconds: z.number().int().positive(),
  id: z.string().min(1),
  priceJpy: z.number().int().positive(),
  title: z.string().min(1),
});

const unlockSurfaceSchema = z.object({
  access: accessSchema,
  creator: creatorSchema,
  main: mainSummarySchema,
  purchase: purchaseSchema,
  setup: setupSchema,
  short: shortSchema,
  unlockCta: unlockCtaSchema,
});

type RawUnlockState = {
  access: UnlockSurfaceModel["access"];
  purchase: UnlockSurfaceModel["purchase"];
  setup: UnlockSurfaceModel["setup"];
  unlockCta: UnlockSurfaceModel["unlockCta"];
};

const rawUnlockStateByShortId: Readonly<Record<string, RawUnlockState>> = {
  afterrain: {
    access: {
      mainId: "main_sora_after_rain",
      reason: "purchase_required",
      status: "locked",
    },
    purchase: {
      mainId: "main_sora_after_rain",
      status: "not_purchased",
    },
    setup: {
      required: false,
      requiresAgeConfirmation: false,
      requiresTermsAcceptance: false,
    },
    unlockCta: {
      mainDurationSeconds: 540,
      priceJpy: 2100,
      resumePositionSeconds: null,
      state: "unlock_available",
    },
  },
  balcony: {
    access: {
      mainId: "main_aoi_blue_balcony",
      reason: "owner_preview",
      status: "owner",
    },
    purchase: {
      mainId: "main_aoi_blue_balcony",
      status: "not_purchased",
    },
    setup: {
      required: false,
      requiresAgeConfirmation: false,
      requiresTermsAcceptance: false,
    },
    unlockCta: {
      mainDurationSeconds: null,
      priceJpy: null,
      resumePositionSeconds: null,
      state: "owner_preview",
    },
  },
  mirror: {
    access: {
      mainId: "main_mina_hotel_mirror",
      reason: "purchase_required",
      status: "locked",
    },
    purchase: {
      mainId: "main_mina_hotel_mirror",
      status: "not_purchased",
    },
    setup: {
      required: true,
      requiresAgeConfirmation: true,
      requiresTermsAcceptance: true,
    },
    unlockCta: {
      mainDurationSeconds: 660,
      priceJpy: 2400,
      resumePositionSeconds: null,
      state: "setup_required",
    },
  },
  poolcut: {
    access: {
      mainId: "main_sora_poolside_cut",
      reason: "purchase_required",
      status: "locked",
    },
    purchase: {
      mainId: "main_sora_poolside_cut",
      status: "not_purchased",
    },
    setup: {
      required: false,
      requiresAgeConfirmation: false,
      requiresTermsAcceptance: false,
    },
    unlockCta: {
      mainDurationSeconds: 480,
      priceJpy: 1900,
      resumePositionSeconds: null,
      state: "unlock_available",
    },
  },
  rooftop: {
    access: {
      mainId: "main_mina_quiet_rooftop",
      reason: "purchase_required",
      status: "locked",
    },
    purchase: {
      mainId: "main_mina_quiet_rooftop",
      status: "not_purchased",
    },
    setup: {
      required: true,
      requiresAgeConfirmation: true,
      requiresTermsAcceptance: true,
    },
    unlockCta: {
      mainDurationSeconds: 480,
      priceJpy: 1800,
      resumePositionSeconds: null,
      state: "setup_required",
    },
  },
  softlight: {
    access: {
      mainId: "main_aoi_blue_balcony",
      reason: "purchased_access",
      status: "purchased",
    },
    purchase: {
      mainId: "main_aoi_blue_balcony",
      status: "purchased",
    },
    setup: {
      required: false,
      requiresAgeConfirmation: false,
      requiresTermsAcceptance: false,
    },
    unlockCta: {
      mainDurationSeconds: null,
      priceJpy: null,
      resumePositionSeconds: 198,
      state: "continue_main",
    },
  },
};

/**
 * short ごとの unlock surface 用 mock を取得する。
 */
export function getUnlockSurfaceByShortId(shortId: string): UnlockSurfaceModel | undefined {
  const short = getShortById(shortId);

  if (!short) {
    return undefined;
  }

  const creator = getCreatorById(short.creatorId);

  if (!creator) {
    throw new Error(`Unknown creator for unlock surface: ${short.creatorId}`);
  }

  const main = getMainById(short.canonicalMainId);

  if (!main) {
    throw new Error(`Unknown main for unlock surface: ${short.canonicalMainId}`);
  }

  const rawState = rawUnlockStateByShortId[short.id];

  if (!rawState) {
    throw new Error(`Unknown unlock state for short: ${short.id}`);
  }

  return unlockSurfaceSchema.parse({
    access: rawState.access,
    creator,
    main: {
      durationSeconds: main.durationSeconds,
      id: main.id,
      priceJpy: main.priceJpy,
      title: main.title,
    },
    purchase: rawState.purchase,
    setup: rawState.setup,
    short,
    unlockCta: rawState.unlockCta,
  });
}
