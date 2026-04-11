import { z } from "zod";

import { creatorSummarySchema } from "@/entities/creator";
import { requestJson } from "@/shared/api";

const creatorWorkspaceOwnerPreviewAccessSchema = z.object({
  mainId: z.string().min(1),
  reason: z.literal("owner_preview"),
  status: z.literal("owner"),
});

const creatorWorkspacePreviewVideoDisplayAssetSchema = z.object({
  durationSeconds: z.number().int().positive().nullable(),
  id: z.string().min(1),
  kind: z.literal("video"),
  posterUrl: z.string().min(1).nullable(),
  url: z.string().min(1),
});

const creatorWorkspacePreviewShortDetailItemSchema = z.object({
  caption: z.string(),
  canonicalMainId: z.string().min(1),
  creatorId: z.string().min(1),
  id: z.string().min(1),
  media: creatorWorkspacePreviewVideoDisplayAssetSchema,
  previewDurationSeconds: z.number().int().positive(),
});

const creatorWorkspacePreviewMainDetailItemSchema = z.object({
  durationSeconds: z.number().int().positive(),
  id: z.string().min(1),
  media: creatorWorkspacePreviewVideoDisplayAssetSchema,
  priceJpy: z.number().int().positive(),
});

const creatorWorkspacePreviewShortDetailResponseSchema = z.object({
  data: z.object({
    preview: z.object({
      access: creatorWorkspaceOwnerPreviewAccessSchema,
      creator: creatorSummarySchema,
      short: creatorWorkspacePreviewShortDetailItemSchema,
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
      access: creatorWorkspaceOwnerPreviewAccessSchema,
      creator: creatorSummarySchema,
      entryShort: creatorWorkspacePreviewShortDetailItemSchema,
      main: creatorWorkspacePreviewMainDetailItemSchema,
    }),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorWorkspacePreviewShortDetail = z.output<
  typeof creatorWorkspacePreviewShortDetailResponseSchema
>["data"]["preview"];
export type CreatorWorkspacePreviewMainDetail = z.output<
  typeof creatorWorkspacePreviewMainDetailResponseSchema
>["data"]["preview"];

type GetCreatorWorkspacePreviewDetailOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  signal?: AbortSignal;
};

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
    path: `/api/creator/workspace/shorts/${shortId}/preview`,
    schema: creatorWorkspacePreviewShortDetailResponseSchema,
  });

  return response.data.preview;
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
    path: `/api/creator/workspace/mains/${mainId}/preview`,
    schema: creatorWorkspacePreviewMainDetailResponseSchema,
  });

  return response.data.preview;
}
