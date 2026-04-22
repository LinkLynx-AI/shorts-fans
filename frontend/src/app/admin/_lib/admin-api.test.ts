import {
  createAdminAPIFetcher,
  getAdminAPIToken,
} from "./admin-api";

describe("admin api fetcher", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("attaches the server-side admin token header", async () => {
    vi.stubEnv("ADMIN_API_TOKEN", " test-admin-token ");
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await createAdminAPIFetcher(fetcher)("https://api.example.com/admin", {
      headers: {
        Accept: "application/json",
      },
    });

    const requestInit = fetcher.mock.calls[0]?.[1];
    const headers = new Headers(requestInit?.headers);
    expect(getAdminAPIToken()).toBe("test-admin-token");
    expect(headers.get("X-Shorts-Fans-Admin-Token")).toBe("test-admin-token");
    expect(headers.get("Accept")).toBe("application/json");
  });

  it("does not attach an empty token", async () => {
    vi.stubEnv("ADMIN_API_TOKEN", "   ");
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await createAdminAPIFetcher(fetcher)("https://api.example.com/admin");

    const requestInit = fetcher.mock.calls[0]?.[1];
    const headers = new Headers(requestInit?.headers);
    expect(headers.get("X-Shorts-Fans-Admin-Token")).toBeNull();
  });
});
