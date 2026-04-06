import { z } from "zod";

import { loadCreatorSearchState } from "@/features/creator-search";
import { getSingleQueryParam } from "@/shared/lib";
import { SearchShell } from "@/widgets/search-shell";

const searchParamsSchema = z.object({
  q: z.string().optional().default(""),
});

export default async function SearchPage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string | string[] }>;
}) {
  const rawSearchParams = await searchParams;
  const { q } = searchParamsSchema.parse({
    q: getSingleQueryParam(rawSearchParams.q),
  });
  const initialState = await loadCreatorSearchState(q);

  return <SearchShell initialState={initialState} query={q} />;
}
