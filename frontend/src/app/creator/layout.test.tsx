import {
  render,
  screen,
} from "@testing-library/react";

import {
  useCurrentViewer,
  useHasViewerSession,
} from "@/entities/viewer";

import CreatorLayout from "./layout";

vi.mock("@/features/fan-auth-gate", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth-gate")>("@/features/fan-auth-gate");

  return {
    ...actual,
    getFanAuthGateState: vi.fn(),
  };
});

function ProviderProbe() {
  const currentViewer = useCurrentViewer();
  const hasSession = useHasViewerSession();

  return (
    <>
      <span>{hasSession ? "session-present" : "session-missing"}</span>
      <span>{currentViewer?.id ?? "anonymous"}</span>
    </>
  );
}

describe("CreatorLayout", () => {
  it("provides viewer and session context to creator routes", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      },
      hasSession: true,
    });

    render(await CreatorLayout({ children: <ProviderProbe /> }));

    expect(screen.getByText("session-present")).toBeInTheDocument();
    expect(screen.getByText("viewer_creator_001")).toBeInTheDocument();
  });

  it("keeps anonymous creator routes consistent when no session exists", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: false,
    });

    render(await CreatorLayout({ children: <ProviderProbe /> }));

    expect(screen.getByText("session-missing")).toBeInTheDocument();
    expect(screen.getByText("anonymous")).toBeInTheDocument();
  });
});
