import { z } from "zod";

const submissionReviewHandlePattern = /^@[a-z0-9._]+$/;
const submissionReviewUUIDPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const submissionReviewUUIDSchema = z.string().regex(submissionReviewUUIDPattern);

export const submissionReviewDecisions = [
  "approved",
  "revision_requested",
  "rejected",
] as const;

export const submissionReviewDecisionTargetStates = [
  "approved_for_unlock",
  "approved_for_publish",
  "revision_requested",
  "rejected",
] as const;

export const submissionReviewIntakeStatuses = [
  "pending_review",
  "decision_applied",
] as const;

export const submissionReviewObjectStates = [
  "draft",
  "pending_review",
  "approved_for_unlock",
  "approved_for_publish",
  "revision_requested",
  "rejected",
] as const;

export const submissionReviewSubmitKinds = [
  "initial_submit",
  "resubmit",
] as const;

export const submissionReviewAvatarAssetSchema = z.object({
  durationSeconds: z.null(),
  id: z.string().min(1),
  kind: z.literal("image"),
  posterUrl: z.string().min(1).nullable(),
  url: z.string().min(1),
});

export const submissionReviewCreatorSchema = z.object({
  avatar: submissionReviewAvatarAssetSchema.nullable(),
  bio: z.string(),
  displayName: z.string().min(1),
  handle: z.custom<`@${string}`>(
    (value) => typeof value === "string" && submissionReviewHandlePattern.test(value),
  ),
  id: z.string().min(1),
});

export const submissionReviewMediaSchema = z.object({
  durationSeconds: z.number().int().positive(),
  id: z.string().min(1),
  kind: z.literal("video"),
  posterUrl: z.string().min(1),
  url: z.string().min(1),
});

export const submissionReviewProvenanceSchema = z.object({
  decisionSource: z.string().min(1).nullable(),
  decisionedAt: z.string().datetime().nullable(),
  reasonCode: z.string().min(1).nullable(),
  reviewNote: z.string().min(1).nullable(),
});

export const submissionReviewDecisionLogSchema = z.object({
  decisionSource: z.string().min(1),
  decisionedAt: z.string().datetime(),
  reasonCode: z.string().min(1).nullable(),
  reviewNote: z.string().min(1).nullable(),
  targetState: z.enum(submissionReviewDecisionTargetStates),
});

export const submissionReviewQueueItemSchema = z.object({
  creator: submissionReviewCreatorSchema,
  intakeId: submissionReviewUUIDSchema,
  mainDecisionRequired: z.boolean(),
  pendingShortCount: z.number().int().nonnegative(),
  shortCount: z.number().int().nonnegative(),
  submitKind: z.enum(submissionReviewSubmitKinds),
  submittedAt: z.string().datetime(),
});

export const submissionReviewIntakeSchema = z.object({
  canonicalMainId: submissionReviewUUIDSchema,
  consentConfirmed: z.boolean(),
  creatorUserId: submissionReviewUUIDSchema,
  id: submissionReviewUUIDSchema,
  mainMediaAssetId: submissionReviewUUIDSchema,
  mainPriceJpy: z.number().int().nonnegative(),
  ownershipConfirmed: z.boolean(),
  previousIntakeId: submissionReviewUUIDSchema.nullable(),
  status: z.enum(submissionReviewIntakeStatuses),
  submitKind: z.enum(submissionReviewSubmitKinds),
  submittedAt: z.string().datetime(),
});

export const submissionReviewMainSchema = z.object({
  currencyCode: z.string().min(1),
  decisionRequired: z.boolean(),
  id: submissionReviewUUIDSchema,
  intakeDecisionLog: submissionReviewDecisionLogSchema.nullable(),
  media: submissionReviewMediaSchema,
  priceJpy: z.number().int().nonnegative(),
  review: submissionReviewProvenanceSchema,
  state: z.enum(submissionReviewObjectStates),
});

export const submissionReviewShortSchema = z.object({
  caption: z.string().nullable(),
  decisionRequired: z.boolean(),
  id: submissionReviewUUIDSchema,
  intakeDecisionLog: submissionReviewDecisionLogSchema.nullable(),
  media: submissionReviewMediaSchema,
  review: submissionReviewProvenanceSchema,
  state: z.enum(submissionReviewObjectStates),
});

export const submissionReviewCaseSchema = z.object({
  creator: submissionReviewCreatorSchema,
  intake: submissionReviewIntakeSchema,
  main: submissionReviewMainSchema,
  shorts: z.array(submissionReviewShortSchema),
});

export const submissionReviewQueueResponseSchema = z.object({
  data: z.object({
    items: z.array(submissionReviewQueueItemSchema),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const submissionReviewCaseResponseSchema = z.object({
  data: z.object({
    case: submissionReviewCaseSchema,
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export function isSubmissionReviewIntakeId(value: string): boolean {
  return submissionReviewUUIDSchema.safeParse(value).success;
}

export type SubmissionReviewCase = z.output<typeof submissionReviewCaseSchema>;
export type SubmissionReviewDecision = (typeof submissionReviewDecisions)[number];
export type SubmissionReviewDecisionLog = z.output<typeof submissionReviewDecisionLogSchema>;
export type SubmissionReviewDecisionTargetState = (typeof submissionReviewDecisionTargetStates)[number];
export type SubmissionReviewIntakeStatus = (typeof submissionReviewIntakeStatuses)[number];
export type SubmissionReviewMain = z.output<typeof submissionReviewMainSchema>;
export type SubmissionReviewObjectState = (typeof submissionReviewObjectStates)[number];
export type SubmissionReviewQueueItem = z.output<typeof submissionReviewQueueItemSchema>;
export type SubmissionReviewShort = z.output<typeof submissionReviewShortSchema>;
export type SubmissionReviewSubmitKind = (typeof submissionReviewSubmitKinds)[number];
