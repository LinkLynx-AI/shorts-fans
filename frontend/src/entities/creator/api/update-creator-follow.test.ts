import { ApiError } from "@/shared/api";

import {
  CreatorFollowApiError,
  updateCreatorFollow,
} from "./update-creator-follow";

describe("updateCreatorFollow", () => {
  it("sends a follow request with include credentials and returns the post-condition state", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            stats: {
              fanCount: 11,
            },
            viewer: {
              isFollowing: true,
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_follow_put_success_001",
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
      updateCreatorFollow({
        action: "follow",
        baseUrl: "http://127.0.0.1:3201",
        creatorId: "creator_aoi_n",
        fetcher,
      }),
    ).resolves.toEqual({
      stats: {
        fanCount: 11,
      },
      viewer: {
        isFollowing: true,
      },
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("http://127.0.0.1:3201/api/fan/creators/creator_aoi_n/follow"),
      {
        credentials: "include",
        headers: {
          Accept: "application/json",
        },
        method: "PUT",
      },
    );
  });

  it("sends DELETE when unfollowing", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            stats: {
              fanCount: 10,
            },
            viewer: {
              isFollowing: false,
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_follow_delete_success_001",
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

    await updateCreatorFollow({
      action: "unfollow",
      baseUrl: "http://127.0.0.1:3201",
      creatorId: "creator_mina_rei",
      credentials: "same-origin",
      fetcher,
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("http://127.0.0.1:3201/api/fan/creators/creator_mina_rei/follow"),
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
            message: "creator follow requires authentication",
          },
          meta: {
            page: null,
            requestId: "req_creator_follow_put_auth_required_001",
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
      updateCreatorFollow({
        action: "follow",
        baseUrl: "http://127.0.0.1:3201",
        creatorId: "creator_aoi_n",
        fetcher,
      }),
    ).rejects.toEqual(
      new CreatorFollowApiError("auth_required", "creator follow requires authentication", {
        requestId: "req_creator_follow_put_auth_required_001",
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
      updateCreatorFollow({
        action: "follow",
        baseUrl: "http://127.0.0.1:3201",
        creatorId: "creator_aoi_n",
        fetcher,
      }),
    ).rejects.toBeInstanceOf(ApiError);
  });
});
