import { z } from "zod";

import { creatorSummarySchema } from "@/entities/creator";

export const fanFeedTabSchema = z.enum(["following", "recommended"]);

export const shortVideoDisplayAssetSchema = z.object({
  durationSeconds: z.number().int().positive().nullable(),
  id: z.string().min(1),
  kind: z.literal("video"),
  posterUrl: z.string().min(1).nullable(),
  url: z.string().min(1),
});

export const publicShortSummarySchema = z.object({
  caption: z.string(),
  canonicalMainId: z.string().min(1),
  creatorId: z.string().min(1),
  id: z.string().min(1),
  media: shortVideoDisplayAssetSchema,
  previewDurationSeconds: z.number().int().positive(),
});

export const unlockCtaStateSchema = z.object({
  mainDurationSeconds: z.number().int().positive().nullable(),
  priceJpy: z.number().int().positive().nullable(),
  resumePositionSeconds: z.number().int().nonnegative().nullable(),
  state: z.enum(["continue_main", "owner_preview", "setup_required", "unavailable", "unlock_available"]),
});

export const fanFeedItemSchema = z.object({
  creator: creatorSummarySchema,
  short: publicShortSummarySchema,
  unlockCta: unlockCtaStateSchema,
  viewer: z.object({
    isFollowingCreator: z.boolean(),
    isPinned: z.boolean(),
  }),
});

export const fanFeedResponseSchema = z.object({
  data: z.object({
    items: z.array(fanFeedItemSchema),
    tab: fanFeedTabSchema,
  }),
  error: z.null(),
  meta: z.object({
    page: z.object({
      hasNext: z.boolean(),
      nextCursor: z.string().min(1).nullable(),
    }),
    requestId: z.string().min(1),
  }),
});

export const publicShortDetailSchema = z.object({
  creator: creatorSummarySchema,
  short: publicShortSummarySchema,
  unlockCta: unlockCtaStateSchema,
  viewer: z.object({
    isFollowingCreator: z.boolean(),
    isPinned: z.boolean(),
  }),
});

export const publicShortDetailResponseSchema = z.object({
  data: z.object({
    detail: publicShortDetailSchema,
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type FanFeedItem = z.output<typeof fanFeedItemSchema>;
export type FanFeedTab = z.output<typeof fanFeedTabSchema>;
export type PublicShortDetail = z.output<typeof publicShortDetailSchema>;
