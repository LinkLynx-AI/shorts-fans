import {
  buildShortContinuationCopy,
  buildShortPaywallTitle,
  getFeedShortForTab,
  getShortById,
  getShortIds,
  getShortThemeStyle,
  getShortsByCreatorId,
  normalizeShortCaptionForTitle,
} from "@/entities/short";

describe("short model", () => {
  it("returns feed shorts for each tab", () => {
    expect(getFeedShortForTab("recommended").id).toBe("rooftop");
    expect(getFeedShortForTab("following").id).toBe("softlight");
  });

  it("returns fixture collections", () => {
    expect(getShortIds()).toContain("afterrain");
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

  it("resolves API alias IDs to the local short fixtures", () => {
    expect(getShortById("short_mina_rooftop")?.id).toBe("rooftop");
    expect(getShortById("short_aoi_softlight")?.id).toBe("softlight");
    expect(getShortThemeStyle("short_mina_rooftop")).toMatchObject({
      "--short-bg-mid": "#2a648f",
      "--short-tile-top": "#d8f3ff",
    });
  });

  it("falls back to a stable theme for unknown short ids", () => {
    const first = getShortThemeStyle("short_dbcc1756d3d9406988e6860c7348609c");
    const second = getShortThemeStyle("short_dbcc1756d3d9406988e6860c7348609c");

    expect(first).toEqual(second);
    expect(first).toMatchObject({
      "--short-bg-accent": expect.any(String),
      "--short-bg-mid": expect.any(String),
      "--short-tile-top": expect.any(String),
    });
  });

  it("normalizes caption-based copy and falls back when caption is empty", () => {
    expect(normalizeShortCaptionForTitle(" quiet rooftop preview。 ")).toBe("quiet rooftop preview");
    expect(buildShortPaywallTitle(" quiet rooftop preview。 ")).toBe("quiet rooftop preview の続きを見る");
    expect(buildShortPaywallTitle("   ")).toBe("この short の続きを見る");
    expect(buildShortContinuationCopy(" quiet rooftop preview。 ")).toBe("quiet rooftop preview の続き。");
    expect(buildShortContinuationCopy("   ")).toBe("short の続きから再生中。");
  });
});
