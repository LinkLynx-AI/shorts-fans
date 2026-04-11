import { creatorSummarySchema } from "@/entities/creator";
import {
  publicShortSummarySchema,
  unlockCtaStateSchema,
} from "@/entities/short";
import { z } from "zod";

export const mainAccessStateSchema = z.object({
  mainId: z.string().min(1),
  reason: z.enum(["owner_preview", "session_unlocked", "unlock_required"]),
  status: z.enum(["locked", "owner", "unlocked"]),
});

export const unlockSetupStateSchema = z.object({
  required: z.boolean(),
  requiresAgeConfirmation: z.boolean(),
  requiresTermsAcceptance: z.boolean(),
});

export const unlockMainSummarySchema = z.object({
  durationSeconds: z.number().int().positive(),
  id: z.string().min(1),
  priceJpy: z.number().int().positive(),
});

export const unlockShortSummarySchema = publicShortSummarySchema;

export const unlockSurfaceSchema = z.object({
  access: mainAccessStateSchema,
  creator: creatorSummarySchema,
  mainAccessEntry: z.object({
    routePath: z.string().min(1),
    token: z.string().min(1),
  }),
  main: unlockMainSummarySchema,
  setup: unlockSetupStateSchema,
  short: unlockShortSummarySchema,
  unlockCta: unlockCtaStateSchema,
});

export const unlockSurfaceResponseSchema = z.object({
  data: unlockSurfaceSchema,
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const mainAccessEntryResponseSchema = z.object({
  data: z.object({
    href: z.string().min(1),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});
