import type { CreatorSummary } from "@/entities/creator";

import {
  consumePendingCreatorSearchHistorySelection,
  createCreatorSearchHistoryScope,
  creatorSearchHistoryLimit,
  creatorSearchHistoryPendingNavigationKey,
  markPendingCreatorSearchHistorySelection,
  readCreatorSearchHistory,
  recordCreatorSearchHistory,
} from "./creator-search-history";

const guestHistoryScope = createCreatorSearchHistoryScope({
  hasViewerSession: false,
  viewerId: null,
});

function createCreator(index: number): CreatorSummary {
  return {
    avatar: null,
    bio: `creator ${index} bio`,
    displayName: `Creator ${index}`,
    handle: `@creator${index}`,
    id: `creator_test_${index}`,
  };
}

describe("creator search history", () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
  });

  it("returns an empty history when the stored value is missing or invalid", () => {
    expect(readCreatorSearchHistory({ scope: guestHistoryScope })).toEqual([]);

    expect(guestHistoryScope).not.toBeNull();
    window.sessionStorage.setItem(guestHistoryScope?.storageKey ?? "", JSON.stringify([{ id: 1 }]));

    expect(readCreatorSearchHistory({ scope: guestHistoryScope })).toEqual([]);
    expect(window.sessionStorage.getItem(guestHistoryScope?.storageKey ?? "")).toBeNull();
  });

  it("deduplicates creators and moves a revisited creator to the front", () => {
    const mina = createCreator(1);
    const aoi = createCreator(2);

    recordCreatorSearchHistory(mina, { scope: guestHistoryScope });
    recordCreatorSearchHistory(aoi, { scope: guestHistoryScope });
    recordCreatorSearchHistory(mina, { scope: guestHistoryScope });

    expect(readCreatorSearchHistory({ scope: guestHistoryScope }).map((creator) => creator.id)).toEqual([
      mina.id,
      aoi.id,
    ]);
  });

  it("keeps only the most recent creators up to the history cap", () => {
    for (let index = 0; index < creatorSearchHistoryLimit + 2; index += 1) {
      recordCreatorSearchHistory(createCreator(index), { scope: guestHistoryScope });
    }

    const history = readCreatorSearchHistory({ scope: guestHistoryScope });

    expect(history).toHaveLength(creatorSearchHistoryLimit);
    expect(history[0]?.id).toBe("creator_test_11");
    expect(history.at(-1)?.id).toBe("creator_test_2");
  });

  it("consumes a pending creator selection only once when creator and query match", () => {
    markPendingCreatorSearchHistorySelection(
      {
        creatorId: "creator_test_1",
        query: "mina",
      },
      window.sessionStorage,
    );

    expect(
      consumePendingCreatorSearchHistorySelection(
        {
          creatorId: "creator_test_1",
          query: "mina",
        },
        window.sessionStorage,
      ),
    ).toBe(true);
    expect(window.sessionStorage.getItem(creatorSearchHistoryPendingNavigationKey)).toBeNull();
  });
});
