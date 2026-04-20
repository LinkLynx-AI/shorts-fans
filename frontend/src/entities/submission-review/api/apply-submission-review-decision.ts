import { requestJson } from "@/shared/api";
import { z } from "zod";

import {
  type SubmissionReviewDecision,
} from "./contracts";

type SubmissionReviewTargetDecisionInput = {
  decision: SubmissionReviewDecision;
  reasonCode?: string;
  reviewNote?: string;
};

type ApplySubmissionReviewDecisionOptions = {
  baseUrl?: string;
  decisionSource?: string;
  fetcher?: typeof fetch;
  intakeId: string;
  mainDecision?: SubmissionReviewTargetDecisionInput;
  shortDecisions: Array<SubmissionReviewTargetDecisionInput & { shortId: string }>;
};

/**
 * object-level decision を intake に適用する。
 */
export async function applySubmissionReviewDecision({
  baseUrl,
  decisionSource,
  fetcher,
  intakeId,
  mainDecision,
  shortDecisions,
}: ApplySubmissionReviewDecisionOptions): Promise<void> {
  const path = `/api/admin/submission-reviews/${intakeId}/decision` as `/${string}`;
  await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      body: JSON.stringify({
        ...(decisionSource ? { decisionSource } : {}),
        ...(mainDecision
          ? {
              mainDecision: {
                decision: mainDecision.decision,
                reasonCode: mainDecision.reasonCode ?? "",
                reviewNote: mainDecision.reviewNote ?? "",
              },
            }
          : {}),
        shortDecisions: shortDecisions.map((decision) => ({
          decision: decision.decision,
          reasonCode: decision.reasonCode ?? "",
          reviewNote: decision.reviewNote ?? "",
          shortId: decision.shortId,
        })),
      }),
      cache: "no-store",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      method: "POST",
    },
    path,
    schema: z.undefined(),
  });
}
