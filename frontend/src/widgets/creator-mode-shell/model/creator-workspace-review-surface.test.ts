import {
  buildCreatorWorkspaceReviewPackageHeadline,
  deriveCreatorWorkspaceReviewNotifications,
  hasCreatorWorkspaceReviewIssue,
  resolveCreatorWorkspaceObjectReviewBadge,
  resolveCreatorWorkspaceReviewBlockerLabel,
  resolveCreatorWorkspaceReviewReasonCopy,
} from "./creator-workspace-review-surface";

describe("creator workspace review surface model helpers", () => {
  it("maps only problematic object states to tile badges", () => {
    expect(resolveCreatorWorkspaceObjectReviewBadge("approved_for_publish")).toBeNull();
    expect(resolveCreatorWorkspaceObjectReviewBadge("approved_for_unlock")).toBeNull();
    expect(resolveCreatorWorkspaceObjectReviewBadge("pending_review")).toBeNull();
    expect(resolveCreatorWorkspaceObjectReviewBadge("draft")).toBeNull();

    expect(resolveCreatorWorkspaceObjectReviewBadge("rejected")).toEqual({
      label: "却下",
      tone: "removed",
    });
    expect(resolveCreatorWorkspaceObjectReviewBadge("revision_requested")).toEqual({
      label: "差し戻し",
      tone: "revision",
    });
  });

  it("builds package headlines", () => {
    expect(buildCreatorWorkspaceReviewPackageHeadline({
      blockers: [],
      canonicalMainId: "main_approved",
      linkedShortCount: 1,
      readiness: "none",
      reviewStatus: "approved",
      submitAction: "none",
    })).toBeNull();

    expect(buildCreatorWorkspaceReviewPackageHeadline({
      blockers: [],
      canonicalMainId: "main_pending",
      linkedShortCount: 1,
      readiness: "none",
      reviewStatus: "pending_review",
      submitAction: "none",
    })).toBeNull();

    expect(buildCreatorWorkspaceReviewPackageHeadline({
      blockers: [],
      canonicalMainId: "main_aoi_blue_balcony",
      linkedShortCount: 2,
      readiness: "ready",
      reviewStatus: "changes_requested",
      submitAction: "resubmit",
    })).toBe("修正内容を確認してください。");

    expect(buildCreatorWorkspaceReviewPackageHeadline({
      blockers: ["main_price_missing"],
      canonicalMainId: "main_aoi_blue_balcony",
      linkedShortCount: 2,
      readiness: "blocked",
      reviewStatus: "changes_requested",
      submitAction: "none",
    })).toBe("再審査前に必要項目を満たしてください。");
  });

  it("derives dashboard notifications from package summaries", () => {
    expect(deriveCreatorWorkspaceReviewNotifications([
      {
        blockers: [],
        canonicalMainId: "main_changes_requested",
        linkedShortCount: 2,
        readiness: "ready",
        reviewStatus: "changes_requested",
        submitAction: "resubmit",
      },
      {
        blockers: [],
        canonicalMainId: "main_approved",
        linkedShortCount: 1,
        readiness: "none",
        reviewStatus: "approved",
        submitAction: "none",
      },
      {
        blockers: [],
        canonicalMainId: "main_pending_review",
        linkedShortCount: 1,
        readiness: "none",
        reviewStatus: "pending_review",
        submitAction: "none",
      },
      {
        blockers: ["main_price_missing"],
        canonicalMainId: "main_blocked",
        linkedShortCount: 1,
        readiness: "blocked",
        reviewStatus: "draft",
        submitAction: "none",
      },
    ])).toEqual([
      expect.objectContaining({
        key: "changes_requested",
        label: "差し戻し",
      }),
      expect.objectContaining({
        key: "blocked_draft",
        label: "要確認",
      }),
    ]);
  });

  it("detects review issues only for states that need action", () => {
    expect(hasCreatorWorkspaceReviewIssue({
      package: {
        blockers: [],
        canonicalMainId: "main_approved",
        linkedShortCount: 1,
        readiness: "none",
        reviewStatus: "approved",
        submitAction: "none",
      },
      requestId: "req_approved",
      review: {
        reasonCode: null,
        state: "approved_for_publish",
      },
      target: {
        canonicalMainId: "main_approved",
        id: "short_approved",
        kind: "short",
      },
    })).toBe(false);

    expect(hasCreatorWorkspaceReviewIssue({
      package: {
        blockers: [],
        canonicalMainId: "main_revision",
        linkedShortCount: 1,
        readiness: "ready",
        reviewStatus: "changes_requested",
        submitAction: "resubmit",
      },
      requestId: "req_revision",
      review: {
        reasonCode: "caption_mismatch",
        state: "revision_requested",
      },
      target: {
        canonicalMainId: "main_revision",
        id: "short_revision",
        kind: "short",
      },
    })).toBe(true);
  });

  it("does not double-count blocked changes_requested packages as blocked drafts", () => {
    expect(deriveCreatorWorkspaceReviewNotifications([
      {
        blockers: ["main_price_missing"],
        canonicalMainId: "main_changes_requested_blocked",
        linkedShortCount: 2,
        readiness: "blocked",
        reviewStatus: "changes_requested",
        submitAction: "none",
      },
    ])).toEqual([
      expect.objectContaining({
        key: "changes_requested",
        label: "差し戻し",
      }),
    ]);
  });

  it("maps blocker codes to readable copy", () => {
    expect(resolveCreatorWorkspaceReviewBlockerLabel("main_price_missing")).toBe("本編価格が未設定です。");
    expect(resolveCreatorWorkspaceReviewBlockerLabel("linked_short_missing")).toBe("linked short がまだありません。");
  });

  it("maps reason codes to creator-facing copy without exposing raw codes", () => {
    expect(resolveCreatorWorkspaceReviewReasonCopy("content_safety_issue")).toEqual({
      description: "コンテンツ安全性の観点で追加対応が必要です。",
      label: "安全性の懸念",
    });
    expect(resolveCreatorWorkspaceReviewReasonCopy("unknown_reason")).toEqual({
      description: "詳細は運営からの案内を確認してください。",
      label: "審査基準の確認が必要です",
    });
    expect(resolveCreatorWorkspaceReviewReasonCopy(null)).toBeNull();
  });
});
