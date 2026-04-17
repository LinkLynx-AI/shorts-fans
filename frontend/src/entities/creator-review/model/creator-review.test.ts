import {
  buildCreatorReviewAvatarFallback,
  getCreatorReviewRejectHandling,
  creatorReviewReasonOptions,
  creatorReviewRejectHandlingOptions,
  formatCreatorReviewFileSize,
  formatCreatorReviewTimestamp,
  getCreatorReviewAvailableDecisions,
  getCreatorReviewDecisionLabel,
  getSuggestedCreatorReviewRejectHandling,
  getCreatorReviewReasonOption,
  getCreatorReviewStateLabel,
  normalizeCreatorReviewState,
} from "./creator-review";

describe("creator review model helpers", () => {
  it("normalizes unknown state values to submitted", () => {
    expect(normalizeCreatorReviewState(undefined)).toBe("submitted");
    expect(normalizeCreatorReviewState("unexpected")).toBe("submitted");
    expect(normalizeCreatorReviewState(["approved", "submitted"])).toBe("approved");
  });

  it("returns the available decision set per state", () => {
    expect(getCreatorReviewAvailableDecisions("submitted")).toEqual(["approved", "rejected"]);
    expect(getCreatorReviewAvailableDecisions("approved")).toEqual(["suspended"]);
    expect(getCreatorReviewAvailableDecisions("rejected")).toEqual([]);
  });

  it("formats labels and display helpers", () => {
    expect(getCreatorReviewStateLabel("submitted")).toBe("審査待ち");
    expect(getCreatorReviewDecisionLabel("suspended")).toBe("停止する");
    expect(buildCreatorReviewAvatarFallback("Mina Rei")).toBe("MR");
    expect(formatCreatorReviewTimestamp("2026-04-18T09:00:00Z")).toBe("2026/04/18 09:00 UTC");
    expect(formatCreatorReviewTimestamp(null)).toBe("未記録");
    expect(formatCreatorReviewFileSize(1536)).toBe("1.5 KB");
  });

  it("returns configured reason metadata when present", () => {
    expect(getCreatorReviewReasonOption("documents_blurry")).toEqual(creatorReviewReasonOptions[1]);
    expect(getCreatorReviewReasonOption("unknown_reason")).toBeNull();
    expect(getCreatorReviewReasonOption(null)).toBeNull();
  });

  it("suggests and resolves reject handling metadata", () => {
    expect(getSuggestedCreatorReviewRejectHandling("documents_blurry")).toBe("resubmit_eligible");
    expect(getSuggestedCreatorReviewRejectHandling("fraud_suspected")).toBe("support_review_required");
    expect(getCreatorReviewRejectHandling("support_review_required")).toEqual(creatorReviewRejectHandlingOptions[1]);
  });
});
