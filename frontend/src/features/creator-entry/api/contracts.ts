import { z } from "zod";

const creatorRegistrationHandlePattern = /^@[a-z0-9._]+$/;

const creatorRegistrationHandleSchema = z.custom<`@${string}`>(
  (value) => typeof value === "string" && creatorRegistrationHandlePattern.test(value),
);

export const creatorRegistrationEvidenceKinds = [
  "government_id",
  "payout_proof",
] as const;

export const creatorRegistrationPayoutRecipientTypes = [
  "self",
  "business",
] as const;

export const creatorRegistrationAvatarAssetSchema = z.object({
  durationSeconds: z.null(),
  id: z.string().min(1),
  kind: z.literal("image"),
  posterUrl: z.string().min(1).nullable(),
  url: z.string().url(),
});

export const creatorRegistrationSharedProfileSchema = z.object({
  avatar: creatorRegistrationAvatarAssetSchema.nullable(),
  displayName: z.string().min(1),
  handle: creatorRegistrationHandleSchema,
});

export const creatorRegistrationEvidenceSchema = z.object({
  fileName: z.string().min(1),
  fileSizeBytes: z.number().int().positive(),
  kind: z.enum(creatorRegistrationEvidenceKinds),
  mimeType: z.string().min(1),
  uploadedAt: z.string().datetime(),
});

export const creatorRegistrationIntakeSchema = z.object({
  acceptsConsentResponsibility: z.boolean(),
  birthDate: z.string().min(1).nullable(),
  canSubmit: z.boolean(),
  creatorBio: z.string(),
  declaresNoProhibitedCategory: z.boolean(),
  evidences: z.array(creatorRegistrationEvidenceSchema),
  isReadOnly: z.boolean(),
  legalName: z.string(),
  payoutRecipientName: z.string(),
  payoutRecipientType: z.enum(creatorRegistrationPayoutRecipientTypes).nullable(),
  registrationState: z.string().min(1).nullable(),
  sharedProfile: creatorRegistrationSharedProfileSchema,
});

const creatorRegistrationActionsSchema = z.object({
  canEnterCreatorMode: z.boolean(),
  canResubmit: z.boolean(),
  canSubmit: z.boolean(),
});

const creatorRegistrationReviewSchema = z.object({
  approvedAt: z.string().datetime().nullable(),
  rejectedAt: z.string().datetime().nullable(),
  submittedAt: z.string().datetime().nullable(),
  suspendedAt: z.string().datetime().nullable(),
});

const creatorRegistrationSurfaceSchema = z.object({
  kind: z.string().min(1),
  workspacePreview: z.string().min(1).nullable(),
});

const creatorRegistrationRejectionSchema = z.object({
  isResubmitEligible: z.boolean(),
  isSupportReviewRequired: z.boolean(),
  reasonCode: z.string().min(1).nullable(),
  selfServeResubmitCount: z.number().int().min(0),
  selfServeResubmitRemaining: z.number().int().min(0),
});

export const creatorRegistrationStatusSchema = z.object({
  actions: creatorRegistrationActionsSchema,
  creatorDraft: z.object({
    bio: z.string(),
  }),
  rejection: creatorRegistrationRejectionSchema.nullable(),
  review: creatorRegistrationReviewSchema,
  sharedProfile: creatorRegistrationSharedProfileSchema,
  state: z.string().min(1),
  surface: creatorRegistrationSurfaceSchema,
});

const creatorRegistrationDirectUploadSchema = z.object({
  headers: z.record(z.string(), z.string()),
  method: z.literal("PUT"),
  url: z.string().url(),
});

export const creatorRegistrationEvidenceUploadTargetSchema = z.object({
  fileName: z.string().min(1),
  mimeType: z.string().min(1),
  upload: creatorRegistrationDirectUploadSchema,
});

export const creatorRegistrationIntakeResponseSchema = z.object({
  data: z.object({
    intake: creatorRegistrationIntakeSchema,
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const creatorRegistrationStatusResponseSchema = z.object({
  data: z.object({
    registration: creatorRegistrationStatusSchema.nullable(),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const creatorRegistrationEvidenceUploadCreateResponseSchema = z.object({
  data: z.object({
    evidenceKind: z.enum(creatorRegistrationEvidenceKinds),
    evidenceUploadToken: z.string().min(1),
    expiresAt: z.string().datetime(),
    uploadTarget: creatorRegistrationEvidenceUploadTargetSchema,
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const creatorRegistrationEvidenceUploadCompleteResponseSchema = z.object({
  data: z.object({
    evidence: creatorRegistrationEvidenceSchema,
    evidenceKind: z.enum(creatorRegistrationEvidenceKinds),
    evidenceUploadToken: z.string().min(1),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorRegistrationEvidence = z.output<typeof creatorRegistrationEvidenceSchema>;
export type CreatorRegistrationEvidenceKind = (typeof creatorRegistrationEvidenceKinds)[number];
export type CreatorRegistrationEvidenceUploadTarget = z.output<typeof creatorRegistrationEvidenceUploadTargetSchema>;
export type CreatorRegistrationIntake = z.output<typeof creatorRegistrationIntakeSchema>;
export type CreatorRegistrationPayoutRecipientType = (typeof creatorRegistrationPayoutRecipientTypes)[number];
export type CreatorRegistrationStatus = z.output<typeof creatorRegistrationStatusSchema>;
