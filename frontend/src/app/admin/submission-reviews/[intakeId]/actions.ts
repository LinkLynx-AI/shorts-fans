"use server";

import {
  applySubmissionReviewDecision,
  type SubmissionReviewDecision,
} from "@/entities/submission-review";

import { createAdminAPIFetcher } from "../../_lib/admin-api";
import { assertAdminUiAccess } from "../../_lib/admin-ui-access";

type SubmissionReviewTargetDecisionActionInput = {
  decision: SubmissionReviewDecision;
  reasonCode?: string;
  reviewNote?: string;
};

export type SubmitSubmissionReviewDecisionInput = {
  intakeId: string;
  mainDecision?: SubmissionReviewTargetDecisionActionInput;
  shortDecisions: Array<SubmissionReviewTargetDecisionActionInput & { shortId: string }>;
};

export async function applySubmissionReviewDecisionFromAdmin(
  input: SubmitSubmissionReviewDecisionInput,
): Promise<void> {
  await assertAdminUiAccess();
  await applySubmissionReviewDecision({
    ...input,
    fetcher: createAdminAPIFetcher(),
  });
}
