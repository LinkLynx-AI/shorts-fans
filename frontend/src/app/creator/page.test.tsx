import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { switchViewerActiveMode } from "@/features/creator-entry/api/switch-viewer-active-mode";
import { ApiError } from "@/shared/api";
import { getCreatorWorkspaceSummary } from "@/widgets/creator-mode-shell/api/get-creator-workspace-summary";
import {
  CreatorModeShell,
  getMockCreatorModeShellState,
} from "@/widgets/creator-mode-shell";

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

vi.mock("@/widgets/creator-mode-shell/api/get-creator-workspace-summary", () => ({
  getCreatorWorkspaceSummary: vi.fn(),
}));

type CreatorWorkspaceSummary = Awaited<ReturnType<typeof getCreatorWorkspaceSummary>>;

function createDeferredPromise<TResult = void>() {
  let resolvePromise: (value: TResult | PromiseLike<TResult>) => void = () => {};
  let rejectPromise: (reason?: unknown) => void = () => {};
  const promise = new Promise<TResult>((resolve, reject) => {
    resolvePromise = resolve;
    rejectPromise = reject;
  });

  return {
    promise,
    reject: rejectPromise,
    resolve: resolvePromise,
  };
}

function createCreatorWorkspaceSummary(
  overrides: Partial<CreatorWorkspaceSummary> = {},
): CreatorWorkspaceSummary {
  return {
    creator: {
      avatar: {
        durationSeconds: null,
        id: "asset_creator_mina_avatar",
        kind: "image",
        posterUrl: null,
        url: "https://cdn.example.com/creator/mina/avatar.jpg",
      },
      bio: "quiet rooftop と hotel light の preview を軸に投稿。",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_mina_rei",
    },
    overviewMetrics: {
      grossUnlockRevenueJpy: 120000,
      unlockCount: 238,
      uniquePurchaserCount: 164,
    },
    revisionRequestedSummary: {
      mainCount: 0,
      shortCount: 1,
      totalCount: 1,
    },
    ...overrides,
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
    vi.mocked(getCreatorWorkspaceSummary).mockReset();
    vi.mocked(getCreatorWorkspaceSummary).mockResolvedValue(createCreatorWorkspaceSummary());
  });

  it("renders contract-backed summary data for creator-mode viewers", async () => {
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
    vi.mocked(getCreatorWorkspaceSummary).mockResolvedValue(
      createCreatorWorkspaceSummary({
        creator: {
          avatar: {
            durationSeconds: null,
            id: "asset_creator_contract_avatar",
            kind: "image",
            posterUrl: null,
            url: "https://cdn.example.com/creator/contract/avatar.jpg",
          },
          bio: "contract-backed creator bio",
          displayName: "Contract Mina",
          handle: "@contractmina",
          id: "creator_mina_rei",
        },
        overviewMetrics: {
          grossUnlockRevenueJpy: 82000,
          unlockCount: 91,
          uniquePurchaserCount: 74,
        },
        revisionRequestedSummary: {
          mainCount: 1,
          shortCount: 1,
          totalCount: 2,
        },
      }),
    );

    render(await CreatorPage());

    expect(await screen.findByText("@contractmina")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "動画を追加" })).toHaveAttribute("href", "/creator/upload");
    expect(screen.getByRole("button", { name: "Account menu" })).toBeEnabled();
    expect(screen.queryByText("Dashboard")).not.toBeInTheDocument();
    expect(screen.queryByText("@minarei")).not.toBeInTheDocument();
    expect(screen.getByText("contract-backed creator bio")).toBeInTheDocument();
    expect(screen.getByText("¥82,000")).toBeInTheDocument();
    expect(screen.getByText("差し戻しが2件あります")).toBeInTheDocument();
    expect(screen.getByText("short 1件 / main 1件を確認してください")).toBeInTheDocument();
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

    await screen.findByText("@minarei");
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

    await screen.findByText("@minarei");
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

  it("shows a summary loading state while the workspace request is pending", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");
    const deferred = createDeferredPromise<CreatorWorkspaceSummary>();

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      },
      hasSession: true,
    });
    vi.mocked(getCreatorWorkspaceSummary).mockReturnValue(deferred.promise);

    render(await CreatorPage());

    expect(screen.getByRole("status")).toHaveTextContent("workspace summary を読み込んでいます...");
    expect(screen.queryByText("@minarei")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Top main" })).toBeInTheDocument();
  });

  it("shows a local summary error and retries without hiding managed mock sections", async () => {
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
    vi.mocked(getCreatorWorkspaceSummary)
      .mockRejectedValueOnce(new Error("boom"))
      .mockResolvedValueOnce(createCreatorWorkspaceSummary({
        creator: {
          avatar: {
            durationSeconds: null,
            id: "asset_creator_retry_avatar",
            kind: "image",
            posterUrl: null,
            url: "https://cdn.example.com/creator/retry/avatar.jpg",
          },
          bio: "retry success bio",
          displayName: "Retry Mina",
          handle: "@retrymina",
          id: "creator_mina_rei",
        },
      }));

    render(await CreatorPage());

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "creator workspace summary を読み込めませんでした。少し時間を置いてから再読み込みしてください。",
    );
    expect(screen.getByRole("button", { name: "Top main" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "再読み込み" }));

    expect(await screen.findByText("@retrymina")).toBeInTheDocument();
  });

  it("hides the revision notice when the contract response has no revision summary", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      },
      hasSession: true,
    });
    vi.mocked(getCreatorWorkspaceSummary).mockResolvedValue(
      createCreatorWorkspaceSummary({
        revisionRequestedSummary: null,
      }),
    );

    render(await CreatorPage());

    await screen.findByText("@minarei");

    expect(screen.queryByText("差し戻し")).not.toBeInTheDocument();
    expect(screen.queryByText(/差し戻しが/)).not.toBeInTheDocument();
  });

  it("falls back to a generic revision message when revision counts are inconsistent", async () => {
    vi.mocked(getCreatorWorkspaceSummary).mockResolvedValue(
      createCreatorWorkspaceSummary({
        revisionRequestedSummary: {
          mainCount: 0,
          shortCount: 0,
          totalCount: 0,
        },
      }),
    );

    render(<CreatorModeShell state={getMockCreatorModeShellState()} />);

    expect(await screen.findByText("差し戻しが0件あります")).toBeInTheDocument();
    expect(screen.getByText("修正依頼内容を確認してください")).toBeInTheDocument();
  });

  it("falls back to the unauthenticated blocked state when the summary API returns 401", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      },
      hasSession: true,
    });
    vi.mocked(getCreatorWorkspaceSummary).mockRejectedValue(
      new ApiError("API request failed with a non-success status.", {
        code: "http",
        details: JSON.stringify({
          error: {
            code: "auth_required",
            message: "creator workspace requires authentication",
          },
        }),
        status: 401,
      }),
    );

    render(await CreatorPage());

    expect(
      await screen.findByRole("heading", { name: "creator mode を開くにはログインが必要です。" }),
    ).toBeInTheDocument();
  });

  it("falls back to the capability blocked state when the summary API returns 403", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      },
      hasSession: true,
    });
    vi.mocked(getCreatorWorkspaceSummary).mockRejectedValue(
      new ApiError("API request failed with a non-success status.", {
        code: "http",
        details: JSON.stringify({
          error: {
            code: "creator_mode_unavailable",
            message: "creator mode is not available",
          },
        }),
        status: 403,
      }),
    );

    render(await CreatorPage());

    expect(
      await screen.findByRole("heading", { name: "creator mode はまだ利用できません。" }),
    ).toBeInTheDocument();
  });
});
