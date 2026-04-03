import { Search } from "lucide-react";

import { RouteStructurePanel } from "@/widgets/route-structure-panel";

type SearchShellProps = {
  query: string;
};

/**
 * creator search 用の route shell を表示する。
 */
export function SearchShell({ query }: SearchShellProps) {
  return (
    <section className="min-h-full overflow-y-auto px-4 pb-28 pt-4 text-foreground">
      <h1 className="font-display text-[30px] font-semibold tracking-[-0.05em] text-foreground">Creator search structure</h1>
      <form action="/search" className="relative mt-4">
        <Search className="pointer-events-none absolute left-4 top-1/2 size-4 -translate-y-1/2 text-muted" strokeWidth={2} />
        <input
          className="h-12 w-full rounded-full border border-border bg-white/88 pl-11 pr-4 text-sm text-foreground shadow-[0_12px_28px_rgba(36,94,132,0.1)] outline-none backdrop-blur-md placeholder:text-muted focus-visible:ring-4 focus-visible:ring-ring/70"
          defaultValue={query}
          name="q"
          placeholder="creator を検索"
          type="search"
        />
      </form>

      <RouteStructurePanel
        description={
          query
            ? `query "${query}" を保持できる route だけを先に定義しています。`
            : "creator search only の route と input frame を先に定義しています。"
        }
        items={[
          {
            description: "display name / handle 検索をここに接続する",
            key: "input",
            label: "Search input contract",
          },
          {
            description: "検索結果 row と creator profile 遷移をここに差し込む",
            key: "results",
            label: "Result list slot",
          },
        ]}
        title="Search route blueprint"
      />
    </section>
  );
}
