import { z } from "zod";

import { requestJson } from "@/shared/api";

const creatorWorkspacePreviewMediaSchema = z.object({
  durationSeconds: z.number().int().positive(),
  id: z.string().min(1),
  kind: z.literal("video"),
  posterUrl: z.string().min(1),
});

const creatorWorkspacePreviewShortItemSchema = z.object({
  canonicalMainId: z.string().min(1),
  id: z.string().min(1),
  media: creatorWorkspacePreviewMediaSchema,
  previewDurationSeconds: z.number().int().positive(),
});

const creatorWorkspacePreviewMainItemSchema = z.object({
  durationSeconds: z.number().int().positive(),
  id: z.string().min(1),
  leadShortId: z.string().min(1),
  media: creatorWorkspacePreviewMediaSchema,
  priceJpy: z.number().int().nonnegative(),
});

const creatorWorkspacePreviewPageSchema = z.object({
  hasNext: z.boolean(),
  nextCursor: z.string().min(1).nullable(),
});

const creatorWorkspacePreviewShortListResponseSchema = z.object({
  data: z.object({
    items: z.array(creatorWorkspacePreviewShortItemSchema),
  }),
  error: z.null(),
  meta: z.object({
    page: creatorWorkspacePreviewPageSchema,
    requestId: z.string().min(1),
  }),
});

const creatorWorkspacePreviewMainListResponseSchema = z.object({
  data: z.object({
    items: z.array(creatorWorkspacePreviewMainItemSchema),
  }),
  error: z.null(),
  meta: z.object({
    page: creatorWorkspacePreviewPageSchema,
    requestId: z.string().min(1),
  }),
});

export type CreatorWorkspacePreviewShortItem = z.output<typeof creatorWorkspacePreviewShortItemSchema>;
export type CreatorWorkspacePreviewMainItem = z.output<typeof creatorWorkspacePreviewMainItemSchema>;

export type CreatorWorkspacePreviewShortListPage = {
  items: readonly CreatorWorkspacePreviewShortItem[];
  page: {
    hasNext: boolean;
    nextCursor: string | null;
  };
  requestId: string;
};

export type CreatorWorkspacePreviewMainListPage = {
  items: readonly CreatorWorkspacePreviewMainItem[];
  page: {
    hasNext: boolean;
    nextCursor: string | null;
  };
  requestId: string;
};

type GetCreatorWorkspacePreviewListOptions = {
  baseUrl?: string;
  cursor?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  signal?: AbortSignal;
};

function buildCreatorWorkspacePreviewPath(
  pathname: "/api/creator/workspace/shorts" | "/api/creator/workspace/mains",
  cursor?: string,
): `/${string}` {
  const searchParams = new URLSearchParams();
  const trimmedCursor = cursor?.trim();

  if (trimmedCursor) {
    searchParams.set("cursor", trimmedCursor);
  }

  const queryString = searchParams.toString();

  return queryString.length > 0
    ? (`${pathname}?${queryString}` as `/${string}`)
    : pathname;
}

/**
 * creator workspace owner preview の short list を取得する。
 */
export async function getCreatorWorkspacePreviewShorts({
  baseUrl,
  cursor,
  credentials = "include",
  fetcher,
  signal,
}: GetCreatorWorkspacePreviewListOptions = {}): Promise<CreatorWorkspacePreviewShortListPage> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials,
      ...(signal ? { signal } : {}),
    },
    path: buildCreatorWorkspacePreviewPath("/api/creator/workspace/shorts", cursor),
    schema: creatorWorkspacePreviewShortListResponseSchema,
  });

  return {
    items: response.data.items,
    page: response.meta.page,
    requestId: response.meta.requestId,
  };
}

/**
 * creator workspace owner preview の main list を取得する。
 */
export async function getCreatorWorkspacePreviewMains({
  baseUrl,
  cursor,
  credentials = "include",
  fetcher,
  signal,
}: GetCreatorWorkspacePreviewListOptions = {}): Promise<CreatorWorkspacePreviewMainListPage> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials,
      ...(signal ? { signal } : {}),
    },
    path: buildCreatorWorkspacePreviewPath("/api/creator/workspace/mains", cursor),
    schema: creatorWorkspacePreviewMainListResponseSchema,
  });

  return {
    items: response.data.items,
    page: response.meta.page,
    requestId: response.meta.requestId,
  };
}
