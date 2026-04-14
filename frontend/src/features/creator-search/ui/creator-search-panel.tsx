"use client";

import Link from "next/link";
import { Search } from "lucide-react";

import { CreatorAvatar, getCreatorById, type CreatorSummary } from "@/entities/creator";
import { buildCreatorProfileHref } from "@/features/creator-navigation";

import type { CreatorSearchState } from "../model/creator-search-state";
import { useCreatorSearch } from "../model/use-creator-search";

type CreatorSearchPanelProps = {
  initialState: CreatorSearchState;
  initialQuery: string;
};

function CreatorSearchResultRow({ creator, query }: { creator: CreatorSummary; query: string }) {
  return (
    <Link
      className="group flex items-center gap-3 text-foreground transition hover:opacity-90 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-ring/70"
      href={buildCreatorProfileHref(creator.id, {
        ...(getCreatorById(creator.id)
          ? {}
          : {
              creatorDisplayName: creator.displayName,
              creatorHandle: creator.handle,
            }),
        from: "search",
        q: query,
      })}
    >
      <CreatorAvatar
        className="size-11 shrink-0 rounded-full border border-black/[0.04] bg-[#f3f4f6] shadow-[0_1px_2px_rgba(15,23,42,0.06)]"
        creator={creator}
      />
      <span className="min-w-0">
        <span className="block truncate text-[16px] font-semibold tracking-[-0.03em] text-[#101828]">
          {creator.displayName}
        </span>
        <span className="mt-0.5 block truncate text-[14px] font-semibold tracking-[-0.01em] text-[#7f8795]">
          {creator.handle}
        </span>
      </span>
    </Link>
  );
}

function LoadingRows() {
  return (
    <div aria-hidden="true" className="mt-6 space-y-6">
      {Array.from({ length: 4 }).map((_, index) => (
        <div key={`search-loading-row-${index}`} className="flex items-center gap-3">
          <div className="size-11 shrink-0 rounded-full bg-[#eef1f5]" />
          <div className="min-w-0 flex-1">
            <div className="h-[18px] w-[88px] rounded-full bg-[#eef1f5]" />
            <div className="mt-1.5 h-[14px] w-[72px] rounded-full bg-[#f3f5f8]" />
          </div>
        </div>
      ))}
    </div>
  );
}

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
    <div className="min-h-full bg-white">
      <div className="sticky top-0 z-10 border-b border-black/[0.06] bg-white/[0.96] px-4 pb-5 pt-5 backdrop-blur-md">
        <label className="relative block">
          <Search
            aria-hidden="true"
            className="pointer-events-none absolute left-5 top-1/2 size-5 -translate-y-1/2 text-[#9ca3af]"
            strokeWidth={2.2}
          />
          <input
            aria-label="クリエイターを検索"
            className="h-[46px] w-full appearance-none rounded-full border border-black/[0.04] bg-[#f4f5f7] pl-[50px] pr-5 text-[16px] font-semibold tracking-[-0.02em] text-[#101828] outline-none shadow-[inset_0_1px_2px_rgba(15,23,42,0.04),0_2px_10px_rgba(15,23,42,0.04)] placeholder:font-semibold placeholder:text-[#a9b0bc] focus-visible:ring-4 focus-visible:ring-ring/70"
            onChange={(event) => {
              setQuery(event.currentTarget.value);
            }}
            placeholder="検索"
            type="search"
            value={query}
          />
        </label>
      </div>

      <div className="px-5 pb-10 pt-6">
        {showsRecentLabel ? (
          <p className="text-[13px] font-bold tracking-[-0.02em] text-[#a1a8b3]">最近</p>
        ) : null}

        {state.kind === "loading" ? (
          <>
            <p className="sr-only" role="status">
              読み込み中...
            </p>
            <LoadingRows />
          </>
        ) : null}

        {state.kind === "error" ? (
          <div className="mt-6 rounded-[22px] border border-[#f2d7dc] bg-[#fff8f9] px-5 py-4 text-foreground shadow-[0_10px_28px_rgba(15,23,42,0.05)]">
            <p className="text-[14px] leading-6 text-[#7b5560]">{state.message}</p>
            <button
              className="mt-4 inline-flex min-h-10 items-center rounded-full bg-[#101828] px-4 text-[13px] font-semibold text-white focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-ring/70"
              onClick={retry}
              type="button"
            >
              再読み込み
            </button>
          </div>
        ) : null}

        {state.kind === "empty" ? (
          <p className="mt-6 text-[15px] leading-7 text-[#7f8795]">
            {showsRecentLabel ? "表示できる creator がまだいません。" : "一致する creator は見つかりませんでした。"}
          </p>
        ) : null}

        {state.kind === "ready" ? (
          <div className={showsRecentLabel ? "mt-6 space-y-6" : "mt-2 space-y-6"}>
            {creators.map((creator) => (
              <CreatorSearchResultRow key={creator.id} creator={creator} query={state.query} />
            ))}
          </div>
        ) : null}
      </div>
    </div>
  );
}
