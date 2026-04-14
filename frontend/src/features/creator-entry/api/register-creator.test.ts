import { registerCreator } from "@/features/creator-entry";

describe("registerCreator", () => {
  it("posts an empty creator registration payload with credentials included", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await expect(
      registerCreator({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/viewer/creator-registration"),
      expect.objectContaining({
        body: JSON.stringify({}),
        credentials: "include",
        method: "POST",
      }),
    );
  });
});
