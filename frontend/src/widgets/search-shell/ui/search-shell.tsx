import { CreatorSearchPanel } from "@/features/creator-search";
import type { CreatorSearchState } from "@/features/creator-search";

type SearchShellProps = {
  initialState: CreatorSearchState;
  query: string;
};

/**
 * creator search 用の route shell を表示する。
 */
export function SearchShell({ initialState, query }: SearchShellProps) {
  return (
    <section className="min-h-full overflow-y-auto bg-white pb-28 text-foreground">
      <h1 className="sr-only">Creator search</h1>
      <CreatorSearchPanel initialQuery={query} initialState={initialState} key={query} />
    </section>
  );
}
