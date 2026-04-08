import { fetchFanProfileOverview } from "@/entities/fan-profile";

describe("fetchFanProfileOverview", () => {
  it("returns the overview payload from the API", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            fanProfile: {
              counts: {
                following: 3,
                library: 2,
                pinnedShorts: 1,
              },
              title: "My archive",
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_fan_profile_overview_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      fetchFanProfileOverview({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual({
      counts: {
        following: 3,
        library: 2,
        pinnedShorts: 1,
      },
      title: "My archive",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe("https://api.example.com/api/fan/profile");
  });

  it("forwards the session cookie when a token is provided", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            fanProfile: {
              counts: {
                following: 0,
                library: 0,
                pinnedShorts: 0,
              },
              title: "My archive",
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_fan_profile_overview_002",
          },
        }),
        { status: 200 },
      ),
    );

    await fetchFanProfileOverview({
      baseUrl: "https://api.example.com",
      fetcher,
      sessionToken: "raw-session-token",
    });

    const requestInit = fetcher.mock.calls[0]?.[1];
    const headers = new Headers(requestInit?.headers);

    expect(headers.get("Cookie")).toBe("shorts_fans_session=raw-session-token");
  });
});
