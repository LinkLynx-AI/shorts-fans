import {
  buildCreatorProfileHref,
  buildCreatorShortDetailHref,
  buildFanProfileLibraryMainHref,
  buildFanProfileShortDetailHref,
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
    expect(buildCreatorProfileHref("mina", { from: "short", shortFanTab: "pinned", shortId: "rooftop" })).toBe(
      "/creators/mina?from=short&shortFanTab=pinned&shortId=rooftop",
    );
  });

  it("resolves creator profile back hrefs", () => {
    expect(resolveCreatorProfileBackHref({ from: "search", q: "mina" })).toBe("/search?q=mina");
    expect(resolveCreatorProfileBackHref({ from: "feed", tab: "following" })).toBe("/?tab=following");
    expect(resolveCreatorProfileBackHref({ from: "short", shortId: "rooftop" })).toBe("/shorts/rooftop");
    expect(resolveCreatorProfileBackHref({ from: "short", shortFanTab: "pinned", shortId: "rooftop" })).toBe(
      "/shorts/rooftop?fanTab=pinned&from=fan",
    );
    expect(resolveCreatorProfileBackHref({})).toBe("/");
  });

  it("keeps creator context when moving between profile and short detail", () => {
    expect(
      buildCreatorShortDetailHref("mirror", "mina", {
        from: "short",
        shortFanTab: "pinned",
        shortId: "rooftop",
      }),
    ).toBe("/shorts/mirror?creatorId=mina&from=creator&profileFrom=short&profileShortFanTab=pinned&profileShortId=rooftop");

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
        profileShortFanTab: "pinned",
        profileFrom: "short",
        profileShortId: "rooftop",
      }),
    ).toBe("/creators/mina?from=short&shortFanTab=pinned&shortId=rooftop");

    expect(
      resolveShortDetailBackHref({
        creatorId: "../../fan",
        from: "creator",
      }),
    ).toBe("/");
  });

  it("returns to fan profile when short detail came from pinned shorts", () => {
    expect(buildFanProfileShortDetailHref("mirror", "pinned")).toBe("/shorts/mirror?fanTab=pinned&from=fan");
    expect(buildFanProfileLibraryMainHref("main_mina_rooftop", "mirror")).toBe(
      "/mains/main_mina_rooftop?fanTab=library&from=fan&fromShortId=mirror",
    );

    expect(
      resolveShortDetailBackHref({
        fanTab: "pinned",
        from: "fan",
      }),
    ).toBe("/fan?tab=pinned");
  });
});
