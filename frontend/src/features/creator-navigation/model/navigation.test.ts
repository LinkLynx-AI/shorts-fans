import {
  buildCreatorProfileHref,
  buildCreatorShortDetailHref,
  resolveCreatorProfileBackHref,
  resolveShortDetailBackHref,
} from "@/features/creator-navigation";

describe("creator navigation", () => {
  it("builds creator profile hrefs by origin", () => {
    expect(buildCreatorProfileHref("mina", { from: "search", q: "mina" })).toBe("/creators/mina?from=search&q=mina");
    expect(buildCreatorProfileHref("mina", { from: "feed", tab: "recommended" })).toBe(
      "/creators/mina?from=feed&tab=recommended",
    );
    expect(buildCreatorProfileHref("mina", { from: "short", shortId: "rooftop" })).toBe(
      "/creators/mina?from=short&shortId=rooftop",
    );
  });

  it("resolves creator profile back hrefs", () => {
    expect(resolveCreatorProfileBackHref({ from: "search", q: "mina" })).toBe("/search?q=mina");
    expect(resolveCreatorProfileBackHref({ from: "feed", tab: "following" })).toBe("/?tab=following");
    expect(resolveCreatorProfileBackHref({ from: "short", shortId: "rooftop" })).toBe("/shorts/rooftop");
    expect(resolveCreatorProfileBackHref({})).toBe("/");
  });

  it("keeps creator context when moving between profile and short detail", () => {
    expect(
      buildCreatorShortDetailHref("mirror", "mina", {
        from: "search",
        q: "mina",
      }),
    ).toBe("/shorts/mirror?creatorId=mina&from=creator&profileFrom=search&profileQ=mina");

    expect(
      resolveShortDetailBackHref({
        creatorId: "mina",
        from: "creator",
        profileFrom: "feed",
        profileTab: "recommended",
      }),
    ).toBe("/creators/mina?from=feed&tab=recommended");

    expect(
      resolveShortDetailBackHref({
        creatorId: "mina",
        from: "creator",
        profileFrom: "short",
      }),
    ).toBe("/creators/mina");

    expect(
      resolveShortDetailBackHref({
        creatorId: "../../fan",
        from: "creator",
      }),
    ).toBe("/");
  });
});
