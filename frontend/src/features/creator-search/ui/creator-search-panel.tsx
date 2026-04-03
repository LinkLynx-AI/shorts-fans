"use client";

import Link from "next/link";
import { Search } from "lucide-react";
import { useState } from "react";

import { CreatorAvatar, searchCreators } from "@/entities/creator";
import { buildCreatorProfileHref } from "@/features/creator-navigation";

import { useDebouncedValue } from "../model/use-debounced-value";

type CreatorSearchPanelProps = {
  initialQuery: string;
};

const searchDelayMs = 250;

/**
 * creator search の入力と結果一覧を表示する。
 */
export function CreatorSearchPanel({ initialQuery }: CreatorSearchPanelProps) {
  const [query, setQuery] = useState(initialQuery);
  const resolvedQuery = useDebouncedValue(query, searchDelayMs);
  const creators = searchCreators(resolvedQuery);
  const showsRecentLabel = query.trim().length === 0;

  return (
    <div className="mt-1">
      <div className="relative">
        <Search
          aria-hidden="true"
          className="pointer-events-none absolute left-4 top-1/2 size-[15px] -translate-y-1/2 text-accent-strong/54"
          strokeWidth={2}
        />
        <input
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

      <div className={showsRecentLabel ? "mt-[10px] grid gap-[10px]" : "mt-4 grid gap-[10px]"}>
        {creators.map((creator) => (
          <Link
            key={creator.id}
            className="flex items-center gap-3 rounded-[18px] bg-white/80 px-3 py-3 text-foreground transition hover:bg-white/90"
            href={buildCreatorProfileHref(creator.id, {
              from: "search",
              q: query,
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
