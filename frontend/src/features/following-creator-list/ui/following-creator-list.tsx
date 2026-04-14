"use client";

import Link from "next/link";
import { ArrowLeft, Search } from "lucide-react";
import { useDeferredValue, useState } from "react";

import type { FanFollowingItem } from "@/entities/fan-profile";
import {
  CreatorFollowApiError,
  CreatorAvatar,
  CreatorFollowButton,
  CreatorIdentity,
  updateCreatorFollow,
} from "@/entities/creator";
import { ApiError } from "@/shared/api";
import { cn } from "@/shared/lib";
import { Button } from "@/shared/ui";

import {
  useFollowingCreatorRows,
  type UpdateFollowingCreatorRelation,
} from "../model/use-following-creator-rows";

export type FollowingCreatorListProps = {
  layout?: "embedded" | "standalone";
  items: readonly FanFollowingItem[];
  onAuthRequired?: (() => void) | undefined;
  updateFollowingCreatorRelation?: UpdateFollowingCreatorRelation | undefined;
};

function getFollowingCreatorRelationErrorMessage(error: unknown): string {
  if (error instanceof CreatorFollowApiError) {
    if (error.code === "not_found") {
      return "この creator は現在利用できません。";
    }

    return "フォロー状態を更新できませんでした。少し時間を置いてから再度お試しください。";
  }

  if (error instanceof ApiError) {
    if (error.code === "network") {
      return "フォロー状態を更新できませんでした。通信状態を確認してから再度お試しください。";
    }

    return "フォロー状態を更新できませんでした。少し時間を置いてから再度お試しください。";
  }

  return "フォロー状態を更新できませんでした。少し時間を置いてから再度お試しください。";
}

function removeCreatorErrorMessage(
  errorMessageByCreatorId: Record<string, string>,
  creatorId: string,
): Record<string, string> {
  if (!(creatorId in errorMessageByCreatorId)) {
    return errorMessageByCreatorId;
  }

  const nextErrorMessageByCreatorId = { ...errorMessageByCreatorId };

  delete nextErrorMessageByCreatorId[creatorId];

  return nextErrorMessageByCreatorId;
}

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
 * following 一覧の検索と row actions を表示する。
 */
export function FollowingCreatorList({
  layout = "standalone",
  items,
  onAuthRequired,
  updateFollowingCreatorRelation,
}: FollowingCreatorListProps) {
  const [query, setQuery] = useState("");
  const [errorMessageByCreatorId, setErrorMessageByCreatorId] = useState<Record<string, string>>({});
  const deferredQuery = useDeferredValue(query);
  const resolveUpdateFollowingCreatorRelation =
    updateFollowingCreatorRelation ??
    (async ({ action, creatorId }) => {
      setErrorMessageByCreatorId((currentErrorMessageByCreatorId) =>
        removeCreatorErrorMessage(currentErrorMessageByCreatorId, creatorId),
      );

      try {
        await updateCreatorFollow({
          action,
          creatorId,
        });
      } catch (error) {
        if (error instanceof CreatorFollowApiError && error.code === "auth_required") {
          onAuthRequired?.();
          throw error;
        }

        setErrorMessageByCreatorId((currentErrorMessageByCreatorId) => ({
          ...currentErrorMessageByCreatorId,
          [creatorId]: getFollowingCreatorRelationErrorMessage(error),
        }));
        throw error;
      }
    });
  const { rows, toggleFollowing } = useFollowingCreatorRows({
    items,
    updateFollowingCreatorRelation: resolveUpdateFollowingCreatorRelation,
  });
  const visibleRows = rows.filter((row) => matchesCreatorQuery(row, deferredQuery));
  const isEmbedded = layout === "embedded";

  return (
    <section
      className={cn(
        "font-sans text-foreground",
        isEmbedded ? "bg-white pb-8 pt-4" : "min-h-full overflow-y-auto px-4 pb-28 pt-4",
      )}
    >
      {isEmbedded ? null : (
        <>
          <div className="flex items-center justify-between gap-3">
            <Button asChild size="icon" variant="ghost">
              <Link aria-label="Back" href="/fan">
                <ArrowLeft className="size-5" strokeWidth={2.1} />
              </Link>
            </Button>
            <span aria-hidden="true" className="size-10" />
          </div>

          <h1 className="mt-2 font-display text-[28px] font-semibold tracking-[-0.04em] text-foreground">
            following
          </h1>
        </>
      )}

      <div
        className={cn(
          isEmbedded
            ? "mt-1 px-4"
            : "mt-4 rounded-2xl border border-border/80 bg-white/82 px-4 py-3 shadow-[0_12px_28px_rgba(36,94,132,0.08)] backdrop-blur-md",
        )}
      >
        <label className="relative block">
          <Search
            className="pointer-events-none absolute left-4 top-1/2 size-5 -translate-y-1/2 text-[#9ca3af]"
            strokeWidth={2.2}
          />
          <input
            className="h-[46px] w-full rounded-full border border-black/[0.04] bg-[#f4f5f7] pl-[50px] pr-5 text-[16px] font-semibold tracking-[-0.02em] text-[#101828] outline-none shadow-[inset_0_1px_2px_rgba(15,23,42,0.04),0_2px_10px_rgba(15,23,42,0.04)] placeholder:font-semibold placeholder:text-[#a9b0bc] focus-visible:ring-4 focus-visible:ring-ring/70"
            onChange={(event) => setQuery(event.currentTarget.value)}
            placeholder="検索"
            type="search"
            value={query}
          />
        </label>
      </div>

      <div className={cn("mt-8 flex items-end justify-between gap-3", isEmbedded ? "px-5" : "px-1")}>
        <p
          className={cn(
            "uppercase",
            isEmbedded
              ? "text-[12px] font-bold tracking-[0.22em] text-[#6fa9ed]"
              : "font-display text-[11px] font-semibold tracking-[0.24em] text-accent",
          )}
        >
          all creators
        </p>
        <span className={cn(isEmbedded ? "text-[13px] font-semibold text-[#99a0ad]" : "text-[12px] text-muted")}>
          {visibleRows.length} creators
        </span>
      </div>

      <div className="mt-3">
        {visibleRows.length ? (
          visibleRows.map((row, index) => (
            <div
              key={row.creator.id}
              className={cn(
                isEmbedded
                  ? "mx-4 border-b border-[#edf0f4] px-1 py-3"
                  : "border-b border-border/55 py-3.5",
                isEmbedded ? index === 0 && "border-t border-[#edf0f4]" : index === 0 && "border-t border-border/55",
              )}
            >
              <div className="flex items-center justify-between gap-3">
                <Link className="min-w-0 flex-1 text-left transition hover:opacity-90" href={`/creators/${row.creator.id}`}>
                  <span className="flex items-center gap-3">
                    <CreatorAvatar
                      className={cn(
                        isEmbedded
                          ? "size-[38px] rounded-full border-white/72 shadow-[0_8px_20px_rgba(7,19,29,0.2)]"
                          : "size-10 rounded-[16px]",
                      )}
                      creator={row.creator}
                    />
                    <CreatorIdentity
                      className={cn(
                        "text-foreground",
                        isEmbedded &&
                          "[&_p:first-child]:text-[14px] [&_p:first-child]:font-bold [&_p:last-child]:mt-0.5 [&_p:last-child]:text-[12px] [&_p:last-child]:text-muted",
                      )}
                      creator={row.creator}
                    />
                  </span>
                </Link>
                <CreatorFollowButton
                  className={cn(
                    "shrink-0",
                    isEmbedded
                      ? "h-9 min-w-[96px] border-[#dfe4eb] bg-white px-4 text-[14px] font-bold text-[#1f2430] shadow-none"
                      : "h-9 px-4",
                  )}
                  isFollowing={row.isFollowing}
                  isPending={row.isPending}
                  labels={{
                    follow: "フォロー",
                    followPending: "フォロー中...",
                    following: "フォロー中",
                    unfollowPending: "フォロー解除中...",
                  }}
                  onClick={() => {
                    void toggleFollowing(row.creator.id).catch(() => undefined);
                  }}
                />
              </div>
              {errorMessageByCreatorId[row.creator.id] ? (
                <p
                  className={cn(
                    "mt-3 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]",
                  )}
                  role="alert"
                >
                  {errorMessageByCreatorId[row.creator.id]}
                </p>
              ) : null}
            </div>
          ))
        ) : (
          <p className={cn("mt-4 text-[13px] leading-6", isEmbedded ? "px-1 text-[#7b8393]" : "text-muted")}>
            一致する creator は見つかりませんでした。
          </p>
        )}
      </div>
    </section>
  );
}
