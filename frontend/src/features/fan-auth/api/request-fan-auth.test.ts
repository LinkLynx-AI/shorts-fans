import { authenticateFanWithEmail } from "@/features/fan-auth";

describe("authenticateFanWithEmail", () => {
  it("issues a challenge and starts a session with credentials included", async () => {
    const fetcher = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              challengeToken: "challenge-token",
              expiresAt: "2026-04-07T21:00:00+09:00",
            },
            error: null,
            meta: {
              page: null,
              requestId: "req_sign_in_challenge_001",
            },
          }),
          {
            headers: {
              "Content-Type": "application/json",
            },
            status: 200,
          },
        ),
      )
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    await expect(
      authenticateFanWithEmail({
        email: "fan@example.com",
        mode: "sign-in",
      }, {
        baseUrl: "http://127.0.0.1:8080",
        fetcher,
      }),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenNthCalledWith(
      1,
      new URL("http://127.0.0.1:8080/api/fan/auth/sign-in/challenges"),
      expect.objectContaining({
        body: JSON.stringify({
          email: "fan@example.com",
        }),
        credentials: "include",
        method: "POST",
      }),
    );
    expect(fetcher).toHaveBeenNthCalledWith(
      2,
      new URL("http://127.0.0.1:8080/api/fan/auth/sign-in/session"),
      expect.objectContaining({
        body: JSON.stringify({
          challengeToken: "challenge-token",
          email: "fan@example.com",
        }),
        credentials: "include",
        method: "POST",
      }),
    );
  });

  it("surfaces contract errors from the challenge boundary", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          data: null,
          error: {
            code: "email_already_registered",
            message: "email is already registered",
          },
          meta: {
            page: null,
            requestId: "req_sign_up_challenge_001",
          },
        }),
        {
          headers: {
            "Content-Type": "application/json",
          },
          status: 409,
        },
      ),
    );

    await expect(
      authenticateFanWithEmail({
        displayName: "Mina",
        email: "fan@example.com",
        handle: "mina",
        mode: "sign-up",
      }, {
        baseUrl: "http://127.0.0.1:8080",
        fetcher,
      }),
    ).rejects.toEqual(
      expect.objectContaining({
        code: "email_already_registered",
        message: "email is already registered",
        name: "FanAuthApiError",
        requestId: "req_sign_up_challenge_001",
        status: 409,
      }),
    );
  });

  it("surfaces malformed transport failures as shared API errors", async () => {
    const fetcher = vi.fn<typeof fetch>().mockRejectedValueOnce(new Error("network down"));

    await expect(
      authenticateFanWithEmail({
        email: "fan@example.com",
        mode: "sign-in",
      }, {
        baseUrl: "http://127.0.0.1:8080",
        fetcher,
      }),
    ).rejects.toEqual(
      expect.objectContaining({
        code: "network",
        message: "API request failed before a response was received.",
        name: "ApiError",
      }),
    );
  });

  it("includes displayName and handle when starting a sign-up session", async () => {
    const fetcher = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              challengeToken: "challenge-token",
              expiresAt: "2026-04-07T21:00:00+09:00",
            },
            error: null,
            meta: {
              page: null,
              requestId: "req_sign_up_challenge_002",
            },
          }),
          {
            headers: {
              "Content-Type": "application/json",
            },
            status: 200,
          },
        ),
      )
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    await expect(
      authenticateFanWithEmail({
        displayName: "Mina",
        email: "fan@example.com",
        handle: "@mina",
        mode: "sign-up",
      }, {
        baseUrl: "http://127.0.0.1:8080",
        fetcher,
      }),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenNthCalledWith(
      2,
      new URL("http://127.0.0.1:8080/api/fan/auth/sign-up/session"),
      expect.objectContaining({
        body: JSON.stringify({
          challengeToken: "challenge-token",
          displayName: "Mina",
          email: "fan@example.com",
          handle: "@mina",
        }),
      }),
    );
  });
});
