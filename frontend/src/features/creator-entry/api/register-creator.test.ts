import { registerCreator } from "@/features/creator-entry";

describe("registerCreator", () => {
  it("posts the creator registration payload with credentials included", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await expect(
      registerCreator(
        {
          bio: "quiet rooftop の continuation を中心に投稿します。",
          displayName: "Mina Rei",
          handle: "@mina.rei",
        },
        {
          baseUrl: "https://api.example.com",
          fetcher,
        },
      ),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/viewer/creator-registration"),
      expect.objectContaining({
        body: JSON.stringify({
          bio: "quiet rooftop の continuation を中心に投稿します。",
          displayName: "Mina Rei",
          handle: "@mina.rei",
        }),
        credentials: "include",
        method: "POST",
      }),
    );
  });

  it("includes avatarUploadToken when a completed avatar upload exists", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await expect(
      registerCreator(
        {
          avatarUploadToken: "vcupl_token",
          bio: "quiet rooftop の continuation を中心に投稿します。",
          displayName: "Mina Rei",
          handle: "@mina.rei",
        },
        {
          baseUrl: "https://api.example.com",
          fetcher,
        },
      ),
    ).resolves.toBeUndefined();

    const [calledURL, calledInit] = fetcher.mock.calls[0] ?? [];

    expect(calledURL).toEqual(new URL("https://api.example.com/api/viewer/creator-registration"));
    expect(JSON.parse(String(calledInit?.body))).toEqual({
      avatarUploadToken: "vcupl_token",
      bio: "quiet rooftop の continuation を中心に投稿します。",
      displayName: "Mina Rei",
      handle: "@mina.rei",
    });
  });
});
