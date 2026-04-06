import { getCurrentViewerBootstrap } from "@/entities/viewer/api";

describe("getCurrentViewerBootstrap", () => {
  it("returns the authenticated current viewer", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            currentViewer: {
              activeMode: "fan",
              canAccessCreatorMode: false,
              id: "viewer_123",
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_123",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getCurrentViewerBootstrap({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual({
      activeMode: "fan",
      canAccessCreatorMode: false,
      id: "viewer_123",
    });
  });

  it("returns null for unauthenticated bootstrap", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            currentViewer: null,
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_123",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getCurrentViewerBootstrap({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toBeNull();
  });

  it("forwards the session cookie when a token is provided", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            currentViewer: null,
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_123",
          },
        }),
        { status: 200 },
      ),
    );

    await getCurrentViewerBootstrap({
      baseUrl: "https://api.example.com",
      fetcher,
      sessionToken: "raw-session-token",
    });

    const requestInit = fetcher.mock.calls[0]?.[1];
    const headers = new Headers(requestInit?.headers);

    expect(headers.get("Cookie")).toBe("shorts_fans_session=raw-session-token");
  });
});
