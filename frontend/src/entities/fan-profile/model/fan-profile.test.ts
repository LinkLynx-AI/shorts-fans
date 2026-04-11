import {
  getFanHubState,
  getFanProfileOverview,
  listFanSettingsSections,
  listFollowingItems,
  normalizeFanHubTab,
} from "@/entities/fan-profile";

describe("fan profile model", () => {
  it("normalizes the hub tab", () => {
    expect(normalizeFanHubTab("library")).toBe("library");
    expect(normalizeFanHubTab("pinned")).toBe("pinned");
    expect(normalizeFanHubTab("unknown")).toBe("pinned");
  });

  it("returns overview and collection state", () => {
    const overview = getFanProfileOverview();
    const libraryState = getFanHubState("library");

    expect(overview.title).toBe("My archive");
    expect(overview.counts).toEqual({
      following: 3,
      library: 3,
      pinnedShorts: 3,
    });
    expect(libraryState.activeTab).toBe("library");
    expect(libraryState.libraryItems[0]?.main.id).toBe("main_aoi_soft_light");
    expect(libraryState.pinnedItems).toHaveLength(3);
  });

  it("returns following rows and settings sections", () => {
    expect(listFollowingItems()).toHaveLength(3);
    expect(listFollowingItems()[1]?.creator.handle).toBe("@minarei");
    expect(listFanSettingsSections()).toEqual([
      { available: true, key: "account", label: "Account" },
      { available: true, key: "payment", label: "Payment" },
      { available: true, key: "safety", label: "Safety" },
    ]);
  });
});
