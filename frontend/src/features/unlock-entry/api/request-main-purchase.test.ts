import { viewerSessionCookieName } from "@/entities/viewer";

import { requestMainPurchase } from "./request-main-purchase";

describe("requestMainPurchase", () => {
  it("posts a saved-card purchase payload and returns the purchase outcome", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            access: {
              mainId: "main_mina_quiet_rooftop",
              reason: "purchased",
              status: "unlocked",
            },
            entryContext: {
              accessEntryPath: "/api/fan/mains/main_mina_quiet_rooftop/access-entry",
              purchasePath: "/api/fan/mains/main_mina_quiet_rooftop/purchase",
              token: "purchase-entry-token",
            },
            purchase: {
              canRetry: false,
              failureReason: null,
              status: "succeeded",
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_main_purchase_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      requestMainPurchase({
        acceptedAge: true,
        acceptedTerms: true,
        baseUrl: "https://api.example.com",
        entryToken: "signed-entry-token",
        fetcher,
        fromShortId: "short_mina_rooftop",
        mainId: "main_mina_quiet_rooftop",
        paymentMethod: {
          mode: "saved_card",
          paymentMethodId: "paymeth_saved_visa",
        },
        sessionToken: "raw-session-token",
      }),
    ).resolves.toEqual({
      access: {
        mainId: "main_mina_quiet_rooftop",
        reason: "purchased",
        status: "unlocked",
      },
      entryContext: {
        accessEntryPath: "/api/fan/mains/main_mina_quiet_rooftop/access-entry",
        purchasePath: "/api/fan/mains/main_mina_quiet_rooftop/purchase",
        token: "purchase-entry-token",
      },
      purchase: {
        canRetry: false,
        failureReason: null,
        status: "succeeded",
      },
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/mains/main_mina_quiet_rooftop/purchase",
    );
    expect(fetcher.mock.calls[0]?.[1]?.method).toBe("POST");
    expect(new Headers(fetcher.mock.calls[0]?.[1]?.headers).get("Cookie")).toBe(
      `${viewerSessionCookieName}=raw-session-token`,
    );
    expect(fetcher.mock.calls[0]?.[1]?.body).toBe(
      JSON.stringify({
        acceptedAge: true,
        acceptedTerms: true,
        entryToken: "signed-entry-token",
        fromShortId: "short_mina_rooftop",
        paymentMethod: {
          mode: "saved_card",
          paymentMethodId: "paymeth_saved_visa",
        },
      }),
    );
  });

  it("posts a new-card purchase payload with the exchanged card setup token", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            access: {
              mainId: "main_mina_quiet_rooftop",
              reason: "unlock_required",
              status: "locked",
            },
            entryContext: null,
            purchase: {
              canRetry: true,
              failureReason: "purchase_declined",
              status: "failed",
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_main_purchase_new_card_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      requestMainPurchase({
        acceptedAge: true,
        acceptedTerms: true,
        baseUrl: "https://api.example.com",
        entryToken: "signed-entry-token",
        fetcher,
        fromShortId: "short_mina_rooftop",
        mainId: "main_mina_quiet_rooftop",
        paymentMethod: {
          cardSetupToken: "opaque-card-setup-token",
          mode: "new_card",
        },
      }),
    ).resolves.toEqual({
      access: {
        mainId: "main_mina_quiet_rooftop",
        reason: "unlock_required",
        status: "locked",
      },
      entryContext: null,
      purchase: {
        canRetry: true,
        failureReason: "purchase_declined",
        status: "failed",
      },
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/mains/main_mina_quiet_rooftop/purchase",
    );
    expect(fetcher.mock.calls[0]?.[1]?.body).toBe(
      JSON.stringify({
        acceptedAge: true,
        acceptedTerms: true,
        entryToken: "signed-entry-token",
        fromShortId: "short_mina_rooftop",
        paymentMethod: {
          cardSetupToken: "opaque-card-setup-token",
          mode: "new_card",
        },
      }),
    );
  });
});
