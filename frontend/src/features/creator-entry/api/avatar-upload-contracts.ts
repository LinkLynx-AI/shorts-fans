import { z } from "zod";

const creatorRegistrationAvatarDirectUploadSchema = z.object({
  headers: z.record(z.string(), z.string()),
  method: z.literal("PUT"),
  url: z.string().url(),
});

export const creatorRegistrationAvatarUploadTargetSchema = z.object({
  fileName: z.string().min(1),
  mimeType: z.string().min(1),
  upload: creatorRegistrationAvatarDirectUploadSchema,
});

export const creatorRegistrationAvatarCreateResponseSchema = z.object({
  data: z.object({
    avatarUploadToken: z.string().min(1),
    expiresAt: z.string().datetime(),
    uploadTarget: creatorRegistrationAvatarUploadTargetSchema,
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const creatorRegistrationAvatarCompleteResponseSchema = z.object({
  data: z.object({
    avatar: z.object({
      durationSeconds: z.null(),
      id: z.string().min(1),
      kind: z.literal("image"),
      posterUrl: z.null(),
      url: z.string().url(),
    }),
    avatarUploadToken: z.string().min(1),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorRegistrationAvatarCreateResponse = z.output<typeof creatorRegistrationAvatarCreateResponseSchema>;
export type CreatorRegistrationAvatarCompleteResponse = z.output<typeof creatorRegistrationAvatarCompleteResponseSchema>;
export type CreatorRegistrationAvatarUploadTarget = z.output<typeof creatorRegistrationAvatarUploadTargetSchema>;
