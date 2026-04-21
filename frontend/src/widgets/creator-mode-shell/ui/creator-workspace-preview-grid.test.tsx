import {
  render,
  screen,
} from "@testing-library/react";

import type {
  CreatorWorkspacePreviewMainItem,
  CreatorWorkspacePreviewShortItem,
} from "../api/get-creator-workspace-preview-collections";
import type { CreatorWorkspacePreviewCollectionsState } from "../model/creator-workspace-preview-collections";
import type { CreatorWorkspaceReviewSurfaceState } from "../model/creator-workspace-review-surface";
import { CreatorWorkspacePreviewGrid } from "./creator-workspace-preview-grid";

function buildPreviewShort(): CreatorWorkspacePreviewShortItem {
  return {
    canonicalMainId: "main_rejected",
    id: "short_rejected",
    media: {
      durationSeconds: 16,
      id: "asset_short_rejected",
      kind: "video",
      posterUrl: "https://cdn.example.com/creator/preview/shorts/rejected-poster.jpg",
    },
    previewDurationSeconds: 16,
  };
}

function buildPreviewMain(): CreatorWorkspacePreviewMainItem {
  return {
    durationSeconds: 720,
    id: "main_rejected",
    leadShortId: "short_rejected",
    media: {
      durationSeconds: 720,
      id: "asset_main_rejected",
      kind: "video",
      posterUrl: "https://cdn.example.com/creator/preview/mains/rejected-poster.jpg",
    },
    priceJpy: 1800,
  };
}

function buildPreviewCollectionsState(): CreatorWorkspacePreviewCollectionsState {
  return {
    collections: {
      mains: {
        items: [buildPreviewMain()],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_preview_mains",
      },
      shorts: {
        items: [buildPreviewShort()],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_preview_shorts",
      },
    },
    kind: "ready",
  };
}

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
      packages: [],
      requestId: "req_review_surface",
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

describe("CreatorWorkspacePreviewGrid", () => {
  it("marks rejected videos with an in-thumbnail icon without showing a status label outside the thumbnail", () => {
    render(
      <CreatorWorkspacePreviewGrid
        activeTab="shorts"
        activeTabLabel="ショート"
        onOpenDetail={() => {}}
        onRetry={() => {}}
        reviewSurfaceState={buildRejectedReviewSurfaceState()}
        state={buildPreviewCollectionsState()}
      />,
    );

    const tile = screen.getByTestId("creator-workspace-preview-tile");

    expect(tile.closest("section")).toHaveClass("grid-cols-2");
    expect(tile).not.toHaveTextContent("却下");
    expect(tile).not.toHaveTextContent(/\bShort\b|\bMain\b/);
    expect(screen.getByTestId("creator-workspace-rejected-marker")).toBeInTheDocument();
  });
});
