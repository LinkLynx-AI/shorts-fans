"use client";

import {
  CreatorAvatar,
  type CreatorSummary,
} from "@/entities/creator";

import type {
  ApprovedCreatorWorkspaceOverviewMetrics,
  ApprovedCreatorWorkspaceRevisionRequestedSummary,
} from "../model/approved-creator-workspace";
import type { CreatorWorkspaceSummaryState } from "../model/creator-workspace-summary";
import {
  buildRevisionRequestedDetail,
  formatCount,
  formatJpy,
} from "../lib/creator-mode-shell-ui";

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
  if (revisionRequestedSummary === null) {
    return null;
  }

  return (
    <div className="mt-[14px] flex items-center gap-3 rounded-[18px] border border-[rgba(244,152,45,0.18)] bg-[linear-gradient(180deg,rgba(255,248,238,0.96),rgba(252,242,224,0.92))] px-[14px] py-3 text-foreground">
      <span className="inline-flex min-h-7 items-center justify-center rounded-full bg-[rgba(244,152,45,0.14)] px-3 text-[10px] font-bold uppercase tracking-[0.14em] text-[#8e4e0a]">
        差し戻し
      </span>
      <div className="grid gap-1">
        <b className="text-[13px] leading-[1.35] text-foreground">
          差し戻しが{formatCount(revisionRequestedSummary.totalCount)}件あります
        </b>
        <span className="text-[11px] text-muted">
          {buildRevisionRequestedDetail({
            mainCount: revisionRequestedSummary.mainCount,
            shortCount: revisionRequestedSummary.shortCount,
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
              <div
                aria-hidden="true"
                className="mx-auto h-5 w-14 animate-pulse rounded-full bg-[rgba(167,220,249,0.34)]"
              />
              <div
                aria-hidden="true"
                className="mx-auto h-3 w-12 animate-pulse rounded-full bg-[rgba(167,220,249,0.24)]"
              />
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

export function CreatorWorkspaceSummarySection({
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
