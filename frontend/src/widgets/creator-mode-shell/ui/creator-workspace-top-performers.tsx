"use client";

import { Button } from "@/shared/ui";

import type {
  CreatorWorkspaceTopMainPerformer,
  CreatorWorkspaceTopShortPerformer,
} from "../api/get-creator-workspace-top-performers";
import type { ApprovedCreatorWorkspaceState } from "../model/approved-creator-workspace";
import type { CreatorWorkspacePreviewCollectionsState } from "../model/creator-workspace-preview-collections";
import type { CreatorWorkspaceTopPerformersState } from "../model/creator-workspace-top-performers";
import {
  createVideoPosterStyle,
  formatUnlockMetric,
} from "../lib/creator-mode-shell-ui";
import type { CreatorWorkspaceDetailSelection } from "./creator-mode-shell.types";

function CreatorWorkspaceTopPerformerThumb({ posterUrl }: { posterUrl: string }) {
  return (
    <span
      aria-hidden="true"
      className="block h-10 w-[30px] shrink-0 rounded-[8px] bg-[#dbeaf2] shadow-[inset_0_0_0_1px_rgba(255,255,255,0.56)]"
      style={createVideoPosterStyle(posterUrl)}
    />
  );
}

function CreatorWorkspaceTopPerformerRow({
  index,
  label,
  metric,
  onOpenDetail,
  posterUrl,
  selection,
}: {
  index: number;
  label: string;
  metric: string;
  onOpenDetail: (selection: CreatorWorkspaceDetailSelection) => void;
  posterUrl: string;
  selection: CreatorWorkspaceDetailSelection | null;
}) {
  return (
    <button
      aria-label={`${label} ${metric}`}
      className={`flex min-h-[58px] w-full items-center justify-between gap-[14px] bg-transparent px-0 text-left text-foreground disabled:cursor-default disabled:opacity-100 ${
        index === 0 ? "" : "border-t border-[rgba(167,220,249,0.32)]"
      }`}
      disabled={selection === null}
      onClick={() => {
        if (selection) {
          onOpenDetail(selection);
        }
      }}
      type="button"
    >
      <span className="text-[11px] font-bold uppercase tracking-[0.1em] text-muted">{label}</span>
      <span className="flex min-w-0 items-center gap-3">
        <span className="text-sm font-bold leading-[1.3] text-foreground">{metric}</span>
        <CreatorWorkspaceTopPerformerThumb posterUrl={posterUrl} />
      </span>
    </button>
  );
}

function CreatorWorkspaceTopPerformersLoading() {
  return (
    <section aria-label="Top performers" className="mt-[18px] border-y border-[rgba(167,220,249,0.48)]">
      {["Top main", "Top short"].map((label, index) => (
        <button
          aria-label={label}
          className={`flex min-h-[58px] w-full items-center justify-between gap-[14px] bg-transparent px-0 text-left text-foreground disabled:cursor-default disabled:opacity-100 ${
            index === 0 ? "" : "border-t border-[rgba(167,220,249,0.32)]"
          }`}
          disabled
          key={label}
          type="button"
        >
          <span className="text-[11px] font-bold uppercase tracking-[0.1em] text-muted">{label}</span>
          <span className="flex min-w-0 items-center gap-3">
            <span aria-hidden="true" className="h-4 w-20 animate-pulse rounded-full bg-[rgba(167,220,249,0.28)]" />
            <span aria-hidden="true" className="block h-10 w-[30px] animate-pulse rounded-[8px] bg-[rgba(167,220,249,0.28)]" />
          </span>
        </button>
      ))}
    </section>
  );
}

function CreatorWorkspaceTopPerformersError({
  message,
  onRetry,
}: {
  message: string;
  onRetry: () => void;
}) {
  return (
    <section className="mt-[18px] rounded-[20px] border border-[rgba(167,220,249,0.4)] bg-[#f8fbfd] px-4 py-4 text-foreground">
      <p className="text-sm leading-6 text-muted" role="alert">
        {message}
      </p>
      <div className="mt-3">
        <Button onClick={onRetry} size="sm" type="button" variant="secondary">
          再読み込み
        </Button>
      </div>
    </section>
  );
}

function buildTopPerformerRows(
  topMain: CreatorWorkspaceTopMainPerformer | null,
  topShort: CreatorWorkspaceTopShortPerformer | null,
  previewCollectionsState: CreatorWorkspacePreviewCollectionsState,
  workspace: ApprovedCreatorWorkspaceState,
): readonly {
  label: string;
  metric: string;
  posterUrl: string;
  selection: CreatorWorkspaceDetailSelection | null;
}[] {
  const rows: {
    label: string;
    metric: string;
    posterUrl: string;
    selection: CreatorWorkspaceDetailSelection | null;
  }[] = [];
  const readyCollections = previewCollectionsState.kind === "ready" ? previewCollectionsState.collections : null;

  if (topMain) {
    const leadShortId = readyCollections?.mains.items.find((item) => item.id === topMain.id)?.leadShortId ?? null;

    rows.push({
      label: "Top main",
      metric: formatUnlockMetric(topMain.unlockCount),
      posterUrl: topMain.media.posterUrl,
      selection: leadShortId && workspace.detailsByTab.main[leadShortId]
        ? {
            kind: "mock",
            shortId: leadShortId,
            tab: "main",
          }
        : null,
    });
  }

  if (topShort) {
    rows.push({
      label: "Top short",
      metric: formatUnlockMetric(topShort.attributedUnlockCount),
      posterUrl: topShort.media.posterUrl,
      selection: workspace.detailsByTab.shorts[topShort.id]
        ? {
            kind: "mock",
            shortId: topShort.id,
            tab: "shorts",
          }
        : null,
    });
  }

  return rows;
}

export function CreatorWorkspaceTopPerformers({
  onOpenDetail,
  onRetry,
  previewCollectionsState,
  state,
  workspace,
}: {
  onOpenDetail: (selection: CreatorWorkspaceDetailSelection) => void;
  onRetry: () => void;
  previewCollectionsState: CreatorWorkspacePreviewCollectionsState;
  state: CreatorWorkspaceTopPerformersState;
  workspace: ApprovedCreatorWorkspaceState;
}) {
  if (state.kind === "loading") {
    return <CreatorWorkspaceTopPerformersLoading />;
  }

  if (state.kind === "error") {
    return <CreatorWorkspaceTopPerformersError message={state.message} onRetry={onRetry} />;
  }

  const rows = buildTopPerformerRows(
    state.topPerformers.topMain,
    state.topPerformers.topShort,
    previewCollectionsState,
    workspace,
  );

  if (rows.length === 0) {
    return null;
  }

  return (
    <section aria-label="Top performers" className="mt-[18px] border-y border-[rgba(167,220,249,0.48)]">
      {rows.map((item, index) => (
        <CreatorWorkspaceTopPerformerRow
          index={index}
          key={item.label}
          label={item.label}
          metric={item.metric}
          onOpenDetail={onOpenDetail}
          posterUrl={item.posterUrl}
          selection={item.selection}
        />
      ))}
    </section>
  );
}
