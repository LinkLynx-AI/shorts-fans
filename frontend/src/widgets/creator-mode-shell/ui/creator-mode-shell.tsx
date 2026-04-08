"use client";

import Link from "next/link";
import {
  useState,
  type CSSProperties,
  type ReactNode,
} from "react";
import { ArrowLeft } from "lucide-react";

import { CreatorAvatar } from "@/entities/creator";
import { Button, SurfacePanel } from "@/shared/ui";

import type {
  ApprovedCreatorWorkspaceDetailState,
  ApprovedCreatorWorkspaceManagedItem,
  ApprovedCreatorWorkspaceManagedItemTone,
  ApprovedCreatorWorkspaceManagedTab,
  ApprovedCreatorWorkspacePoster,
  ApprovedCreatorWorkspaceState,
} from "../model/approved-creator-workspace";
import type { CreatorModeShellState } from "../model/creator-mode-shell";

type CreatorWorkspaceDetailSelection = {
  shortId: string;
  tab: ApprovedCreatorWorkspaceManagedTab;
};

function formatCount(value: number): string {
  return value.toLocaleString("ja-JP");
}

function formatJpy(value: number): string {
  return `¥${value.toLocaleString("ja-JP")}`;
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

function CreatorWorkspaceTopBar() {
  return (
    <div className="flex items-center justify-between gap-2.5">
      <CreatorWorkspaceActionButton
        ariaLabel="動画を追加"
        className="inline-flex size-7 items-center justify-center bg-transparent text-[#1082c8] disabled:cursor-default disabled:opacity-100"
      >
        <span aria-hidden="true" className="text-[34px] font-extralight leading-none">
          +
        </span>
      </CreatorWorkspaceActionButton>
      <CreatorWorkspaceActionButton
        ariaLabel="Account menu"
        className="inline-flex size-[34px] items-center justify-center bg-transparent text-[#1082c8] disabled:cursor-default disabled:opacity-100"
      >
        <AccountMenuIcon />
      </CreatorWorkspaceActionButton>
    </div>
  );
}

function CreatorWorkspaceMetricStrip({ workspace }: { workspace: ApprovedCreatorWorkspaceState }) {
  const metrics = [
    {
      label: "revenue",
      value: formatJpy(workspace.overviewMetrics.grossUnlockRevenueJpy),
    },
    {
      label: "unlocks",
      value: formatCount(workspace.overviewMetrics.unlockCount),
    },
    {
      label: "purchasers",
      value: formatCount(workspace.overviewMetrics.uniquePurchaserCount),
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

function CreatorWorkspaceHeader({ state }: { state: Extract<CreatorModeShellState, { kind: "ready" }> }) {
  return (
    <section className="mt-[18px] text-foreground">
      <div className="flex items-center gap-[18px]">
        <CreatorAvatar
          className="size-[86px] rounded-full border-white/70 shadow-[0_10px_24px_rgba(36,92,129,0.16)]"
          creator={state.creator}
        />
        <CreatorWorkspaceMetricStrip workspace={state.workspace} />
      </div>

      <div className="mt-[14px]">
        <p className="m-0 text-[15px] font-bold text-foreground">{state.creator.handle}</p>
        <p className="mt-1 text-[13px] font-bold text-foreground">{state.creator.displayName}</p>
        {state.creator.bio.trim().length > 0 ? (
          <p className="mt-[6px] text-[13px] leading-[1.55] text-muted">{state.creator.bio}</p>
        ) : null}
      </div>
    </section>
  );
}

function CreatorWorkspaceRevisionNotice({ workspace }: { workspace: ApprovedCreatorWorkspaceState }) {
  const summary = workspace.revisionRequestedSummary;

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
              onOpenDetail({ shortId: item.shortId, tab: item.kind });
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
        onOpenDetail({ shortId: item.shortId, tab });
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
  onOpenDetail,
  workspace,
}: {
  activeTab: ApprovedCreatorWorkspaceManagedTab;
  onChangeTab: (tab: ApprovedCreatorWorkspaceManagedTab) => void;
  onOpenDetail: (selection: CreatorWorkspaceDetailSelection) => void;
  workspace: ApprovedCreatorWorkspaceState;
}) {
  const activeItems = workspace.managedCollections.itemsByTab[activeTab];

  return (
    <>
      <div
        aria-label="Managed posts"
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

      <div className="mt-[18px] grid grid-cols-3 gap-[3px]">
        {activeItems.flatMap((item) => {
          const poster = workspace.posters[item.shortId];

          return poster
            ? [
                <CreatorWorkspaceManagedTile
                  item={item}
                  key={`${activeTab}:${item.shortId}`}
                  onOpenDetail={onOpenDetail}
                  poster={poster}
                  tab={activeTab}
                />,
              ]
            : [];
        })}
      </div>
    </>
  );
}

function CreatorWorkspaceDetailMedia({
  detail,
  poster,
}: {
  detail: ApprovedCreatorWorkspaceDetailState;
  poster: ApprovedCreatorWorkspacePoster;
}) {
  const pendingLike = detail.statusTone === "pending" || detail.statusTone === "revision" || detail.statusTone === "paused";
  const mutedLike = detail.statusTone === "hidden" || detail.statusTone === "removed";

  return (
    <div className="relative overflow-hidden rounded-[32px]">
      <span
        aria-hidden="true"
        className={`block aspect-[4/5] bg-[linear-gradient(180deg,var(--creator-workspace-tile-top),var(--creator-workspace-tile-mid)_42%,var(--creator-workspace-tile-bottom)_100%)] ${
          detail.statusTone === "approved" ? "" : "brightness-[0.72] saturate-[0.82]"
        }`}
        style={createPosterStyle(poster)}
      />
      <div
        className={`absolute inset-0 flex flex-col justify-between p-4 ${
          detail.statusTone === "approved"
            ? "bg-[linear-gradient(180deg,rgba(6,21,33,0.12)_0%,rgba(6,21,33,0.04)_34%,rgba(6,21,33,0.58)_100%)]"
            : "bg-[linear-gradient(180deg,rgba(6,21,33,0.34)_0%,rgba(6,21,33,0.16)_34%,rgba(6,21,33,0.7)_100%)]"
        }`}
      >
        <div className="flex items-center justify-between gap-2.5">
          <span className="inline-flex min-h-[30px] items-center justify-center rounded-full bg-white/18 px-3 text-[11px] font-bold tracking-[0.08em] text-[#f8fcff] backdrop-blur-[10px]">
            {detail.kindLabel}
          </span>
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

function CreatorWorkspaceDetailMetrics({ detail }: { detail: ApprovedCreatorWorkspaceDetailState }) {
  return (
    <div className="grid grid-cols-2 border-y border-[rgba(167,220,249,0.48)] py-1">
      {detail.metrics.map((metric, index) => (
        <div
          className={`grid gap-1 px-2 py-3 text-center ${detail.metrics.length % 2 === 1 && index === detail.metrics.length - 1 ? "col-span-2" : ""}`}
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

function CreatorWorkspaceDetailSettings({ detail }: { detail: ApprovedCreatorWorkspaceDetailState }) {
  return (
    <div className="grid border-t border-[rgba(167,220,249,0.4)]">
      {detail.settings.map((item) => (
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

function CreatorWorkspaceDetailView({
  detailSelection,
  onBack,
  onOpenDetail,
  state,
}: {
  detailSelection: CreatorWorkspaceDetailSelection;
  onBack: () => void;
  onOpenDetail: (selection: CreatorWorkspaceDetailSelection) => void;
  state: Extract<CreatorModeShellState, { kind: "ready" }>;
}) {
  const detail = state.workspace.detailsByTab[detailSelection.tab][detailSelection.shortId];
  const poster = state.workspace.posters[detailSelection.shortId];

  if (!detail || !poster) {
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
            creator={state.creator}
          />
          <div className="min-w-0">
            <p className="m-0 text-[13px] font-bold text-muted">{state.creator.handle}</p>
          </div>
        </div>

        <CreatorWorkspaceDetailMedia detail={detail} poster={poster} />

        <div className="grid gap-1.5">
          <p className="m-0 text-[15px] leading-[1.6] text-foreground">{detail.summary}</p>
        </div>

        <CreatorWorkspaceDetailMetrics detail={detail} />

        <CreatorWorkspaceDetailSection title="設定">
          <CreatorWorkspaceDetailSettings detail={detail} />
        </CreatorWorkspaceDetailSection>

        {detailSelection.tab === "main" && detail.linkedShortIds.length > 0 ? (
          <CreatorWorkspaceDetailSection title="紐づくショート">
            <CreatorWorkspaceDetailLinkedGrid
              items={detail.linkedShortIds}
              onOpenDetail={onOpenDetail}
              tab="shorts"
              workspace={state.workspace}
            />
          </CreatorWorkspaceDetailSection>
        ) : null}

        {detailSelection.tab === "shorts" && detail.linkedMainShortId ? (
          <CreatorWorkspaceDetailSection title="紐づく本編">
            <CreatorWorkspaceDetailLinkedGrid
              items={[detail.linkedMainShortId]}
              onOpenDetail={onOpenDetail}
              tab="main"
              workspace={state.workspace}
            />
          </CreatorWorkspaceDetailSection>
        ) : null}
      </section>
    </section>
  );
}

function CreatorWorkspaceDashboard({
  activeTab,
  onChangeTab,
  onOpenDetail,
  state,
}: {
  activeTab: ApprovedCreatorWorkspaceManagedTab;
  onChangeTab: (tab: ApprovedCreatorWorkspaceManagedTab) => void;
  onOpenDetail: (selection: CreatorWorkspaceDetailSelection) => void;
  state: Extract<CreatorModeShellState, { kind: "ready" }>;
}) {
  return (
    <section className="relative z-[2] min-h-svh overflow-y-auto px-4 pb-24 pt-[14px] text-foreground">
      <h1 className="sr-only">{state.creator.displayName} creator workspace</h1>
      <CreatorWorkspaceTopBar />
      <CreatorWorkspaceHeader state={state} />
      <CreatorWorkspaceRevisionNotice workspace={state.workspace} />
      <CreatorWorkspaceTopPerformers onOpenDetail={onOpenDetail} workspace={state.workspace} />
      <CreatorWorkspaceManagedPosts
        activeTab={activeTab}
        onChangeTab={onChangeTab}
        onOpenDetail={onOpenDetail}
        workspace={state.workspace}
      />
    </section>
  );
}

function CreatorWorkspaceReadyState({ state }: { state: Extract<CreatorModeShellState, { kind: "ready" }> }) {
  const [activeTab, setActiveTab] = useState<ApprovedCreatorWorkspaceManagedTab>(state.workspace.managedCollections.defaultTab);
  const [detailSelection, setDetailSelection] = useState<CreatorWorkspaceDetailSelection | null>(null);

  return (
    <CreatorModeWorkspaceFrame>
      {detailSelection ? (
        <CreatorWorkspaceDetailView
          detailSelection={detailSelection}
          onBack={() => {
            setDetailSelection(null);
          }}
          onOpenDetail={setDetailSelection}
          state={state}
        />
      ) : (
        <CreatorWorkspaceDashboard
          activeTab={activeTab}
          onChangeTab={setActiveTab}
          onOpenDetail={(selection) => {
            setActiveTab(selection.tab);
            setDetailSelection(selection);
          }}
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
