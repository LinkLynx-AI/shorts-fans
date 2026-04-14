import { z } from "zod";

const viewerProfileHandlePattern = /^@[a-z0-9._]+$/;

const viewerProfileHandleSchema = z.custom<`@${string}`>(
  (value) => typeof value === "string" && viewerProfileHandlePattern.test(value),
);

export const viewerProfileAvatarAssetSchema = z.object({
  durationSeconds: z.null(),
  id: z.string().min(1),
  kind: z.literal("image"),
  posterUrl: z.string().min(1).nullable(),
  url: z.string().url(),
});

export const viewerProfileSchema = z.object({
  avatar: viewerProfileAvatarAssetSchema.nullable(),
  displayName: z.string().min(1),
  handle: viewerProfileHandleSchema,
});

const viewerProfileAvatarDirectUploadSchema = z.object({
  headers: z.record(z.string(), z.string()),
  method: z.literal("PUT"),
  url: z.string().url(),
});

export const viewerProfileAvatarUploadTargetSchema = z.object({
  fileName: z.string().min(1),
  mimeType: z.string().min(1),
  upload: viewerProfileAvatarDirectUploadSchema,
});

export const viewerProfileAvatarCreateResponseSchema = z.object({
  data: z.object({
    avatarUploadToken: z.string().min(1),
    expiresAt: z.string().datetime(),
    uploadTarget: viewerProfileAvatarUploadTargetSchema,
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const viewerProfileAvatarCompleteResponseSchema = z.object({
  data: z.object({
    avatar: viewerProfileAvatarAssetSchema,
    avatarUploadToken: z.string().min(1),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const viewerProfileResponseSchema = z.object({
  data: z.object({
    profile: viewerProfileSchema,
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type ViewerProfile = z.output<typeof viewerProfileSchema>;
export type ViewerProfileAvatarCreateResponse = z.output<typeof viewerProfileAvatarCreateResponseSchema>;
export type ViewerProfileAvatarCompleteResponse = z.output<typeof viewerProfileAvatarCompleteResponseSchema>;
export type ViewerProfileAvatarUploadTarget = z.output<typeof viewerProfileAvatarUploadTargetSchema>;
