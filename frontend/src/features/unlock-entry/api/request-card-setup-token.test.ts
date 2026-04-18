import { viewerSessionCookieName } from "@/entities/viewer";

import { requestCardSetupToken } from "./request-card-setup-token";

describe("requestCardSetupToken", () => {
  it("posts the provider token and returns the opaque card-setup token", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            cardSetupToken: "opaque-card-setup-token",
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_card_setup_token_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      requestCardSetupToken({
        baseUrl: "https://api.example.com",
        cardSetupSessionToken: "card-setup-session-token",
        entryToken: "signed-entry-token",
        fetcher,
        fromShortId: "short_mina_rooftop",
        mainId: "main_mina_quiet_rooftop",
        paymentTokenId: "provider-payment-token",
        sessionToken: "raw-session-token",
      }),
    ).resolves.toEqual({
      cardSetupToken: "opaque-card-setup-token",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/mains/main_mina_quiet_rooftop/card-setup-token",
    );
    expect(new Headers(fetcher.mock.calls[0]?.[1]?.headers).get("Cookie")).toBe(
      `${viewerSessionCookieName}=raw-session-token`,
    );
    expect(fetcher.mock.calls[0]?.[1]?.body).toBe(
      JSON.stringify({
        entryToken: "signed-entry-token",
        fromShortId: "short_mina_rooftop",
        paymentTokenId: "provider-payment-token",
        sessionToken: "card-setup-session-token",
      }),
    );
  });
});
