import { z } from "zod";

const creatorHandleSchema = z.custom<`@${string}`>((value) => typeof value === "string" && value.startsWith("@"));

export const creatorAvatarAssetSchema = z.object({
  durationSeconds: z.null(),
  id: z.string().min(1),
  kind: z.literal("image"),
  posterUrl: z.string().min(1).nullable(),
  url: z.string().min(1),
});

export const creatorSummarySchema = z.object({
  avatar: creatorAvatarAssetSchema.nullable(),
  bio: z.string(),
  displayName: z.string().min(1),
  handle: creatorHandleSchema,
  id: z.string().min(1),
});

export const creatorProfileStatsSchema = z.object({
  fanCount: z.number().int().nonnegative(),
  shortCount: z.number().int().nonnegative(),
});

export const creatorProfileViewerSchema = z.object({
  isFollowing: z.boolean(),
});

export const creatorProfileHeaderSchema = z.object({
  creator: creatorSummarySchema,
  stats: creatorProfileStatsSchema,
  viewer: creatorProfileViewerSchema,
});

export const creatorProfileShortGridMediaSchema = z.object({
  durationSeconds: z.number().int().positive().nullable(),
  id: z.string().min(1),
  kind: z.literal("video"),
  posterUrl: z.string().min(1).nullable(),
  url: z.string().min(1),
});

export const creatorProfileShortGridItemSchema = z.object({
  canonicalMainId: z.string().min(1),
  creatorId: z.string().min(1),
  id: z.string().min(1),
  media: creatorProfileShortGridMediaSchema,
  previewDurationSeconds: z.number().int().positive(),
});
