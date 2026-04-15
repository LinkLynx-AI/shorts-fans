"use client";

import Link from "next/link";
import { ArrowLeft, EllipsisVertical } from "lucide-react";

import {
  CreatorAvatar,
  CreatorFollowButton,
} from "@/entities/creator";
import {
  useCurrentViewer,
  useHasViewerSession,
} from "@/entities/viewer";
import { useCreatorModeEntry } from "@/features/creator-entry";
import {
  buildCreatorShortDetailHref,
  resolveCreatorProfileBackHref,
  type CreatorProfileRouteState,
} from "@/features/creator-navigation";
import { useFanAuthDialog } from "@/features/fan-auth";
import { Button } from "@/shared/ui";

import type {
  CreatorProfileShellShortItem,
  CreatorProfileShellState,
} from "../model/load-creator-profile-shell-state";
import { useCreatorProfileFollow } from "../model/use-creator-profile-follow";

type CreatorProfileShellProps = {
  routeState: CreatorProfileRouteState;
  state: CreatorProfileShellState;
};

type CreatorProfileShortGridTileProps = {
  creatorDisplayName: string;
  creatorId: string;
  routeState: CreatorProfileRouteState;
  short: CreatorProfileShellShortItem;
};

function formatPreviewDuration(seconds: number): string {
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;

  return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}`;
}

function formatProfileCount(value: number): string {
  return new Intl.NumberFormat("en", {
    maximumFractionDigits: 1,
    notation: "compact",
  }).format(value);
}

function normalizeViewerIdForSelfComparison(viewerId: string | undefined): string | null {
  if (!viewerId) {
    return null;
  }

  const normalizedViewerId = viewerId.trim().toLowerCase().replaceAll("-", "");

  return /^[0-9a-f]{32}$/.test(normalizedViewerId) ? normalizedViewerId : null;
}

function normalizeCreatorIdForSelfComparison(creatorId: string): string | null {
  const normalizedCreatorId = creatorId.trim().toLowerCase();

  if (!normalizedCreatorId.startsWith("creator_")) {
    return null;
  }

  const creatorUserId = normalizedCreatorId.slice("creator_".length);

  return /^[0-9a-f]{32}$/.test(creatorUserId) ? creatorUserId : null;
}

function isSelfCreatorProfile(currentViewerId: string | undefined, creatorId: string): boolean {
  const normalizedViewerId = normalizeViewerIdForSelfComparison(currentViewerId);
  const normalizedCreatorUserId = normalizeCreatorIdForSelfComparison(creatorId);

  return normalizedViewerId !== null && normalizedCreatorUserId !== null && normalizedViewerId === normalizedCreatorUserId;
}

type CreatorProfileStatProps = {
  label: string;
  value: string;
};

function CreatorProfileStat({ label, value }: CreatorProfileStatProps) {
  return (
    <div className="min-w-[84px] text-center">
      <strong className="block font-display text-[22px] font-bold leading-none tracking-[-0.04em] text-foreground">
        {value}
      </strong>
      <span className="mt-2 block text-[11px] font-semibold uppercase tracking-[0.16em] text-muted">
        {label}
      </span>
    </div>
  );
}

function CreatorProfileShortGridTile({
  creatorDisplayName,
  creatorId,
  routeState,
  short,
}: CreatorProfileShortGridTileProps) {
  const previewDuration = formatPreviewDuration(short.previewDurationSeconds);

  return (
    <Link
      aria-label={`${creatorDisplayName} preview ${previewDuration}`}
      className="group relative block aspect-[3/4] overflow-hidden bg-[linear-gradient(180deg,#d7f4ff_0%,#81c7f1_44%,#1f4f73_100%)] focus-visible:z-10 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-ring/70"
      href={buildCreatorShortDetailHref(short.routeShortId, creatorId, routeState)}
    >
      <video
        aria-hidden="true"
        className="absolute inset-0 size-full object-cover transition duration-200 group-hover:scale-[1.02]"
        muted
        playsInline
        poster={short.media.posterUrl ?? undefined}
        preload="metadata"
        src={short.media.url}
      />
      <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(10,22,35,0.02)_0%,rgba(10,22,35,0.18)_68%,rgba(10,22,35,0.4)_100%)]" />
      <span className="absolute bottom-2 right-2 rounded-full bg-black/42 px-2 py-1 text-[11px] font-semibold text-white backdrop-blur-sm">
        {previewDuration}
      </span>
    </Link>
  );
}

/**
 * creator profile の header と short grid を表示する。
 */
export function CreatorProfileShell({
  routeState,
  state,
}: CreatorProfileShellProps) {
  const { creator, shorts, stats, viewer } = state;
  const currentViewer = useCurrentViewer();
  const { openFanAuthDialog } = useFanAuthDialog();
  const hasViewerSession = useHasViewerSession();
  const {
    enterCreatorMode,
    errorMessage: creatorModeEntryErrorMessage,
    isSubmitting: isCreatorModeEntrySubmitting,
  } = useCreatorModeEntry();
  const backHref = resolveCreatorProfileBackHref(routeState);
  const isSelfProfile = isSelfCreatorProfile(currentViewer?.id, creator.id);
  const {
    errorMessage,
    fanCount,
    isFollowing,
    isPending,
    toggleFollow,
  } = useCreatorProfileFollow({
    creatorId: creator.id,
    hasViewerSession,
    initialFanCount: stats.fanCount,
    initialIsFollowing: viewer.isFollowing,
    onAuthRequired: () => {
      openFanAuthDialog();
    },
    onUnauthenticated: () => {
      openFanAuthDialog();
    },
  });
  const primaryActionErrorMessage = isSelfProfile ? creatorModeEntryErrorMessage : errorMessage;

  return (
    <section className="min-h-full overflow-y-auto bg-white pb-28 text-foreground">
      <header className="sticky top-0 z-10 bg-white/95 backdrop-blur-sm">
        <div className="flex min-h-14 items-center border-b border-border/70 px-3.5">
          <Button
            asChild
            className="size-10 text-foreground hover:bg-surface-subtle"
            size="icon"
            variant="ghost"
          >
            <Link aria-label="Back" href={backHref}>
              <ArrowLeft className="size-5" strokeWidth={2.1} />
            </Link>
          </Button>
          <div className="min-w-0 flex-1 px-3 text-center">
            <p className="truncate text-[18px] font-semibold tracking-[-0.03em] text-foreground">
              {creator.handle}
            </p>
          </div>
          <div aria-hidden="true" className="flex size-10 items-center justify-center text-muted">
            <EllipsisVertical className="size-[18px]" strokeWidth={2.1} />
          </div>
        </div>
      </header>

      <div className="px-6 pb-8 pt-7 text-center">
        <CreatorAvatar
          className="mx-auto size-28 shadow-[0_16px_36px_rgba(36,92,129,0.16)]"
          creator={creator}
        />

        <div className="mt-5">
          <h1 className="text-[20px] font-semibold tracking-[-0.04em] text-foreground">
            {creator.displayName}
          </h1>
          <p className="mt-1 text-[15px] font-semibold tracking-[-0.02em] text-muted-strong">
            {creator.handle}
          </p>
        </div>

        <p className="mx-auto mt-4 max-w-[292px] text-[15px] leading-[1.6] text-muted-strong">
          {creator.bio}
        </p>

        <div className="mt-7 flex justify-center gap-6">
          <CreatorProfileStat label="Followers" value={formatProfileCount(fanCount)} />
          <CreatorProfileStat label="Shorts" value={stats.shortCount.toString()} />
        </div>

        <div className="mx-auto mt-7 w-full max-w-[284px]">
          {isSelfProfile ? (
            <Button
              aria-busy={isCreatorModeEntrySubmitting || undefined}
              className="h-12 w-full rounded-full text-[17px] font-semibold"
              disabled={isCreatorModeEntrySubmitting}
              onClick={() => {
                void enterCreatorMode();
              }}
              type="button"
            >
              {isCreatorModeEntrySubmitting ? "Creatorページを開いています..." : "Creatorページを開く"}
            </Button>
          ) : (
            <CreatorFollowButton
              className="h-12 rounded-full text-[17px] font-semibold shadow-[0_14px_32px_rgba(80,159,224,0.24)]"
              fullWidth
              isFollowing={isFollowing}
              isPending={isPending}
              onClick={() => {
                void toggleFollow();
              }}
            />
          )}
          {primaryActionErrorMessage ? (
            <p
              className="mt-3 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
              role="alert"
            >
              {primaryActionErrorMessage}
            </p>
          ) : null}
        </div>
      </div>

      {state.kind === "empty" ? (
        <p className="border-t border-border/70 px-6 pt-8 text-center text-[14px] leading-6 text-muted">
          まだ公開中の short はありません。
        </p>
      ) : (
        <div className="grid grid-cols-3 gap-px bg-white">
          {shorts.map((short) => (
            <div className="bg-white" key={short.id}>
              <CreatorProfileShortGridTile
                creatorDisplayName={creator.displayName}
                creatorId={creator.id}
                routeState={routeState}
                short={short}
              />
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
