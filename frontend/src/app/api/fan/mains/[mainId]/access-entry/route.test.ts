import { NextRequest } from "next/server";

import {
  buildMockMainAccessEntryContext,
  parseMockMainPlaybackGrantContext,
} from "@/features/unlock-entry";
import {
  getCurrentViewerBootstrap,
  viewerSessionCookieName,
} from "@/entities/viewer";
import { issueMockSignedToken, readMockSignedToken } from "@/shared/lib/mock-signed-token";

import { POST } from "./route";

vi.mock("@/entities/viewer", async () => {
  const actual = await vi.importActual<typeof import("@/entities/viewer")>("@/entities/viewer");

  return {
    ...actual,
    getCurrentViewerBootstrap: vi.fn(),
  };
});

async function postMainAccessEntry(
  mainId: string,
  body: object,
  sessionToken?: string,
) {
  const request = {
    cookies: {
      get(name: string) {
        if (!sessionToken || name !== viewerSessionCookieName) {
          return undefined;
        }

        return {
          name,
          value: sessionToken,
        };
      },
    },
    json: async () => body,
  } as unknown as NextRequest;

  return POST(request, {
    params: Promise.resolve({
      mainId,
    }),
  });
}

async function expectPlaybackGrantResponse(
  response: Response,
  expected: {
    fromShortId: string;
    grantKind: "owner" | "unlocked";
    mainId: string;
  },
) {
  const body = await response.json();
  const href = body.data?.href;
  const playbackUrl = new URL(href, "http://localhost");
  const grant = playbackUrl.searchParams.get("grant");

  expect(response.status).toBe(200);
  expect(playbackUrl.pathname).toBe(`/mains/${expected.mainId}`);
  expect(playbackUrl.searchParams.get("fromShortId")).toBe(expected.fromShortId);
  expect(grant).toBeTruthy();

  const grantPayload = readMockSignedToken(grant!);

  expect(grantPayload).not.toBeNull();

  if (!grantPayload) {
    throw new Error("grant payload missing");
  }

  expect(parseMockMainPlaybackGrantContext(grantPayload.context)).toEqual(expected);
}

describe("POST /api/fan/mains/[mainId]/access-entry", () => {
  beforeEach(() => {
    vi.mocked(getCurrentViewerBootstrap).mockResolvedValue({
      activeMode: "fan",
      canAccessCreatorMode: false,
      id: "viewer_123",
    });
  });

  it("returns auth_required when no fan session exists", async () => {
    const response = await postMainAccessEntry("main_mina_quiet_rooftop", {
      acceptedAge: true,
      acceptedTerms: true,
      entryToken: issueMockSignedToken(
        buildMockMainAccessEntryContext("main_mina_quiet_rooftop", "rooftop"),
      ),
      fromShortId: "rooftop",
    });

    expect(response.status).toBe(401);
    await expect(response.json()).resolves.toEqual({
      data: null,
      error: {
        code: "auth_required",
        message: "main playback requires authentication",
      },
      meta: {
        page: null,
        requestId: "req_main_access_entry_auth_required_001",
      },
    });
  });

  it("rejects setup-required access without the required confirmations", async () => {
    const response = await postMainAccessEntry(
      "main_mina_quiet_rooftop",
      {
        acceptedAge: false,
        acceptedTerms: false,
        entryToken: issueMockSignedToken(
          buildMockMainAccessEntryContext("main_mina_quiet_rooftop", "rooftop"),
        ),
        fromShortId: "rooftop",
      },
      "viewer-session",
    );

    expect(response.status).toBe(403);
    await expect(response.json()).resolves.toEqual({
      data: null,
      error: {
        code: "main_locked",
        message: "main access entry could not be issued",
      },
      meta: {
        page: null,
        requestId: "req_main_access_entry_locked_001",
      },
    });
  });

  it("issues an unlocked playback grant after setup confirmation", async () => {
    const response = await postMainAccessEntry(
      "main_mina_quiet_rooftop",
      {
        acceptedAge: true,
        acceptedTerms: true,
        entryToken: issueMockSignedToken(
          buildMockMainAccessEntryContext("main_mina_quiet_rooftop", "rooftop"),
        ),
        fromShortId: "rooftop",
      },
      "viewer-session",
    );

    await expectPlaybackGrantResponse(response, {
      fromShortId: "rooftop",
      grantKind: "unlocked",
      mainId: "main_mina_quiet_rooftop",
    });
  });

  it("issues an unlocked playback grant for continue_main entries", async () => {
    const response = await postMainAccessEntry(
      "main_aoi_blue_balcony",
      {
        acceptedAge: true,
        acceptedTerms: true,
        entryToken: issueMockSignedToken(
          buildMockMainAccessEntryContext("main_aoi_blue_balcony", "softlight"),
        ),
        fromShortId: "softlight",
      },
      "viewer-session",
    );

    await expectPlaybackGrantResponse(response, {
      fromShortId: "softlight",
      grantKind: "unlocked",
      mainId: "main_aoi_blue_balcony",
    });
  });

  it("issues an owner playback grant for owner_preview entries", async () => {
    const response = await postMainAccessEntry(
      "main_aoi_blue_balcony",
      {
        acceptedAge: true,
        acceptedTerms: true,
        entryToken: issueMockSignedToken(
          buildMockMainAccessEntryContext("main_aoi_blue_balcony", "balcony"),
        ),
        fromShortId: "balcony",
      },
      "viewer-session",
    );

    await expectPlaybackGrantResponse(response, {
      fromShortId: "balcony",
      grantKind: "owner",
      mainId: "main_aoi_blue_balcony",
    });
  });

  it("rejects requests without a valid server-issued entry token", async () => {
    const response = await postMainAccessEntry(
      "main_mina_quiet_rooftop",
      {
        acceptedAge: true,
        acceptedTerms: true,
        entryToken: "invalid",
        fromShortId: "rooftop",
      },
      "viewer-session",
    );

    expect(response.status).toBe(403);
    await expect(response.json()).resolves.toEqual({
      data: null,
      error: {
        code: "main_locked",
        message: "main access entry could not be issued",
      },
      meta: {
        page: null,
        requestId: "req_main_access_entry_locked_001",
      },
    });
  });

  it("returns auth_required when bootstrap cannot resolve the session cookie", async () => {
    vi.mocked(getCurrentViewerBootstrap).mockResolvedValueOnce(null);

    const response = await postMainAccessEntry(
      "main_mina_quiet_rooftop",
      {
        acceptedAge: true,
        acceptedTerms: true,
        entryToken: issueMockSignedToken(
          buildMockMainAccessEntryContext("main_mina_quiet_rooftop", "rooftop"),
        ),
        fromShortId: "rooftop",
      },
      "stale-session",
    );

    expect(response.status).toBe(401);
    await expect(response.json()).resolves.toEqual({
      data: null,
      error: {
        code: "auth_required",
        message: "main playback requires authentication",
      },
      meta: {
        page: null,
        requestId: "req_main_access_entry_auth_required_001",
      },
    });
  });
});
