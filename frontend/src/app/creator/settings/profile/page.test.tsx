import {
  render,
  screen,
} from "@testing-library/react";

import {
  CurrentViewerProvider,
  ViewerSessionProvider,
} from "@/entities/viewer";
import { getCreatorWorkspaceSummary } from "@/widgets/creator-mode-shell/api/get-creator-workspace-summary";

import CreatorProfileSettingsPage from "./page";

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

vi.mock("@/widgets/creator-mode-shell/api/get-creator-workspace-summary", () => ({
  getCreatorWorkspaceSummary: vi.fn(),
}));

type CreatorWorkspaceSummary = Awaited<ReturnType<typeof getCreatorWorkspaceSummary>>;

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

describe("CreatorProfileSettingsPage", () => {
  beforeEach(() => {
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    vi.mocked(getCreatorWorkspaceSummary).mockReset();
    vi.mocked(getCreatorWorkspaceSummary).mockResolvedValue(createCreatorWorkspaceSummary());
  });

  it("renders the profile edit form with creator summary values", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      },
      hasSession: true,
    });

    render(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider
          currentViewer={{
            activeMode: "creator",
            canAccessCreatorMode: true,
            id: "viewer_creator_001",
          }}
        >
          {await CreatorProfileSettingsPage()}
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    expect(await screen.findByDisplayValue("Mina Rei")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "プロフィールを編集" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Handle" })).toHaveValue("@minarei");
    expect(screen.getByRole("textbox", { name: "Bio" })).toHaveValue(
      "quiet rooftop と hotel light の preview を軸に投稿。",
    );
    expect(screen.getByRole("link", { name: "ワークスペースへ戻る" })).toHaveAttribute("href", "/creator");
  });

  it("renders the login-required state for unauthenticated viewers", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: false,
    });

    render(await CreatorProfileSettingsPage());

    expect(screen.getByRole("heading", { name: "creator mode を開くにはログインが必要です。" })).toBeInTheDocument();
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

    render(await CreatorProfileSettingsPage());

    expect(screen.getByRole("heading", { name: "creator mode に切り替えてから開いてください。" })).toBeInTheDocument();
  });
});
