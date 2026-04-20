import {
  buildSubmissionReviewAvatarFallback,
  doesSubmissionReviewDecisionRequireReason,
  formatSubmissionReviewDecisionSource,
  formatSubmissionReviewPrice,
  formatSubmissionReviewTimestamp,
  getSubmissionReviewDecisionLabel,
  getSubmissionReviewReasonOption,
  getSubmissionReviewStateLabel,
  getSubmissionReviewSubmitKindLabel,
  submissionReviewReasonOptions,
} from "./submission-review";

describe("submission review model helpers", () => {
  it("formats labels and display helpers", () => {
    expect(getSubmissionReviewStateLabel("pending_review")).toBe("審査待ち");
    expect(getSubmissionReviewDecisionLabel("revision_requested")).toBe("修正依頼にする");
    expect(getSubmissionReviewSubmitKindLabel("resubmit")).toBe("再 submit");
    expect(buildSubmissionReviewAvatarFallback("Mina Rei")).toBe("MR");
    expect(formatSubmissionReviewTimestamp("2026-04-20T09:00:00Z")).toBe("2026/04/20 09:00 UTC");
    expect(formatSubmissionReviewTimestamp(null)).toBe("未記録");
    expect(formatSubmissionReviewDecisionSource("manual_override")).toBe("manual");
    expect(formatSubmissionReviewDecisionSource(null)).toBe("未記録");
    expect(formatSubmissionReviewPrice(1800)).toBe("1,800 円");
  });

  it("resolves reason metadata and requirement rules", () => {
    expect(doesSubmissionReviewDecisionRequireReason("approved")).toBe(false);
    expect(doesSubmissionReviewDecisionRequireReason("revision_requested")).toBe(true);
    expect(getSubmissionReviewReasonOption("quality_issue")).toEqual(submissionReviewReasonOptions[5]);
    expect(getSubmissionReviewReasonOption("unknown_reason")).toBeNull();
  });
});
