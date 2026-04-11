import { render, screen } from "@testing-library/react";

import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { buildMockMainPlaybackGrantContext } from "@/features/unlock-entry";
import { issueMockSignedToken } from "@/shared/lib/mock-signed-token";
import {
  getMainPlaybackSurfaceById,
  loadMainPlaybackSurface,
} from "@/widgets/main-playback-surface";

import MainPlaybackPage from "./page";

const { cookiesMock, mockedRouter, notFound, redirect } = vi.hoisted(() => ({
  cookiesMock: vi.fn(),
  mockedRouter: {
    back: vi.fn(),
    forward: vi.fn(),
    prefetch: vi.fn(),
    push: vi.fn(),
    refresh: vi.fn(),
    replace: vi.fn(),
  },
  notFound: vi.fn(),
  redirect: vi.fn(),
}));

vi.mock("next/headers", () => ({
  cookies: cookiesMock,
}));

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");

  return {
    ...actual,
    notFound,
    redirect,
    useRouter: () => mockedRouter,
  };
});

vi.mock("@/features/fan-auth-gate", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth-gate")>("@/features/fan-auth-gate");

  return {
    ...actual,
    getFanAuthGateState: vi.fn(),
  };
});

vi.mock("@/widgets/main-playback-surface", async () => {
  const actual = await vi.importActual<typeof import("@/widgets/main-playback-surface")>("@/widgets/main-playback-surface");

  return {
    ...actual,
    loadMainPlaybackSurface: vi.fn(),
  };
});

describe("MainPlaybackPage", () => {
  beforeEach(() => {
    cookiesMock.mockReset();
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    notFound.mockReset();
    redirect.mockReset();
    vi.mocked(loadMainPlaybackSurface).mockReset();
  });

  it("renders the playback surface when a valid grant matches the short context", async () => {
    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: true,
    });
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "viewer-session",
      }),
    });

    const validGrant = issueMockSignedToken(
      buildMockMainPlaybackGrantContext("main_mina_quiet_rooftop", "rooftop", "unlocked"),
    );
    const surface = getMainPlaybackSurfaceById(
      "main_mina_quiet_rooftop",
      "rooftop",
      "unlocked",
    );

    if (!surface) {
      throw new Error("expected mock playback surface");
    }

    vi.mocked(loadMainPlaybackSurface).mockResolvedValue({
      kind: "ready",
      surface,
    });

    render(
      await MainPlaybackPage({
        params: Promise.resolve({
          mainId: "main_mina_quiet_rooftop",
        }),
        searchParams: Promise.resolve({
          fromShortId: "rooftop",
          grant: validGrant,
        }),
      }),
    );

    expect(loadMainPlaybackSurface).toHaveBeenCalledWith("main_mina_quiet_rooftop", {
      fromShortId: "rooftop",
      grant: validGrant,
      sessionToken: "viewer-session",
    });
    expect(screen.getByText("Playing main")).toBeInTheDocument();
    expect(screen.getByLabelText("quiet rooftop main playback")).toBeInTheDocument();
    expect(screen.queryByText("この main はまだ unlock されていません。")).not.toBeInTheDocument();
  });

  it("renders the locked state when a signed grant is replayed for a different short context", async () => {
    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: true,
    });
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "viewer-session",
      }),
    });

    const mismatchedGrant = issueMockSignedToken(
      buildMockMainPlaybackGrantContext("main_mina_quiet_rooftop", "mirror", "unlocked"),
    );
    vi.mocked(loadMainPlaybackSurface).mockResolvedValue({
      kind: "locked",
    });

    render(
      await MainPlaybackPage({
        params: Promise.resolve({
          mainId: "main_mina_quiet_rooftop",
        }),
        searchParams: Promise.resolve({
          fromShortId: "rooftop",
          grant: mismatchedGrant,
        }),
      }),
    );

    expect(screen.getByRole("heading", { name: "この main はまだ unlock されていません。" })).toBeInTheDocument();
    expect(screen.queryByText("Playing main")).not.toBeInTheDocument();
  });

  it("redirects to login when the playback loader reports auth_required", async () => {
    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: true,
    });
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "viewer-session",
      }),
    });
    vi.mocked(loadMainPlaybackSurface).mockResolvedValue({
      kind: "auth_required",
    });

    await MainPlaybackPage({
      params: Promise.resolve({
        mainId: "main_mina_quiet_rooftop",
      }),
      searchParams: Promise.resolve({
        fromShortId: "rooftop",
        grant: "grant_123",
      }),
    });

    expect(redirect).toHaveBeenCalledWith("/login");
  });

  it("delegates not_found loader results to next/navigation.notFound", async () => {
    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: true,
    });
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "viewer-session",
      }),
    });
    vi.mocked(loadMainPlaybackSurface).mockResolvedValue({
      kind: "not_found",
    });

    await MainPlaybackPage({
      params: Promise.resolve({
        mainId: "main_mina_quiet_rooftop",
      }),
      searchParams: Promise.resolve({
        fromShortId: "rooftop",
        grant: "grant_123",
      }),
    });

    expect(notFound).toHaveBeenCalled();
  });
});
