import {
  getCreatorById,
  getCreatorIds,
  getCreatorInitials,
  getCreatorProfileStatsById,
  getRecentCreators,
  listCreators,
  searchCreators,
} from "@/entities/creator";

describe("creator model", () => {
  it("returns all creator ids", () => {
    expect(getCreatorIds()).toEqual(["aoi", "mina", "sora"]);
    expect(listCreators()).toHaveLength(3);
  });

  it("finds creator by id", () => {
    expect(getCreatorById("mina")?.handle).toBe("@minarei");
    expect(getCreatorById("mina")?.displayName).toBe("Mina Rei");
    expect(getCreatorProfileStatsById("mina")?.fanCount).toBe(24000);
    expect(getCreatorProfileStatsById("sora")?.shortCount).toBe(0);
    expect(getCreatorById("unknown")).toBeUndefined();
  });

  it("derives initials from the creator name", () => {
    expect(getCreatorInitials("Mina Rei")).toBe("MR");
    expect(getCreatorInitials("Aoi")).toBe("A");
  });

  it("returns recent creators and filters by display name / handle only", () => {
    expect(getRecentCreators().map((creator) => creator.id)).toEqual(["aoi", "mina"]);
    expect(searchCreators("mina").map((creator) => creator.id)).toEqual(["mina"]);
    expect(searchCreators("@sora").map((creator) => creator.id)).toEqual(["sora"]);
    expect(searchCreators("  ")).toEqual(getRecentCreators());
  });
});
