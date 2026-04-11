import { z } from "zod";

import {
  creatorSummarySchema,
} from "@/entities/creator";
import {
  publicShortSummarySchema,
} from "@/entities/short";
import { viewerSessionCookieName } from "@/entities/viewer";
import { requestJson } from "@/shared/api";

import type { FanPinnedShortItem } from "../model/fan-profile";

const fanPinnedShortItemSchema = z.object({
  creator: creatorSummarySchema,
  short: publicShortSummarySchema,
});

const fanProfilePinnedShortsResponseSchema = z.object({
  data: z.object({
    items: z.array(fanPinnedShortItemSchema),
  }),
  error: z.null(),
  meta: z.object({
    page: z.object({
      hasNext: z.boolean(),
      nextCursor: z.string().min(1).nullable(),
    }),
    requestId: z.string().min(1),
  }),
});

export type FanProfilePinnedShortsPage = {
  items: readonly FanPinnedShortItem[];
  page: {
    hasNext: boolean;
    nextCursor: string | null;
  };
  requestId: string;
};

type FetchFanProfilePinnedShortsPageOptions = {
  baseUrl?: string | undefined;
  cursor?: string | undefined;
  fetcher?: typeof fetch | undefined;
  sessionToken?: string | undefined;
  signal?: AbortSignal | undefined;
};

function buildFanProfilePinnedShortsPath(cursor?: string): `/${string}` {
  const trimmedCursor = cursor?.trim();

  if (!trimmedCursor) {
    return "/api/fan/profile/pinned-shorts";
  }

  const searchParams = new URLSearchParams();
  searchParams.set("cursor", trimmedCursor);

  return `/api/fan/profile/pinned-shorts?${searchParams.toString()}` as `/${string}`;
}

/**
 * fan profile pinned shorts API から short 一覧 1 page を取得する。
 */
export async function fetchFanProfilePinnedShortsPage({
  baseUrl,
  cursor,
  fetcher,
  sessionToken,
  signal,
}: FetchFanProfilePinnedShortsPageOptions = {}): Promise<FanProfilePinnedShortsPage> {
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
      ...(signal ? { signal } : {}),
    },
    path: buildFanProfilePinnedShortsPath(cursor),
    schema: fanProfilePinnedShortsResponseSchema,
  });

  return {
    items: response.data.items,
    page: response.meta.page,
    requestId: response.meta.requestId,
  };
}
