import userEvent from "@testing-library/user-event";
import { render, screen } from "@testing-library/react";

import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { loadLibraryMainReelState } from "@/widgets/main-playback-surface";
import { buildMockMainPlaybackGrantContext } from "@/features/unlock-entry";
import { issueMockSignedToken } from "@/shared/lib/mock-signed-token";

import MainPlaybackPage from "./page";

const { cookiesMock } = vi.hoisted(() => ({
  cookiesMock: vi.fn(),
}));

vi.mock("next/headers", () => ({
  cookies: cookiesMock,
}));

vi.mock("@/features/fan-auth-gate", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth-gate")>("@/features/fan-auth-gate");

  return {
    ...actual,
    getFanAuthGateState: vi.fn(),
  };
});

vi.mock("@/widgets/main-playback-surface", async () => {
  const actual = await vi.importActual<typeof import("@/widgets/main-playback-surface")>(
    "@/widgets/main-playback-surface",
  );

  return {
    ...actual,
    LibraryMainReel: vi.fn((props: { initialIndex: number }) => (
      <div data-testid="library-main-reel">{props.initialIndex}</div>
    )),
    loadLibraryMainReelState: vi.fn(),
  };
});

const mockedLoadLibraryMainReelState = vi.mocked(loadLibraryMainReelState);

describe("MainPlaybackPage", () => {
  beforeEach(() => {
    vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
    cookiesMock.mockReset();
    mockedLoadLibraryMainReelState.mockReset();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders the playback surface for a valid signed grant", async () => {
    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: true,
    });

    const validGrant = issueMockSignedToken(
      buildMockMainPlaybackGrantContext("main_mina_quiet_rooftop", "rooftop", "unlocked"),
    );

    const user = userEvent.setup();

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

    expect(screen.getByLabelText("Main playback video")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "この main はまだ unlock されていません。" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "More options" }));

    expect(screen.getByRole("link", { name: "クリエイターのプロフィールへ" })).toHaveAttribute(
      "href",
      "/creators/creator_mina_rei?from=short&shortId=rooftop",
    );
  });

  it("renders the locked state when a signed grant is replayed for a different short context", async () => {
    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: true,
    });

    const mismatchedGrant = issueMockSignedToken(
      buildMockMainPlaybackGrantContext("main_mina_quiet_rooftop", "mirror", "unlocked"),
    );

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

  it("renders the library main reel when opened from fan profile library without a grant", async () => {
    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: true,
    });
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    mockedLoadLibraryMainReelState.mockResolvedValue({
      initialIndex: 1,
      items: [],
    });

    render(
      await MainPlaybackPage({
        params: Promise.resolve({
          mainId: "main_mina_quiet_rooftop",
        }),
        searchParams: Promise.resolve({
          fanTab: "library",
          from: "fan",
          fromShortId: "short_mina_rooftop",
        }),
      }),
    );

    expect(mockedLoadLibraryMainReelState).toHaveBeenCalledWith({
      mainId: "main_mina_quiet_rooftop",
      sessionToken: "valid-session",
    });
    expect(screen.getByTestId("library-main-reel")).toHaveTextContent("1");
  });
});
