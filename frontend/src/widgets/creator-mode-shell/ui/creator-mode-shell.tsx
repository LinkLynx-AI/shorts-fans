"use client";

import * as Dialog from "@radix-ui/react-dialog";
import Link from "next/link";
import {
  useState,
  type CSSProperties,
  type ReactNode,
} from "react";
import {
  ArrowLeft,
  ChevronRight,
} from "lucide-react";

import {
  CreatorAvatar,
  type CreatorSummary,
} from "@/entities/creator";
import { useFanModeEntry } from "@/features/creator-entry";
import { Button, SurfacePanel } from "@/shared/ui";

import type {
  CreatorWorkspacePreviewMainItem,
  CreatorWorkspacePreviewShortItem,
} from "../api/get-creator-workspace-preview-collections";
import type {
  ApprovedCreatorWorkspaceDetailMetric,
  ApprovedCreatorWorkspaceDetailSetting,
  ApprovedCreatorWorkspaceOverviewMetrics,
  ApprovedCreatorWorkspaceManagedItem,
  ApprovedCreatorWorkspaceManagedItemTone,
  ApprovedCreatorWorkspaceManagedTab,
  ApprovedCreatorWorkspacePoster,
  ApprovedCreatorWorkspaceRevisionRequestedSummary,
  ApprovedCreatorWorkspaceState,
} from "../model/approved-creator-workspace";
import type { CreatorWorkspaceSummaryState } from "../model/creator-workspace-summary";
import type { CreatorWorkspacePreviewCollectionsState } from "../model/creator-workspace-preview-collections";
import type { CreatorModeShellState } from "../model/creator-mode-shell";
import { useCreatorWorkspacePreviewCollections } from "../model/use-creator-workspace-preview-collections";
import { useCreatorWorkspaceSummary } from "../model/use-creator-workspace-summary";

type CreatorWorkspaceDetailSelection = {
  kind: "mock";
  shortId: string;
  tab: ApprovedCreatorWorkspaceManagedTab;
};

type CreatorWorkspacePreviewDetailSelection =
  | {
      index: number;
      item: CreatorWorkspacePreviewMainItem;
      kind: "preview-main";
      tab: "main";
    }
  | {
      index: number;
      item: CreatorWorkspacePreviewShortItem;
      kind: "preview-short";
      tab: "shorts";
    };

type CreatorWorkspaceResolvedDetailState = {
  durationLabel: string;
  kindLabel: string;
  linkedMainShortId: string | null;
  linkedShortIds: readonly string[];
  metrics: readonly ApprovedCreatorWorkspaceDetailMetric[];
  settings: readonly ApprovedCreatorWorkspaceDetailSetting[];
  statusLabel: string | null;
  statusTone: ApprovedCreatorWorkspaceManagedItemTone | null;
  summary: string;
};

type CreatorWorkspaceDetailPoster =
  | {
      kind: "mock";
      poster: ApprovedCreatorWorkspacePoster;
    }
  | {
      kind: "preview";
      posterUrl: string;
    };

function formatCount(value: number): string {
  return value.toLocaleString("ja-JP");
}

function formatJpy(value: number): string {
  return `¥${value.toLocaleString("ja-JP")}`;
}

function formatDurationLabel(totalSeconds: number): string {
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
  }

  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

function buildPreviewShortAriaLabel(item: CreatorWorkspacePreviewShortItem, index: number): string {
  return `ショート詳細を開く ${index + 1}件目 ${formatDurationLabel(item.previewDurationSeconds)}`;
}

function buildPreviewMainAriaLabel(item: CreatorWorkspacePreviewMainItem, index: number): string {
  return `本編詳細を開く ${index + 1}件目 ${formatJpy(item.priceJpy)} ${formatDurationLabel(item.durationSeconds)}`;
}

function buildRevisionRequestedDetail({
  mainCount,
  shortCount,
}: {
  mainCount: number;
  shortCount: number;
}): string {
  const scopes = [
    shortCount > 0 ? `short ${formatCount(shortCount)}件` : null,
    mainCount > 0 ? `main ${formatCount(mainCount)}件` : null,
  ].filter((scope) => scope !== null);

  if (scopes.length === 0) {
    return "修正依頼内容を確認してください";
  }

  return `${scopes.join(" / ")}を確認してください`;
}

function CreatorModeBlockedFrame({ children }: { children: ReactNode }) {
  return (
    <main className="min-h-svh bg-[radial-gradient(circle_at_top,rgba(214,242,247,0.82),transparent_34%),linear-gradient(180deg,#f7fcfd_0%,#eef7f8_42%,#e8eff6_100%)] text-foreground">
      {children}
    </main>
  );
}

function CreatorModeWorkspaceFrame({ children }: { children: ReactNode }) {
  return (
    <main className="bg-background">
      <div className="mx-auto min-h-svh w-full max-w-[408px] overflow-hidden bg-white text-foreground">
        {children}
      </div>
    </main>
  );
}

function CreatorShellBlockedState({ state }: { state: Exclude<CreatorModeShellState, { kind: "ready" }> }) {
  return (
    <CreatorModeBlockedFrame>
      <div className="mx-auto flex min-h-svh max-w-3xl items-center px-4 py-12 sm:px-6">
        <SurfacePanel className="w-full px-6 py-7 sm:px-8 sm:py-9">
          <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[#0f6172]">{state.eyebrow}</p>
          <h1 className="mt-4 font-display text-[30px] font-semibold tracking-[-0.05em] text-foreground">
            {state.title}
          </h1>
          <p className="mt-3 text-sm leading-7 text-muted">{state.description}</p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Button asChild>
              <Link href={state.ctaHref}>{state.ctaLabel}</Link>
            </Button>
          </div>
        </SurfacePanel>
      </div>
    </CreatorModeBlockedFrame>
  );
}

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

function AccountMenuIcon() {
  return (
    <svg
      aria-hidden="true"
      className="size-5 fill-none stroke-current [stroke-linecap:round] [stroke-linejoin:round] [stroke-width:1.7]"
      viewBox="0 0 20 20"
    >
      <line x1="10" x2="10" y1="1.8" y2="4.1" />
      <line x1="10" x2="10" y1="15.9" y2="18.2" />
      <line x1="1.8" x2="4.1" y1="10" y2="10" />
      <line x1="15.9" x2="18.2" y1="10" y2="10" />
      <line x1="4.2" x2="5.9" y1="4.2" y2="5.9" />
      <line x1="14.1" x2="15.8" y1="14.1" y2="15.8" />
      <line x1="14.1" x2="15.8" y1="5.9" y2="4.2" />
      <line x1="4.2" x2="5.9" y1="15.8" y2="14.1" />
      <circle cx="10" cy="10" r="3.1" />
    </svg>
  );
}

function CreatorWorkspaceAccountMenu() {
  const {
    clearError,
    enterFanMode,
    errorMessage,
    isSubmitting,
  } = useFanModeEntry();

  return (
    <Dialog.Root>
      <Dialog.Trigger asChild>
        <button
          aria-label="Account menu"
          className="inline-flex size-[34px] items-center justify-center bg-transparent text-[#1082c8] transition hover:bg-[#1082c8]/10"
          onClick={clearError}
          type="button"
        >
          <AccountMenuIcon />
        </button>
      </Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-y-0 left-1/2 z-40 w-full max-w-[408px] -translate-x-1/2 bg-[rgba(77,132,166,0.22)] backdrop-blur-[8px]" />
        <Dialog.Content className="fixed bottom-3 left-1/2 z-50 w-[calc(100vw-24px)] max-w-[384px] -translate-x-1/2 rounded-[28px] border border-[rgba(217,226,232,0.94)] bg-[rgba(255,255,255,0.98)] p-[10px_10px_14px] shadow-[0_18px_42px_rgba(6,21,33,0.12)]">
          <Dialog.Title className="sr-only">アカウントメニュー</Dialog.Title>
          <Dialog.Description className="sr-only">
            creator workspace から fan mode へ戻るメニュー
          </Dialog.Description>

          <div
            aria-hidden="true"
            className="mx-auto mb-3 h-1 w-10 rounded-full bg-[rgba(6,21,33,0.16)]"
          />

          <div className="rounded-[24px] bg-[#f3f6f8] py-1">
            <Dialog.Close asChild>
              <Link
                className="flex min-h-[54px] w-full items-center justify-between px-[18px] text-left text-sm font-bold text-foreground transition hover:bg-white/65"
                href="/creator/settings/profile"
              >
                <span>プロフィールを編集</span>
                <ChevronRight aria-hidden="true" className="size-4 text-muted" strokeWidth={2.2} />
              </Link>
            </Dialog.Close>
            <button
              className="flex min-h-[54px] w-full items-center justify-between border-t border-[rgba(167,220,249,0.24)] px-[18px] text-left text-sm font-bold text-foreground transition hover:bg-white/65"
              disabled={isSubmitting}
              onClick={() => {
                void enterFanMode();
              }}
              type="button"
            >
              <span>{isSubmitting ? "Fan mode に切り替えています..." : "Fan mode に切り替え"}</span>
              <ChevronRight aria-hidden="true" className="size-4 text-muted" strokeWidth={2.2} />
            </button>
          </div>

          {errorMessage ? (
            <p
              aria-live="polite"
              className="mt-3 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
              role="alert"
            >
              {errorMessage}
            </p>
          ) : null}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
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

function CreatorWorkspaceMetricStrip({
  overviewMetrics,
}: {
  overviewMetrics: ApprovedCreatorWorkspaceOverviewMetrics;
}) {
  const metrics = [
    {
      label: "revenue",
      value: formatJpy(overviewMetrics.grossUnlockRevenueJpy),
    },
    {
      label: "unlocks",
      value: formatCount(overviewMetrics.unlockCount),
    },
    {
      label: "purchasers",
      value: formatCount(overviewMetrics.uniquePurchaserCount),
    },
  ] as const;

  return (
    <div className="grid flex-1 grid-cols-3 gap-x-2 gap-y-2 text-center">
      {metrics.map((item) => (
        <div key={item.label} className="min-w-0">
          <strong className="block text-base font-bold leading-[1.2] tracking-[-0.03em] text-foreground">
            {item.value}
          </strong>
          <span className="mt-1 block text-[11px] text-muted">{item.label}</span>
        </div>
      ))}
    </div>
  );
}

function CreatorWorkspaceHeader({
  creator,
  overviewMetrics,
}: {
  creator: CreatorSummary;
  overviewMetrics: ApprovedCreatorWorkspaceOverviewMetrics;
}) {
  return (
    <section className="mt-[18px] text-foreground">
      <div className="flex items-center gap-[18px]">
        <CreatorAvatar
          className="size-[86px] rounded-full border-white/70 shadow-[0_10px_24px_rgba(36,92,129,0.16)]"
          creator={creator}
        />
        <CreatorWorkspaceMetricStrip overviewMetrics={overviewMetrics} />
      </div>

      <div className="mt-[14px]">
        <p className="m-0 text-[15px] font-bold text-foreground">{creator.handle}</p>
        <p className="mt-1 text-[13px] font-bold text-foreground">{creator.displayName}</p>
        {creator.bio.trim().length > 0 ? (
          <p className="mt-[6px] text-[13px] leading-[1.55] text-muted">{creator.bio}</p>
        ) : null}
      </div>
    </section>
  );
}

function CreatorWorkspaceRevisionNotice({
  revisionRequestedSummary,
}: {
  revisionRequestedSummary: ApprovedCreatorWorkspaceRevisionRequestedSummary | null;
}) {
  const summary = revisionRequestedSummary;
  if (summary === null) {
    return null;
  }

  return (
    <div className="mt-[14px] flex items-center gap-3 rounded-[18px] border border-[rgba(244,152,45,0.18)] bg-[linear-gradient(180deg,rgba(255,248,238,0.96),rgba(252,242,224,0.92))] px-[14px] py-3 text-foreground">
      <span className="inline-flex min-h-7 items-center justify-center rounded-full bg-[rgba(244,152,45,0.14)] px-3 text-[10px] font-bold uppercase tracking-[0.14em] text-[#8e4e0a]">
        差し戻し
      </span>
      <div className="grid gap-1">
        <b className="text-[13px] leading-[1.35] text-foreground">差し戻しが{formatCount(summary.totalCount)}件あります</b>
        <span className="text-[11px] text-muted">
          {buildRevisionRequestedDetail({
            mainCount: summary.mainCount,
            shortCount: summary.shortCount,
          })}
        </span>
      </div>
    </div>
  );
}

function CreatorWorkspaceSummaryLoading() {
  return (
    <section className="mt-[18px] grid gap-[14px] text-foreground">
      <p className="sr-only" role="status">
        workspace summary を読み込んでいます...
      </p>

      <div className="flex items-center gap-[18px]">
        <div
          aria-hidden="true"
          className="size-[86px] animate-pulse rounded-full bg-[rgba(167,220,249,0.34)]"
        />
        <div className="grid flex-1 grid-cols-3 gap-x-2 gap-y-2">
          {Array.from({ length: 3 }).map((_, index) => (
            <div className="grid gap-2 text-center" key={index}>
              <div aria-hidden="true" className="mx-auto h-5 w-14 animate-pulse rounded-full bg-[rgba(167,220,249,0.34)]" />
              <div aria-hidden="true" className="mx-auto h-3 w-12 animate-pulse rounded-full bg-[rgba(167,220,249,0.24)]" />
            </div>
          ))}
        </div>
      </div>

      <div className="grid gap-2">
        <div aria-hidden="true" className="h-4 w-24 animate-pulse rounded-full bg-[rgba(167,220,249,0.32)]" />
        <div aria-hidden="true" className="h-4 w-36 animate-pulse rounded-full bg-[rgba(167,220,249,0.28)]" />
        <div aria-hidden="true" className="h-4 w-full animate-pulse rounded-full bg-[rgba(167,220,249,0.22)]" />
      </div>
    </section>
  );
}

function CreatorWorkspaceSummaryError({
  message,
  onRetry,
}: {
  message: string;
  onRetry: () => void;
}) {
  return (
    <section className="mt-[18px]">
      <div
        className="rounded-[22px] border border-[rgba(255,184,189,0.84)] bg-[linear-gradient(180deg,rgba(255,247,248,0.98),rgba(255,241,243,0.96))] px-4 py-4 text-foreground"
        role="alert"
      >
        <p className="m-0 text-[13px] leading-6 text-muted">{message}</p>
        <button
          className="mt-3 inline-flex min-h-10 items-center rounded-[12px] bg-[#1082c8] px-4 text-[13px] font-bold text-white transition hover:opacity-90"
          onClick={onRetry}
          type="button"
        >
          再読み込み
        </button>
      </div>
    </section>
  );
}

function CreatorWorkspaceSummarySection({
  onRetry,
  state,
}: {
  onRetry: () => void;
  state: CreatorWorkspaceSummaryState;
}) {
  if (state.kind === "loading") {
    return <CreatorWorkspaceSummaryLoading />;
  }

  if (state.kind === "error") {
    return <CreatorWorkspaceSummaryError message={state.message} onRetry={onRetry} />;
  }

  return (
    <>
      <CreatorWorkspaceHeader
        creator={state.summary.creator}
        overviewMetrics={state.summary.overviewMetrics}
      />
      <CreatorWorkspaceRevisionNotice revisionRequestedSummary={state.summary.revisionRequestedSummary} />
    </>
  );
}

function createPosterStyle(poster: ApprovedCreatorWorkspacePoster): CSSProperties {
  return {
    "--creator-workspace-tile-bottom": poster.tile.bottom,
    "--creator-workspace-tile-mid": poster.tile.mid,
    "--creator-workspace-tile-top": poster.tile.top,
  } as CSSProperties;
}

function CreatorWorkspacePosterThumb({ poster }: { poster: ApprovedCreatorWorkspacePoster }) {
  return (
    <span
      aria-hidden="true"
      className="block h-10 w-[30px] shrink-0 rounded-[8px] bg-[linear-gradient(180deg,var(--creator-workspace-tile-top),var(--creator-workspace-tile-mid)_42%,var(--creator-workspace-tile-bottom)_100%)] shadow-[inset_0_0_0_1px_rgba(255,255,255,0.56)]"
      style={createPosterStyle(poster)}
    />
  );
}

function CreatorWorkspaceTopPerformers({
  onOpenDetail,
  workspace,
}: {
  onOpenDetail: (selection: CreatorWorkspaceDetailSelection) => void;
  workspace: ApprovedCreatorWorkspaceState;
}) {
  if (workspace.topPerformers.length === 0) {
    return null;
  }

  return (
    <section aria-label="Top performers" className="mt-[18px] border-y border-[rgba(167,220,249,0.48)]">
      {workspace.topPerformers.map((item, index) => {
        const poster = workspace.posters[item.shortId];

        return (
          <button
            aria-label={item.label}
            className={`flex min-h-[58px] w-full items-center justify-between gap-[14px] bg-transparent px-0 text-left text-foreground disabled:cursor-default disabled:opacity-100 ${
              index === 0 ? "" : "border-t border-[rgba(167,220,249,0.32)]"
            }`}
            key={`${item.kind}:${item.shortId}`}
            onClick={() => {
              onOpenDetail({ kind: "mock", shortId: item.shortId, tab: item.kind });
            }}
            type="button"
          >
            <span className="text-[11px] font-bold uppercase tracking-[0.1em] text-muted">{item.label}</span>
            <span className="flex min-w-0 items-center gap-3">
              <span className="text-sm font-bold leading-[1.3] text-foreground">{item.metric}</span>
              {poster ? <CreatorWorkspacePosterThumb poster={poster} /> : null}
            </span>
          </button>
        );
      })}
    </section>
  );
}

function getManagedTileFrameClassName(tone: ApprovedCreatorWorkspaceManagedItemTone): string {
  if (tone === "approved") {
    return "";
  }

  return "brightness-[0.72] saturate-[0.82]";
}

function getManagedTileOverlayClassName(tone: ApprovedCreatorWorkspaceManagedItemTone): string {
  if (tone === "approved") {
    return "bg-[linear-gradient(180deg,rgba(6,21,33,0.04)_0%,rgba(6,21,33,0.02)_38%,rgba(6,21,33,0.48)_100%)]";
  }

  return "bg-[linear-gradient(180deg,rgba(6,21,33,0.42)_0%,rgba(6,21,33,0.18)_34%,rgba(6,21,33,0.72)_100%)]";
}

function getManagedTileStatusClassName(tone: ApprovedCreatorWorkspaceManagedItemTone): string {
  switch (tone) {
    case "hidden":
      return "bg-[rgba(7,19,29,0.18)] text-[#f6fbff]";
    case "paused":
      return "bg-[rgba(16,130,200,0.18)] text-[#eff8ff]";
    case "pending":
      return "bg-[rgba(16,130,200,0.18)] text-[#eff8ff]";
    case "removed":
      return "bg-[rgba(217,77,77,0.2)] text-[#fff7f7]";
    case "revision":
      return "bg-[rgba(244,152,45,0.18)] text-[#fff7ea]";
    case "approved":
      return "bg-[rgba(52,168,83,0.12)] text-[#effff2]";
  }
}

function CreatorWorkspaceManagedTile({
  item,
  onOpenDetail,
  poster,
  tab,
}: {
  item: ApprovedCreatorWorkspaceManagedItem;
  onOpenDetail: (selection: CreatorWorkspaceDetailSelection) => void;
  poster: ApprovedCreatorWorkspacePoster;
  tab: ApprovedCreatorWorkspaceManagedTab;
}) {
  return (
    <button
      aria-label={poster.title}
      className="relative overflow-hidden rounded-[4px] text-left transition"
      onClick={() => {
        onOpenDetail({ kind: "mock", shortId: item.shortId, tab });
      }}
      type="button"
    >
      <span
        aria-hidden="true"
        className={`block aspect-[3/4] bg-[linear-gradient(180deg,var(--creator-workspace-tile-top),var(--creator-workspace-tile-mid)_42%,var(--creator-workspace-tile-bottom)_100%)] transition ${getManagedTileFrameClassName(item.tone)}`}
        style={createPosterStyle(poster)}
      />
      <span className={`absolute inset-0 grid place-items-center p-2 ${getManagedTileOverlayClassName(item.tone)}`}>
        {item.tone !== "approved" ? (
          <span
            className={`inline-flex min-h-8 items-center justify-center rounded-full px-4 text-[11px] font-bold uppercase tracking-[0.16em] backdrop-blur-[8px] ${getManagedTileStatusClassName(item.tone)}`}
          >
            {item.status}
          </span>
        ) : null}
      </span>
    </button>
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
        aria-label="Your videos"
        className="mt-[18px] grid grid-cols-2 border-t border-[rgba(167,220,249,0.48)]"
        role="tablist"
      >
        {workspace.managedCollections.tabs.map((tab) => {
          const active = tab.key === activeTab;

          return (
            <button
              aria-label={tab.label}
              aria-selected={active}
              className={`inline-flex min-h-[42px] min-w-0 items-center justify-center border-t-2 pt-[10px] text-xs font-bold uppercase tracking-[0.08em] ${
                active ? "border-t-foreground text-foreground" : "border-t-transparent text-muted"
              }`}
              key={tab.key}
              onClick={() => {
                onChangeTab(tab.key);
              }}
              role="tab"
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

function createVideoPosterStyle(posterUrl: string): CSSProperties {
  return {
    backgroundImage: `url("${posterUrl}")`,
    backgroundPosition: "center",
    backgroundSize: "cover",
  };
}

function CreatorWorkspacePreviewTileFrame({
  badge,
  bottomLeft,
  bottomRight,
  posterUrl,
}: {
  badge: string;
  bottomLeft: string | null;
  bottomRight: string;
  posterUrl: string;
}) {
  return (
    <article
      className="relative overflow-hidden rounded-[4px] bg-[#dbeaf2]"
      data-testid="creator-workspace-preview-tile"
    >
      <span
        aria-hidden="true"
        className="block aspect-[3/4] bg-[#dbeaf2]"
        style={createVideoPosterStyle(posterUrl)}
      />
      <div className="absolute inset-0 flex flex-col justify-between bg-[linear-gradient(180deg,rgba(6,21,33,0.12)_0%,rgba(6,21,33,0.03)_34%,rgba(6,21,33,0.66)_100%)] p-2.5">
        <div className="flex items-start justify-between gap-2">
          <span
            aria-hidden="true"
            className="inline-flex min-h-6 items-center justify-center rounded-full bg-white/16 px-2.5 text-[10px] font-bold uppercase tracking-[0.12em] text-white backdrop-blur-[10px]"
          >
            {badge}
          </span>
          <span className="sr-only">{badge}</span>
        </div>
        <div className="flex items-end justify-between gap-2">
          {bottomLeft ? (
            <span className="inline-flex min-h-6 items-center justify-center rounded-full bg-white/16 px-2.5 text-[10px] font-bold tracking-[0.02em] text-white backdrop-blur-[10px]">
              {bottomLeft}
            </span>
          ) : <span />}
          <span className="inline-flex min-h-6 items-center justify-center rounded-full bg-white/16 px-2.5 text-[10px] font-bold tracking-[0.02em] text-white backdrop-blur-[10px]">
            {bottomRight}
          </span>
        </div>
      </div>
    </article>
  );
}

function CreatorWorkspacePreviewTileButton({
  ariaLabel,
  children,
  onClick,
}: {
  ariaLabel: string;
  children: ReactNode;
  onClick: () => void;
}) {
  return (
    <button
      aria-label={ariaLabel}
      className="overflow-hidden rounded-[4px] text-left transition hover:opacity-95"
      onClick={onClick}
      type="button"
    >
      {children}
    </button>
  );
}

function CreatorWorkspacePreviewShortTile({
  index,
  item,
  onOpenDetail,
}: {
  index: number;
  item: CreatorWorkspacePreviewShortItem;
  onOpenDetail: (selection: CreatorWorkspacePreviewDetailSelection) => void;
}) {
  return (
    <CreatorWorkspacePreviewTileButton
      ariaLabel={buildPreviewShortAriaLabel(item, index)}
      onClick={() => {
        onOpenDetail({
          index,
          item,
          kind: "preview-short",
          tab: "shorts",
        });
      }}
    >
      <CreatorWorkspacePreviewTileFrame
        badge="Short"
        bottomLeft={null}
        bottomRight={formatDurationLabel(item.previewDurationSeconds)}
        posterUrl={item.media.posterUrl}
      />
    </CreatorWorkspacePreviewTileButton>
  );
}

function CreatorWorkspacePreviewMainTile({
  index,
  item,
  onOpenDetail,
}: {
  index: number;
  item: CreatorWorkspacePreviewMainItem;
  onOpenDetail: (selection: CreatorWorkspacePreviewDetailSelection) => void;
}) {
  return (
    <CreatorWorkspacePreviewTileButton
      ariaLabel={buildPreviewMainAriaLabel(item, index)}
      onClick={() => {
        onOpenDetail({
          index,
          item,
          kind: "preview-main",
          tab: "main",
        });
      }}
    >
      <CreatorWorkspacePreviewTileFrame
        badge="Main"
        bottomLeft={formatJpy(item.priceJpy)}
        bottomRight={formatDurationLabel(item.durationSeconds)}
        posterUrl={item.media.posterUrl}
      />
    </CreatorWorkspacePreviewTileButton>
  );
}

function CreatorWorkspacePreviewLoading() {
  return (
    <section className="mt-[18px]">
      <p className="sr-only" role="status">
        workspace video list を読み込んでいます...
      </p>
      <div className="grid grid-cols-3 gap-[3px]">
        {Array.from({ length: 6 }).map((_, index) => (
          <div
            aria-hidden="true"
            className="aspect-[3/4] animate-pulse rounded-[4px] bg-[rgba(167,220,249,0.28)]"
            key={index}
          />
        ))}
      </div>
    </section>
  );
}

function CreatorWorkspacePreviewError({
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

function CreatorWorkspacePreviewEmpty({ activeTabLabel }: { activeTabLabel: string }) {
  return (
    <section className="mt-[18px] rounded-[20px] border border-dashed border-[rgba(167,220,249,0.5)] bg-[#fbfdff] px-4 py-6 text-center text-sm leading-6 text-muted">
      表示できる{activeTabLabel}動画はまだありません。
    </section>
  );
}

function CreatorWorkspacePreviewGrid({
  activeTab,
  activeTabLabel,
  onOpenDetail,
  onRetry,
  state,
}: {
  activeTab: ApprovedCreatorWorkspaceManagedTab;
  activeTabLabel: string;
  onOpenDetail: (selection: CreatorWorkspacePreviewDetailSelection) => void;
  onRetry: () => void;
  state: CreatorWorkspacePreviewCollectionsState;
}) {
  if (state.kind === "loading") {
    return <CreatorWorkspacePreviewLoading />;
  }

  if (state.kind === "error") {
    return <CreatorWorkspacePreviewError message={state.message} onRetry={onRetry} />;
  }

  const activeItems = activeTab === "shorts" ? state.collections.shorts.items : state.collections.mains.items;

  if (activeItems.length === 0) {
    return <CreatorWorkspacePreviewEmpty activeTabLabel={activeTabLabel} />;
  }

  return (
    <section className="mt-[18px] grid grid-cols-3 gap-[3px]">
      {activeTab === "shorts"
        ? state.collections.shorts.items.map((item, index) => (
            <CreatorWorkspacePreviewShortTile index={index} item={item} key={item.id} onOpenDetail={onOpenDetail} />
          ))
        : state.collections.mains.items.map((item, index) => (
            <CreatorWorkspacePreviewMainTile index={index} item={item} key={item.id} onOpenDetail={onOpenDetail} />
          ))}
    </section>
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
  const isMockPoster = poster.kind === "mock";
  const hasStatus = detail.statusTone !== null && detail.statusLabel !== null;

  return (
    <div className="relative overflow-hidden rounded-[32px]">
      <span
        aria-hidden="true"
        className={`block aspect-[4/5] bg-[linear-gradient(180deg,var(--creator-workspace-tile-top),var(--creator-workspace-tile-mid)_42%,var(--creator-workspace-tile-bottom)_100%)] ${
          hasStatus && detail.statusTone !== "approved" ? "brightness-[0.72] saturate-[0.82]" : ""
        }`}
        style={isMockPoster ? createPosterStyle(poster.poster) : createVideoPosterStyle(poster.posterUrl)}
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

function buildPreviewShortDetailSettings(item: CreatorWorkspacePreviewShortItem) {
  return [
    { label: "長さ", value: formatDurationLabel(item.previewDurationSeconds) },
  ] as const;
}

function buildPreviewMainDetailSettings(item: CreatorWorkspacePreviewMainItem) {
  return [
    { label: "価格", value: formatJpy(item.priceJpy) },
    { label: "長さ", value: formatDurationLabel(item.durationSeconds) },
  ] as const;
}

function CreatorWorkspacePreviewDetailLinkedGrid({
  items,
  onOpenDetail,
}: {
  items: readonly (CreatorWorkspacePreviewMainItem | CreatorWorkspacePreviewShortItem)[];
  onOpenDetail: (selection: CreatorWorkspacePreviewDetailSelection) => void;
}) {
  return (
    <div className="grid grid-cols-3 gap-[3px]">
      {items.map((item, index) => (
        "priceJpy" in item ? (
          <CreatorWorkspacePreviewMainTile index={index} item={item} key={item.id} onOpenDetail={onOpenDetail} />
        ) : (
          <CreatorWorkspacePreviewShortTile index={index} item={item} key={item.id} onOpenDetail={onOpenDetail} />
        )
      ))}
    </div>
  );
}

function CreatorWorkspaceDetailView({
  creator,
  detailSelection,
  onBack,
  onOpenDetail,
  previewCollections,
  state,
}: {
  creator: CreatorSummary;
  detailSelection: CreatorWorkspaceDetailSelection | CreatorWorkspacePreviewDetailSelection;
  onBack: () => void;
  onOpenDetail: (selection: CreatorWorkspaceDetailSelection | CreatorWorkspacePreviewDetailSelection) => void;
  previewCollections: NonNullable<Extract<CreatorWorkspacePreviewCollectionsState, { kind: "ready" }>["collections"]> | null;
  state: Extract<CreatorModeShellState, { kind: "ready" }>;
}) {
  let detail: CreatorWorkspaceResolvedDetailState | null = null;
  let poster: CreatorWorkspaceDetailPoster | null = null;
  let linkedPreviewItems: readonly (CreatorWorkspacePreviewMainItem | CreatorWorkspacePreviewShortItem)[] = [];

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

    detail = detailSelection.kind === "preview-main"
      ? {
          durationLabel: formatDurationLabel(detailSelection.item.durationSeconds),
          kindLabel: "本編",
          linkedMainShortId: null,
          linkedShortIds: [],
          metrics: [],
          settings: buildPreviewMainDetailSettings(detailSelection.item),
          statusLabel: null,
          statusTone: null,
          summary: "owner preview 一覧から取得した本編データです。",
        }
      : {
          durationLabel: formatDurationLabel(detailSelection.item.previewDurationSeconds),
          kindLabel: "ショート",
          linkedMainShortId: null,
          linkedShortIds: [],
          metrics: [],
          settings: buildPreviewShortDetailSettings(detailSelection.item),
          statusLabel: null,
          statusTone: null,
          summary: "owner preview 一覧から取得したショートデータです。",
        };
    poster = {
      kind: "preview",
      posterUrl: detailSelection.item.media.posterUrl,
    };
    linkedPreviewItems = detailSelection.kind === "preview-main"
      ? previewCollections.shorts.items.filter((item) => item.canonicalMainId === detailSelection.item.id)
      : previewCollections.mains.items.filter((item) => item.id === detailSelection.item.canonicalMainId);
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

function CreatorWorkspaceDashboard({
  activeTab,
  creator,
  onChangeTab,
  onOpenDetail,
  onOpenPreviewDetail,
  onRetryPreviewCollections,
  onRetrySummary,
  previewCollectionsState,
  summaryState,
  state,
}: {
  activeTab: ApprovedCreatorWorkspaceManagedTab;
  creator: CreatorSummary;
  onChangeTab: (tab: ApprovedCreatorWorkspaceManagedTab) => void;
  onOpenDetail: (selection: CreatorWorkspaceDetailSelection) => void;
  onOpenPreviewDetail: (selection: CreatorWorkspacePreviewDetailSelection) => void;
  onRetryPreviewCollections: () => void;
  onRetrySummary: () => void;
  previewCollectionsState: CreatorWorkspacePreviewCollectionsState;
  summaryState: CreatorWorkspaceSummaryState;
  state: Extract<CreatorModeShellState, { kind: "ready" }>;
}) {
  return (
    <section className="relative z-[2] min-h-svh overflow-y-auto px-4 pb-24 pt-[14px] text-foreground">
      <h1 className="sr-only">{creator.displayName} creator workspace</h1>
      <CreatorWorkspaceTopBar />
      <CreatorWorkspaceSummarySection onRetry={onRetrySummary} state={summaryState} />
      <CreatorWorkspaceTopPerformers onOpenDetail={onOpenDetail} workspace={state.workspace} />
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

function CreatorWorkspaceReadyState({ state }: { state: Extract<CreatorModeShellState, { kind: "ready" }> }) {
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
  const [activeTab, setActiveTab] = useState<ApprovedCreatorWorkspaceManagedTab>(state.workspace.managedCollections.defaultTab);
  const [detailSelection, setDetailSelection] = useState<CreatorWorkspaceDetailSelection | CreatorWorkspacePreviewDetailSelection | null>(null);
  const creator = summaryState.kind === "ready" ? summaryState.summary.creator : state.creator;
  const blockedState = summaryBlockedState ?? previewBlockedState;

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
          onOpenDetail={(selection) => {
            setActiveTab(selection.tab);
            setDetailSelection(selection);
          }}
          previewCollections={previewCollectionsState.kind === "ready" ? previewCollectionsState.collections : null}
          state={state}
        />
      ) : (
        <CreatorWorkspaceDashboard
          activeTab={activeTab}
          creator={creator}
          onChangeTab={setActiveTab}
          onOpenDetail={(selection) => {
            setActiveTab(selection.tab);
            setDetailSelection(selection);
          }}
          onOpenPreviewDetail={(selection) => {
            setActiveTab(selection.tab);
            setDetailSelection(selection);
          }}
          onRetryPreviewCollections={retryPreviewCollections}
          onRetrySummary={retrySummary}
          previewCollectionsState={previewCollectionsState}
          summaryState={summaryState}
          state={state}
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
