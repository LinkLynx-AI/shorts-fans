import { render, screen } from "@testing-library/react";

import CreatorPage from "./page";

vi.mock("@/features/fan-auth-gate", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth-gate")>("@/features/fan-auth-gate");

  return {
    ...actual,
    getFanAuthGateState: vi.fn(),
  };
});

describe("CreatorPage", () => {
  it("renders the creator route shell for creator-mode viewers", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      },
      hasSession: true,
    });

    render(await CreatorPage());

    expect(screen.getByRole("heading", { name: "Dashboard shell" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Dashboard/i })).toHaveAttribute("aria-current", "page");
    expect(screen.queryByRole("link", { name: "Upload" })).not.toBeInTheDocument();
    expect(screen.getByText("creator route shell")).toBeInTheDocument();
  });

  it("renders the login-required state for unauthenticated viewers", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: false,
    });

    render(await CreatorPage());

    expect(screen.getByRole("heading", { name: "creator mode を開くにはログインが必要です。" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ログインへ進む" })).toHaveAttribute("href", "/login");
  });

  it("renders the capability-required state for authenticated fan-only viewers", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_fan_001",
      },
      hasSession: true,
    });

    render(await CreatorPage());

    expect(screen.getByRole("heading", { name: "creator mode はまだ利用できません。" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "フィードへ戻る" })).toHaveAttribute("href", "/");
  });

  it("renders a mode mismatch state when the viewer has not switched into creator mode yet", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: true,
        id: "viewer_creator_002",
      },
      hasSession: true,
    });

    render(await CreatorPage());

    expect(screen.getByRole("heading", { name: "creator mode に切り替えてから開いてください。" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "フィードへ戻る" })).toHaveAttribute("href", "/");
  });
});
