import { CreatorSearchPanel } from "@/features/creator-search";

type SearchShellProps = {
  query: string;
};

/**
 * creator search 用の route shell を表示する。
 */
export function SearchShell({ query }: SearchShellProps) {
  return (
    <section className="min-h-full overflow-y-auto px-4 pb-28 pt-4 text-foreground">
      <h1 className="sr-only">Creator search</h1>
      <CreatorSearchPanel initialQuery={query} key={query} />
    </section>
  );
}
