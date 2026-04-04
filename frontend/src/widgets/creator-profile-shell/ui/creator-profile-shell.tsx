"use client";

import { useState } from "react";
import Link from "next/link";

import { CreatorAvatar, CreatorStatList } from "@/entities/creator";
import { ShortPoster } from "@/entities/short";
import {
  buildCreatorShortDetailHref,
  resolveCreatorProfileBackHref,
  type CreatorProfileRouteState,
} from "@/features/creator-navigation";
import { Button } from "@/shared/ui";
import { DetailShell } from "@/widgets/detail-shell";

import type { CreatorProfileShellState } from "../model/mock-creator-profile-shell";

type CreatorProfileShellProps = {
  routeState: CreatorProfileRouteState;
  state: CreatorProfileShellState;
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

/**
 * creator profile の header と short grid を表示する。
 */
export function CreatorProfileShell({
  routeState,
  state,
}: CreatorProfileShellProps) {
  const { creator, shorts, stats, viewer } = state;
  const backHref = resolveCreatorProfileBackHref(routeState);
  const displayHandle = creator.handle.replace(/^@/, "");
  const [isFollowing, setIsFollowing] = useState(viewer.isFollowing);

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
                : "min-h-9 w-full rounded-[10px] bg-accent-strong text-[13px] font-bold text-white shadow-none"
            }
            onClick={() => {
              setIsFollowing((currentValue) => !currentValue);
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
              <Link
                key={short.id}
                aria-label={`${creator.displayName} ${short.title}`}
                className="block focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-ring/70"
                href={buildCreatorShortDetailHref(short.id, creator.id, routeState)}
              >
                <ShortPoster short={short} variant="profile" />
              </Link>
            ))}
          </div>
        )}
      </section>
    </DetailShell>
  );
}
