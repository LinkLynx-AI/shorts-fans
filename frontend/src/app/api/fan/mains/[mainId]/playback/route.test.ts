import { NextRequest } from "next/server";

import {
  buildMockMainPlaybackGrantContext,
} from "@/features/unlock-entry";
import {
  getCurrentViewerBootstrap,
  viewerSessionCookieName,
} from "@/entities/viewer";
import {
  createMockSessionProof,
  issueMockSignedToken,
} from "@/shared/lib/mock-signed-token";

import { GET } from "./route";

vi.mock("@/entities/viewer", async () => {
  const actual = await vi.importActual<typeof import("@/entities/viewer")>("@/entities/viewer");

  return {
    ...actual,
    getCurrentViewerBootstrap: vi.fn(),
  };
});

async function getMainPlayback(
  mainId: string,
  searchParams: {
    fromShortId?: string;
    grant?: string;
  },
  sessionToken?: string,
) {
  const url = new URL(`http://localhost/api/fan/mains/${mainId}/playback`);

  if (searchParams.fromShortId) {
    url.searchParams.set("fromShortId", searchParams.fromShortId);
  }

  if (searchParams.grant) {
    url.searchParams.set("grant", searchParams.grant);
  }

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
    nextUrl: url,
  } as unknown as NextRequest;

  return GET(request, {
    params: Promise.resolve({
      mainId,
    }),
  });
}

describe("GET /api/fan/mains/[mainId]/playback", () => {
  beforeEach(() => {
    vi.mocked(getCurrentViewerBootstrap).mockResolvedValue({
      activeMode: "fan",
      canAccessCreatorMode: false,
      id: "viewer_123",
    });
  });

  it("returns auth_required when no fan session exists", async () => {
    const response = await getMainPlayback("main_mina_quiet_rooftop", {
      fromShortId: "rooftop",
      grant: issueMockSignedToken(
        buildMockMainPlaybackGrantContext("main_mina_quiet_rooftop", "rooftop", "unlocked"),
      ),
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
        requestId: "req_main_playback_auth_required_001",
      },
    });
  });

  it("returns the playback payload when the grant matches the short context", async () => {
    const response = await getMainPlayback(
      "main_mina_quiet_rooftop",
      {
        fromShortId: "rooftop",
        grant: issueMockSignedToken(
          buildMockMainPlaybackGrantContext("main_mina_quiet_rooftop", "rooftop", "unlocked"),
          {
            sessionProof: createMockSessionProof("viewer-session"),
          },
        ),
      },
      "viewer-session",
    );
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.error).toBeNull();
    expect(body.data.main.id).toBe("main_mina_quiet_rooftop");
    expect(body.data.main.media.url).toContain("mina-quiet-rooftop.mp4");
    expect(body.data.entryShort.id).toBe("rooftop");
  });

  it("returns main_locked when the grant is invalid for the requested short", async () => {
    const response = await getMainPlayback(
      "main_mina_quiet_rooftop",
      {
        fromShortId: "rooftop",
        grant: issueMockSignedToken(
          buildMockMainPlaybackGrantContext("main_mina_quiet_rooftop", "mirror", "unlocked"),
          {
            sessionProof: createMockSessionProof("viewer-session"),
          },
        ),
      },
      "viewer-session",
    );

    expect(response.status).toBe(403);
    await expect(response.json()).resolves.toEqual({
      data: null,
      error: {
        code: "main_locked",
        message: "main playback is locked",
      },
      meta: {
        page: null,
        requestId: "req_main_playback_locked_001",
      },
    });
  });

  it("returns not_found when the short does not belong to the requested main", async () => {
    const response = await getMainPlayback(
      "main_mina_quiet_rooftop",
      {
        fromShortId: "softlight",
        grant: issueMockSignedToken(
          buildMockMainPlaybackGrantContext("main_mina_quiet_rooftop", "softlight", "unlocked"),
          {
            sessionProof: createMockSessionProof("viewer-session"),
          },
        ),
      },
      "viewer-session",
    );

    expect(response.status).toBe(404);
    await expect(response.json()).resolves.toEqual({
      data: null,
      error: {
        code: "not_found",
        message: "main or short was not found",
      },
      meta: {
        page: null,
        requestId: "req_main_playback_not_found_001",
      },
    });
  });

  it("returns owner playback when the grant is bound to the current session", async () => {
    const response = await getMainPlayback(
      "main_aoi_blue_balcony",
      {
        fromShortId: "balcony",
        grant: issueMockSignedToken(
          buildMockMainPlaybackGrantContext("main_aoi_blue_balcony", "balcony", "owner"),
          {
            sessionProof: createMockSessionProof("viewer-session"),
          },
        ),
      },
      "viewer-session",
    );
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.error).toBeNull();
    expect(body.data.access.status).toBe("owner");
    expect(body.data.main.id).toBe("main_aoi_blue_balcony");
    expect(body.data.entryShort.id).toBe("balcony");
  });

  it("returns main_locked when the grant is replayed from a different authenticated session", async () => {
    const response = await getMainPlayback(
      "main_mina_quiet_rooftop",
      {
        fromShortId: "rooftop",
        grant: issueMockSignedToken(
          buildMockMainPlaybackGrantContext("main_mina_quiet_rooftop", "rooftop", "unlocked"),
          {
            sessionProof: createMockSessionProof("viewer-session"),
          },
        ),
      },
      "other-session",
    );

    expect(response.status).toBe(403);
    await expect(response.json()).resolves.toEqual({
      data: null,
      error: {
        code: "main_locked",
        message: "main playback is locked",
      },
      meta: {
        page: null,
        requestId: "req_main_playback_locked_001",
      },
    });
  });
});
