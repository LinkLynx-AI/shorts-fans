import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import {
  CreatorFollowApiError,
  updateCreatorFollow,
} from "@/entities/creator";
import {
  useCurrentViewer,
  useHasViewerSession,
} from "@/entities/viewer";
import { useCreatorModeEntry } from "@/features/creator-entry";
import { useFanAuthDialogControls } from "@/features/fan-auth";
import {
  CreatorProfileShell,
  type CreatorProfileShellState,
} from "@/widgets/creator-profile-shell";

vi.mock("@/entities/creator", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/entities/creator")>();

  return {
    ...actual,
    updateCreatorFollow: vi.fn(),
  };
});

vi.mock("@/entities/viewer", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/entities/viewer")>();

  return {
    ...actual,
    useCurrentViewer: vi.fn(),
    useHasViewerSession: vi.fn(),
  };
});

vi.mock("@/features/creator-entry", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/features/creator-entry")>();

  return {
    ...actual,
    useCreatorModeEntry: vi.fn(),
  };
});

vi.mock("@/features/fan-auth", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/features/fan-auth")>();

  return {
    ...actual,
    useFanAuthDialogControls: vi.fn(),
  };
});

const mockedUpdateCreatorFollow = vi.mocked(updateCreatorFollow);
const mockedUseCurrentViewer = vi.mocked(useCurrentViewer);
const mockedUseHasViewerSession = vi.mocked(useHasViewerSession);
const mockedUseCreatorModeEntry = vi.mocked(useCreatorModeEntry);
const mockedUseFanAuthDialogControls = vi.mocked(useFanAuthDialogControls);
const openFanAuthDialog = vi.fn();
const enterCreatorMode = vi.fn();

function buildReadyState(
  overrides?: Partial<Extract<CreatorProfileShellState, { kind: "ready" }>>,
): Extract<CreatorProfileShellState, { kind: "ready" }> {
  return {
    creator: {
      avatar: null,
      bio: "quiet rooftop と hotel light の preview を軸に投稿。",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_mina_rei",
    },
    kind: "ready" as const,
    shorts: [
      {
        canonicalMainId: "main_mina_quiet_rooftop",
        creatorId: "creator_mina_rei",
        id: "short_mina_rooftop",
        media: {
          durationSeconds: 16,
          id: "asset_short_mina_rooftop",
          kind: "video" as const,
          posterUrl: null,
          url: "https://cdn.example.com/shorts/mina-rooftop.mp4",
        },
        previewDurationSeconds: 16,
        routeShortId: "short_mina_rooftop",
      },
    ],
    stats: {
      fanCount: 10,
      shortCount: 2,
    },
    viewer: {
      isFollowing: false,
    },
    ...overrides,
  };
}

function buildEmptyState(
  overrides?: Partial<Extract<CreatorProfileShellState, { kind: "empty" }>>,
): Extract<CreatorProfileShellState, { kind: "empty" }> {
  return {
    creator: {
      avatar: null,
      bio: "quiet rooftop と hotel light の preview を軸に投稿。",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_mina_rei",
    },
    kind: "empty",
    shorts: [],
    stats: {
      fanCount: 10,
      shortCount: 2,
    },
    viewer: {
      isFollowing: false,
    },
    ...overrides,
  };
}

function getFollowerStatValue(): string | null {
  return screen.getByText("Followers").previousElementSibling?.textContent ?? null;
}

describe("CreatorProfileShell", () => {
  beforeEach(() => {
    mockedUpdateCreatorFollow.mockReset();
    mockedUseCurrentViewer.mockReset();
    mockedUseHasViewerSession.mockReset();
    mockedUseCreatorModeEntry.mockReset();
    openFanAuthDialog.mockReset();
    enterCreatorMode.mockReset();
    mockedUseCurrentViewer.mockReturnValue(null);
    mockedUseCreatorModeEntry.mockReturnValue({
      clearError: vi.fn(),
      enterCreatorMode,
      errorMessage: null,
      isSubmitting: false,
    });
    mockedUseFanAuthDialogControls.mockReturnValue({
      closeFanAuthDialog: vi.fn(),
      openFanAuthDialog,
    });
  });

  it("renders the contract-backed creator profile header and short grid", () => {
    mockedUseHasViewerSession.mockReturnValue(true);

    render(
      <CreatorProfileShell
        routeState={{ from: "search", q: "mina" }}
        state={buildReadyState({
          viewer: {
            isFollowing: true,
          },
        })}
      />,
    );

    expect(screen.getAllByText("@minarei")).toHaveLength(2);
    expect(screen.getByRole("heading", { name: "Mina Rei" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Following" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("link", { name: /Mina Rei preview 0:16/i })).toHaveAttribute(
      "href",
      "/shorts/short_mina_rooftop?creatorId=creator_mina_rei&from=creator&profileFrom=search&profileQ=mina",
    );
  });

  it("replaces the follow CTA with a creator page entry button on self profile", async () => {
    const user = userEvent.setup();

    mockedUseCurrentViewer.mockReturnValue({
      activeMode: "fan",
      canAccessCreatorMode: true,
      id: "11111111-1111-1111-1111-111111111111",
    });
    mockedUseHasViewerSession.mockReturnValue(true);

    render(
      <CreatorProfileShell
        routeState={{ from: "search", q: "mika" }}
        state={buildReadyState({
          creator: {
            avatar: null,
            bio: "Public shorts から paid main へつながる creator mock profile.",
            displayName: "Mika Aoi",
            handle: "@mikaaoi",
            id: "creator_11111111111111111111111111111111",
          },
          shorts: [
            {
              canonicalMainId: "main_mika_preview",
              creatorId: "creator_11111111111111111111111111111111",
              id: "short_mika_preview",
              media: {
                durationSeconds: 16,
                id: "asset_short_mika_preview",
                kind: "video",
                posterUrl: null,
                url: "https://cdn.example.com/shorts/mika-preview.mp4",
              },
              previewDurationSeconds: 16,
              routeShortId: "short_mika_preview",
            },
          ],
        })}
      />,
    );

    expect(screen.getByRole("button", { name: "Creatorページを開く" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Follow" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Creatorページを開く" }));

    expect(enterCreatorMode).toHaveBeenCalledTimes(1);
    expect(mockedUpdateCreatorFollow).not.toHaveBeenCalled();
  });

  it("updates the CTA state and fan count after an authenticated unfollow succeeds", async () => {
    const user = userEvent.setup();

    mockedUseHasViewerSession.mockReturnValue(true);
    mockedUpdateCreatorFollow.mockResolvedValue({
      stats: {
        fanCount: 10,
      },
      viewer: {
        isFollowing: false,
      },
    });

    render(
      <CreatorProfileShell
        routeState={{ from: "feed", tab: "recommended" }}
        state={buildReadyState({
          stats: {
            fanCount: 11,
            shortCount: 2,
          },
          viewer: {
            isFollowing: true,
          },
        })}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Following" }));

    await waitFor(() => {
      expect(mockedUpdateCreatorFollow).toHaveBeenCalledWith({
        action: "unfollow",
        creatorId: "creator_mina_rei",
      });
    });

    expect(screen.getByRole("button", { name: "Follow" })).toHaveAttribute("aria-pressed", "false");
    expect(getFollowerStatValue()).toBe("10");
  });

  it("keeps the CTA pending during follow and ignores duplicate clicks", async () => {
    const user = userEvent.setup();
    let resolveFollow: ((value: Awaited<ReturnType<typeof updateCreatorFollow>>) => void) | undefined;

    mockedUseHasViewerSession.mockReturnValue(true);
    mockedUpdateCreatorFollow.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveFollow = resolve;
        }),
    );

    render(
      <CreatorProfileShell
        routeState={{ from: "feed", tab: "recommended" }}
        state={buildReadyState()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Follow" }));
    await user.click(screen.getByRole("button", { name: "Following..." }));

    expect(mockedUpdateCreatorFollow).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("button", { name: "Following..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Following..." })).toHaveAttribute("aria-busy", "true");

    resolveFollow?.({
      stats: {
        fanCount: 11,
      },
      viewer: {
        isFollowing: true,
      },
    });

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Following" })).toBeEnabled();
    });

    expect(getFollowerStatValue()).toBe("11");
  });

  it("opens the shared auth modal when an anonymous viewer presses follow", async () => {
    const user = userEvent.setup();

    mockedUseHasViewerSession.mockReturnValue(false);

    render(
      <CreatorProfileShell
        routeState={{ from: "feed", tab: "recommended" }}
        state={buildEmptyState()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Follow" }));

    expect(openFanAuthDialog).toHaveBeenCalledTimes(1);
    expect(mockedUpdateCreatorFollow).not.toHaveBeenCalled();
  });

  it("reopens the auth modal when the follow request returns auth_required", async () => {
    const user = userEvent.setup();

    mockedUseHasViewerSession.mockReturnValue(true);
    mockedUpdateCreatorFollow.mockRejectedValue(
      new CreatorFollowApiError("auth_required", "creator follow requires authentication", {
        requestId: "req_creator_follow_put_auth_required_001",
        status: 401,
      }),
    );

    render(
      <CreatorProfileShell
        routeState={{ from: "feed", tab: "recommended" }}
        state={buildReadyState()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Follow" }));

    await waitFor(() => {
      expect(openFanAuthDialog).toHaveBeenCalledTimes(1);
    });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("shows an inline error when follow mutation fails", async () => {
    const user = userEvent.setup();

    mockedUseHasViewerSession.mockReturnValue(true);
    mockedUpdateCreatorFollow.mockRejectedValue(
      new CreatorFollowApiError("not_found", "creator was not found", {
        requestId: "req_creator_follow_put_not_found_001",
        status: 404,
      }),
    );

    render(
      <CreatorProfileShell
        routeState={{ from: "feed", tab: "recommended" }}
        state={buildReadyState()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Follow" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("この creator profile は現在利用できません。");
    expect(screen.getByRole("button", { name: "Follow" })).toHaveAttribute("aria-pressed", "false");
  });

  it("renders the empty creator profile state without disabling the follow CTA", () => {
    mockedUseHasViewerSession.mockReturnValue(true);

    render(
      <CreatorProfileShell
        routeState={{ from: "feed", tab: "recommended" }}
        state={buildEmptyState()}
      />,
    );

    expect(screen.getByText("まだ公開中の short はありません。")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Follow" })).toBeEnabled();
    expect(screen.queryByRole("link", { name: /preview/i })).not.toBeInTheDocument();
  });
});
