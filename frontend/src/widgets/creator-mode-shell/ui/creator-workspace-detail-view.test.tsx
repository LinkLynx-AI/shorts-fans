import {
  getMockCreatorModeShellState,
  type CreatorModeShellReadyState,
} from "../index";
import type {
  CreatorWorkspacePreviewMainItem,
  CreatorWorkspacePreviewShortItem,
} from "../api/get-creator-workspace-preview-collections";
import type { CreatorWorkspacePreviewDetailState } from "../model/use-creator-workspace-preview-detail";
import type { CreatorWorkspaceDetailViewSelection } from "./creator-mode-shell.types";
import { resolveCreatorWorkspaceDetailSummary } from "./creator-workspace-detail-view";

function getMockReadyState(): CreatorModeShellReadyState {
  return getMockCreatorModeShellState("dashboard");
}

function buildPreviewShortSelection(): Extract<CreatorWorkspaceDetailViewSelection, { kind: "preview-short" }> {
  const item: CreatorWorkspacePreviewShortItem = {
    canonicalMainId: "main_quiet_rooftop",
    id: "short_quiet_rooftop",
    media: {
      durationSeconds: 16,
      id: "asset_short_quiet_rooftop",
      kind: "video",
      posterUrl: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop-poster.jpg",
    },
    previewDurationSeconds: 16,
  };

  return {
    index: 0,
    item,
    kind: "preview-short",
    tab: "shorts",
  };
}

function buildPreviewMainSelection(): Extract<CreatorWorkspaceDetailViewSelection, { kind: "preview-main" }> {
  const item: CreatorWorkspacePreviewMainItem = {
    durationSeconds: 720,
    id: "main_quiet_rooftop",
    leadShortId: "short_quiet_rooftop",
    media: {
      durationSeconds: 720,
      id: "asset_main_quiet_rooftop",
      kind: "video",
      posterUrl: "https://cdn.example.com/creator/preview/mains/quiet-rooftop-poster.jpg",
    },
    priceJpy: 1800,
  };

  return {
    index: 0,
    item,
    kind: "preview-main",
    tab: "main",
  };
}

function buildReadyPreviewShortDetailState(caption: string): CreatorWorkspacePreviewDetailState {
  const state = getMockReadyState();

  return {
    detail: {
      access: {
        mainId: "main_quiet_rooftop",
        reason: "owner_preview",
        status: "owner",
      },
      creator: state.creator,
      kind: "preview-short",
      requestId: "req_creator_workspace_short_detail_001",
      short: {
        caption,
        canonicalMainId: "main_quiet_rooftop",
        creatorId: state.creator.id,
        id: "short_quiet_rooftop",
        media: {
          durationSeconds: 16,
          id: "asset_short_quiet_rooftop",
          kind: "video",
          posterUrl: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop-poster.jpg",
          url: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop.mp4",
        },
        previewDurationSeconds: 16,
      },
    },
    kind: "ready",
  };
}

describe("resolveCreatorWorkspaceDetailSummary", () => {
  it("returns the ready short caption for owner preview shorts", () => {
    expect(
      resolveCreatorWorkspaceDetailSummary(
        buildPreviewShortSelection(),
        "",
        buildReadyPreviewShortDetailState("quiet rooftop preview."),
      ),
    ).toBe("quiet rooftop preview.");
  });

  it("hides the summary for owner preview mains", () => {
    expect(
      resolveCreatorWorkspaceDetailSummary(
        buildPreviewMainSelection(),
        "owner preview 一覧から取得した本編データです。",
        {
          kind: "loading",
        },
      ),
    ).toBeNull();
  });

  it("keeps mock workspace summaries unchanged", () => {
    const state = getMockReadyState();
    const firstManagedShort = state.workspace.managedCollections.itemsByTab.shorts[0];

    expect(firstManagedShort).toBeDefined();

    if (!firstManagedShort) {
      throw new Error("mock creator workspace short is missing");
    }

    const mockShortId = firstManagedShort.shortId;
    const mockDetail = state.workspace.detailsByTab.shorts[mockShortId];

    expect(mockDetail).toBeDefined();
    expect(
      resolveCreatorWorkspaceDetailSummary(
        {
          kind: "mock",
          shortId: mockShortId,
          tab: "shorts",
        },
        mockDetail?.summary ?? "",
        {
          kind: "idle",
        },
      ),
    ).toBe(mockDetail?.summary);
  });
});
