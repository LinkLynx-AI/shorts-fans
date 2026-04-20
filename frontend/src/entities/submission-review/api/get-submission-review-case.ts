import { requestJson } from "@/shared/api";

import {
  submissionReviewCaseResponseSchema,
  type SubmissionReviewCase,
} from "./contracts";

type GetSubmissionReviewCaseOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
  intakeId: string;
};

/**
 * intake 単位の submission review detail を取得する。
 */
export async function getSubmissionReviewCase({
  baseUrl,
  fetcher,
  intakeId,
}: GetSubmissionReviewCaseOptions): Promise<SubmissionReviewCase> {
  const path = `/api/admin/submission-reviews/${intakeId}` as `/${string}`;
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials: "include",
    },
    path,
    schema: submissionReviewCaseResponseSchema,
  });

  return response.data.case;
}
