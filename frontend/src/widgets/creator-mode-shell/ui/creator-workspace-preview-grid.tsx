"use client";

import type { ReactNode } from "react";

import { Button } from "@/shared/ui";

import type {
  CreatorWorkspacePreviewMainItem,
  CreatorWorkspacePreviewShortItem,
} from "../api/get-creator-workspace-preview-collections";
import type {
  ApprovedCreatorWorkspaceManagedTab,
} from "../model/approved-creator-workspace";
import type { CreatorWorkspacePreviewCollectionsState } from "../model/creator-workspace-preview-collections";
import {
  resolveCreatorWorkspaceObjectReviewBadge,
  type CreatorWorkspaceReviewBadge,
  type CreatorWorkspaceReviewSurfaceState,
} from "../model/creator-workspace-review-surface";
import {
  buildPreviewMainAriaLabel,
  buildPreviewShortAriaLabel,
  createVideoPosterStyle,
  formatDurationLabel,
  formatJpy,
  getManagedTileStatusClassName,
} from "../lib/creator-mode-shell-ui";
import type { CreatorWorkspacePreviewDetailSelection } from "./creator-mode-shell.types";

function CreatorWorkspacePreviewTileFrame({
  badge,
  bottomLeft,
  bottomRight,
  posterUrl,
  statusBadge,
}: {
  badge: string;
  bottomLeft: string | null;
  bottomRight: string;
  posterUrl: string;
  statusBadge: CreatorWorkspaceReviewBadge | null;
}) {
  return (
    <article
      className="relative overflow-hidden rounded-[4px] bg-[#dbeaf2]"
      data-testid="creator-workspace-preview-tile"
    >
      <span
        aria-hidden="true"
        className={`block aspect-[3/4] bg-[#dbeaf2] ${statusBadge && statusBadge.tone !== "approved" ? "brightness-[0.76] saturate-[0.84]" : ""}`}
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
          {statusBadge ? (
            <span
              className={`inline-flex min-h-6 items-center justify-center rounded-full px-2.5 text-[10px] font-bold tracking-[0.02em] backdrop-blur-[10px] ${getManagedTileStatusClassName(statusBadge.tone)}`}
            >
              {statusBadge.label}
            </span>
          ) : null}
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
  statusBadge,
}: {
  index: number;
  item: CreatorWorkspacePreviewShortItem;
  onOpenDetail: (selection: CreatorWorkspacePreviewDetailSelection) => void;
  statusBadge: CreatorWorkspaceReviewBadge | null;
}) {
  return (
    <CreatorWorkspacePreviewTileButton
      ariaLabel={buildPreviewShortAriaLabel(item, index, statusBadge?.label)}
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
        statusBadge={statusBadge}
      />
    </CreatorWorkspacePreviewTileButton>
  );
}

function CreatorWorkspacePreviewMainTile({
  index,
  item,
  onOpenDetail,
  statusBadge,
}: {
  index: number;
  item: CreatorWorkspacePreviewMainItem;
  onOpenDetail: (selection: CreatorWorkspacePreviewDetailSelection) => void;
  statusBadge: CreatorWorkspaceReviewBadge | null;
}) {
  return (
    <CreatorWorkspacePreviewTileButton
      ariaLabel={buildPreviewMainAriaLabel(item, index, statusBadge?.label)}
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
        statusBadge={statusBadge}
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

export function CreatorWorkspacePreviewDetailLinkedGrid({
  items,
  onOpenDetail,
  reviewSurfaceState,
}: {
  items: readonly (CreatorWorkspacePreviewMainItem | CreatorWorkspacePreviewShortItem)[];
  onOpenDetail: (selection: CreatorWorkspacePreviewDetailSelection) => void;
  reviewSurfaceState?: CreatorWorkspaceReviewSurfaceState;
}) {
  const mainReviewBadges = reviewSurfaceState?.kind === "ready"
    ? new Map(reviewSurfaceState.surface.mains.map((item) => [item.id, resolveCreatorWorkspaceObjectReviewBadge(item.state)]))
    : null;
  const shortReviewBadges = reviewSurfaceState?.kind === "ready"
    ? new Map(reviewSurfaceState.surface.shorts.map((item) => [item.id, resolveCreatorWorkspaceObjectReviewBadge(item.state)]))
    : null;

  return (
    <div className="grid grid-cols-3 gap-[3px]">
      {items.map((item, index) => (
        "priceJpy" in item ? (
          <CreatorWorkspacePreviewMainTile
            index={index}
            item={item}
            key={item.id}
            onOpenDetail={onOpenDetail}
            statusBadge={mainReviewBadges?.get(item.id) ?? null}
          />
        ) : (
          <CreatorWorkspacePreviewShortTile
            index={index}
            item={item}
            key={item.id}
            onOpenDetail={onOpenDetail}
            statusBadge={shortReviewBadges?.get(item.id) ?? null}
          />
        )
      ))}
    </div>
  );
}

export function CreatorWorkspacePreviewGrid({
  activeTab,
  activeTabLabel,
  onOpenDetail,
  onRetry,
  reviewSurfaceState,
  state,
}: {
  activeTab: ApprovedCreatorWorkspaceManagedTab;
  activeTabLabel: string;
  onOpenDetail: (selection: CreatorWorkspacePreviewDetailSelection) => void;
  onRetry: () => void;
  reviewSurfaceState: CreatorWorkspaceReviewSurfaceState;
  state: CreatorWorkspacePreviewCollectionsState;
}) {
  if (state.kind === "loading") {
    return <CreatorWorkspacePreviewLoading />;
  }

  if (state.kind === "error") {
    return <CreatorWorkspacePreviewError message={state.message} onRetry={onRetry} />;
  }

  const mainReviewBadges = reviewSurfaceState.kind === "ready"
    ? new Map(reviewSurfaceState.surface.mains.map((item) => [item.id, resolveCreatorWorkspaceObjectReviewBadge(item.state)]))
    : null;
  const shortReviewBadges = reviewSurfaceState.kind === "ready"
    ? new Map(reviewSurfaceState.surface.shorts.map((item) => [item.id, resolveCreatorWorkspaceObjectReviewBadge(item.state)]))
    : null;
  const activeItems = activeTab === "shorts" ? state.collections.shorts.items : state.collections.mains.items;

  if (activeItems.length === 0) {
    return <CreatorWorkspacePreviewEmpty activeTabLabel={activeTabLabel} />;
  }

  return (
    <section className="mt-[18px] grid grid-cols-3 gap-[3px]">
      {activeTab === "shorts"
        ? state.collections.shorts.items.map((item, index) => (
            <CreatorWorkspacePreviewShortTile
              index={index}
              item={item}
              key={item.id}
              onOpenDetail={onOpenDetail}
              statusBadge={shortReviewBadges?.get(item.id) ?? null}
            />
          ))
        : state.collections.mains.items.map((item, index) => (
            <CreatorWorkspacePreviewMainTile
              index={index}
              item={item}
              key={item.id}
              onOpenDetail={onOpenDetail}
              statusBadge={mainReviewBadges?.get(item.id) ?? null}
            />
          ))}
    </section>
  );
}
