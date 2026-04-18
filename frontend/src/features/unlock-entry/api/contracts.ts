import { creatorSummarySchema } from "@/entities/creator";
import {
  publicShortSummarySchema,
  unlockCtaStateSchema,
} from "@/entities/short";
import { z } from "zod";

export const supportedCardBrandSchema = z.enum([
  "visa",
  "mastercard",
  "jcb",
  "american_express",
]);

export const mainAccessStateSchema = z.object({
  mainId: z.string().min(1),
  reason: z.enum(["owner_preview", "purchased", "session_unlocked", "unlock_required"]),
  status: z.enum(["locked", "owner", "unlocked"]),
});

export const entryContextSchema = z.object({
  accessEntryPath: z.string().min(1),
  purchasePath: z.string().min(1),
  token: z.string().min(1),
});

export const purchaseSetupStateSchema = z.object({
  required: z.boolean(),
  requiresAgeConfirmation: z.boolean(),
  requiresCardSetup: z.boolean(),
  requiresTermsAcceptance: z.boolean(),
});

export const savedPaymentMethodSummarySchema = z.object({
  brand: supportedCardBrandSchema,
  last4: z.string().min(1),
  paymentMethodId: z.string().min(1),
});

export const unlockPurchaseStateSchema = z.object({
  pendingReason: z.enum(["provider_processing"]).nullable(),
  savedPaymentMethods: z.array(savedPaymentMethodSummarySchema),
  setup: purchaseSetupStateSchema,
  state: z.enum([
    "setup_required",
    "purchase_ready",
    "purchase_pending",
    "already_purchased",
    "owner_preview",
    "unavailable",
  ]),
  supportedCardBrands: z.array(supportedCardBrandSchema),
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
  entryContext: entryContextSchema,
  main: unlockMainSummarySchema,
  purchase: unlockPurchaseStateSchema,
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

export const mainPurchaseResponseSchema = z.object({
  data: z.object({
    access: mainAccessStateSchema,
    entryContext: entryContextSchema.nullable(),
    purchase: z.object({
      canRetry: z.boolean(),
      failureReason: z.enum(["authentication_failed", "card_brand_unsupported", "purchase_declined"]).nullable(),
      status: z.enum(["succeeded", "pending", "failed", "already_purchased", "owner_preview"]),
    }),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const cardSetupSessionResponseSchema = z.object({
  data: z.object({
    apiBaseUrl: z.string().url(),
    apiKey: z.string().min(1),
    clientAccount: z.string().min(1),
    currency: z.enum(["JPY"]),
    initialPeriod: z.string().min(1),
    initialPrice: z.string().min(1),
    sessionToken: z.string().min(1),
    subAccount: z.string().min(1),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export const cardSetupTokenResponseSchema = z.object({
  data: z.object({
    cardSetupToken: z.string().min(1),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});
