"use client";

import Link from "next/link";

import { CreatorAvatar, CreatorStatList } from "@/entities/creator";
import { useHasViewerSession } from "@/entities/viewer";
import {
  buildCreatorShortDetailHref,
  resolveCreatorProfileBackHref,
  type CreatorProfileRouteState,
} from "@/features/creator-navigation";
import { useFanAuthDialog } from "@/features/fan-auth";
import { Button } from "@/shared/ui";
import { DetailShell } from "@/widgets/detail-shell";

import type {
  CreatorProfileShellShortItem,
  CreatorProfileShellState,
} from "../model/load-creator-profile-shell-state";

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

function ShortsGridTab() {
  return (
    <div className="mt-[18px] flex justify-center border-t border-border/60">
      <div
        aria-hidden="true"
        className="inline-flex min-h-[42px] w-[72px] items-center justify-center border-t-2 border-t-foreground pt-[10px] text-foreground"
      >
        <svg aria-hidden="true" className="size-[18px] fill-current" viewBox="0 0 18 18">
          <rect height="4" rx="1" width="4" x="2" y="2" />
          <rect height="4" rx="1" width="4" x="7" y="2" />
          <rect height="4" rx="1" width="4" x="12" y="2" />
          <rect height="4" rx="1" width="4" x="2" y="7" />
          <rect height="4" rx="1" width="4" x="7" y="7" />
          <rect height="4" rx="1" width="4" x="12" y="7" />
          <rect height="4" rx="1" width="4" x="2" y="12" />
          <rect height="4" rx="1" width="4" x="7" y="12" />
          <rect height="4" rx="1" width="4" x="12" y="12" />
        </svg>
      </div>
    </div>
  );
}

function formatPreviewDuration(seconds: number): string {
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;

  return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}`;
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
      className="group relative block aspect-[3/4] overflow-hidden bg-[linear-gradient(180deg,#d7f4ff_0%,#81c7f1_44%,#1f4f73_100%)] focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-ring/70"
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
      <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(6,21,33,0.04)_0%,rgba(6,21,33,0.28)_72%,rgba(6,21,33,0.62)_100%)]" />
      <span className="absolute bottom-2 right-2 rounded-full bg-black/54 px-2 py-1 text-[11px] font-semibold text-white backdrop-blur-sm">
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
  const { openFanAuthDialog } = useFanAuthDialog();
  const hasViewerSession = useHasViewerSession();
  const backHref = resolveCreatorProfileBackHref(routeState);
  const displayHandle = creator.handle.replace(/^@/, "");
  const isFollowing = viewer.isFollowing;
  const isFollowReadOnly = hasViewerSession && !isFollowing;

  return (
    <DetailShell
      backButtonClassName="-ml-2"
      backHref={backHref}
      headerContent={
        <p className="truncate text-[19px] font-semibold tracking-[-0.03em] text-foreground">
          {displayHandle}
        </p>
      }
      variant="surface"
    >
      <section className="text-foreground">
        <h1 className="sr-only">{creator.displayName} creator profile</h1>
        <div className="flex items-start gap-4">
          <CreatorAvatar
            className="size-[86px] rounded-full border-white/70 shadow-[0_10px_24px_rgba(36,92,129,0.16)]"
            creator={creator}
          />
          <div className="min-w-0 flex-1 pl-1.5">
            <p className="truncate text-[18px] font-semibold tracking-[-0.03em] text-foreground">
              {creator.displayName}
            </p>
            <CreatorStatList className="mt-3 gap-[10px]" stats={stats} variant="creatorProfile" />
          </div>
        </div>

        <p className="mt-4 text-[13px] leading-[1.6] text-muted">{creator.bio}</p>

        <div className="mt-[14px]">
          <Button
            aria-pressed={isFollowing}
            className={
              isFollowing
                ? "min-h-9 w-full rounded-[10px] border-transparent bg-[#edf2f7] text-[13px] font-bold text-foreground shadow-none backdrop-blur-none"
                : "min-h-9 w-full rounded-[10px] bg-accent-strong text-[13px] font-bold text-white shadow-none disabled:brightness-90"
            }
            disabled={isFollowReadOnly}
            onClick={() => {
              if (!hasViewerSession) {
                openFanAuthDialog();
              }
            }}
            type="button"
            variant={isFollowing ? "secondary" : "default"}
          >
            {isFollowing ? "Following" : "Follow"}
          </Button>
        </div>

        <ShortsGridTab />

        {state.kind === "empty" ? (
          <p className="px-1 pt-6 text-center text-[13px] leading-6 text-muted">
            まだ公開中の short はありません。
          </p>
        ) : (
          <div className="mt-0.5 grid grid-cols-3 gap-[3px]">
            {shorts.map((short) => (
              <CreatorProfileShortGridTile
                creatorDisplayName={creator.displayName}
                creatorId={creator.id}
                key={short.id}
                routeState={routeState}
                short={short}
              />
            ))}
          </div>
        )}
      </section>
    </DetailShell>
  );
}
