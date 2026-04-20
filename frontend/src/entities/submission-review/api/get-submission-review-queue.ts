import { requestJson } from "@/shared/api";

import {
  submissionReviewQueueResponseSchema,
  type SubmissionReviewQueueItem,
} from "./contracts";

export type SubmissionReviewQueue = {
  items: SubmissionReviewQueueItem[];
};

type GetSubmissionReviewQueueOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
};

/**
 * pending submission review intake queue を取得する。
 */
export async function getSubmissionReviewQueue({
  baseUrl,
  fetcher,
}: GetSubmissionReviewQueueOptions = {}): Promise<SubmissionReviewQueue> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials: "include",
    },
    path: "/api/admin/submission-reviews",
    schema: submissionReviewQueueResponseSchema,
  });

  return response.data;
}
