import { requestJson } from "@/shared/api";

import {
  creatorReviewCaseResponseSchema,
  type CreatorReviewCase,
} from "./contracts";

type GetCreatorReviewCaseOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
  userId: string;
};

/**
 * admin creator review detail を user 単位で取得する。
 */
export async function getCreatorReviewCase({
  baseUrl,
  fetcher,
  userId,
}: GetCreatorReviewCaseOptions): Promise<CreatorReviewCase> {
  const path = `/api/admin/creator-reviews/${userId}` as `/${string}`;
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials: "include",
    },
    path,
    schema: creatorReviewCaseResponseSchema,
  });

  return response.data.case;
}
