import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { switchViewerActiveMode } from "@/features/creator-entry/api/switch-viewer-active-mode";

import CreatorPage from "./page";

const mockedRouter = vi.hoisted(() => ({
  back: vi.fn(),
  forward: vi.fn(),
  prefetch: vi.fn(),
  push: vi.fn(),
  refresh: vi.fn(),
  replace: vi.fn(),
}));

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");

  return {
    ...actual,
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

vi.mock("@/features/creator-entry/api/switch-viewer-active-mode", () => ({
  switchViewerActiveMode: vi.fn(),
}));

function createDeferredPromise() {
  let resolvePromise: () => void = () => {};
  let rejectPromise: (reason?: unknown) => void = () => {};
  const promise = new Promise<void>((resolve, reject) => {
    resolvePromise = resolve;
    rejectPromise = reject;
  });

  return {
    promise,
    reject: rejectPromise,
    resolve: resolvePromise,
  };
}

describe("CreatorPage", () => {
  beforeEach(() => {
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    vi.mocked(switchViewerActiveMode).mockReset();
  });

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
    expect(screen.getByRole("button", { name: "Account menu" })).toBeEnabled();
    expect(screen.queryByText("Dashboard")).not.toBeInTheDocument();
    expect(screen.getByText("@minarei")).toBeInTheDocument();
    expect(screen.getByText("差し戻しが1件あります")).toBeInTheDocument();
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

  it("switches the viewer back to fan mode home from the account menu", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");
    const user = userEvent.setup();
    const deferred = createDeferredPromise();

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      },
      hasSession: true,
    });
    vi.mocked(switchViewerActiveMode).mockReturnValue(deferred.promise);

    render(await CreatorPage());

    await user.click(screen.getByRole("button", { name: "Account menu" }));
    await user.click(screen.getByRole("button", { name: "Fan mode に切り替え" }));

    expect(switchViewerActiveMode).toHaveBeenCalledWith("fan");
    expect(screen.getByRole("button", { name: "Fan mode に切り替えています..." })).toBeDisabled();

    deferred.resolve();

    await waitFor(() => {
      expect(mockedRouter.push).toHaveBeenCalledWith("/");
    });
  });

  it("shows a retryable error when fan mode switching fails", async () => {
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
    vi.mocked(switchViewerActiveMode)
      .mockRejectedValueOnce(new Error("boom"))
      .mockResolvedValueOnce(undefined);

    render(await CreatorPage());

    await user.click(screen.getByRole("button", { name: "Account menu" }));
    await user.click(screen.getByRole("button", { name: "Fan mode に切り替え" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "fan mode に戻れませんでした。少し時間を置いてからやり直してください。",
    );

    await user.click(screen.getByRole("button", { name: "Fan mode に切り替え" }));

    await waitFor(() => {
      expect(switchViewerActiveMode).toHaveBeenCalledTimes(2);
      expect(mockedRouter.push).toHaveBeenCalledWith("/");
    });
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
