import { requestJson } from "@/shared/api";

import {
  creatorReviewQueueResponseSchema,
  type CreatorReviewQueueItem,
  type CreatorReviewQueueState,
} from "./contracts";

type GetCreatorReviewQueueOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
  state: CreatorReviewQueueState;
};

export type CreatorReviewQueue = {
  items: CreatorReviewQueueItem[];
  state: CreatorReviewQueueState;
};

/**
 * admin creator review queue を state 単位で取得する。
 */
export async function getCreatorReviewQueue({
  baseUrl,
  fetcher,
  state,
}: GetCreatorReviewQueueOptions): Promise<CreatorReviewQueue> {
  const path = `/api/admin/creator-reviews?state=${encodeURIComponent(state)}` as `/${string}`;
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials: "include",
    },
    path,
    schema: creatorReviewQueueResponseSchema,
  });

  return response.data;
}
