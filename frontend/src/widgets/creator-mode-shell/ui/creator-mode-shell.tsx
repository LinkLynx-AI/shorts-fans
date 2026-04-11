"use client";

import { useState } from "react";

import type { ApprovedCreatorWorkspaceManagedTab } from "../model/approved-creator-workspace";
import type {
  CreatorModeShellReadyState,
  CreatorModeShellState,
} from "../model/creator-mode-shell";
import { useCreatorWorkspacePreviewDetail } from "../model/use-creator-workspace-preview-detail";
import { useCreatorWorkspacePreviewCollections } from "../model/use-creator-workspace-preview-collections";
import { useCreatorWorkspaceSummary } from "../model/use-creator-workspace-summary";
import { useCreatorWorkspaceTopPerformers } from "../model/use-creator-workspace-top-performers";
import { CreatorShellBlockedState, CreatorModeWorkspaceFrame } from "./creator-mode-shell-blocked-state";
import type { CreatorWorkspaceDetailViewSelection } from "./creator-mode-shell.types";
import { CreatorWorkspaceDashboard } from "./creator-workspace-dashboard";
import { CreatorWorkspaceDetailView } from "./creator-workspace-detail-view";

function CreatorWorkspaceReadyState({ state }: { state: CreatorModeShellReadyState }) {
  const {
    blockedState: summaryBlockedState,
    retry: retrySummary,
    state: summaryState,
  } = useCreatorWorkspaceSummary();
  const {
    blockedState: previewBlockedState,
    retry: retryPreviewCollections,
    state: previewCollectionsState,
  } = useCreatorWorkspacePreviewCollections();
  const {
    blockedState: topPerformersBlockedState,
    retry: retryTopPerformers,
    state: topPerformersState,
  } = useCreatorWorkspaceTopPerformers();
  const [activeTab, setActiveTab] = useState<ApprovedCreatorWorkspaceManagedTab>(state.workspace.managedCollections.defaultTab);
  const [detailSelection, setDetailSelection] = useState<CreatorWorkspaceDetailViewSelection | null>(null);
  const previewDetailSelection = detailSelection?.kind === "mock" ? null : detailSelection;
  const {
    blockedState: previewDetailBlockedState,
    retry: retryPreviewDetail,
    state: previewDetailState,
  } = useCreatorWorkspacePreviewDetail(previewDetailSelection);
  const creator = summaryState.kind === "ready" ? summaryState.summary.creator : state.creator;
  const blockedState = summaryBlockedState ?? topPerformersBlockedState ?? previewBlockedState ?? previewDetailBlockedState;

  function handleOpenDetail(selection: CreatorWorkspaceDetailViewSelection) {
    setActiveTab(selection.tab);
    setDetailSelection(selection);
  }

  if (blockedState) {
    return <CreatorShellBlockedState state={blockedState} />;
  }

  return (
    <CreatorModeWorkspaceFrame>
      {detailSelection ? (
        <CreatorWorkspaceDetailView
          creator={creator}
          detailSelection={detailSelection}
          onBack={() => {
            setDetailSelection(null);
          }}
          onOpenDetail={handleOpenDetail}
          onRetryPreviewDetail={retryPreviewDetail}
          previewDetailState={previewDetailState}
          previewCollections={previewCollectionsState.kind === "ready" ? previewCollectionsState.collections : null}
          state={state}
        />
      ) : (
        <CreatorWorkspaceDashboard
          activeTab={activeTab}
          creator={creator}
          onChangeTab={setActiveTab}
          onOpenDetail={handleOpenDetail}
          onOpenPreviewDetail={handleOpenDetail}
          onRetryPreviewCollections={retryPreviewCollections}
          onRetrySummary={retrySummary}
          onRetryTopPerformers={retryTopPerformers}
          previewCollectionsState={previewCollectionsState}
          state={state}
          summaryState={summaryState}
          topPerformersState={topPerformersState}
        />
      )}
    </CreatorModeWorkspaceFrame>
  );
}

/**
 * `/creator` の route shell を表示する。
 */
export function CreatorModeShell({ state }: { state: CreatorModeShellState }) {
  if (state.kind !== "ready") {
    return <CreatorShellBlockedState state={state} />;
  }

  return <CreatorWorkspaceReadyState state={state} />;
}
