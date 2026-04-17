import { requestJson } from "@/shared/api";

import {
  creatorReviewCaseResponseSchema,
  type CreatorReviewCase,
  type CreatorReviewDecision,
} from "./contracts";

type ApplyCreatorReviewDecisionOptions = {
  baseUrl?: string;
  decision: CreatorReviewDecision;
  fetcher?: typeof fetch;
  isResubmitEligible?: boolean;
  isSupportReviewRequired?: boolean;
  reasonCode?: string;
  userId: string;
};

/**
 * admin creator review decision を適用し、更新後の case を返す。
 */
export async function applyCreatorReviewDecision({
  baseUrl,
  decision,
  fetcher,
  isResubmitEligible = false,
  isSupportReviewRequired = false,
  reasonCode = "",
  userId,
}: ApplyCreatorReviewDecisionOptions): Promise<CreatorReviewCase> {
  const path = `/api/admin/creator-reviews/${userId}/decision` as `/${string}`;
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      body: JSON.stringify({
        decision,
        isResubmitEligible,
        isSupportReviewRequired,
        reasonCode,
      }),
      cache: "no-store",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      method: "POST",
    },
    path,
    schema: creatorReviewCaseResponseSchema,
  });

  return response.data.case;
}
