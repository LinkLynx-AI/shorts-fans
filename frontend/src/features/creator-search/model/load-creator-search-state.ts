import { getCreatorSearchResults } from "@/entities/creator";

import {
  buildEmptyCreatorSearchState,
  buildErrorCreatorSearchState,
  buildReadyCreatorSearchState,
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
      return buildEmptyCreatorSearchState(response.query);
    }

    return buildReadyCreatorSearchState(response.query, response.items);
  } catch {
    return buildErrorCreatorSearchState(normalizedQuery);
  }
}
