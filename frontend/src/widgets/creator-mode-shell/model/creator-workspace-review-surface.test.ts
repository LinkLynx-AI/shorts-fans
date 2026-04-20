import {
  buildCreatorWorkspaceReviewActionLabel,
  buildCreatorWorkspaceReviewPackageHeadline,
  deriveCreatorWorkspaceReviewNotifications,
  resolveCreatorWorkspaceObjectReviewBadge,
  resolveCreatorWorkspaceReviewBlockerLabel,
} from "./creator-workspace-review-surface";

describe("creator workspace review surface model helpers", () => {
  it("maps object states to tile badges", () => {
    expect(resolveCreatorWorkspaceObjectReviewBadge("pending_review")).toEqual({
      label: "審査中",
      tone: "pending",
    });
    expect(resolveCreatorWorkspaceObjectReviewBadge("rejected")).toEqual({
      label: "却下",
      tone: "removed",
    });
    expect(resolveCreatorWorkspaceObjectReviewBadge("draft")).toEqual({
      label: "未申請",
      tone: "paused",
    });
  });

  it("builds package headlines and submit labels", () => {
    expect(buildCreatorWorkspaceReviewActionLabel("submit")).toBe("審査へ申請");
    expect(buildCreatorWorkspaceReviewActionLabel("resubmit")).toBe("再申請する");
    expect(buildCreatorWorkspaceReviewActionLabel("none")).toBeNull();

    expect(buildCreatorWorkspaceReviewPackageHeadline({
      blockers: [],
      canonicalMainId: "main_aoi_blue_balcony",
      linkedShortCount: 2,
      readiness: "ready",
      reviewStatus: "changes_requested",
      submitAction: "resubmit",
    })).toBe("修正後に再申請できます。");

    expect(buildCreatorWorkspaceReviewPackageHeadline({
      blockers: ["main_price_missing"],
      canonicalMainId: "main_aoi_blue_balcony",
      linkedShortCount: 2,
      readiness: "blocked",
      reviewStatus: "changes_requested",
      submitAction: "none",
    })).toBe("再申請前に必要項目を満たしてください。");
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
      expect.objectContaining({
        key: "pending_review",
        label: "審査中",
      }),
    ]);
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
});
