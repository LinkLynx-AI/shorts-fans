"use server";

import {
  applyCreatorReviewDecision,
  type CreatorReviewDecision,
} from "@/entities/creator-review";

import { createAdminAPIFetcher } from "../../_lib/admin-api";
import { assertAdminUiAccess } from "../../_lib/admin-ui-access";

export type SubmitCreatorReviewDecisionInput = {
  decision: CreatorReviewDecision;
  isResubmitEligible: boolean;
  isSupportReviewRequired: boolean;
  reasonCode: string;
  userId: string;
};

export async function applyCreatorReviewDecisionFromAdmin(
  input: SubmitCreatorReviewDecisionInput,
): Promise<void> {
  await assertAdminUiAccess();
  await applyCreatorReviewDecision({
    ...input,
    fetcher: createAdminAPIFetcher(),
  });
}
