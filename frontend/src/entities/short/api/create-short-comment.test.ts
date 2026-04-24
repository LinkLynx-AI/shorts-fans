import { ApiError } from "@/shared/api";

import {
  createShortComment,
  ShortCommentApiError,
} from "./create-short-comment";

describe("createShortComment", () => {
  it("posts a comment body and returns the created comment", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            comment: {
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
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_short_comment_create_001",
          },
        }),
        { status: 201 },
      ),
    );

    await expect(
      createShortComment({
        baseUrl: "https://api.example.com",
        body: "hello",
        fetcher,
        shortId: "short_11111111111111111111111111111111",
      }),
    ).resolves.toMatchObject({
      body: "hello",
      id: "comment_22222222222222222222222222222222",
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/fan/shorts/short_11111111111111111111111111111111/comments"),
      {
        body: JSON.stringify({ body: "hello" }),
        credentials: "include",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        method: "POST",
      },
    );
  });

  it("throws a contract error when the API returns validation_error", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: null,
          error: {
            code: "validation_error",
            message: "comment body is invalid",
          },
          meta: {
            page: null,
            requestId: "req_short_comment_validation_001",
          },
        }),
        { status: 400 },
      ),
    );

    await expect(
      createShortComment({
        baseUrl: "https://api.example.com",
        body: " ",
        fetcher,
        shortId: "short_11111111111111111111111111111111",
      }),
    ).rejects.toEqual(
      new ShortCommentApiError("validation_error", "comment body is invalid", {
        requestId: "req_short_comment_validation_001",
        status: 400,
      }),
    );
  });

  it("rejects image avatar payloads with non-null posterUrl", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            comment: {
              author: {
                avatar: {
                  durationSeconds: null,
                  id: "avatar_11111111111111111111111111111111",
                  kind: "image",
                  posterUrl: "https://cdn.example.com/avatar-poster.jpg",
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
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_short_comment_create_001",
          },
        }),
        { status: 201 },
      ),
    );

    await expect(
      createShortComment({
        baseUrl: "https://api.example.com",
        body: "hello",
        fetcher,
        shortId: "short_11111111111111111111111111111111",
      }),
    ).rejects.toMatchObject({
      code: "parse",
      status: 201,
    });
  });

  it("throws ApiError when an error response does not match the contract", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("server exploded", {
        headers: {
          "Content-Type": "text/plain",
        },
        status: 500,
      }),
    );

    await expect(
      createShortComment({
        baseUrl: "https://api.example.com",
        body: "hello",
        fetcher,
        shortId: "short_11111111111111111111111111111111",
      }),
    ).rejects.toBeInstanceOf(ApiError);
  });
});
