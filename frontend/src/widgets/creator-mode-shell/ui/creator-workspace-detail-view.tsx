"use client";

import { ArrowLeft } from "lucide-react";
import type { ReactNode } from "react";
import { useState } from "react";

import {
  CreatorAvatar,
  type CreatorSummary,
} from "@/entities/creator";
import { Button } from "@/shared/ui";

import { formatJpy, createPosterStyle, createVideoPosterStyle, formatDurationLabel } from "../lib/creator-mode-shell-ui";
import type { CreatorModeShellReadyState } from "../model/creator-mode-shell";
import { useCreatorWorkspacePreviewDetail } from "../model/use-creator-workspace-preview-detail";
import type {
  ApprovedCreatorWorkspaceDetailMetric,
  ApprovedCreatorWorkspaceDetailSetting,
  ApprovedCreatorWorkspaceManagedTab,
  ApprovedCreatorWorkspaceState,
} from "../model/approved-creator-workspace";
import type {
  CreatorWorkspaceDetailPoster,
  CreatorWorkspaceDetailSelection,
  CreatorWorkspaceDetailViewSelection,
  CreatorWorkspaceLinkedPreviewItems,
  CreatorWorkspacePreviewDetailData,
  CreatorWorkspacePreviewDetailSelection,
  CreatorWorkspaceReadyPreviewCollections,
  CreatorWorkspaceResolvedDetailState,
} from "./creator-mode-shell.types";
import { CreatorWorkspaceManagedTile } from "./creator-workspace-managed-tile";
import { CreatorWorkspaceMetadataEditSheet } from "./creator-workspace-metadata-edit-sheet";
import { CreatorWorkspacePreviewDetailLinkedGrid } from "./creator-workspace-preview-grid";

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

function buildPreviewShortResolvedDetailState(
  detail: CreatorWorkspacePreviewDetailData,
): CreatorWorkspaceResolvedDetailState {
  if (detail.kind === "preview-main") {
    return {
      durationLabel: formatDurationLabel(detail.detail.main.durationSeconds),
      kindLabel: "本編",
      linkedMainShortId: null,
      linkedShortIds: [],
      metrics: [],
      settings: [
        { label: "価格", value: formatJpy(detail.detail.main.priceJpy) },
        { label: "長さ", value: formatDurationLabel(detail.detail.main.durationSeconds) },
      ],
      statusLabel: null,
      statusTone: null,
      summary: detail.detail.entryShort.caption.trim()
        ? `${detail.detail.entryShort.caption.trim()}から流入する本編です。`
        : "owner preview 用の本編です。",
    };
  }

  return {
    durationLabel: formatDurationLabel(detail.detail.short.previewDurationSeconds),
    kindLabel: "ショート",
    linkedMainShortId: null,
    linkedShortIds: [],
    metrics: [],
    settings: [
      { label: "長さ", value: formatDurationLabel(detail.detail.short.previewDurationSeconds) },
    ],
    statusLabel: null,
    statusTone: null,
    summary: detail.detail.short.caption.trim() || "caption はまだ設定されていません。",
  };
}

function resolvePreviewDetailPoster(detail: CreatorWorkspacePreviewDetailData): CreatorWorkspaceDetailPoster {
  return detail.kind === "preview-main"
    ? {
        kind: "preview",
        posterUrl: detail.detail.main.media.posterUrl ?? "",
      }
    : {
        kind: "preview",
        posterUrl: detail.detail.short.media.posterUrl ?? "",
      };
}

function resolvePreviewLinkedItems(
  detail: CreatorWorkspacePreviewDetailData,
  previewCollections: CreatorWorkspaceReadyPreviewCollections | null,
): CreatorWorkspaceLinkedPreviewItems {
  if (!previewCollections) {
    return [];
  }

  if (detail.kind === "preview-main") {
    return previewCollections.shorts.items.filter((item) => item.canonicalMainId === detail.detail.main.id);
  }

  return previewCollections.mains.items.filter((item) => item.id === detail.detail.short.canonicalMainId);
}

function CreatorWorkspaceDetailScreen({
  actionSlot,
  creator,
  detail,
  linkedContent,
  onBack,
  poster,
}: {
  actionSlot?: ReactNode;
  creator: CreatorSummary;
  detail: CreatorWorkspaceResolvedDetailState;
  linkedContent: ReactNode;
  onBack: () => void;
  poster: CreatorWorkspaceDetailPoster;
}) {
  return (
    <section className="relative z-[2] min-h-svh overflow-y-auto px-4 pb-10 pt-[14px] text-foreground">
      <div className="flex items-center justify-between gap-3">
        <Button className="-ml-2" onClick={onBack} size="icon" variant="ghost">
          <span className="sr-only">Back</span>
          <ArrowLeft className="size-5" strokeWidth={2.1} />
        </Button>
        {actionSlot ?? <span aria-hidden="true" className="block min-h-8 min-w-7" />}
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

        {linkedContent}
      </section>
    </section>
  );
}

function CreatorWorkspacePreviewDetailPane({
  detailSelection,
  onBack,
  onMainPriceSaved,
  onOpenDetail,
  previewCollections,
}: {
  detailSelection: CreatorWorkspacePreviewDetailSelection;
  onBack: () => void;
  onMainPriceSaved: (mainId: string, priceJpy: number) => void;
  onOpenDetail: (selection: CreatorWorkspaceDetailViewSelection) => void;
  previewCollections: CreatorWorkspaceReadyPreviewCollections | null;
}) {
  const { retry, state } = useCreatorWorkspacePreviewDetail(detailSelection);
  const [savedDetail, setSavedDetail] = useState<CreatorWorkspacePreviewDetailData | null>(null);

  if (state.kind === "loading") {
    return (
      <section className="relative z-[2] min-h-svh overflow-y-auto px-4 pb-10 pt-[14px] text-foreground">
        <div className="flex items-center justify-between gap-3">
          <Button className="-ml-2" onClick={onBack} size="icon" variant="ghost">
            <span className="sr-only">Back</span>
            <ArrowLeft className="size-5" strokeWidth={2.1} />
          </Button>
          <span aria-hidden="true" className="block min-h-8 min-w-7" />
        </div>

        <div className="mt-[18px] grid gap-4">
          <p className="m-0 text-sm text-muted" role="status">
            動画詳細を読み込んでいます...
          </p>
          <div aria-hidden="true" className="aspect-[4/5] animate-pulse rounded-[32px] bg-[rgba(167,220,249,0.28)]" />
          <div aria-hidden="true" className="h-20 animate-pulse rounded-[24px] bg-[rgba(167,220,249,0.16)]" />
        </div>
      </section>
    );
  }

  if (state.kind === "error") {
    return (
      <section className="relative z-[2] min-h-svh overflow-y-auto px-4 pb-10 pt-[14px] text-foreground">
        <div className="flex items-center justify-between gap-3">
          <Button className="-ml-2" onClick={onBack} size="icon" variant="ghost">
            <span className="sr-only">Back</span>
            <ArrowLeft className="size-5" strokeWidth={2.1} />
          </Button>
          <span aria-hidden="true" className="block min-h-8 min-w-7" />
        </div>

        <section className="mt-[18px] rounded-[24px] border border-[rgba(167,220,249,0.4)] bg-[#f8fbfd] px-4 py-5">
          <p className="m-0 text-sm leading-6 text-muted" role="alert">
            {state.message}
          </p>
          <div className="mt-3">
            <Button onClick={retry} size="sm" type="button" variant="secondary">
              再読み込み
            </Button>
          </div>
        </section>
      </section>
    );
  }

  const previewDetail = savedDetail ?? state.detail;
  const resolvedDetail = buildPreviewShortResolvedDetailState(previewDetail);
  const linkedPreviewItems = resolvePreviewLinkedItems(previewDetail, previewCollections);

  return (
    <CreatorWorkspaceDetailScreen
      actionSlot={(
        <CreatorWorkspaceMetadataEditSheet
          detail={previewDetail}
          onDetailSaved={setSavedDetail}
          onMainPriceSaved={onMainPriceSaved}
        />
      )}
      creator={previewDetail.detail.creator}
      detail={resolvedDetail}
      linkedContent={linkedPreviewItems.length > 0 ? (
        <CreatorWorkspaceDetailSection title={previewDetail.kind === "preview-main" ? "紐づくショート" : "紐づく本編"}>
          <CreatorWorkspacePreviewDetailLinkedGrid items={linkedPreviewItems} onOpenDetail={onOpenDetail} />
        </CreatorWorkspaceDetailSection>
      ) : null}
      onBack={onBack}
      poster={resolvePreviewDetailPoster(previewDetail)}
    />
  );
}

export function CreatorWorkspaceDetailView({
  creator,
  detailSelection,
  onBack,
  onMainPriceSaved,
  onOpenDetail,
  previewCollections,
  state,
}: {
  creator: CreatorSummary;
  detailSelection: CreatorWorkspaceDetailViewSelection;
  onBack: () => void;
  onMainPriceSaved: (mainId: string, priceJpy: number) => void;
  onOpenDetail: (selection: CreatorWorkspaceDetailViewSelection) => void;
  previewCollections: CreatorWorkspaceReadyPreviewCollections | null;
  state: CreatorModeShellReadyState;
}) {
  if (detailSelection.kind !== "mock") {
    return (
      <CreatorWorkspacePreviewDetailPane
        detailSelection={detailSelection}
        key={`${detailSelection.kind}:${detailSelection.id}`}
        onBack={onBack}
        onMainPriceSaved={onMainPriceSaved}
        onOpenDetail={onOpenDetail}
        previewCollections={previewCollections}
      />
    );
  }

  const mockDetail = state.workspace.detailsByTab[detailSelection.tab][detailSelection.shortId];
  const mockPoster = state.workspace.posters[detailSelection.shortId];

  if (!mockDetail || !mockPoster) {
    return null;
  }

  return (
    <CreatorWorkspaceDetailScreen
      creator={creator}
      detail={mockDetail}
      linkedContent={(
        <>
          {detailSelection.tab === "main" && mockDetail.linkedShortIds.length > 0 ? (
            <CreatorWorkspaceDetailSection title="紐づくショート">
              <CreatorWorkspaceDetailLinkedGrid
                items={mockDetail.linkedShortIds}
                onOpenDetail={onOpenDetail}
                tab="shorts"
                workspace={state.workspace}
              />
            </CreatorWorkspaceDetailSection>
          ) : null}

          {detailSelection.tab === "shorts" && mockDetail.linkedMainShortId ? (
            <CreatorWorkspaceDetailSection title="紐づく本編">
              <CreatorWorkspaceDetailLinkedGrid
                items={[mockDetail.linkedMainShortId]}
                onOpenDetail={onOpenDetail}
                tab="main"
                workspace={state.workspace}
              />
            </CreatorWorkspaceDetailSection>
          ) : null}
        </>
      )}
      onBack={onBack}
      poster={{
        kind: "mock",
        poster: mockPoster,
      }}
    />
  );
}
