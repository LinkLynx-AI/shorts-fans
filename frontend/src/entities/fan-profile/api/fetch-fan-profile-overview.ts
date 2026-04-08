import { z } from "zod";

import { requestJson } from "@/shared/api";
import { viewerSessionCookieName } from "@/entities/viewer";

import type { FanProfileOverview } from "../model/fan-profile";

const fanProfileOverviewResponseSchema = z.object({
  data: z.object({
    fanProfile: z.object({
      counts: z.object({
        following: z.number().int().nonnegative(),
        library: z.number().int().nonnegative(),
        pinnedShorts: z.number().int().nonnegative(),
      }),
      title: z.string().min(1),
    }),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

type FetchFanProfileOverviewOptions = {
  baseUrl?: string | undefined;
  fetcher?: typeof fetch | undefined;
  sessionToken?: string | undefined;
};

/**
 * fan profile overview API から counts-only payload を取得する。
 */
export async function fetchFanProfileOverview({
  baseUrl,
  fetcher,
  sessionToken,
}: FetchFanProfileOverviewOptions = {}): Promise<FanProfileOverview> {
  const headers = new Headers();

  if (sessionToken) {
    headers.set("Cookie", `${viewerSessionCookieName}=${sessionToken}`);
  }

  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      headers,
    },
    path: "/api/fan/profile",
    schema: fanProfileOverviewResponseSchema,
  });

  return response.data.fanProfile;
}
