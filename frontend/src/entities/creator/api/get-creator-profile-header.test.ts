import { getCreatorProfileHeader } from "@/entities/creator";

describe("getCreatorProfileHeader", () => {
  it("requests the creator profile header and forwards the viewer session cookie", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            profile: {
              creator: {
                avatar: null,
                bio: "quiet rooftop と hotel light の preview を軸に投稿。",
                displayName: "Mina Rei",
                handle: "@minarei",
                id: "creator_mina_rei",
              },
              stats: {
                fanCount: 24000,
                shortCount: 2,
              },
              viewer: {
                isFollowing: true,
              },
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_profile_header_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getCreatorProfileHeader({
        baseUrl: "https://api.example.com",
        creatorId: "creator_mina_rei",
        fetcher,
        sessionToken: "raw-session-token",
      }),
    ).resolves.toEqual({
      creator: {
        avatar: null,
        bio: "quiet rooftop と hotel light の preview を軸に投稿。",
        displayName: "Mina Rei",
        handle: "@minarei",
        id: "creator_mina_rei",
      },
      stats: {
        fanCount: 24000,
        shortCount: 2,
      },
      viewer: {
        isFollowing: true,
      },
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe("https://api.example.com/api/fan/creators/creator_mina_rei");
    expect(fetcher.mock.calls[0]?.[1]?.headers).toBeInstanceOf(Headers);
    expect(new Headers(fetcher.mock.calls[0]?.[1]?.headers).get("Cookie")).toBe("shorts_fans_session=raw-session-token");
  });
});
