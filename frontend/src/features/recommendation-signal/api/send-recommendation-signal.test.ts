import { sendRecommendationSignal } from "./send-recommendation-signal";

describe("sendRecommendationSignal", () => {
  it("posts the signal payload with credentials and keepalive", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(null, {
        status: 200,
      }),
    );

    await expect(
      sendRecommendationSignal({
        baseUrl: "https://api.example.com",
        fetcher,
        input: {
          eventKind: "impression",
          idempotencyKey: "impression:short_1:session_1",
          shortId: "short_11111111111111111111111111111111",
        },
      }),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenCalledWith(
      new URL("/api/fan/recommendation/events", "https://api.example.com"),
      expect.objectContaining({
        body: JSON.stringify({
          eventKind: "impression",
          idempotencyKey: "impression:short_1:session_1",
          shortId: "short_11111111111111111111111111111111",
        }),
        credentials: "include",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        keepalive: true,
        method: "POST",
      }),
    );
  });
});
