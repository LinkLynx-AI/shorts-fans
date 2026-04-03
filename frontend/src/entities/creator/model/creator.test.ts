import {
  getCreatorById,
  getCreatorIds,
  getCreatorInitials,
  listCreators,
} from "@/entities/creator";

describe("creator model", () => {
  it("returns all creator ids", () => {
    expect(getCreatorIds()).toEqual(["aoi", "mina", "sora"]);
    expect(listCreators()).toHaveLength(3);
  });

  it("finds creator by id", () => {
    expect(getCreatorById("mina")?.handle).toBe("@minarei");
    expect(getCreatorById("unknown")).toBeUndefined();
  });

  it("derives initials from the creator name", () => {
    expect(getCreatorInitials("Mina Rei")).toBe("MR");
    expect(getCreatorInitials("Aoi")).toBe("A");
  });
});
