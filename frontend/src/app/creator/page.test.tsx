import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { switchViewerActiveMode } from "@/features/creator-entry/api/switch-viewer-active-mode";
import { ApiError } from "@/shared/api";
import { getCreatorWorkspaceSummary } from "@/widgets/creator-mode-shell/api/get-creator-workspace-summary";
import { getCreatorWorkspaceTopPerformers } from "@/widgets/creator-mode-shell/api/get-creator-workspace-top-performers";
import {
  getCreatorWorkspacePreviewMains,
  getCreatorWorkspacePreviewShorts,
} from "@/widgets/creator-mode-shell/api/get-creator-workspace-preview-collections";
import {
  getCreatorWorkspacePreviewMainDetail,
  getCreatorWorkspacePreviewShortDetail,
} from "@/widgets/creator-mode-shell/api/get-creator-workspace-preview-detail";
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

vi.mock("@/widgets/creator-mode-shell/api/get-creator-workspace-top-performers", () => ({
  getCreatorWorkspaceTopPerformers: vi.fn(),
}));

vi.mock("@/widgets/creator-mode-shell/api/get-creator-workspace-preview-collections", () => ({
  getCreatorWorkspacePreviewMains: vi.fn(),
  getCreatorWorkspacePreviewShorts: vi.fn(),
}));

vi.mock("@/widgets/creator-mode-shell/api/get-creator-workspace-preview-detail", () => ({
  getCreatorWorkspacePreviewMainDetail: vi.fn(),
  getCreatorWorkspacePreviewShortDetail: vi.fn(),
}));

type CreatorWorkspaceSummary = Awaited<ReturnType<typeof getCreatorWorkspaceSummary>>;
type CreatorWorkspaceTopPerformers = Awaited<ReturnType<typeof getCreatorWorkspaceTopPerformers>>;
type CreatorWorkspacePreviewShorts = Awaited<ReturnType<typeof getCreatorWorkspacePreviewShorts>>;
type CreatorWorkspacePreviewMains = Awaited<ReturnType<typeof getCreatorWorkspacePreviewMains>>;
type CreatorWorkspacePreviewShortDetail = Awaited<ReturnType<typeof getCreatorWorkspacePreviewShortDetail>>;
type CreatorWorkspacePreviewMainDetail = Awaited<ReturnType<typeof getCreatorWorkspacePreviewMainDetail>>;

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

function createCreatorWorkspacePreviewShorts(
  overrides: Partial<CreatorWorkspacePreviewShorts> = {},
): CreatorWorkspacePreviewShorts {
  return {
    items: [
      {
        canonicalMainId: "main_quiet_rooftop",
        id: "short_quiet_rooftop",
        media: {
          durationSeconds: 16,
          id: "asset_short_quiet_rooftop",
          kind: "video",
          posterUrl: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop-poster.jpg",
        },
        previewDurationSeconds: 16,
      },
    ],
    page: {
      hasNext: false,
      nextCursor: null,
    },
    requestId: "req_creator_workspace_shorts_001",
    ...overrides,
  };
}

function createCreatorWorkspaceTopPerformers(
  overrides: Partial<CreatorWorkspaceTopPerformers> = {},
): CreatorWorkspaceTopPerformers {
  return {
    topMain: {
      id: "main_quiet_rooftop",
      media: {
        durationSeconds: 720,
        id: "asset_main_quiet_rooftop",
        kind: "video",
        posterUrl: "https://signed.example.com/mains/top-quiet-rooftop.jpg",
      },
      unlockCount: 238,
    },
    topShort: {
      attributedUnlockCount: 238,
      id: "short_quiet_rooftop",
      media: {
        durationSeconds: 16,
        id: "asset_short_quiet_rooftop",
        kind: "video",
        posterUrl: "https://cdn.example.com/shorts/top-quiet-rooftop.jpg",
      },
    },
    ...overrides,
  };
}

function createCreatorWorkspacePreviewShortDetail(
  overrides: Partial<CreatorWorkspacePreviewShortDetail> = {},
): CreatorWorkspacePreviewShortDetail {
  return {
    access: {
      mainId: "main_quiet_rooftop",
      reason: "owner_preview",
      status: "owner",
    },
    creator: {
      avatar: null,
      bio: "quiet rooftop と hotel light の preview を軸に投稿。",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_mina_rei",
    },
    kind: "preview-short",
    requestId: "req_creator_workspace_short_detail_001",
    short: {
      caption: "quiet rooftop preview.",
      canonicalMainId: "main_quiet_rooftop",
      creatorId: "creator_mina_rei",
      id: "short_quiet_rooftop",
      media: {
        durationSeconds: 16,
        id: "asset_short_quiet_rooftop",
        kind: "video",
        posterUrl: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop-poster.jpg",
        url: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop.mp4",
      },
      previewDurationSeconds: 16,
      title: "quiet rooftop preview",
    },
    ...overrides,
  };
}

function createCreatorWorkspacePreviewMainDetail(
  overrides: Partial<CreatorWorkspacePreviewMainDetail> = {},
): CreatorWorkspacePreviewMainDetail {
  return {
    access: {
      mainId: "main_quiet_rooftop",
      reason: "owner_preview",
      status: "owner",
    },
    creator: {
      avatar: null,
      bio: "quiet rooftop と hotel light の preview を軸に投稿。",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_mina_rei",
    },
    entryShort: {
      caption: "quiet rooftop preview.",
      canonicalMainId: "main_quiet_rooftop",
      creatorId: "creator_mina_rei",
      id: "short_quiet_rooftop",
      media: {
        durationSeconds: 16,
        id: "asset_short_quiet_rooftop",
        kind: "video",
        posterUrl: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop-poster.jpg",
        url: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop.mp4",
      },
      previewDurationSeconds: 16,
      title: "quiet rooftop preview",
    },
    kind: "preview-main",
    main: {
      durationSeconds: 720,
      id: "main_quiet_rooftop",
      media: {
        durationSeconds: 720,
        id: "asset_main_quiet_rooftop",
        kind: "video",
        posterUrl: "https://cdn.example.com/creator/preview/mains/quiet-rooftop-poster.jpg",
        url: "https://cdn.example.com/creator/preview/mains/quiet-rooftop.mp4",
      },
      priceJpy: 1800,
      title: "quiet rooftop main",
    },
    requestId: "req_creator_workspace_main_detail_001",
    ...overrides,
  };
}

function createCreatorWorkspacePreviewMains(
  overrides: Partial<CreatorWorkspacePreviewMains> = {},
): CreatorWorkspacePreviewMains {
  return {
    items: [
      {
        durationSeconds: 720,
        id: "main_quiet_rooftop",
        leadShortId: "short_quiet_rooftop",
        media: {
          durationSeconds: 720,
          id: "asset_main_quiet_rooftop",
          kind: "video",
          posterUrl: "https://cdn.example.com/creator/preview/mains/quiet-rooftop-poster.jpg",
        },
        priceJpy: 1800,
      },
    ],
    page: {
      hasNext: false,
      nextCursor: null,
    },
    requestId: "req_creator_workspace_mains_001",
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
    vi.mocked(getCreatorWorkspaceTopPerformers).mockReset();
    vi.mocked(getCreatorWorkspacePreviewMains).mockReset();
    vi.mocked(getCreatorWorkspacePreviewShorts).mockReset();
    vi.mocked(getCreatorWorkspacePreviewMainDetail).mockReset();
    vi.mocked(getCreatorWorkspacePreviewShortDetail).mockReset();
    vi.mocked(getCreatorWorkspaceSummary).mockResolvedValue(createCreatorWorkspaceSummary());
    vi.mocked(getCreatorWorkspaceTopPerformers).mockResolvedValue(createCreatorWorkspaceTopPerformers());
    vi.mocked(getCreatorWorkspacePreviewShorts).mockResolvedValue(createCreatorWorkspacePreviewShorts());
    vi.mocked(getCreatorWorkspacePreviewMains).mockResolvedValue(createCreatorWorkspacePreviewMains());
    vi.mocked(getCreatorWorkspacePreviewShortDetail).mockResolvedValue(createCreatorWorkspacePreviewShortDetail());
    vi.mocked(getCreatorWorkspacePreviewMainDetail).mockResolvedValue(createCreatorWorkspacePreviewMainDetail());
  });

  it("opens top performers with the same preview detail flow as the lower preview cards", async () => {
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
    expect(screen.getByRole("button", { name: /^Top main\b/ })).toBeEnabled();
    expect(screen.getByRole("button", { name: /^Top short\b/ })).toBeEnabled();
    expect(screen.getAllByText("238 unlocks")).toHaveLength(2);
    expect(await screen.findByTestId("creator-workspace-preview-tile")).toBeInTheDocument();
    expect(screen.queryByText("owner preview 一覧から取得した本編データです。")).not.toBeInTheDocument();
    expect(screen.queryByText("owner preview 一覧から取得したショートデータです。")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /^Top main\b/ }));

    expect(await screen.findByText("owner preview 一覧から取得した本編データです。")).toBeInTheDocument();
    expect(screen.getByText("¥1,800")).toBeInTheDocument();
    expect(screen.getAllByText("12:00")).toHaveLength(2);

    await user.click(screen.getByRole("button", { name: "Back" }));

    expect(await screen.findByRole("button", { name: "Main" })).toHaveAttribute("aria-pressed", "true");

    await user.click(screen.getByRole("button", { name: /^Top short\b/ }));

    expect(await screen.findByText("owner preview 一覧から取得したショートデータです。")).toBeInTheDocument();
    expect(screen.getAllByText("0:16")).toHaveLength(2);

    await user.click(screen.getByRole("button", { name: "Back" }));

    expect(await screen.findByRole("button", { name: "Shorts" })).toHaveAttribute("aria-pressed", "true");
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
    expect(screen.getByRole("link", { name: "プロフィールを編集" })).toHaveAttribute(
      "href",
      "/creator/settings/profile",
    );
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
    const topPerformersDeferred = createDeferredPromise<CreatorWorkspaceTopPerformers>();
    const previewShortsDeferred = createDeferredPromise<CreatorWorkspacePreviewShorts>();
    const previewMainsDeferred = createDeferredPromise<CreatorWorkspacePreviewMains>();

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      },
      hasSession: true,
    });
    vi.mocked(getCreatorWorkspaceSummary).mockReturnValue(deferred.promise);
    vi.mocked(getCreatorWorkspaceTopPerformers).mockReturnValue(topPerformersDeferred.promise);
    vi.mocked(getCreatorWorkspacePreviewShorts).mockReturnValue(previewShortsDeferred.promise);
    vi.mocked(getCreatorWorkspacePreviewMains).mockReturnValue(previewMainsDeferred.promise);

    render(await CreatorPage());

    expect(screen.getByText("workspace summary を読み込んでいます...")).toBeInTheDocument();
    expect(screen.queryByText("@minarei")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^Top main\b/ })).toBeInTheDocument();
  });

  it("renders contract-backed preview lists and allows inline playback in lower detail cards", async () => {
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

    expect(await screen.findByText("0:16")).toBeInTheDocument();
    const previewTiles = screen.getAllByTestId("creator-workspace-preview-tile");
    expect(previewTiles).toHaveLength(1);

    await user.click(screen.getByRole("button", { name: "ショート詳細を開く 1件目 0:16" }));

    expect(screen.getByText("owner preview 一覧から取得したショートデータです。")).toBeInTheDocument();
    expect(screen.queryByText("short_quiet_rooftop")).not.toBeInTheDocument();
    expect(screen.queryByText("main_quiet_rooftop")).not.toBeInTheDocument();
    expect(screen.queryByText("asset_short_quiet_rooftop")).not.toBeInTheDocument();
    expect(await screen.findByRole("button", { name: "ショートを再生" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "ショートを再生" }));

    expect(screen.getByLabelText("ショート動画")).toHaveAttribute(
      "src",
      "https://cdn.example.com/creator/preview/shorts/quiet-rooftop.mp4",
    );

    await user.click(screen.getByRole("button", { name: "本編詳細を開く 1件目 ¥1,800 12:00" }));

    expect(screen.getByText("owner preview 一覧から取得した本編データです。")).toBeInTheDocument();
    expect(screen.queryByText("asset_main_quiet_rooftop")).not.toBeInTheDocument();
    expect(screen.getByText("¥1,800")).toBeInTheDocument();
    expect(screen.getAllByText("12:00")).toHaveLength(2);
    expect(await screen.findByRole("button", { name: "本編を再生" })).toBeInTheDocument();
    expect(screen.queryByLabelText("ショート動画")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "本編を再生" }));

    expect(screen.getByLabelText("本編動画")).toHaveAttribute(
      "src",
      "https://cdn.example.com/creator/preview/mains/quiet-rooftop.mp4",
    );

    await user.click(screen.getByRole("button", { name: "Back" }));

    await user.click(screen.getByRole("button", { name: "Main" }));

    expect(await screen.findByText("12:00")).toBeInTheDocument();
    expect(await screen.findByText("¥1,800")).toBeInTheDocument();
    expect(screen.queryByText("owner preview 一覧から取得した本編データです。")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "本編詳細を開く 1件目 ¥1,800 12:00" }));

    expect(await screen.findByRole("button", { name: "本編を再生" })).toBeInTheDocument();
    expect(screen.queryByLabelText("本編動画")).not.toBeInTheDocument();
  });

  it("shows a retryable lower-list error without hiding the rest of the workspace", async () => {
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
    vi.mocked(getCreatorWorkspacePreviewShorts)
      .mockRejectedValueOnce(new Error("boom"))
      .mockResolvedValueOnce(createCreatorWorkspacePreviewShorts());
    vi.mocked(getCreatorWorkspacePreviewMains)
      .mockRejectedValueOnce(new Error("boom"))
      .mockResolvedValueOnce(createCreatorWorkspacePreviewMains());

    render(await CreatorPage());

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "動画一覧を読み込めませんでした。少し時間を置いてから再読み込みしてください。",
    );
    expect(screen.getByRole("button", { name: /^Top main\b/ })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "再読み込み" }));

    expect(await screen.findByText("0:16")).toBeInTheDocument();
  });

  it("shows a local preview detail error and retries inside the detail screen", async () => {
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
    vi.mocked(getCreatorWorkspacePreviewShortDetail)
      .mockRejectedValueOnce(new Error("boom"))
      .mockResolvedValueOnce(createCreatorWorkspacePreviewShortDetail());

    render(await CreatorPage());

    await screen.findByText("0:16");
    await user.click(screen.getByRole("button", { name: "ショート詳細を開く 1件目 0:16" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "動画詳細を読み込めませんでした。少し時間を置いてから再読み込みしてください。",
    );
    expect(screen.getByRole("button", { name: "Back" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "再読み込み" }));

    expect(await screen.findByRole("button", { name: "ショートを再生" })).toBeInTheDocument();
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
    expect(screen.getByRole("button", { name: /^Top main\b/ })).toBeInTheDocument();

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

  it("shows a retryable top performers error without hiding the rest of the workspace", async () => {
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
    vi.mocked(getCreatorWorkspaceTopPerformers)
      .mockRejectedValueOnce(new Error("boom"))
      .mockResolvedValueOnce(createCreatorWorkspaceTopPerformers());

    render(await CreatorPage());

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "top performers を読み込めませんでした。少し時間を置いてから再読み込みしてください。",
    );
    expect(screen.getByText("@minarei")).toBeInTheDocument();
    expect(screen.getAllByTestId("creator-workspace-preview-tile")).toHaveLength(1);

    await user.click(screen.getByRole("button", { name: "再読み込み" }));

    expect(await screen.findAllByText("238 unlocks")).toHaveLength(2);
  });

  it("hides the top performers section when the contract response is empty", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      },
      hasSession: true,
    });
    vi.mocked(getCreatorWorkspaceTopPerformers).mockResolvedValue(
      createCreatorWorkspaceTopPerformers({
        topMain: null,
        topShort: null,
      }),
    );

    render(await CreatorPage());

    await screen.findByText("@minarei");

    expect(screen.queryByRole("button", { name: /^Top main\b/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /^Top short\b/ })).not.toBeInTheDocument();
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
