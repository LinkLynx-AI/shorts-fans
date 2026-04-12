import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
} from "@testing-library/react";

import { getFanHubState, type FanHubState, type FanHubTab } from "@/entities/fan-profile";
import {
  CurrentViewerProvider,
  ViewerSessionProvider,
} from "@/entities/viewer";
import { useFanLogoutEntry } from "@/features/fan-auth";

import { FanHubShell } from "./fan-hub-shell";

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

vi.mock("@/features/fan-auth", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth")>("@/features/fan-auth");

  return {
    ...actual,
    useFanLogoutEntry: vi.fn(),
  };
});

function renderFanHubShell(currentViewer: {
  activeMode: "fan" | "creator";
  canAccessCreatorMode: boolean;
  id: string;
} | null, activeTab: FanHubTab = "library") {
  return renderFanHubShellWithState(currentViewer, getFanHubState(activeTab));
}

function renderFanHubShellWithState(currentViewer: {
  activeMode: "fan" | "creator";
  canAccessCreatorMode: boolean;
  id: string;
} | null, state: FanHubState) {
  return render(
    <ViewerSessionProvider hasSession>
      <CurrentViewerProvider currentViewer={currentViewer}>
        <FanHubShell state={state} />
      </CurrentViewerProvider>
    </ViewerSessionProvider>,
  );
}

describe("FanHubShell account menu", () => {
  beforeEach(() => {
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    vi.mocked(useFanLogoutEntry).mockReset();
    vi.mocked(useFanLogoutEntry).mockReturnValue({
      clearError: vi.fn(),
      errorMessage: null,
      isSubmitting: false,
      logout: vi.fn().mockResolvedValue(true),
    });
  });

  it("shows the registration entry for viewers without creator access", async () => {
    const user = userEvent.setup();

    renderFanHubShell({
      activeMode: "fan",
      canAccessCreatorMode: false,
      id: "viewer_123",
    });

    await user.click(screen.getByRole("button", { name: "Account menu" }));

    expect(screen.getByRole("link", { name: "Creator登録を始める" })).toHaveAttribute(
      "href",
      "/fan/creator/register",
    );
    expect(screen.getByRole("button", { name: "ログアウト" })).toBeInTheDocument();
  });

  it("shows the creator switch entry for creator-capable viewers", async () => {
    const user = userEvent.setup();

    renderFanHubShell({
      activeMode: "fan",
      canAccessCreatorMode: true,
      id: "viewer_123",
    });

    await user.click(screen.getByRole("button", { name: "Account menu" }));

    expect(screen.getByRole("button", { name: "Creator mode に切り替え" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "ログアウト" })).toBeInTheDocument();
  });

  it("shows a pending logout label while the action is in flight", async () => {
    const user = userEvent.setup();
    const logout = vi.fn().mockResolvedValue(true);

    vi.mocked(useFanLogoutEntry).mockReturnValue({
      clearError: vi.fn(),
      errorMessage: null,
      isSubmitting: true,
      logout,
    });

    renderFanHubShell({
      activeMode: "fan",
      canAccessCreatorMode: true,
      id: "viewer_123",
    });

    await user.click(screen.getByRole("button", { name: "Account menu" }));

    expect(screen.getByRole("button", { name: "ログアウトしています..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Creator mode に切り替え" })).toBeDisabled();
    expect(logout).not.toHaveBeenCalled();
  });

  it("shows a retryable error when logout fails", async () => {
    const user = userEvent.setup();

    vi.mocked(useFanLogoutEntry).mockReturnValue({
      clearError: vi.fn(),
      errorMessage: "ログアウトできませんでした。少し時間を置いてからやり直してください。",
      isSubmitting: false,
      logout: vi.fn().mockResolvedValue(false),
    });

    renderFanHubShell({
      activeMode: "fan",
      canAccessCreatorMode: false,
      id: "viewer_123",
    });

    await user.click(screen.getByRole("button", { name: "Account menu" }));

    expect(screen.getByRole("alert")).toHaveTextContent(
      "ログアウトできませんでした。少し時間を置いてからやり直してください。",
    );
  });

  it("invokes logout from the account menu action", async () => {
    const user = userEvent.setup();
    const logout = vi.fn().mockResolvedValue(true);

    vi.mocked(useFanLogoutEntry).mockReturnValue({
      clearError: vi.fn(),
      errorMessage: null,
      isSubmitting: false,
      logout,
    });

    renderFanHubShell({
      activeMode: "fan",
      canAccessCreatorMode: false,
      id: "viewer_123",
    });

    await user.click(screen.getByRole("button", { name: "Account menu" }));
    await user.click(screen.getByRole("button", { name: "ログアウト" }));

    expect(logout).toHaveBeenCalledTimes(1);
  });

  it("renders the pinned short poster in the fan hub grid", () => {
    renderFanHubShell(
      {
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_123",
      },
      "pinned",
    );

    const pinnedLink = screen.getByRole("link", { name: /雨上がりの balcony preview/i });
    const pinnedPoster = pinnedLink.querySelector("[style*='background-image']");

    expect(pinnedPoster).not.toBeNull();
    expect(pinnedPoster).toHaveStyle({
      backgroundImage: 'url("https://cdn.example.com/shorts/sora-after-rain-poster.jpg")',
    });
  });

  it("uses an owner-preview fallback label when the library entry short caption is blank", () => {
    const libraryState = getFanHubState("library");
    const firstItem = libraryState.libraryItems[0];

    if (!firstItem) {
      throw new Error("library fixture missing");
    }

    renderFanHubShellWithState(
      {
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_123",
      },
      {
        ...libraryState,
        libraryItems: [
          {
            ...firstItem,
            access: {
              ...firstItem.access,
              reason: "owner_preview",
              status: "owner",
            },
            entryShort: {
              ...firstItem.entryShort,
              caption: "   ",
            },
          },
        ],
      },
    );

    expect(screen.getByRole("button", { name: `${firstItem.creator.displayName} owner preview main` })).toBeInTheDocument();
  });
});
