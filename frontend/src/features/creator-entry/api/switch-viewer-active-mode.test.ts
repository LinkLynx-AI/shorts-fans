import { switchViewerActiveMode } from "@/features/creator-entry";

describe("switchViewerActiveMode", () => {
  it("puts the next active mode with credentials included", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await expect(
      switchViewerActiveMode("creator", {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/viewer/active-mode"),
      expect.objectContaining({
        body: JSON.stringify({
          activeMode: "creator",
        }),
        credentials: "include",
        method: "PUT",
      }),
    );
  });
});
