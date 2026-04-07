import { z } from "zod";

import { requestJson } from "@/shared/api";

import type { CreatorSummary } from "../model/creator";

const creatorAvatarAssetSchema = z.object({
  durationSeconds: z.null(),
  id: z.string().min(1),
  kind: z.literal("image"),
  posterUrl: z.string().min(1).nullable(),
  url: z.string().min(1),
});

const creatorSummarySchema = z.object({
  avatar: creatorAvatarAssetSchema.nullable(),
  bio: z.string(),
  displayName: z.string().min(1),
  handle: z.custom<`@${string}`>((value) => typeof value === "string" && value.startsWith("@")),
  id: z.string().min(1),
});

const creatorSearchResponseSchema = z.object({
  data: z.object({
    items: z.array(
      z.object({
        creator: creatorSummarySchema,
      }),
    ),
    query: z.string(),
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

export type CreatorSearchPage = {
  items: readonly CreatorSummary[];
  page: {
    hasNext: boolean;
    nextCursor: string | null;
  };
  query: string;
  requestId: string;
};

type GetCreatorSearchResultsOptions = {
  baseUrl?: string | undefined;
  cursor?: string | undefined;
  fetcher?: typeof fetch | undefined;
  query?: string | undefined;
  signal?: AbortSignal | undefined;
};

function buildCreatorSearchPath(query: string, cursor?: string): `/${string}` {
  const searchParams = new URLSearchParams();
  const trimmedQuery = query.trim();
  const trimmedCursor = cursor?.trim();

  if (trimmedQuery.length > 0) {
    searchParams.set("q", trimmedQuery);
  }
  if (trimmedCursor) {
    searchParams.set("cursor", trimmedCursor);
  }

  const queryString = searchParams.toString();

  return queryString.length > 0
    ? (`/api/fan/creators/search?${queryString}` as `/${string}`)
    : "/api/fan/creators/search";
}

/**
 * creator search API から recent / filtered result を取得する。
 */
export async function getCreatorSearchResults({
  baseUrl,
  cursor,
  fetcher,
  query = "",
  signal,
}: GetCreatorSearchResultsOptions = {}): Promise<CreatorSearchPage> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      ...(signal ? { signal } : {}),
    },
    path: buildCreatorSearchPath(query, cursor),
    schema: creatorSearchResponseSchema,
  });

  return {
    items: response.data.items.map((item) => item.creator),
    page: response.meta.page,
    query: response.data.query,
    requestId: response.meta.requestId,
  };
}
