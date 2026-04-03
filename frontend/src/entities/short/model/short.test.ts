import {
  getFeedShortForTab,
  getLibraryShorts,
  getPinnedShorts,
  getShortById,
  getShortIds,
  getShortThemeStyle,
  getShortsByCreatorId,
} from "@/entities/short";

describe("short model", () => {
  it("returns feed shorts for each tab", () => {
    expect(getFeedShortForTab("recommended").id).toBe("rooftop");
    expect(getFeedShortForTab("following").id).toBe("softlight");
  });

  it("returns fixture collections", () => {
    expect(getShortIds()).toContain("afterrain");
    expect(getPinnedShorts()).toHaveLength(3);
    expect(getLibraryShorts()).toHaveLength(3);
    expect(getShortsByCreatorId("sora")).toHaveLength(2);
  });

  it("builds theme style variables", () => {
    const short = getShortById("rooftop");

    if (!short) {
      throw new Error("fixture missing");
    }

    expect(getShortThemeStyle(short)).toMatchObject({
      "--short-bg-mid": "#2a648f",
      "--short-tile-top": "#d8f3ff",
    });
  });

  it("falls back to a deterministic theme for unknown short IDs", () => {
    const style = getShortThemeStyle("unknown-short-id") as Record<string, string | undefined>;

    expect(style["--short-bg-start"]).toBeTypeOf("string");
    expect(style["--short-tile-bottom"]).toBeTypeOf("string");
  });
});
