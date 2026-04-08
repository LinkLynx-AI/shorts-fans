"use client";

import Link from "next/link";
import { ArrowLeft, Search } from "lucide-react";
import { useDeferredValue, useState } from "react";

import type { FanFollowingItem } from "@/entities/fan-profile";
import {
  CreatorAvatar,
  CreatorFollowButton,
  CreatorIdentity,
} from "@/entities/creator";
import { cn } from "@/shared/lib";
import { Button } from "@/shared/ui";

import {
  useFollowingCreatorRows,
  type UpdateFollowingCreatorRelation,
} from "../model/use-following-creator-rows";

type FollowingShellProps = {
  items: readonly FanFollowingItem[];
  updateFollowingCreatorRelation?: UpdateFollowingCreatorRelation | undefined;
};

function matchesCreatorQuery(
  item: {
    creator: FanFollowingItem["creator"];
  },
  query: string,
): boolean {
  const normalizedQuery = query.trim().toLowerCase();

  if (!normalizedQuery) {
    return true;
  }

  return `${item.creator.displayName} ${item.creator.handle}`.toLowerCase().includes(normalizedQuery);
}

/**
 * following 詳細画面を表示する。
 */
export function FollowingShell({
  items,
  updateFollowingCreatorRelation,
}: FollowingShellProps) {
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const { rows, toggleFollowing } = useFollowingCreatorRows({
    items,
    updateFollowingCreatorRelation,
  });
  const visibleRows = rows.filter((row) => matchesCreatorQuery(row, deferredQuery));

  return (
    <section className="min-h-full overflow-y-auto px-4 pb-28 pt-4 text-foreground">
      <div className="flex items-center justify-between gap-3">
        <Button asChild size="icon" variant="ghost">
          <Link aria-label="Back" href="/fan">
            <ArrowLeft className="size-5" strokeWidth={2.1} />
          </Link>
        </Button>
        <span aria-hidden="true" className="size-10" />
      </div>

      <h1 className="mt-2 font-display text-[28px] font-semibold tracking-[-0.04em] text-foreground">following</h1>

      <div className="mt-4 rounded-[22px] border border-border/80 bg-white/82 px-4 py-3 shadow-[0_12px_28px_rgba(36,94,132,0.08)] backdrop-blur-md">
        <label className="relative block">
          <Search className="pointer-events-none absolute left-0 top-1/2 size-4 -translate-y-1/2 text-muted" strokeWidth={2} />
          <input
            className="h-6 w-full border-0 bg-transparent pl-7 pr-0 text-sm text-foreground outline-none placeholder:text-muted"
            onChange={(event) => setQuery(event.currentTarget.value)}
            placeholder="検索"
            type="search"
            value={query}
          />
        </label>
      </div>

      <div className="mt-6 flex items-end justify-between gap-3">
        <p className="font-display text-[11px] font-semibold uppercase tracking-[0.24em] text-accent">all creators</p>
        <span className="text-[12px] text-muted">{visibleRows.length} creators</span>
      </div>

      <div className="mt-3">
        {visibleRows.length ? (
          visibleRows.map((row, index) => (
            <div
              key={row.creator.id}
              className={cn(
                "flex items-center justify-between gap-3 border-b border-border/55 py-3.5",
                index === 0 && "border-t border-border/55",
              )}
            >
              <Link className="min-w-0 flex-1 text-left transition hover:opacity-90" href={`/creators/${row.creator.id}`}>
                <span className="flex items-center gap-3">
                  <CreatorAvatar className="size-10 rounded-[16px]" creator={row.creator} />
                  <CreatorIdentity className="text-foreground" creator={row.creator} />
                </span>
              </Link>
              <CreatorFollowButton
                className="h-9 shrink-0 px-4"
                isFollowing={row.isFollowing}
                isPending={row.isPending}
                labels={{
                  follow: "フォロー",
                  followPending: "フォロー中...",
                  following: "フォロー中",
                  unfollowPending: "フォロー解除中...",
                }}
                onClick={() => {
                  void toggleFollowing(row.creator.id);
                }}
              />
            </div>
          ))
        ) : (
          <p className="mt-4 text-[13px] leading-6 text-muted">一致する creator は見つかりませんでした。</p>
        )}
      </div>
    </section>
  );
}
