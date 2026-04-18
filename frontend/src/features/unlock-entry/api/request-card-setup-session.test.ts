import { viewerSessionCookieName } from "@/entities/viewer";

import { requestCardSetupSession } from "./request-card-setup-session";

describe("requestCardSetupSession", () => {
  it("posts the card-setup session request and returns widget config", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            apiBaseUrl: "https://api.ccbill.test",
            apiKey: "widget-api-key",
            clientAccount: "900000",
            currency: "JPY",
            initialPeriod: "1",
            initialPrice: "1800.00",
            sessionToken: "card-setup-session-token",
            subAccount: "0001",
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_card_setup_session_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      requestCardSetupSession({
        baseUrl: "https://api.example.com",
        entryToken: "signed-entry-token",
        fetcher,
        fromShortId: "short_mina_rooftop",
        mainId: "main_mina_quiet_rooftop",
        sessionToken: "raw-session-token",
      }),
    ).resolves.toEqual({
      apiBaseUrl: "https://api.ccbill.test",
      apiKey: "widget-api-key",
      clientAccount: "900000",
      currency: "JPY",
      initialPeriod: "1",
      initialPrice: "1800.00",
      sessionToken: "card-setup-session-token",
      subAccount: "0001",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/mains/main_mina_quiet_rooftop/card-setup-session",
    );
    expect(new Headers(fetcher.mock.calls[0]?.[1]?.headers).get("Cookie")).toBe(
      `${viewerSessionCookieName}=raw-session-token`,
    );
    expect(fetcher.mock.calls[0]?.[1]?.body).toBe(
      JSON.stringify({
        entryToken: "signed-entry-token",
        fromShortId: "short_mina_rooftop",
      }),
    );
  });
});
