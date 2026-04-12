"use client";

import {
  ChevronRight,
  Menu,
} from "lucide-react";
import Link from "next/link";

import { useFanModeEntry } from "@/features/creator-entry";
import type { CreatorSummary } from "@/entities/creator";
import {
  BottomSheetMenu,
  BottomSheetMenuAction,
  BottomSheetMenuClose,
  BottomSheetMenuGroup,
} from "@/shared/ui";

import type { CreatorModeShellReadyState } from "../model/creator-mode-shell";
import type {
  ApprovedCreatorWorkspaceManagedTab,
  ApprovedCreatorWorkspaceState,
} from "../model/approved-creator-workspace";
import type { CreatorWorkspacePreviewCollectionsState } from "../model/creator-workspace-preview-collections";
import type { CreatorWorkspaceSummaryState } from "../model/creator-workspace-summary";
import type { CreatorWorkspaceTopPerformersState } from "../model/creator-workspace-top-performers";
import type {
  CreatorWorkspacePreviewDetailSelection,
} from "./creator-mode-shell.types";
import { CreatorWorkspacePreviewGrid } from "./creator-workspace-preview-grid";
import { CreatorWorkspaceSummarySection } from "./creator-workspace-summary-section";
import { CreatorWorkspaceTopPerformers } from "./creator-workspace-top-performers";

function CreatorWorkspaceAccountMenu() {
  const {
    clearError,
    enterFanMode,
    errorMessage,
    isSubmitting,
  } = useFanModeEntry();

  return (
    <BottomSheetMenu
      description="creator workspace から fan mode へ戻るメニュー"
      title="アカウントメニュー"
      trigger={
        <button
          aria-label="Account menu"
          className="inline-flex size-[34px] items-center justify-center bg-transparent text-[#1082c8] transition hover:bg-[#1082c8]/10"
          onClick={clearError}
          type="button"
        >
          <Menu aria-hidden="true" className="size-5" strokeWidth={1.9} />
        </button>
      }
    >
      <BottomSheetMenuGroup>
        <BottomSheetMenuClose asChild>
          <BottomSheetMenuAction asChild>
            <Link href="/creator/settings/profile">
              <span>プロフィールを編集</span>
              <ChevronRight aria-hidden="true" className="size-4 text-muted" strokeWidth={2.2} />
            </Link>
          </BottomSheetMenuAction>
        </BottomSheetMenuClose>
        <BottomSheetMenuAction
          disabled={isSubmitting}
          onClick={() => {
            void enterFanMode();
          }}
          withDivider
        >
          <span>{isSubmitting ? "Fan mode に切り替えています..." : "Fan mode に切り替え"}</span>
          <ChevronRight aria-hidden="true" className="size-4 text-muted" strokeWidth={2.2} />
        </BottomSheetMenuAction>
      </BottomSheetMenuGroup>

      {errorMessage ? (
        <p
          aria-live="polite"
          className="mt-3 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
          role="alert"
        >
          {errorMessage}
        </p>
      ) : null}
    </BottomSheetMenu>
  );
}

function CreatorWorkspaceTopBar() {
  return (
    <div className="flex items-center justify-between gap-2.5">
      <Link
        aria-label="動画を追加"
        className="inline-flex size-7 items-center justify-center bg-transparent text-[#1082c8] transition hover:opacity-80"
        href="/creator/upload"
      >
        <span aria-hidden="true" className="text-[34px] font-extralight leading-none">
          +
        </span>
      </Link>
      <CreatorWorkspaceAccountMenu />
    </div>
  );
}

function CreatorWorkspaceManagedPosts({
  activeTab,
  onChangeTab,
  onOpenPreviewDetail,
  onRetry,
  state,
  workspace,
}: {
  activeTab: ApprovedCreatorWorkspaceManagedTab;
  onChangeTab: (tab: ApprovedCreatorWorkspaceManagedTab) => void;
  onOpenPreviewDetail: (selection: CreatorWorkspacePreviewDetailSelection) => void;
  onRetry: () => void;
  state: CreatorWorkspacePreviewCollectionsState;
  workspace: ApprovedCreatorWorkspaceState;
}) {
  const activeTabLabel = activeTab === "shorts" ? "ショート" : "本編";

  return (
    <>
      <div
        className="mt-[18px] grid grid-cols-2 border-t border-[rgba(167,220,249,0.48)]"
      >
        {workspace.managedCollections.tabs.map((tab) => {
          const active = tab.key === activeTab;

          return (
            <button
              aria-label={tab.label}
              aria-pressed={active}
              className={`inline-flex min-h-[42px] min-w-0 items-center justify-center border-t-2 pt-[10px] text-xs font-bold uppercase tracking-[0.08em] ${
                active ? "border-t-foreground text-foreground" : "border-t-transparent text-muted"
              }`}
              key={tab.key}
              onClick={() => {
                onChangeTab(tab.key);
              }}
              type="button"
            >
              {tab.label}
            </button>
          );
        })}
      </div>

      <CreatorWorkspacePreviewGrid
        activeTab={activeTab}
        activeTabLabel={activeTabLabel}
        onOpenDetail={onOpenPreviewDetail}
        onRetry={onRetry}
        state={state}
      />
    </>
  );
}

export function CreatorWorkspaceDashboard({
  activeTab,
  creator,
  onChangeTab,
  onOpenPreviewDetail,
  onRetryPreviewCollections,
  onRetrySummary,
  onRetryTopPerformers,
  previewCollectionsState,
  state,
  summaryState,
  topPerformersState,
}: {
  activeTab: ApprovedCreatorWorkspaceManagedTab;
  creator: CreatorSummary;
  onChangeTab: (tab: ApprovedCreatorWorkspaceManagedTab) => void;
  onOpenPreviewDetail: (selection: CreatorWorkspacePreviewDetailSelection) => void;
  onRetryPreviewCollections: () => void;
  onRetrySummary: () => void;
  onRetryTopPerformers: () => void;
  previewCollectionsState: CreatorWorkspacePreviewCollectionsState;
  state: CreatorModeShellReadyState;
  summaryState: CreatorWorkspaceSummaryState;
  topPerformersState: CreatorWorkspaceTopPerformersState;
}) {
  return (
    <section className="relative z-[2] min-h-svh overflow-y-auto px-4 pb-24 pt-[14px] text-foreground">
      <h1 className="sr-only">{creator.displayName} creator workspace</h1>
      <CreatorWorkspaceTopBar />
      <CreatorWorkspaceSummarySection onRetry={onRetrySummary} state={summaryState} />
      <CreatorWorkspaceTopPerformers
        onOpenDetail={onOpenPreviewDetail}
        onRetry={onRetryTopPerformers}
        previewCollectionsState={previewCollectionsState}
        state={topPerformersState}
      />
      <CreatorWorkspaceManagedPosts
        activeTab={activeTab}
        onChangeTab={onChangeTab}
        onOpenPreviewDetail={onOpenPreviewDetail}
        onRetry={onRetryPreviewCollections}
        state={previewCollectionsState}
        workspace={state.workspace}
      />
    </section>
  );
}
