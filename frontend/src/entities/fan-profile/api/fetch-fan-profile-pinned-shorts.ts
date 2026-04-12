import { z } from "zod";

import {
  creatorSummarySchema,
} from "@/entities/creator";
import {
  publicShortSummarySchema,
} from "@/entities/short";
import { requestJson } from "@/shared/api";

import {
  buildFanProfileCursorPath,
  buildFanProfileRequestHeaders,
  fanProfileCursorMetaSchema,
  type FanProfileCursorPage,
} from "./shared";
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
  meta: fanProfileCursorMetaSchema,
});

export type FanProfilePinnedShortsPage = {
  items: readonly FanPinnedShortItem[];
  page: FanProfileCursorPage;
  requestId: string;
};

type FetchFanProfilePinnedShortsPageOptions = {
  baseUrl?: string | undefined;
  cursor?: string | undefined;
  fetcher?: typeof fetch | undefined;
  sessionToken?: string | undefined;
  signal?: AbortSignal | undefined;
};

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
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      headers: buildFanProfileRequestHeaders(sessionToken),
      ...(signal ? { signal } : {}),
    },
    path: buildFanProfileCursorPath("/api/fan/profile/pinned-shorts", cursor),
    schema: fanProfilePinnedShortsResponseSchema,
  });

  return {
    items: response.data.items,
    page: response.meta.page,
    requestId: response.meta.requestId,
  };
}
