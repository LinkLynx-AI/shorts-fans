"use client";

import Link from "next/link";
import { Search } from "lucide-react";

import { CreatorAvatar, getCreatorById } from "@/entities/creator";
import { buildCreatorProfileHref } from "@/features/creator-navigation";

import type { CreatorSearchState } from "../model/creator-search-state";
import { useCreatorSearch } from "../model/use-creator-search";

type CreatorSearchPanelProps = {
  initialState: CreatorSearchState;
  initialQuery: string;
};

/**
 * creator search の入力と結果一覧を表示する。
 */
export function CreatorSearchPanel({
  initialState,
  initialQuery,
}: CreatorSearchPanelProps) {
  const { query, retry, setQuery, state } = useCreatorSearch({
    initialQuery,
    initialState,
  });
  const showsRecentLabel = query.trim().length === 0;
  const creators = state.kind === "ready" ? state.items : [];

  return (
    <div className="mt-1">
      <div className="relative">
        <Search
          aria-hidden="true"
          className="pointer-events-none absolute left-4 top-1/2 size-[15px] -translate-y-1/2 text-accent-strong/54"
          strokeWidth={2}
        />
        <input
          aria-label="クリエイターを検索"
          className="h-[46px] w-full rounded-[18px] border border-[rgba(186,231,255,0.9)] bg-white/94 pl-11 pr-4 text-sm text-foreground outline-none shadow-[0_12px_32px_rgba(26,69,98,0.12)] placeholder:text-muted focus-visible:ring-4 focus-visible:ring-ring/70"
          onChange={(event) => {
            setQuery(event.currentTarget.value);
          }}
          placeholder="検索"
          type="search"
          value={query}
        />
      </div>

      {showsRecentLabel ? <p className="mt-4 text-[13px] font-bold text-muted">最近</p> : null}

      {state.kind === "loading" ? (
        <p className="mt-4 text-[13px] font-bold text-muted" role="status">
          読み込み中...
        </p>
      ) : null}

      {state.kind === "error" ? (
        <div className="mt-4 rounded-[18px] bg-white/84 px-4 py-4 text-foreground shadow-[0_12px_32px_rgba(26,69,98,0.12)]">
          <p className="text-[13px] leading-6 text-muted">{state.message}</p>
          <button
            className="mt-3 inline-flex min-h-9 items-center rounded-[10px] bg-accent-strong px-3 text-[13px] font-bold text-white focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-ring/70"
            onClick={retry}
            type="button"
          >
            再読み込み
          </button>
        </div>
      ) : null}

      {state.kind === "empty" ? (
        <p className="mt-4 text-[13px] leading-6 text-muted">
          {showsRecentLabel ? "表示できる creator がまだいません。" : "一致する creator は見つかりませんでした。"}
        </p>
      ) : null}

      <div className={showsRecentLabel ? "mt-[10px] grid gap-[10px]" : "mt-4 grid gap-[10px]"}>
        {creators.map((creator) => (
          <Link
            key={creator.id}
            className="flex items-center gap-3 rounded-[18px] bg-white/80 px-3 py-3 text-foreground transition hover:bg-white/90 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-ring/70"
            href={buildCreatorProfileHref(creator.id, {
              ...(getCreatorById(creator.id)
                ? {}
                : {
                    creatorDisplayName: creator.displayName,
                    creatorHandle: creator.handle,
                  }),
              from: "search",
              q: state.query,
            })}
          >
            <span className="flex min-w-0 items-center gap-3">
              <CreatorAvatar
                className="size-[38px] rounded-full border-white/72 shadow-[0_8px_20px_rgba(7,19,29,0.2)]"
                creator={creator}
              />
              <span className="min-w-0">
                <span className="block truncate text-[14px] font-bold">{creator.displayName}</span>
                <span className="mt-0.5 block truncate text-[12px] text-muted">{creator.handle}</span>
              </span>
            </span>
          </Link>
        ))}
      </div>
    </div>
  );
}
