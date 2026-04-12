import { z } from "zod";

import { requestJson } from "@/shared/api";

const creatorWorkspaceShortCaptionResponseSchema = z.object({
  data: z.object({
    short: z.object({
      caption: z.string(),
      id: z.string().min(1),
    }),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorWorkspaceShortCaptionUpdate = {
  requestId: string;
  short: z.output<typeof creatorWorkspaceShortCaptionResponseSchema>["data"]["short"];
};

type UpdateCreatorWorkspaceShortCaptionOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  signal?: AbortSignal;
};

function buildCreatorWorkspaceShortCaptionPath(shortId: string): `/${string}` {
  const resolvedShortId = encodeURIComponent(shortId.trim());

  return `/api/creator/workspace/shorts/${resolvedShortId}/caption`;
}

/**
 * creator workspace owner preview の short caption を更新する。
 */
export async function updateCreatorWorkspaceShortCaption(
  shortId: string,
  caption: string,
  {
    baseUrl,
    credentials = "include",
    fetcher,
    signal,
  }: UpdateCreatorWorkspaceShortCaptionOptions = {},
): Promise<CreatorWorkspaceShortCaptionUpdate> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      body: JSON.stringify({
        caption,
      }),
      credentials,
      headers: {
        "Content-Type": "application/json",
      },
      method: "PUT",
      ...(signal ? { signal } : {}),
    },
    path: buildCreatorWorkspaceShortCaptionPath(shortId),
    schema: creatorWorkspaceShortCaptionResponseSchema,
  });

  return {
    requestId: response.meta.requestId,
    short: response.data.short,
  };
}
