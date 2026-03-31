import { z } from "zod";

import { createApiUrl, requestJson } from "@/shared/api";

describe("createApiUrl", () => {
  it("joins the API base URL and path", () => {
    expect(createApiUrl("https://api.example.com", "/v1/feed").toString()).toBe(
      "https://api.example.com/v1/feed",
    );
  });
});

describe("requestJson", () => {
  it("parses successful JSON responses", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), {
        status: 200,
      }),
    );

    await expect(
      requestJson({
        baseUrl: "https://api.example.com",
        fetcher,
        path: "/health",
        schema: z.object({
          ok: z.boolean(),
        }),
      }),
    ).resolves.toEqual({
      ok: true,
    });

    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it("accepts a URL instance as the path input", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), {
        status: 200,
      }),
    );

    await requestJson({
      baseUrl: "https://api.example.com",
      fetcher,
      path: new URL("https://api.example.com/direct"),
      schema: z.object({
        ok: z.boolean(),
      }),
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe("https://api.example.com/direct");
  });

  it("throws a http error for non-success statuses", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("server error", {
        status: 500,
      }),
    );

    await expect(
      requestJson({
        baseUrl: "https://api.example.com",
        fetcher,
        path: "/health",
        schema: z.object({
          ok: z.boolean(),
        }),
      }),
    ).rejects.toMatchObject({
      code: "http",
      status: 500,
    });
  });

  it("throws a network error when fetch rejects", async () => {
    const fetcher = vi.fn<typeof fetch>().mockRejectedValue(new Error("boom"));

    await expect(
      requestJson({
        baseUrl: "https://api.example.com",
        fetcher,
        path: "/health",
        schema: z.object({
          ok: z.boolean(),
        }),
      }),
    ).rejects.toMatchObject({
      code: "network",
    });
  });

  it("throws a parse error when the response shape does not match", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify({ ok: "yes" }), {
        status: 200,
      }),
    );

    await expect(
      requestJson({
        baseUrl: "https://api.example.com",
        fetcher,
        path: "/health",
        schema: z.object({
          ok: z.boolean(),
        }),
      }),
    ).rejects.toMatchObject({
      code: "parse",
    });
  });

  it("throws a parse error when the body is not valid JSON", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("not-json", {
        headers: {
          "Content-Type": "application/json",
        },
        status: 200,
      }),
    );

    await expect(
      requestJson({
        baseUrl: "https://api.example.com",
        fetcher,
        path: "/health",
        schema: z.object({
          ok: z.boolean(),
        }),
      }),
    ).rejects.toMatchObject({
      code: "parse",
    });
  });
});
