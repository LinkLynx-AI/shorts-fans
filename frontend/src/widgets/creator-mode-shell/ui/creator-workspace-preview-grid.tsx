"use client";

import { X } from "lucide-react";
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
} from "../lib/creator-mode-shell-ui";
import type { CreatorWorkspacePreviewDetailSelection } from "./creator-mode-shell.types";

const previewGridClassName = "grid grid-cols-2 gap-3";

function CreatorWorkspacePreviewTileFrame({
  durationLabel,
  posterUrl,
  priceLabel,
  statusBadge,
}: {
  durationLabel: string;
  posterUrl: string;
  priceLabel: string | null;
  statusBadge: CreatorWorkspaceReviewBadge | null;
}) {
  const isRejected = statusBadge?.tone === "removed";

  return (
    <span className="grid gap-2" data-testid="creator-workspace-preview-tile">
      <span className="relative block overflow-hidden rounded-[18px] border border-[rgba(7,19,29,0.06)] bg-[#eef4f7] shadow-[0_8px_22px_rgba(36,92,129,0.08)]">
        <span
          aria-hidden="true"
          className={`block aspect-[3/4] bg-[#dbeaf2] transition duration-500 group-hover:scale-[1.025] ${
            isRejected ? "brightness-[0.62] saturate-[0.74]" : ""
          }`}
          style={createVideoPosterStyle(posterUrl)}
        />
        {isRejected ? (
          <span
            aria-hidden="true"
            className="absolute inset-0 grid place-items-center bg-black/10"
            data-testid="creator-workspace-rejected-marker"
          >
            <span className="inline-flex size-12 items-center justify-center rounded-full bg-black/40 text-white shadow-[0_10px_24px_rgba(0,0,0,0.24)] backdrop-blur-[8px]">
              <X className="size-7" strokeWidth={2.3} />
            </span>
          </span>
        ) : null}
        <span className="absolute bottom-2 right-2 inline-flex min-h-6 items-center justify-center rounded-[8px] bg-black/60 px-2 text-[11px] font-bold tracking-[0.02em] text-white backdrop-blur-[8px]">
          {durationLabel}
        </span>
      </span>

      {priceLabel ? (
        <span className="flex min-h-6 items-center px-0.5">
          <span className="truncate text-[13px] font-bold text-foreground">{priceLabel}</span>
        </span>
      ) : null}
    </span>
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
      className="group min-w-0 rounded-[18px] text-left transition hover:opacity-95 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-[#1082c8]/20"
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
        durationLabel={formatDurationLabel(item.previewDurationSeconds)}
        posterUrl={item.media.posterUrl}
        priceLabel={null}
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
        durationLabel={formatDurationLabel(item.durationSeconds)}
        posterUrl={item.media.posterUrl}
        priceLabel={formatJpy(item.priceJpy)}
        statusBadge={statusBadge}
      />
    </CreatorWorkspacePreviewTileButton>
  );
}

function CreatorWorkspacePreviewLoadingTile() {
  return (
    <div className="grid gap-2">
      <div
        aria-hidden="true"
        className="aspect-[3/4] animate-pulse rounded-[18px] bg-[rgba(167,220,249,0.28)]"
      />
      <span className="h-5 w-20 animate-pulse rounded-[8px] bg-[rgba(167,220,249,0.22)]" />
    </div>
  );
}

function CreatorWorkspacePreviewLoading() {
  return (
    <section className="mt-[18px]">
      <p className="sr-only" role="status">
        workspace video list を読み込んでいます...
      </p>
      <div className={previewGridClassName}>
        {Array.from({ length: 4 }).map((_, index) => (
          <CreatorWorkspacePreviewLoadingTile key={index} />
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
    <div className={previewGridClassName}>
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
    <section className={`mt-[18px] ${previewGridClassName}`}>
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
