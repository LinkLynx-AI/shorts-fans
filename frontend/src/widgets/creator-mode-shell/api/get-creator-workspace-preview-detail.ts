import { z } from "zod";

import { creatorSummarySchema } from "@/entities/creator";
import { requestJson } from "@/shared/api";

const creatorWorkspacePreviewVideoDisplayAssetSchema = z.object({
  durationSeconds: z.number().int().positive(),
  id: z.string().min(1),
  kind: z.literal("video"),
  posterUrl: z.string().min(1),
  url: z.string().min(1),
});

const creatorWorkspacePreviewShortSummarySchema = z.object({
  caption: z.string(),
  canonicalMainId: z.string().min(1),
  creatorId: z.string().min(1),
  id: z.string().min(1),
  media: creatorWorkspacePreviewVideoDisplayAssetSchema,
  previewDurationSeconds: z.number().int().positive(),
});

const creatorWorkspacePreviewAccessSchema = z.object({
  mainId: z.string().min(1),
  reason: z.enum(["owner_preview", "session_unlocked", "unlock_required"]),
  status: z.enum(["locked", "owner", "unlocked"]),
});

const creatorWorkspacePreviewShortDetailResponseSchema = z.object({
  data: z.object({
    preview: z.object({
      access: creatorWorkspacePreviewAccessSchema,
      creator: creatorSummarySchema,
      short: creatorWorkspacePreviewShortSummarySchema,
    }),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

const creatorWorkspacePreviewMainDetailResponseSchema = z.object({
  data: z.object({
    preview: z.object({
      access: creatorWorkspacePreviewAccessSchema,
      creator: creatorSummarySchema,
      entryShort: creatorWorkspacePreviewShortSummarySchema,
      main: z.object({
        durationSeconds: z.number().int().positive(),
        id: z.string().min(1),
        media: creatorWorkspacePreviewVideoDisplayAssetSchema,
        priceJpy: z.number().int().nonnegative(),
      }),
    }),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorWorkspacePreviewShortDetail = {
  access: z.output<typeof creatorWorkspacePreviewAccessSchema>;
  creator: z.output<typeof creatorSummarySchema>;
  kind: "preview-short";
  requestId: string;
  short: z.output<typeof creatorWorkspacePreviewShortSummarySchema>;
};

export type CreatorWorkspacePreviewMainDetail = {
  access: z.output<typeof creatorWorkspacePreviewAccessSchema>;
  creator: z.output<typeof creatorSummarySchema>;
  entryShort: z.output<typeof creatorWorkspacePreviewShortSummarySchema>;
  kind: "preview-main";
  main: z.output<typeof creatorWorkspacePreviewMainDetailResponseSchema>["data"]["preview"]["main"];
  requestId: string;
};

export type CreatorWorkspacePreviewDetail =
  | CreatorWorkspacePreviewMainDetail
  | CreatorWorkspacePreviewShortDetail;

type GetCreatorWorkspacePreviewDetailOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  signal?: AbortSignal;
};

function buildCreatorWorkspacePreviewDetailPath(
  pathname: "/api/creator/workspace/shorts" | "/api/creator/workspace/mains",
  id: string,
): `/${string}` {
  const resolvedId = encodeURIComponent(id.trim());

  return `${pathname}/${resolvedId}/preview`;
}

/**
 * creator workspace owner preview の short detail を取得する。
 */
export async function getCreatorWorkspacePreviewShortDetail(
  shortId: string,
  {
    baseUrl,
    credentials = "include",
    fetcher,
    signal,
  }: GetCreatorWorkspacePreviewDetailOptions = {},
): Promise<CreatorWorkspacePreviewShortDetail> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials,
      ...(signal ? { signal } : {}),
    },
    path: buildCreatorWorkspacePreviewDetailPath("/api/creator/workspace/shorts", shortId),
    schema: creatorWorkspacePreviewShortDetailResponseSchema,
  });

  return {
    access: response.data.preview.access,
    creator: response.data.preview.creator,
    kind: "preview-short",
    requestId: response.meta.requestId,
    short: response.data.preview.short,
  };
}

/**
 * creator workspace owner preview の main detail を取得する。
 */
export async function getCreatorWorkspacePreviewMainDetail(
  mainId: string,
  {
    baseUrl,
    credentials = "include",
    fetcher,
    signal,
  }: GetCreatorWorkspacePreviewDetailOptions = {},
): Promise<CreatorWorkspacePreviewMainDetail> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials,
      ...(signal ? { signal } : {}),
    },
    path: buildCreatorWorkspacePreviewDetailPath("/api/creator/workspace/mains", mainId),
    schema: creatorWorkspacePreviewMainDetailResponseSchema,
  });

  return {
    access: response.data.preview.access,
    creator: response.data.preview.creator,
    entryShort: response.data.preview.entryShort,
    kind: "preview-main",
    main: response.data.preview.main,
    requestId: response.meta.requestId,
  };
}
