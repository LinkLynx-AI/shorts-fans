import {
  render,
  screen,
} from "@testing-library/react";

import CreatorUploadPage from "./page";

vi.mock("@/features/fan-auth-gate", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth-gate")>("@/features/fan-auth-gate");

  return {
    ...actual,
    getFanAuthGateState: vi.fn(),
  };
});

describe("CreatorUploadPage", () => {
  it("renders the creator upload route for creator-mode viewers", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      },
      hasSession: true,
    });

    render(await CreatorUploadPage());

    expect(screen.getByRole("link", { name: "Back" })).toHaveAttribute("href", "/creator");
    expect(screen.getByText("新しい動画")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "本編とショートを追加" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "保存してアップロード" })).toBeDisabled();
  });

  it("renders the login-required state for unauthenticated viewers", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: false,
    });

    render(await CreatorUploadPage());

    expect(screen.getByRole("heading", { name: "creator mode を開くにはログインが必要です。" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ログインへ進む" })).toHaveAttribute("href", "/login");
  });

  it("renders the capability-required state for fan-only viewers", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_fan_001",
      },
      hasSession: true,
    });

    render(await CreatorUploadPage());

    expect(screen.getByRole("heading", { name: "creator mode はまだ利用できません。" })).toBeInTheDocument();
  });

  it("renders the mode-required state when the viewer has not switched into creator mode yet", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: true,
        id: "viewer_creator_002",
      },
      hasSession: true,
    });

    render(await CreatorUploadPage());

    expect(screen.getByRole("heading", { name: "creator mode に切り替えてから開いてください。" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "フィードへ戻る" })).toHaveAttribute("href", "/");
  });
});
