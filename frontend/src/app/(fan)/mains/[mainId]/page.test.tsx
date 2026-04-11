import { render, screen } from "@testing-library/react";

import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { buildMockMainPlaybackGrantContext } from "@/features/unlock-entry";
import { issueMockSignedToken } from "@/shared/lib/mock-signed-token";

import MainPlaybackPage from "./page";

vi.mock("@/features/fan-auth-gate", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth-gate")>("@/features/fan-auth-gate");

  return {
    ...actual,
    getFanAuthGateState: vi.fn(),
  };
});

describe("MainPlaybackPage", () => {
  beforeEach(() => {
    vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
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
});
