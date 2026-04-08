import {
  render,
  screen,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  CreatorModeShell,
  getMockCreatorModeShellState,
} from "@/widgets/creator-mode-shell";

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
    const user = userEvent.setup();

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      },
      hasSession: true,
    });

    render(await CreatorPage());

    expect(screen.getByRole("button", { name: "動画を追加" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Account menu" })).toBeDisabled();
    expect(screen.queryByText("Dashboard")).not.toBeInTheDocument();
    expect(screen.getByText("@minarei")).toBeInTheDocument();
    expect(screen.getByText("quiet rooftop と hotel light の preview を軸に投稿。")).toBeInTheDocument();
    expect(screen.getByText("¥120,000")).toBeInTheDocument();
    expect(screen.getByText("差し戻しが1件あります")).toBeInTheDocument();
    expect(screen.getByText("short 1件を確認してください")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Top main" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Top main" }));

    expect(screen.getByText("linked short からの流入を unlock に変えている本編です。")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "rooftop side preview" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "rooftop side preview" }));

    expect(screen.getByText("同じ main に送る別導線として比較しているショートです。")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Back" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Back" }));

    expect(screen.getByRole("button", { name: "Top main" })).toBeInTheDocument();
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

  it("falls back to a generic revision message when revision counts are inconsistent", () => {
    const state = getMockCreatorModeShellState();

    render(
      <CreatorModeShell
        state={{
          ...state,
          workspace: {
            ...state.workspace,
            revisionRequestedSummary: {
              mainCount: 0,
              shortCount: 0,
              totalCount: 0,
            },
          },
        }}
      />,
    );

    expect(screen.getByText("差し戻しが0件あります")).toBeInTheDocument();
    expect(screen.getByText("修正依頼内容を確認してください")).toBeInTheDocument();
  });
});
