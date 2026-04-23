import { ApiError } from "@/shared/api";

import { getShortComments } from "./get-short-comments";
import { ShortCommentApiError } from "./short-comment-error";

describe("getShortComments", () => {
  it("requests the first page without an empty cursor query", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [],
          },
          error: null,
          meta: {
            page: {
              hasNext: false,
              nextCursor: null,
            },
            requestId: "req_short_comments_first_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getShortComments({
        baseUrl: "https://api.example.com",
        fetcher,
        shortId: "short_11111111111111111111111111111111",
      }),
    ).resolves.toEqual({
      items: [],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_short_comments_first_001",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/shorts/short_11111111111111111111111111111111/comments",
    );
  });

  it("requests a cursor page for a public short", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                author: {
                  avatar: null,
                  displayName: "Kana Mori",
                  handle: "@kanamori",
                },
                body: "hello",
                createdAt: "2026-01-02T14:05:00Z",
                id: "comment_22222222222222222222222222222222",
                shortId: "short_11111111111111111111111111111111",
              },
            ],
          },
          error: null,
          meta: {
            page: {
              hasNext: true,
              nextCursor: "cursor_next_001",
            },
            requestId: "req_short_comments_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getShortComments({
        baseUrl: "https://api.example.com",
        cursor: "cursor_prev_001",
        fetcher,
        shortId: "short_11111111111111111111111111111111",
      }),
    ).resolves.toEqual({
      items: [
        {
          author: {
            avatar: null,
            displayName: "Kana Mori",
            handle: "@kanamori",
          },
          body: "hello",
          createdAt: "2026-01-02T14:05:00Z",
          id: "comment_22222222222222222222222222222222",
          shortId: "short_11111111111111111111111111111111",
        },
      ],
      page: {
        hasNext: true,
        nextCursor: "cursor_next_001",
      },
      requestId: "req_short_comments_001",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/shorts/short_11111111111111111111111111111111/comments?cursor=cursor_prev_001",
    );
    expect(fetcher.mock.calls[0]?.[1]?.credentials).toBe("include");
    expect(fetcher.mock.calls[0]?.[1]?.cache).toBe("no-store");
  });

  it("throws a contract error when the API returns not_found", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: null,
          error: {
            code: "not_found",
            message: "short was not found",
          },
          meta: {
            page: null,
            requestId: "req_short_comments_not_found_001",
          },
        }),
        { status: 404 },
      ),
    );

    await expect(
      getShortComments({
        baseUrl: "https://api.example.com",
        fetcher,
        shortId: "short_missing",
      }),
    ).rejects.toEqual(
      new ShortCommentApiError("not_found", "short was not found", {
        requestId: "req_short_comments_not_found_001",
        status: 404,
      }),
    );
  });

  it("rejects image avatar payloads with posterUrl", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                author: {
                  avatar: {
                    durationSeconds: null,
                    id: "asset_short_comment_author_avatar_001",
                    kind: "image",
                    posterUrl: "https://cdn.example.com/poster.jpg",
                    url: "https://cdn.example.com/avatar.jpg",
                  },
                  displayName: "Kana Mori",
                  handle: "@kanamori",
                },
                body: "hello",
                createdAt: "2026-01-02T14:05:00Z",
                id: "comment_22222222222222222222222222222222",
                shortId: "short_11111111111111111111111111111111",
              },
            ],
          },
          error: null,
          meta: {
            page: {
              hasNext: false,
              nextCursor: null,
            },
            requestId: "req_short_comments_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getShortComments({
        baseUrl: "https://api.example.com",
        fetcher,
        shortId: "short_11111111111111111111111111111111",
      }),
    ).rejects.toBeInstanceOf(ApiError);
  });

  it("rejects non-RFC3339 createdAt payloads", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                author: {
                  avatar: null,
                  displayName: "Kana Mori",
                  handle: "@kanamori",
                },
                body: "hello",
                createdAt: "2026/01/02 14:05",
                id: "comment_22222222222222222222222222222222",
                shortId: "short_11111111111111111111111111111111",
              },
            ],
          },
          error: null,
          meta: {
            page: {
              hasNext: false,
              nextCursor: null,
            },
            requestId: "req_short_comments_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getShortComments({
        baseUrl: "https://api.example.com",
        fetcher,
        shortId: "short_11111111111111111111111111111111",
      }),
    ).rejects.toBeInstanceOf(ApiError);
  });
});
