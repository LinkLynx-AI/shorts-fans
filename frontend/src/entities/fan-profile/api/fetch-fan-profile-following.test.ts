import {
  fetchFanProfileFollowingPage,
  type FanProfileFollowingPage,
} from "@/entities/fan-profile";

describe("fetchFanProfileFollowingPage", () => {
  it("returns a parsed following page", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                creator: {
                  avatar: null,
                  bio: "Public shorts から paid main へつながる creator mock profile.",
                  displayName: "Mika Aoi",
                  handle: "@mikaaoi",
                  id: "creator_11111111111111111111111111111111",
                },
                viewer: {
                  isFollowing: true,
                },
              },
            ],
          },
          error: null,
          meta: {
            page: {
              hasNext: false,
              nextCursor: null,
            },
            requestId: "req_fan_profile_following_001",
          },
        }),
        { status: 200 },
      ),
    );
    const expectedPage = {
      items: [
        {
          creator: {
            avatar: null,
            bio: "Public shorts から paid main へつながる creator mock profile.",
            displayName: "Mika Aoi",
            handle: "@mikaaoi",
            id: "creator_11111111111111111111111111111111",
          },
          viewer: {
            isFollowing: true,
          },
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_fan_profile_following_001",
    } satisfies FanProfileFollowingPage;

    await expect(
      fetchFanProfileFollowingPage({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual(expectedPage);
  });

  it("forwards the cursor and session cookie", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [],
          },
          error: null,
          meta: {
            page: {
              hasNext: true,
              nextCursor: "cursor_2",
            },
            requestId: "req_fan_profile_following_002",
          },
        }),
        { status: 200 },
      ),
    );

    await fetchFanProfileFollowingPage({
      baseUrl: "https://api.example.com",
      cursor: "cursor_1",
      fetcher,
      sessionToken: "raw-session-token",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/profile/following?cursor=cursor_1",
    );

    const requestInit = fetcher.mock.calls[0]?.[1];
    const headers = new Headers(requestInit?.headers);

    expect(headers.get("Cookie")).toBe("shorts_fans_session=raw-session-token");
  });
});
