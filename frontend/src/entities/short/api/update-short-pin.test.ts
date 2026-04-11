import { ApiError } from "@/shared/api";

import {
  ShortPinApiError,
  updateShortPin,
} from "./update-short-pin";

describe("updateShortPin", () => {
  it("sends a pin request with include credentials and returns the post-condition state", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            viewer: {
              isPinned: true,
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_short_pin_put_success_001",
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
      updateShortPin({
        action: "pin",
        baseUrl: "http://127.0.0.1:3201",
        fetcher,
        shortId: "short_rooftop",
      }),
    ).resolves.toEqual({
      viewer: {
        isPinned: true,
      },
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("http://127.0.0.1:3201/api/fan/shorts/short_rooftop/pin"),
      {
        credentials: "include",
        headers: {
          Accept: "application/json",
        },
        method: "PUT",
      },
    );
  });

  it("sends DELETE when unpinning", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            viewer: {
              isPinned: false,
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_short_pin_delete_success_001",
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

    await updateShortPin({
      action: "unpin",
      baseUrl: "http://127.0.0.1:3201",
      credentials: "same-origin",
      fetcher,
      shortId: "short_softlight",
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("http://127.0.0.1:3201/api/fan/shorts/short_softlight/pin"),
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
            message: "short pin requires authentication",
          },
          meta: {
            page: null,
            requestId: "req_short_pin_put_auth_required_001",
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
      updateShortPin({
        action: "pin",
        baseUrl: "http://127.0.0.1:3201",
        fetcher,
        shortId: "short_rooftop",
      }),
    ).rejects.toEqual(
      new ShortPinApiError("auth_required", "short pin requires authentication", {
        requestId: "req_short_pin_put_auth_required_001",
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
      updateShortPin({
        action: "pin",
        baseUrl: "http://127.0.0.1:3201",
        fetcher,
        shortId: "short_rooftop",
      }),
    ).rejects.toBeInstanceOf(ApiError);
  });
});
