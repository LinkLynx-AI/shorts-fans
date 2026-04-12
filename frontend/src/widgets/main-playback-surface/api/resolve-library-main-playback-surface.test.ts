import { getFanHubState } from "@/entities/fan-profile";

import { resolveLibraryMainPlaybackSurface } from "./resolve-library-main-playback-surface";

describe("resolveLibraryMainPlaybackSurface", () => {
  it("builds a local playback surface for legacy library items", async () => {
    const item = getFanHubState("library").libraryItems[1];

    if (!item) {
      throw new Error("library item fixture missing");
    }

    await expect(resolveLibraryMainPlaybackSurface(item)).resolves.toMatchObject({
      access: {
        mainId: item.main.id,
        status: "unlocked",
      },
      entryShort: {
        id: item.entryShort.id,
      },
      main: {
        id: item.main.id,
      },
    });
  });
});
