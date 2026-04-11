"use client";

import { ArrowLeft } from "lucide-react";
import type { ReactNode } from "react";

import {
  CreatorAvatar,
  type CreatorSummary,
} from "@/entities/creator";
import { Button } from "@/shared/ui";

import type { CreatorModeShellReadyState } from "../model/creator-mode-shell";
import type {
  ApprovedCreatorWorkspaceDetailMetric,
  ApprovedCreatorWorkspaceDetailSetting,
  ApprovedCreatorWorkspaceManagedTab,
  ApprovedCreatorWorkspaceState,
} from "../model/approved-creator-workspace";
import {
  buildPreviewMainDetailSettings,
  buildPreviewShortDetailSettings,
  createPosterStyle,
  createVideoPosterStyle,
  formatDurationLabel,
} from "../lib/creator-mode-shell-ui";
import type {
  CreatorWorkspaceDetailPoster,
  CreatorWorkspaceDetailSelection,
  CreatorWorkspaceDetailViewSelection,
  CreatorWorkspaceLinkedPreviewItems,
  CreatorWorkspacePreviewDetailSelection,
  CreatorWorkspaceReadyPreviewCollections,
  CreatorWorkspaceResolvedDetailState,
} from "./creator-mode-shell.types";
import { CreatorWorkspaceManagedTile } from "./creator-workspace-managed-tile";
import { CreatorWorkspacePreviewDetailLinkedGrid } from "./creator-workspace-preview-grid";

function CreatorWorkspaceActionButton({
  ariaLabel,
  children,
  className,
  disabled = true,
  onClick,
}: {
  ariaLabel: string;
  children: ReactNode;
  className: string;
  disabled?: boolean;
  onClick?: () => void;
}) {
  return (
    <button
      aria-label={ariaLabel}
      className={className}
      disabled={disabled}
      onClick={onClick}
      type="button"
    >
      {children}
    </button>
  );
}

function CreatorWorkspaceDetailMedia({
  detail,
  poster,
}: {
  detail: CreatorWorkspaceResolvedDetailState;
  poster: CreatorWorkspaceDetailPoster;
}) {
  const pendingLike = detail.statusTone === "pending" || detail.statusTone === "revision" || detail.statusTone === "paused";
  const mutedLike = detail.statusTone === "hidden" || detail.statusTone === "removed";
  const hasStatus = detail.statusTone !== null && detail.statusLabel !== null;

  return (
    <div className="relative overflow-hidden rounded-[32px]">
      <span
        aria-hidden="true"
        className={`block aspect-[4/5] bg-[linear-gradient(180deg,var(--creator-workspace-tile-top),var(--creator-workspace-tile-mid)_42%,var(--creator-workspace-tile-bottom)_100%)] ${
          hasStatus && detail.statusTone !== "approved" ? "brightness-[0.72] saturate-[0.82]" : ""
        }`}
        style={poster.kind === "mock" ? createPosterStyle(poster.poster) : createVideoPosterStyle(poster.posterUrl)}
      />
      <div
        className={`absolute inset-0 flex flex-col justify-between p-4 ${
          detail.statusTone === "approved" || !hasStatus
            ? "bg-[linear-gradient(180deg,rgba(6,21,33,0.12)_0%,rgba(6,21,33,0.04)_34%,rgba(6,21,33,0.58)_100%)]"
            : "bg-[linear-gradient(180deg,rgba(6,21,33,0.34)_0%,rgba(6,21,33,0.16)_34%,rgba(6,21,33,0.7)_100%)]"
        }`}
      >
        <div className="flex items-center justify-between gap-2.5">
          <span className="inline-flex min-h-[30px] items-center justify-center rounded-full bg-white/18 px-3 text-[11px] font-bold tracking-[0.08em] text-[#f8fcff] backdrop-blur-[10px]">
            {detail.kindLabel}
          </span>
          {hasStatus ? (
            <span
              className={`inline-flex min-h-[30px] items-center justify-center rounded-full px-3 text-[11px] font-bold tracking-[0.08em] backdrop-blur-[10px] ${
                detail.statusTone === "approved"
                  ? "bg-[rgba(52,168,83,0.18)] text-[#f4fff7]"
                  : pendingLike
                    ? "bg-[rgba(16,130,200,0.18)] text-[#eff8ff]"
                    : mutedLike
                      ? "bg-[rgba(7,19,29,0.18)] text-[#f6fbff]"
                      : "bg-[rgba(217,77,77,0.2)] text-[#fff6f6]"
              }`}
            >
              {detail.statusLabel}
            </span>
          ) : null}
        </div>

        <span className="relative mx-auto block size-[74px] rounded-full bg-white/18 backdrop-blur-[14px]">
          <span className="absolute left-1/2 top-1/2 -ml-[6px] -mt-3 h-0 w-0 border-y-[12px] border-y-transparent border-l-[18px] border-l-white" />
        </span>

        <span className="inline-flex min-h-[30px] w-fit items-center justify-center rounded-full bg-white/18 px-3 text-[11px] font-bold tracking-[0.08em] text-[#f8fcff] backdrop-blur-[10px]">
          {detail.durationLabel}
        </span>
      </div>
    </div>
  );
}

function CreatorWorkspaceDetailMetrics({ metrics }: { metrics: readonly ApprovedCreatorWorkspaceDetailMetric[] }) {
  return (
    <div className="grid grid-cols-2 border-y border-[rgba(167,220,249,0.48)] py-1">
      {metrics.map((metric, index) => (
        <div
          className={`grid gap-1 px-2 py-3 text-center ${metrics.length % 2 === 1 && index === metrics.length - 1 ? "col-span-2" : ""}`}
          key={metric.label}
        >
          <strong className="text-[18px] font-bold text-foreground">{metric.value}</strong>
          <span className="text-[11px] leading-[1.35] text-muted">{metric.label}</span>
        </div>
      ))}
    </div>
  );
}

function CreatorWorkspaceDetailSection({
  children,
  title,
}: {
  children: ReactNode;
  title: string;
}) {
  return (
    <section className="grid gap-3">
      <h3 className="m-0 text-sm font-bold text-foreground">{title}</h3>
      {children}
    </section>
  );
}

function CreatorWorkspaceDetailSettings({
  settings,
}: {
  settings: readonly ApprovedCreatorWorkspaceDetailSetting[];
}) {
  return (
    <div className="grid border-t border-[rgba(167,220,249,0.4)]">
      {settings.map((item) => (
        <div
          className="flex min-h-12 items-center justify-between gap-3 border-b border-[rgba(167,220,249,0.4)]"
          key={item.label}
        >
          <span className="text-[13px] text-muted">{item.label}</span>
          <strong className="text-right text-[13px] font-bold text-foreground">{item.value}</strong>
        </div>
      ))}
    </div>
  );
}

function CreatorWorkspaceDetailLinkedGrid({
  items,
  onOpenDetail,
  tab,
  workspace,
}: {
  items: readonly string[];
  onOpenDetail: (selection: CreatorWorkspaceDetailSelection) => void;
  tab: ApprovedCreatorWorkspaceManagedTab;
  workspace: ApprovedCreatorWorkspaceState;
}) {
  return (
    <div className="grid grid-cols-3 gap-[3px]">
      {items.flatMap((shortId) => {
        const poster = workspace.posters[shortId];
        const detail = workspace.detailsByTab[tab][shortId];

        if (!poster || !detail) {
          return [];
        }

        return [
          <CreatorWorkspaceManagedTile
            item={{
              detail: detail.summary,
              metric: "",
              shortId,
              status: detail.statusLabel,
              title: poster.title,
              tone: detail.statusTone,
            }}
            key={`${tab}:${shortId}`}
            onOpenDetail={onOpenDetail}
            poster={poster}
            tab={tab}
          />,
        ];
      })}
    </div>
  );
}

function resolvePreviewDetailState(
  detailSelection: CreatorWorkspacePreviewDetailSelection,
  previewCollections: CreatorWorkspaceReadyPreviewCollections,
): {
  detail: CreatorWorkspaceResolvedDetailState;
  linkedPreviewItems: CreatorWorkspaceLinkedPreviewItems;
  poster: CreatorWorkspaceDetailPoster;
} {
  if (detailSelection.kind === "preview-main") {
    return {
      detail: {
        durationLabel: formatDurationLabel(detailSelection.item.durationSeconds),
        kindLabel: "本編",
        linkedMainShortId: null,
        linkedShortIds: [],
        metrics: [],
        settings: buildPreviewMainDetailSettings(detailSelection.item),
        statusLabel: null,
        statusTone: null,
        summary: "owner preview 一覧から取得した本編データです。",
      },
      linkedPreviewItems: previewCollections.shorts.items.filter((item) => item.canonicalMainId === detailSelection.item.id),
      poster: {
        kind: "preview",
        posterUrl: detailSelection.item.media.posterUrl,
      },
    };
  }

  return {
    detail: {
      durationLabel: formatDurationLabel(detailSelection.item.previewDurationSeconds),
      kindLabel: "ショート",
      linkedMainShortId: null,
      linkedShortIds: [],
      metrics: [],
      settings: buildPreviewShortDetailSettings(detailSelection.item),
      statusLabel: null,
      statusTone: null,
      summary: "owner preview 一覧から取得したショートデータです。",
    },
    linkedPreviewItems: previewCollections.mains.items.filter((item) => item.id === detailSelection.item.canonicalMainId),
    poster: {
      kind: "preview",
      posterUrl: detailSelection.item.media.posterUrl,
    },
  };
}

export function CreatorWorkspaceDetailView({
  creator,
  detailSelection,
  onBack,
  onOpenDetail,
  previewCollections,
  state,
}: {
  creator: CreatorSummary;
  detailSelection: CreatorWorkspaceDetailViewSelection;
  onBack: () => void;
  onOpenDetail: (selection: CreatorWorkspaceDetailViewSelection) => void;
  previewCollections: CreatorWorkspaceReadyPreviewCollections | null;
  state: CreatorModeShellReadyState;
}) {
  let detail: CreatorWorkspaceResolvedDetailState | null = null;
  let linkedPreviewItems: CreatorWorkspaceLinkedPreviewItems = [];
  let poster: CreatorWorkspaceDetailPoster | null = null;

  if (detailSelection.kind === "mock") {
    const mockDetail = state.workspace.detailsByTab[detailSelection.tab][detailSelection.shortId];
    const mockPoster = state.workspace.posters[detailSelection.shortId];

    if (!mockDetail || !mockPoster) {
      return null;
    }

    detail = mockDetail;
    poster = {
      kind: "mock",
      poster: mockPoster,
    };
  } else {
    if (!previewCollections) {
      return null;
    }

    const previewDetail = resolvePreviewDetailState(detailSelection, previewCollections);
    detail = previewDetail.detail;
    linkedPreviewItems = previewDetail.linkedPreviewItems;
    poster = previewDetail.poster;
  }

  if (detail === null || poster === null) {
    return null;
  }

  return (
    <section className="relative z-[2] min-h-svh overflow-y-auto px-4 pb-10 pt-[14px] text-foreground">
      <div className="flex items-center justify-between gap-3">
        <Button className="-ml-2" onClick={onBack} size="icon" variant="ghost">
          <span className="sr-only">Back</span>
          <ArrowLeft className="size-5" strokeWidth={2.1} />
        </Button>
        <CreatorWorkspaceActionButton
          ariaLabel="投稿操作"
          className="inline-flex min-h-8 min-w-7 items-center justify-center gap-1 bg-transparent text-[#1082c8] disabled:cursor-default disabled:opacity-100"
        >
          <span className="size-1 rounded-full bg-current" />
          <span className="size-1 rounded-full bg-current" />
          <span className="size-1 rounded-full bg-current" />
        </CreatorWorkspaceActionButton>
      </div>

      <section className="mt-[18px] grid gap-[18px] pb-10">
        <div className="flex items-center gap-3">
          <CreatorAvatar
            className="size-[42px] rounded-full border-white/70 shadow-[0_10px_24px_rgba(36,92,129,0.16)]"
            creator={creator}
          />
          <div className="min-w-0">
            <p className="m-0 text-[13px] font-bold text-muted">{creator.handle}</p>
          </div>
        </div>

        <CreatorWorkspaceDetailMedia detail={detail} poster={poster} />

        <div className="grid gap-1.5">
          <p className="m-0 text-[15px] leading-[1.6] text-foreground">{detail.summary}</p>
        </div>

        {detail.metrics.length > 0 ? <CreatorWorkspaceDetailMetrics metrics={detail.metrics} /> : null}

        {detail.settings.length > 0 ? (
          <CreatorWorkspaceDetailSection title="設定">
            <CreatorWorkspaceDetailSettings settings={detail.settings} />
          </CreatorWorkspaceDetailSection>
        ) : null}

        {detailSelection.kind === "mock" && detailSelection.tab === "main" && detail.linkedShortIds.length > 0 ? (
          <CreatorWorkspaceDetailSection title="紐づくショート">
            <CreatorWorkspaceDetailLinkedGrid
              items={detail.linkedShortIds}
              onOpenDetail={onOpenDetail}
              tab="shorts"
              workspace={state.workspace}
            />
          </CreatorWorkspaceDetailSection>
        ) : null}

        {detailSelection.kind === "mock" && detailSelection.tab === "shorts" && detail.linkedMainShortId ? (
          <CreatorWorkspaceDetailSection title="紐づく本編">
            <CreatorWorkspaceDetailLinkedGrid
              items={[detail.linkedMainShortId]}
              onOpenDetail={onOpenDetail}
              tab="main"
              workspace={state.workspace}
            />
          </CreatorWorkspaceDetailSection>
        ) : null}

        {detailSelection.kind !== "mock" && linkedPreviewItems.length > 0 ? (
          <CreatorWorkspaceDetailSection title={detailSelection.kind === "preview-main" ? "紐づくショート" : "紐づく本編"}>
            <CreatorWorkspacePreviewDetailLinkedGrid items={linkedPreviewItems} onOpenDetail={onOpenDetail} />
          </CreatorWorkspaceDetailSection>
        ) : null}
      </section>
    </section>
  );
}
