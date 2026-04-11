import {
  updateCreatorWorkspaceMainPrice,
  updateCreatorWorkspaceShortCaption,
} from "./update-creator-workspace-metadata";

describe("creator workspace metadata mutation fetchers", () => {
  it("updates the owner preview main price", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await expect(
      updateCreatorWorkspaceMainPrice("main_quiet_rooftop", 2400, {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/creator/workspace/mains/main_quiet_rooftop/price"),
      expect.objectContaining({
        body: JSON.stringify({ priceJpy: 2400 }),
        credentials: "include",
        method: "PUT",
      }),
    );
  });

  it("updates the owner preview short caption", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await expect(
      updateCreatorWorkspaceShortCaption("short_quiet_rooftop", null, {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/creator/workspace/shorts/short_quiet_rooftop/caption"),
      expect.objectContaining({
        body: JSON.stringify({ caption: null }),
        credentials: "include",
        method: "PUT",
      }),
    );
  });
});
