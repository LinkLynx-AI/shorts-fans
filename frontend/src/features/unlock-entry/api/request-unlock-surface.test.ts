import { viewerSessionCookieName } from "@/entities/viewer";

import { requestUnlockSurfaceByShortId } from "./request-unlock-surface";

describe("requestUnlockSurfaceByShortId", () => {
  it("requests the unlock surface and forwards the session cookie", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            access: {
              mainId: "main_mina_quiet_rooftop",
              reason: "unlock_required",
              status: "locked",
            },
            creator: {
              avatar: null,
              bio: "night preview specialist",
              displayName: "Mina Rei",
              handle: "@minarei",
              id: "creator_mina_rei",
            },
            main: {
              durationSeconds: 480,
              id: "main_mina_quiet_rooftop",
              priceJpy: 1800,
            },
            mainAccessEntry: {
              routePath: "/api/fan/mains/main_mina_quiet_rooftop/access-entry",
              token: "signed-token",
            },
            setup: {
              required: true,
              requiresAgeConfirmation: true,
              requiresTermsAcceptance: true,
            },
            short: {
              caption: "quiet rooftop preview",
              canonicalMainId: "main_mina_quiet_rooftop",
              creatorId: "creator_mina_rei",
              id: "short_mina_rooftop",
              media: {
                durationSeconds: 16,
                id: "asset_short_mina_rooftop",
                kind: "video",
                posterUrl: "https://cdn.example.com/shorts/poster.jpg",
                url: "https://cdn.example.com/shorts/playback.mp4",
              },
              previewDurationSeconds: 16,
            },
            unlockCta: {
              mainDurationSeconds: 480,
              priceJpy: 1800,
              resumePositionSeconds: null,
              state: "setup_required",
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_short_unlock_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      requestUnlockSurfaceByShortId({
        baseUrl: "https://api.example.com",
        fetcher,
        sessionToken: "raw-session-token",
        shortId: "short_mina_rooftop",
      }),
    ).resolves.toMatchObject({
      mainAccessEntry: {
        routePath: "/api/fan/mains/main_mina_quiet_rooftop/access-entry",
        token: "signed-token",
      },
      short: {
        id: "short_mina_rooftop",
      },
      unlockCta: {
        state: "setup_required",
      },
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/shorts/short_mina_rooftop/unlock",
    );
    expect(fetcher.mock.calls[0]?.[1]?.credentials).toBe("include");
    expect(new Headers(fetcher.mock.calls[0]?.[1]?.headers).get("Cookie")).toBe(
      `${viewerSessionCookieName}=raw-session-token`,
    );
  });
});
