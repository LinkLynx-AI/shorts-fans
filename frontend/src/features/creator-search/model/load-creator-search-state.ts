import { getCreatorSearchResults } from "@/entities/creator";

import {
  buildErrorCreatorSearchState,
  type CreatorSearchState,
  normalizeCreatorSearchQuery,
} from "./creator-search-state";

type LoadCreatorSearchStateOptions = {
  signal?: AbortSignal | undefined;
};

/**
 * creator search API response を panel state に変換する。
 */
export async function loadCreatorSearchState(
  query: string,
  options: LoadCreatorSearchStateOptions = {},
): Promise<CreatorSearchState> {
  const normalizedQuery = normalizeCreatorSearchQuery(query);

  try {
    const response = await getCreatorSearchResults({
      query: normalizedQuery,
      signal: options.signal,
    });

    if (response.items.length === 0) {
      return {
        items: [],
        kind: "empty",
        query: response.query,
      };
    }

    return {
      items: response.items,
      kind: "ready",
      query: response.query,
    };
  } catch {
    return buildErrorCreatorSearchState(normalizedQuery);
  }
}
