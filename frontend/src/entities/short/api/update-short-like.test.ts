import { ApiError } from "@/shared/api";

import {
  ShortLikeApiError,
  updateShortLike,
} from "./update-short-like";

describe("updateShortLike", () => {
  it("sends a like request with include credentials and returns the post-condition state", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            engagement: {
              likeCount: 42,
            },
            viewer: {
              hasLiked: true,
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_short_like_put_success_001",
          },
        }),
        {
          headers: {
            "Content-Type": "application/json",
          },
          status: 200,
        },
      ),
    );

    await expect(
      updateShortLike({
        action: "like",
        baseUrl: "http://127.0.0.1:3201",
        fetcher,
        shortId: "short_rooftop",
      }),
    ).resolves.toEqual({
      engagement: {
        likeCount: 42,
      },
      viewer: {
        hasLiked: true,
      },
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("http://127.0.0.1:3201/api/fan/shorts/short_rooftop/like"),
      {
        credentials: "include",
        headers: {
          Accept: "application/json",
        },
        method: "PUT",
      },
    );
  });

  it("sends DELETE when unliking", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            engagement: {
              likeCount: 41,
            },
            viewer: {
              hasLiked: false,
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_short_like_delete_success_001",
          },
        }),
        {
          headers: {
            "Content-Type": "application/json",
          },
          status: 200,
        },
      ),
    );

    await updateShortLike({
      action: "unlike",
      baseUrl: "http://127.0.0.1:3201",
      credentials: "same-origin",
      fetcher,
      shortId: "short_softlight",
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("http://127.0.0.1:3201/api/fan/shorts/short_softlight/like"),
      {
        credentials: "same-origin",
        headers: {
          Accept: "application/json",
        },
        method: "DELETE",
      },
    );
  });

  it("throws a contract error when the API returns auth_required", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: null,
          error: {
            code: "auth_required",
            message: "short like requires authentication",
          },
          meta: {
            page: null,
            requestId: "req_short_like_put_auth_required_001",
          },
        }),
        {
          headers: {
            "Content-Type": "application/json",
          },
          status: 401,
        },
      ),
    );

    await expect(
      updateShortLike({
        action: "like",
        baseUrl: "http://127.0.0.1:3201",
        fetcher,
        shortId: "short_rooftop",
      }),
    ).rejects.toEqual(
      new ShortLikeApiError("auth_required", "short like requires authentication", {
        requestId: "req_short_like_put_auth_required_001",
        status: 401,
      }),
    );
  });

  it("throws ApiError when a non-success response body does not match the contract", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("server exploded", {
        headers: {
          "Content-Type": "text/plain",
        },
        status: 500,
      }),
    );

    await expect(
      updateShortLike({
        action: "like",
        baseUrl: "http://127.0.0.1:3201",
        fetcher,
        shortId: "short_rooftop",
      }),
    ).rejects.toBeInstanceOf(ApiError);
  });
});
