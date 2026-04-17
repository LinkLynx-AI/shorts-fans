import {
  confirmFanPasswordReset,
  confirmFanSignUp,
  FanAuthApiError,
  reAuthenticateFan,
  signInFan,
  signUpFan,
  startFanPasswordReset,
} from "@/features/fan-auth";

describe("fan auth requests", () => {
  it("posts credentials to the sign-in endpoint with cookies enabled", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await expect(
      signInFan({
        email: "fan@example.com",
        password: "VeryStrongPass123!",
      }, {
        baseUrl: "http://127.0.0.1:8080",
        fetcher,
      }),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenCalledWith(
      new URL("http://127.0.0.1:8080/api/fan/auth/sign-in"),
      expect.objectContaining({
        body: JSON.stringify({
          email: "fan@example.com",
          password: "VeryStrongPass123!",
        }),
        credentials: "include",
        method: "POST",
      }),
    );
  });

  it("returns the accepted next step for sign-up", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            deliveryDestinationHint: "f***@example.com",
            nextStep: "confirm_sign_up",
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_sign_up_accepted_001",
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
      signUpFan({
        displayName: "Mina",
        email: "fan@example.com",
        handle: "@mina",
        password: "VeryStrongPass123!",
      }, {
        baseUrl: "http://127.0.0.1:8080",
        fetcher,
      }),
    ).resolves.toEqual({
      deliveryDestinationHint: "f***@example.com",
      nextStep: "confirm_sign_up",
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("http://127.0.0.1:8080/api/fan/auth/sign-up"),
      expect.objectContaining({
        body: JSON.stringify({
          displayName: "Mina",
          email: "fan@example.com",
          handle: "@mina",
          password: "VeryStrongPass123!",
        }),
      }),
    );
  });

  it("surfaces contract errors from confirmation endpoints", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: null,
          error: {
            code: "invalid_confirmation_code",
            message: "confirmation code is invalid",
          },
          meta: {
            page: null,
            requestId: "req_sign_up_confirm_invalid_code_001",
          },
        }),
        {
          headers: {
            "Content-Type": "application/json",
          },
          status: 400,
        },
      ),
    );

    await expect(
      confirmFanSignUp({
        confirmationCode: "123456",
        email: "fan@example.com",
      }, {
        baseUrl: "http://127.0.0.1:8080",
        fetcher,
      }),
    ).rejects.toEqual(
      new FanAuthApiError("invalid_confirmation_code", "confirmation code is invalid", {
        requestId: "req_sign_up_confirm_invalid_code_001",
        status: 400,
      }),
    );
  });

  it("returns the accepted next step for password reset", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            deliveryDestinationHint: "f***@example.com",
            nextStep: "confirm_password_reset",
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_password_reset_accepted_001",
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
      startFanPasswordReset({
        email: "fan@example.com",
      }, {
        baseUrl: "http://127.0.0.1:8080",
        fetcher,
      }),
    ).resolves.toEqual({
      deliveryDestinationHint: "f***@example.com",
      nextStep: "confirm_password_reset",
    });
  });

  it("posts the new password when confirming a password reset", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await expect(
      confirmFanPasswordReset({
        confirmationCode: "123456",
        email: "fan@example.com",
        newPassword: "EvenStrongerPass123!",
      }, {
        baseUrl: "http://127.0.0.1:8080",
        fetcher,
      }),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenCalledWith(
      new URL("http://127.0.0.1:8080/api/fan/auth/password-reset/confirm"),
      expect.objectContaining({
        body: JSON.stringify({
          confirmationCode: "123456",
          email: "fan@example.com",
          newPassword: "EvenStrongerPass123!",
        }),
      }),
    );
  });

  it("surfaces network failures from re-auth as shared API errors", async () => {
    const fetcher = vi.fn<typeof fetch>().mockRejectedValue(new Error("network down"));

    await expect(
      reAuthenticateFan({
        password: "VeryStrongPass123!",
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
});
