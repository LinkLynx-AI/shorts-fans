import { z } from "zod";

export const creatorWorkspaceMainPriceErrorCodeSchema = z.enum([
  "auth_required",
  "creator_mode_unavailable",
  "internal_error",
  "invalid_request",
  "not_found",
  "validation_error",
]);

export const creatorWorkspaceMainPriceMutationResultSchema = z.object({
  main: z.object({
    id: z.string().min(1),
    priceJpy: z.number().int().positive(),
  }),
});
