import { z } from "zod";

import { creatorSummarySchema } from "@/entities/creator";
import { requestJson } from "@/shared/api";

const creatorWorkspaceSummaryResponseSchema = z.object({
  data: z.object({
    workspace: z.object({
      creator: creatorSummarySchema,
      overviewMetrics: z.object({
        grossUnlockRevenueJpy: z.number().int().nonnegative(),
        unlockCount: z.number().int().nonnegative(),
        uniquePurchaserCount: z.number().int().nonnegative(),
      }),
      revisionRequestedSummary: z.object({
        mainCount: z.number().int().nonnegative(),
        shortCount: z.number().int().nonnegative(),
        totalCount: z.number().int().nonnegative(),
      }).nullable(),
    }),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorWorkspaceSummaryResponse = z.output<typeof creatorWorkspaceSummaryResponseSchema>["data"]["workspace"];

type GetCreatorWorkspaceSummaryOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  signal?: AbortSignal;
};

/**
 * creator workspace summary を contract-backed API から取得する。
 */
export async function getCreatorWorkspaceSummary({
  baseUrl,
  credentials = "include",
  fetcher,
  signal,
}: GetCreatorWorkspaceSummaryOptions = {}): Promise<CreatorWorkspaceSummaryResponse> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials,
      ...(signal ? { signal } : {}),
    },
    path: "/api/creator/workspace",
    schema: creatorWorkspaceSummaryResponseSchema,
  });

  return response.data.workspace;
}
