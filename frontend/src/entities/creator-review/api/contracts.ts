import { z } from "zod";

const creatorReviewHandlePattern = /^@[a-z0-9._]+$/;
const creatorReviewUserIDPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export const creatorReviewStates = [
  "submitted",
  "approved",
  "rejected",
  "suspended",
] as const;

export const creatorReviewDecisions = [
  "approved",
  "rejected",
  "suspended",
] as const;

export const creatorReviewPayoutRecipientTypes = [
  "self",
  "business",
] as const;

export const creatorReviewAvatarAssetSchema = z.object({
  durationSeconds: z.null(),
  id: z.string().min(1),
  kind: z.literal("image"),
  posterUrl: z.string().min(1).nullable(),
  url: z.string().url(),
});

export const creatorReviewSharedProfileSchema = z.object({
  avatar: creatorReviewAvatarAssetSchema.nullable(),
  displayName: z.string().min(1),
  handle: z.custom<`@${string}`>(
    (value) => typeof value === "string" && creatorReviewHandlePattern.test(value),
  ),
});

export const creatorReviewTimelineSchema = z.object({
  approvedAt: z.string().datetime().nullable(),
  rejectedAt: z.string().datetime().nullable(),
  submittedAt: z.string().datetime().nullable(),
  suspendedAt: z.string().datetime().nullable(),
});

export const creatorReviewRejectionSchema = z.object({
  isResubmitEligible: z.boolean(),
  isSupportReviewRequired: z.boolean(),
  reasonCode: z.string().min(1).nullable(),
  selfServeResubmitCount: z.number().int().min(0),
  selfServeResubmitRemaining: z.number().int().min(0),
});

export const creatorReviewIntakeSchema = z.object({
  acceptsConsentResponsibility: z.boolean(),
  birthDate: z.string().min(1).nullable(),
  declaresNoProhibitedCategory: z.boolean(),
  legalName: z.string(),
  payoutRecipientName: z.string(),
  payoutRecipientType: z.enum(creatorReviewPayoutRecipientTypes).nullable(),
});

export const creatorReviewEvidenceSchema = z.object({
  accessUrl: z.string().url(),
  fileName: z.string().min(1),
  fileSizeBytes: z.number().int().positive(),
  kind: z.string().min(1),
  mimeType: z.string().min(1),
  uploadedAt: z.string().datetime(),
});

export const creatorReviewQueueItemSchema = z.object({
  creatorBio: z.string(),
  legalName: z.string(),
  review: creatorReviewTimelineSchema,
  sharedProfile: creatorReviewSharedProfileSchema,
  state: z.enum(creatorReviewStates),
  userId: z.string().regex(creatorReviewUserIDPattern),
});

export const creatorReviewCaseSchema = z.object({
  creatorBio: z.string(),
  evidences: z.array(creatorReviewEvidenceSchema),
  intake: creatorReviewIntakeSchema,
  rejection: creatorReviewRejectionSchema.nullable(),
  review: creatorReviewTimelineSchema,
  sharedProfile: creatorReviewSharedProfileSchema,
  state: z.enum(creatorReviewStates),
  userId: z.string().regex(creatorReviewUserIDPattern),
});

export const creatorReviewQueueResponseSchema = z.object({
  data: z.object({
    items: z.array(creatorReviewQueueItemSchema),
    state: z.enum(creatorReviewStates),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const creatorReviewCaseResponseSchema = z.object({
  data: z.object({
    case: creatorReviewCaseSchema,
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorReviewCase = z.output<typeof creatorReviewCaseSchema>;
export type CreatorReviewDecision = (typeof creatorReviewDecisions)[number];
export type CreatorReviewQueueItem = z.output<typeof creatorReviewQueueItemSchema>;
export type CreatorReviewQueueState = (typeof creatorReviewStates)[number];
