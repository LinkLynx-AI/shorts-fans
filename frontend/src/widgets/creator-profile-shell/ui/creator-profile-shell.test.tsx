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
import { useHasViewerSession } from "@/entities/viewer";
import { useFanAuthDialog } from "@/features/fan-auth";
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
    useHasViewerSession: vi.fn(),
  };
});

vi.mock("@/features/fan-auth", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/features/fan-auth")>();

  return {
    ...actual,
    useFanAuthDialog: vi.fn(),
  };
});

const mockedUpdateCreatorFollow = vi.mocked(updateCreatorFollow);
const mockedUseHasViewerSession = vi.mocked(useHasViewerSession);
const mockedUseFanAuthDialog = vi.mocked(useFanAuthDialog);
const openFanAuthDialog = vi.fn();

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
        routeShortId: "rooftop",
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

function getFanStatValue(): string | null {
  return screen.getByText("fans").previousElementSibling?.textContent ?? null;
}

describe("CreatorProfileShell", () => {
  beforeEach(() => {
    mockedUpdateCreatorFollow.mockReset();
    mockedUseHasViewerSession.mockReset();
    openFanAuthDialog.mockReset();
    mockedUseFanAuthDialog.mockReturnValue({
      closeFanAuthDialog: vi.fn(),
      isFanAuthDialogOpen: false,
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

    expect(screen.getByText("minarei")).toBeInTheDocument();
    expect(screen.queryByText("@minarei")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Following" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("link", { name: /Mina Rei preview 0:16/i })).toHaveAttribute(
      "href",
      "/shorts/rooftop?creatorId=creator_mina_rei&from=creator&profileFrom=search&profileQ=mina",
    );
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
    expect(getFanStatValue()).toBe("10");
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

    expect(getFanStatValue()).toBe("11");
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
