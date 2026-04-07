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
    expect(getCreatorIds()).toEqual([
      "creator_aoi_n",
      "creator_mina_rei",
      "creator_sora_vale",
      "creator_11111111111111111111111111111111",
      "aoi",
      "mina",
      "sora",
    ]);
    expect(listCreators()).toHaveLength(4);
  });

  it("finds creator by id", () => {
    expect(getCreatorById("mina")?.handle).toBe("@minarei");
    expect(getCreatorById("creator_mina_rei")?.handle).toBe("@minarei");
    expect(getCreatorById("mina")?.displayName).toBe("Mina Rei");
    expect(getCreatorById("creator_mina_rei")?.displayName).toBe("Mina Rei");
    expect(getCreatorProfileStatsById("mina")?.fanCount).toBe(24000);
    expect(getCreatorProfileStatsById("creator_mina_rei")?.fanCount).toBe(24000);
    expect(getCreatorProfileStatsById("sora")?.shortCount).toBe(0);
    expect(getCreatorById("unknown")).toBeUndefined();
  });

  it("derives initials from the creator name", () => {
    expect(getCreatorInitials("Mina Rei")).toBe("MR");
    expect(getCreatorInitials("Aoi")).toBe("A");
  });

  it("returns recent creators and filters by display name / handle only", () => {
    expect(getRecentCreators().map((creator) => creator.id)).toEqual(["creator_aoi_n", "creator_mina_rei"]);
    expect(searchCreators("mina").map((creator) => creator.id)).toEqual(["creator_mina_rei"]);
    expect(searchCreators("@sora").map((creator) => creator.id)).toEqual(["creator_sora_vale"]);
    expect(searchCreators("  ")).toEqual(getRecentCreators());
  });
});
