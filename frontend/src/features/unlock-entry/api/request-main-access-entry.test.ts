import { viewerSessionCookieName } from "@/entities/viewer";

import { requestMainAccessEntry } from "./request-main-access-entry";

describe("requestMainAccessEntry", () => {
  it("posts the access-entry payload and returns the playback href", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            href: "/mains/main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa?fromShortId=short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb&grant=test-grant",
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_main_access_entry_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      requestMainAccessEntry({
        acceptedAge: true,
        acceptedTerms: false,
        baseUrl: "https://api.example.com",
        entryToken: "signed-entry-token",
        fetcher,
        fromShortId: "short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
        mainId: "main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        sessionToken: "raw-session-token",
      }),
    ).resolves.toEqual({
      href: "/mains/main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa?fromShortId=short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb&grant=test-grant",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/mains/main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/access-entry",
    );
    expect(fetcher.mock.calls[0]?.[1]?.method).toBe("POST");
    expect(fetcher.mock.calls[0]?.[1]?.credentials).toBe("include");
    expect(new Headers(fetcher.mock.calls[0]?.[1]?.headers).get("Cookie")).toBe(
      `${viewerSessionCookieName}=raw-session-token`,
    );
    expect(fetcher.mock.calls[0]?.[1]?.body).toBe(
      JSON.stringify({
        entryToken: "signed-entry-token",
        fromShortId: "short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
        acceptedAge: true,
        acceptedTerms: false,
      }),
    );
  });

  it("omits optional consent fields when they are not provided", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            href: "/mains/main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa?fromShortId=short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb&grant=test-grant",
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_main_access_entry_002",
          },
        }),
        { status: 200 },
      ),
    );

    await requestMainAccessEntry({
      entryToken: "signed-entry-token",
      fetcher,
      fromShortId: "short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
      mainId: "main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    });

    expect(fetcher.mock.calls[0]?.[1]?.body).toBe(
      JSON.stringify({
        entryToken: "signed-entry-token",
        fromShortId: "short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
      }),
    );
  });
});
