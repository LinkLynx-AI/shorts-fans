import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { createCreatorWorkspaceSubmissionReview } from "@/features/creator-workspace-submission-review";

import type { CreatorWorkspaceItemReviewSurfaceState } from "../model/creator-workspace-review-surface";
import { CreatorWorkspaceReviewPanel } from "./creator-workspace-review-panel";

vi.mock("@/features/creator-workspace-submission-review", () => ({
  createCreatorWorkspaceSubmissionReview: vi.fn(),
  CreatorWorkspaceSubmissionReviewApiError: class CreatorWorkspaceSubmissionReviewApiError extends Error {
    readonly code: string;

    constructor(code: string, message: string) {
      super(message);
      this.name = "CreatorWorkspaceSubmissionReviewApiError";
      this.code = code;
    }
  },
}));

type ReadyReviewState = Extract<CreatorWorkspaceItemReviewSurfaceState, { kind: "ready" }>;

function createDeferredPromise<TResult = void>() {
  let resolvePromise: (value: TResult | PromiseLike<TResult>) => void = () => {};
  let rejectPromise: (reason?: unknown) => void = () => {};
  const promise = new Promise<TResult>((resolve, reject) => {
    resolvePromise = resolve;
    rejectPromise = reject;
  });

  return {
    promise,
    reject: rejectPromise,
    resolve: resolvePromise,
  };
}

function buildRejectedShortReviewState(): ReadyReviewState {
  return {
    kind: "ready",
    surface: {
      package: {
        blockers: [],
        canonicalMainId: "main_rejected",
        linkedShortCount: 1,
        readiness: "none",
        reviewStatus: "rejected",
        submitAction: "none",
      },
      requestId: "req_rejected_review",
      review: {
        reasonCode: "content_safety_issue",
        state: "rejected",
      },
      target: {
        canonicalMainId: "main_rejected",
        id: "short_rejected",
        kind: "short",
      },
    },
  };
}

function buildApprovedShortReviewState(): ReadyReviewState {
  return {
    kind: "ready",
    surface: {
      package: {
        blockers: [],
        canonicalMainId: "main_approved",
        linkedShortCount: 1,
        readiness: "none",
        reviewStatus: "approved",
        submitAction: "none",
      },
      requestId: "req_approved_review",
      review: {
        reasonCode: null,
        state: "approved_for_publish",
      },
      target: {
        canonicalMainId: "main_approved",
        id: "short_approved",
        kind: "short",
      },
    },
  };
}

function buildApprovedConflictReviewState(): ReadyReviewState {
  const state = buildApprovedShortReviewState();
  state.surface.package.readiness = "conflict";
  return state;
}

function buildRevisionReadyReviewState(): ReadyReviewState {
  return {
    kind: "ready",
    surface: {
      package: {
        blockers: [],
        canonicalMainId: "main_revision_ready",
        linkedShortCount: 1,
        readiness: "ready",
        reviewStatus: "changes_requested",
        submitAction: "resubmit",
      },
      requestId: "req_revision_ready_review",
      review: {
        reasonCode: "caption_context_missing",
        state: "revision_requested",
      },
      target: {
        canonicalMainId: "main_revision_ready",
        id: "short_revision_ready",
        kind: "short",
      },
    },
  };
}

function buildRevisionReadyReviewStateWithRequestId(requestId: string): ReadyReviewState {
  const state = buildRevisionReadyReviewState();

  return {
    ...state,
    surface: {
      ...state.surface,
      requestId,
    },
  };
}

describe("CreatorWorkspaceReviewPanel", () => {
  beforeEach(() => {
    vi.mocked(createCreatorWorkspaceSubmissionReview).mockReset();
    vi.mocked(createCreatorWorkspaceSubmissionReview).mockResolvedValue(undefined);
  });

  it("renders rejected review state as a quiet detail status block", () => {
    render(<CreatorWorkspaceReviewPanel onRetry={() => {}} onSync={() => {}} state={buildRejectedShortReviewState()} />);

    expect(screen.getByText("公開できません")).toBeInTheDocument();
    expect(screen.getByText("却下")).toBeInTheDocument();
    expect(screen.getByText("却下されたため、この package は self-serve で再申請できません。")).toBeInTheDocument();
    expect(screen.getByText("対象")).toBeInTheDocument();
    expect(screen.getByText("package / ショート")).toBeInTheDocument();
    expect(screen.getByText("理由")).toBeInTheDocument();
    expect(screen.getByText("安全性の懸念")).toBeInTheDocument();
    expect(screen.getByText("コンテンツ安全性の観点で追加対応が必要です。")).toBeInTheDocument();
    expect(screen.queryByText("content_safety_issue")).not.toBeInTheDocument();
    expect(screen.queryByText("package 却下")).not.toBeInTheDocument();
    expect(screen.queryByText("ショート 却下")).not.toBeInTheDocument();
    expect(screen.queryByText("reason code")).not.toBeInTheDocument();
  });

  it("does not render normal approved review state", () => {
    const { container } = render(<CreatorWorkspaceReviewPanel onRetry={() => {}} onSync={() => {}} state={buildApprovedShortReviewState()} />);

    expect(container).toBeEmptyDOMElement();
  });

  it("does not render approved non-actionable conflict state", () => {
    const { container } = render(<CreatorWorkspaceReviewPanel onRetry={() => {}} onSync={() => {}} state={buildApprovedConflictReviewState()} />);

    expect(container).toBeEmptyDOMElement();
    expect(screen.queryByText("確認が必要です")).not.toBeInTheDocument();
    expect(screen.queryByText("要確認")).not.toBeInTheDocument();
  });

  it("resubmits revision-ready packages and syncs review state", async () => {
    const user = userEvent.setup();
    const onSync = vi.fn();

    render(<CreatorWorkspaceReviewPanel onRetry={() => {}} onSync={onSync} state={buildRevisionReadyReviewState()} />);

    expect(screen.queryByRole("button", { name: "審査へ申請" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "再申請する" }));

    await waitFor(() => {
      expect(createCreatorWorkspaceSubmissionReview).toHaveBeenCalledWith({
        mainId: "main_revision_ready",
      });
    });
    await waitFor(() => {
      expect(onSync).toHaveBeenCalledTimes(1);
    });
  });

  it("clears stale resubmit errors when the review surface changes", async () => {
    const user = userEvent.setup();
    vi.mocked(createCreatorWorkspaceSubmissionReview).mockRejectedValueOnce(new Error("boom"));

    const { rerender } = render(<CreatorWorkspaceReviewPanel onRetry={() => {}} onSync={() => {}} state={buildRevisionReadyReviewState()} />);

    await user.click(screen.getByRole("button", { name: "再申請する" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "再申請できませんでした。少し時間を置いてからやり直してください。",
    );

    rerender(<CreatorWorkspaceReviewPanel onRetry={() => {}} onSync={() => {}} state={buildRejectedShortReviewState()} />);

    await waitFor(() => {
      expect(screen.queryByText("再申請できませんでした。少し時間を置いてからやり直してください。")).not.toBeInTheDocument();
    });
    expect(screen.getByText("公開できません")).toBeInTheDocument();
  });

  it("ignores stale resubmit results after the review surface changes while submitting", async () => {
    const user = userEvent.setup();
    const deferred = createDeferredPromise();
    vi.mocked(createCreatorWorkspaceSubmissionReview).mockReturnValueOnce(deferred.promise);

    const { rerender } = render(<CreatorWorkspaceReviewPanel onRetry={() => {}} onSync={() => {}} state={buildRevisionReadyReviewState()} />);

    await user.click(screen.getByRole("button", { name: "再申請する" }));

    expect(screen.getByRole("button", { name: "再申請中..." })).toBeDisabled();

    rerender(<CreatorWorkspaceReviewPanel onRetry={() => {}} onSync={() => {}} state={buildRejectedShortReviewState()} />);
    deferred.reject(new Error("boom"));

    await waitFor(() => {
      expect(screen.queryByText("再申請できませんでした。少し時間を置いてからやり直してください。")).not.toBeInTheDocument();
    });
    expect(screen.queryByRole("button", { name: "再申請中..." })).not.toBeInTheDocument();
    expect(screen.getByText("公開できません")).toBeInTheDocument();
  });

  it("treats same-status review surface refetches as new surfaces", async () => {
    const user = userEvent.setup();
    const deferred = createDeferredPromise();
    vi.mocked(createCreatorWorkspaceSubmissionReview).mockReturnValueOnce(deferred.promise);

    const { rerender } = render(
      <CreatorWorkspaceReviewPanel
        onRetry={() => {}}
        onSync={() => {}}
        state={buildRevisionReadyReviewStateWithRequestId("req_revision_ready_review_1")}
      />,
    );

    await user.click(screen.getByRole("button", { name: "再申請する" }));

    expect(screen.getByRole("button", { name: "再申請中..." })).toBeDisabled();

    rerender(
      <CreatorWorkspaceReviewPanel
        onRetry={() => {}}
        onSync={() => {}}
        state={buildRevisionReadyReviewStateWithRequestId("req_revision_ready_review_2")}
      />,
    );
    deferred.reject(new Error("boom"));

    await waitFor(() => {
      expect(screen.queryByText("再申請できませんでした。少し時間を置いてからやり直してください。")).not.toBeInTheDocument();
    });
    expect(screen.getByRole("button", { name: "再申請する" })).toBeEnabled();
  });
});
