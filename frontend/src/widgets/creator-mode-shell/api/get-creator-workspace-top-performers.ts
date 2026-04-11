import { z } from "zod";

import { requestJson } from "@/shared/api";

const creatorWorkspaceTopPerformerMediaSchema = z.object({
  durationSeconds: z.number().int().positive(),
  id: z.string().min(1),
  kind: z.literal("video"),
  posterUrl: z.string().min(1),
});

const creatorWorkspaceTopMainPerformerSchema = z.object({
  id: z.string().min(1),
  media: creatorWorkspaceTopPerformerMediaSchema,
  unlockCount: z.number().int().nonnegative(),
});

const creatorWorkspaceTopShortPerformerSchema = z.object({
  attributedUnlockCount: z.number().int().nonnegative(),
  id: z.string().min(1),
  media: creatorWorkspaceTopPerformerMediaSchema,
});

const creatorWorkspaceTopPerformersResponseSchema = z.object({
  data: z.object({
    topPerformers: z.object({
      topMain: creatorWorkspaceTopMainPerformerSchema.nullable(),
      topShort: creatorWorkspaceTopShortPerformerSchema.nullable(),
    }),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorWorkspaceTopMainPerformer = z.output<typeof creatorWorkspaceTopMainPerformerSchema>;
export type CreatorWorkspaceTopShortPerformer = z.output<typeof creatorWorkspaceTopShortPerformerSchema>;
export type CreatorWorkspaceTopPerformersResponse = z.output<
  typeof creatorWorkspaceTopPerformersResponseSchema
>["data"]["topPerformers"];

type GetCreatorWorkspaceTopPerformersOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  signal?: AbortSignal;
};

/**
 * creator workspace top performers を contract-backed API から取得する。
 */
export async function getCreatorWorkspaceTopPerformers({
  baseUrl,
  credentials = "include",
  fetcher,
  signal,
}: GetCreatorWorkspaceTopPerformersOptions = {}): Promise<CreatorWorkspaceTopPerformersResponse> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials,
      ...(signal ? { signal } : {}),
    },
    path: "/api/creator/workspace/top-performers",
    schema: creatorWorkspaceTopPerformersResponseSchema,
  });

  return response.data.topPerformers;
}
