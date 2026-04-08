import { switchViewerActiveMode } from "@/features/creator-entry";

describe("switchViewerActiveMode", () => {
  it.each(["creator", "fan"] as const)("puts %s as the next active mode with credentials included", async (activeMode) => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await expect(
      switchViewerActiveMode(activeMode, {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/viewer/active-mode"),
      expect.objectContaining({
        body: JSON.stringify({
          activeMode,
        }),
        credentials: "include",
        method: "PUT",
      }),
    );
  });
});
