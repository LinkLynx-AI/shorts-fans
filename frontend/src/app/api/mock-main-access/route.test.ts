import {
  buildMockMainAccessEntryContext,
  parseMockMainPlaybackGrantContext,
} from "@/features/unlock-entry";
import { issueMockSignedToken, readMockSignedToken } from "@/shared/lib/mock-signed-token";

import { POST } from "./route";

async function postMainAccess(body: object) {
  return POST(
    new Request("http://localhost/api/mock-main-access", {
      body: JSON.stringify(body),
      headers: {
        "Content-Type": "application/json",
      },
      method: "POST",
    }),
  );
}

describe("POST /api/mock-main-access", () => {
  it("rejects setup-required access without the required confirmations", async () => {
    const response = await postMainAccess({
      acceptedAge: false,
      acceptedTerms: false,
      entryToken: issueMockSignedToken(
        buildMockMainAccessEntryContext("main_mina_quiet_rooftop", "rooftop"),
      ),
      fromShortId: "rooftop",
      mainId: "main_mina_quiet_rooftop",
    });

    expect(response.status).toBe(403);
    await expect(response.json()).resolves.toEqual({
      fallbackHref: "/shorts/rooftop",
    });
  });

  it("issues a purchased playback grant after setup confirmation", async () => {
    const response = await postMainAccess({
      acceptedAge: true,
      acceptedTerms: true,
      entryToken: issueMockSignedToken(
        buildMockMainAccessEntryContext("main_mina_quiet_rooftop", "rooftop"),
      ),
      fromShortId: "rooftop",
      mainId: "main_mina_quiet_rooftop",
    });
    const body = await response.json();
    const playbackUrl = new URL(body.href, "http://localhost");
    const grant = playbackUrl.searchParams.get("grant");

    expect(response.status).toBe(200);
    expect(playbackUrl.pathname).toBe("/mains/main_mina_quiet_rooftop");
    expect(playbackUrl.searchParams.get("fromShortId")).toBe("rooftop");
    expect(grant).toBeTruthy();
    const grantPayload = readMockSignedToken(grant!);

    expect(grantPayload).not.toBeNull();

    if (!grantPayload) {
      throw new Error("grant payload missing");
    }

    expect(parseMockMainPlaybackGrantContext(grantPayload.context)).toEqual({
      fromShortId: "rooftop",
      grantKind: "purchased",
      mainId: "main_mina_quiet_rooftop",
    });
  });

  it("rejects requests without a valid server-issued entry token", async () => {
    const response = await postMainAccess({
      acceptedAge: true,
      acceptedTerms: true,
      entryToken: "invalid",
      fromShortId: "rooftop",
      mainId: "main_mina_quiet_rooftop",
    });

    expect(response.status).toBe(403);
    await expect(response.json()).resolves.toEqual({
      fallbackHref: "/shorts/rooftop",
    });
  });
});
