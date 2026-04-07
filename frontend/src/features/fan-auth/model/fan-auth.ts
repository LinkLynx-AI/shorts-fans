import { z } from "zod";

export const fanLoginPath = "/login" as const;

const authRequiredResponseSchema = z.object({
  error: z.object({
    code: z.literal("auth_required"),
    message: z.string().min(1),
  }),
});

export type AuthRequiredResponse = z.infer<typeof authRequiredResponseSchema>;

/**
 * fan login entry の route path を返す。
 */
export function buildFanLoginHref(): string {
  return fanLoginPath;
}

/**
 * payload が auth_required 応答かを判定する。
 */
export function isAuthRequiredResponse(value: unknown): value is AuthRequiredResponse {
  return authRequiredResponseSchema.safeParse(value).success;
}
