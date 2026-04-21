import {
  render,
  screen,
} from "@testing-library/react";

import type { CreatorWorkspaceReviewSurfaceState } from "../model/creator-workspace-review-surface";
import { CreatorWorkspaceReviewNotices } from "./creator-workspace-review-notices";

function buildRejectedReviewSurfaceState(): CreatorWorkspaceReviewSurfaceState {
  return {
    kind: "ready",
    surface: {
      mains: [
        {
          id: "main_rejected",
          state: "rejected",
        },
      ],
      packages: [
        {
          blockers: [],
          canonicalMainId: "main_rejected",
          linkedShortCount: 1,
          readiness: "none",
          reviewStatus: "rejected",
          submitAction: "none",
        },
      ],
      requestId: "req_rejected_review_surface",
      shorts: [
        {
          canonicalMainId: "main_rejected",
          id: "short_rejected",
          state: "rejected",
        },
      ],
    },
  };
}

describe("CreatorWorkspaceReviewNotices", () => {
  it("renders rejected packages as a compact status row without using the state label as a marker", () => {
    render(<CreatorWorkspaceReviewNotices onRetry={() => {}} state={buildRejectedReviewSurfaceState()} />);

    expect(screen.getByText("公開不可 1件")).toBeInTheDocument();
    expect(screen.getByText("該当動画の確認をお願いします")).toBeInTheDocument();
    expect(screen.queryByText("却下")).not.toBeInTheDocument();
    expect(screen.queryByText(/self-serve/)).not.toBeInTheDocument();
  });
});
