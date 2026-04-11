import { z } from "zod";

import { requestJson } from "@/shared/api";

type UpdateCreatorWorkspaceMetadataOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
};

/**
 * creator workspace の main price を更新する。
 */
export async function updateCreatorWorkspaceMainPrice(
  mainId: string,
  priceJpy: number,
  options: UpdateCreatorWorkspaceMetadataOptions = {},
): Promise<void> {
  await requestJson({
    ...(options.baseUrl ? { baseUrl: options.baseUrl } : {}),
    ...(options.fetcher ? { fetcher: options.fetcher } : {}),
    init: {
      body: JSON.stringify({ priceJpy }),
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      method: "PUT",
    },
    path: `/api/creator/workspace/mains/${mainId}/price`,
    schema: z.undefined(),
  });
}

/**
 * creator workspace の short caption を更新する。
 */
export async function updateCreatorWorkspaceShortCaption(
  shortId: string,
  caption: string | null,
  options: UpdateCreatorWorkspaceMetadataOptions = {},
): Promise<void> {
  await requestJson({
    ...(options.baseUrl ? { baseUrl: options.baseUrl } : {}),
    ...(options.fetcher ? { fetcher: options.fetcher } : {}),
    init: {
      body: JSON.stringify({ caption }),
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      method: "PUT",
    },
    path: `/api/creator/workspace/shorts/${shortId}/caption`,
    schema: z.undefined(),
  });
}
