import { updateCreatorWorkspaceShortCaption } from "./update-creator-workspace-short-caption";

describe("updateCreatorWorkspaceShortCaption", () => {
  it("puts the short caption update with credentials included", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            short: {
              caption: "updated caption",
              id: "short_quiet_rooftop",
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_workspace_short_caption_put_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      updateCreatorWorkspaceShortCaption("short_quiet_rooftop", "updated caption", {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual({
      requestId: "req_creator_workspace_short_caption_put_001",
      short: {
        caption: "updated caption",
        id: "short_quiet_rooftop",
      },
    });

    expect(fetcher).toHaveBeenCalledTimes(1);

    const [url, init] = fetcher.mock.calls[0] ?? [];

    expect(url).toEqual(new URL("https://api.example.com/api/creator/workspace/shorts/short_quiet_rooftop/caption"));
    expect(init).toMatchObject({
      body: JSON.stringify({
        caption: "updated caption",
      }),
      credentials: "include",
      method: "PUT",
    });
    expect(new Headers(init?.headers).get("Content-Type")).toBe("application/json");
  });

  it("surfaces non-success statuses as API errors", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("server error", {
        status: 500,
      }),
    );

    await expect(
      updateCreatorWorkspaceShortCaption("short_quiet_rooftop", "updated caption", {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).rejects.toMatchObject({
      code: "http",
      status: 500,
    });
  });
});
