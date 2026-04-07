import { z } from "zod";

import { requestJson } from "@/shared/api";

import {
  creatorProfileShortGridItemSchema,
} from "./contracts";

const creatorProfileShortGridResponseSchema = z.object({
  data: z.object({
    items: z.array(creatorProfileShortGridItemSchema),
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

export type CreatorProfileShortGridItem = z.output<typeof creatorProfileShortGridItemSchema>;

export type CreatorProfileShortGridPage = {
  items: readonly CreatorProfileShortGridItem[];
  page: {
    hasNext: boolean;
    nextCursor: string | null;
  };
  requestId: string;
};

type GetCreatorProfileShortGridOptions = {
  baseUrl?: string | undefined;
  creatorId: string;
  cursor?: string | undefined;
  fetcher?: typeof fetch | undefined;
  signal?: AbortSignal | undefined;
};

function buildCreatorProfileShortGridPath(creatorId: string, cursor?: string): `/${string}` {
  const searchParams = new URLSearchParams();
  const trimmedCursor = cursor?.trim();

  if (trimmedCursor) {
    searchParams.set("cursor", trimmedCursor);
  }

  const queryString = searchParams.toString();
  const pathname = `/api/fan/creators/${encodeURIComponent(creatorId)}/shorts`;

  return queryString.length > 0
    ? (`${pathname}?${queryString}` as `/${string}`)
    : (pathname as `/${string}`);
}

/**
 * creator profile short grid を contract-backed API から取得する。
 */
export async function getCreatorProfileShortGrid({
  baseUrl,
  creatorId,
  cursor,
  fetcher,
  signal,
}: GetCreatorProfileShortGridOptions): Promise<CreatorProfileShortGridPage> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      ...(signal ? { signal } : {}),
    },
    path: buildCreatorProfileShortGridPath(creatorId, cursor),
    schema: creatorProfileShortGridResponseSchema,
  });

  return {
    items: response.data.items,
    page: response.meta.page,
    requestId: response.meta.requestId,
  };
}
