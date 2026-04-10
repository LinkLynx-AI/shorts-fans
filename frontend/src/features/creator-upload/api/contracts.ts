import { z } from "zod";

export const creatorUploadErrorCodeSchema = z.enum([
  "auth_required",
  "capability_required",
  "internal_error",
  "invalid_request",
  "storage_failure",
  "upload_expired",
  "upload_failure",
  "validation_error",
]);

export const creatorUploadDirectUploadSchema = z.object({
  headers: z.record(z.string(), z.string()),
  method: z.literal("PUT"),
  url: z.string().url(),
});

const creatorUploadTargetBaseSchema = z.object({
  fileName: z.string().min(1),
  mimeType: z.string().min(1),
  upload: creatorUploadDirectUploadSchema,
  uploadEntryId: z.string().min(1),
});

export const creatorUploadMainTargetSchema = creatorUploadTargetBaseSchema.extend({
  role: z.literal("main"),
});

export const creatorUploadShortTargetSchema = creatorUploadTargetBaseSchema.extend({
  role: z.literal("short"),
});

export const creatorUploadCreateResponseSchema = z.object({
  data: z.object({
    expiresAt: z.string().datetime(),
    packageToken: z.string().min(1),
    uploadTargets: z.object({
      main: creatorUploadMainTargetSchema,
      shorts: z.array(creatorUploadShortTargetSchema).min(1),
    }),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const creatorUploadCreatedMediaAssetSchema = z.object({
  id: z.string().min(1),
  mimeType: z.string().min(1),
  processingState: z.literal("uploaded"),
});

export const creatorUploadCompleteResponseSchema = z.object({
  data: z.object({
    main: z.object({
      id: z.string().min(1),
      mediaAsset: creatorUploadCreatedMediaAssetSchema,
      state: z.literal("draft"),
    }),
    shorts: z.array(
      z.object({
        canonicalMainId: z.string().min(1),
        id: z.string().min(1),
        mediaAsset: creatorUploadCreatedMediaAssetSchema,
        state: z.literal("draft"),
      }),
    ).min(1),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const creatorUploadErrorResponseSchema = z.object({
  data: z.null(),
  error: z.object({
    code: creatorUploadErrorCodeSchema,
    message: z.string().min(1),
  }),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorUploadCreateResponse = z.output<typeof creatorUploadCreateResponseSchema>;
export type CreatorUploadCompleteResponse = z.output<typeof creatorUploadCompleteResponseSchema>;
export type CreatorUploadErrorCode = z.infer<typeof creatorUploadErrorCodeSchema>;
export type CreatorUploadErrorResponse = z.output<typeof creatorUploadErrorResponseSchema>;
export type CreatorUploadTarget = z.output<typeof creatorUploadMainTargetSchema> | z.output<typeof creatorUploadShortTargetSchema>;
