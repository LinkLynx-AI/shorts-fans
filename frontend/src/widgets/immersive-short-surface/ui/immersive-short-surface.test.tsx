import userEvent from "@testing-library/user-event";
import { render, screen, waitFor } from "@testing-library/react";

import {
  CreatorFollowApiError,
  updateCreatorFollow,
} from "@/entities/creator";
import { ViewerSessionProvider } from "@/entities/viewer";
import { useFanAuthDialog } from "@/features/fan-auth";
import { getFeedSurfaceByTab, getShortSurfaceById } from "@/widgets/immersive-short-surface";

import { ImmersiveShortSurface } from "./immersive-short-surface";

const push = vi.fn();

vi.mock("@/entities/creator", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/entities/creator")>();

  return {
    ...actual,
    updateCreatorFollow: vi.fn(),
  };
});

vi.mock("@/features/fan-auth", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/features/fan-auth")>();

  return {
    ...actual,
    useFanAuthDialog: vi.fn(),
  };
});

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push,
  }),
}));

const mockedUpdateCreatorFollow = vi.mocked(updateCreatorFollow);
const mockedUseFanAuthDialog = vi.mocked(useFanAuthDialog);
const openFanAuthDialog = vi.fn();

function renderWithViewerSession(
  ui: React.ReactElement,
  { hasSession }: { hasSession: boolean },
) {
  return render(
    <ViewerSessionProvider hasSession={hasSession}>
      {ui}
    </ViewerSessionProvider>,
  );
}

describe("ImmersiveShortSurface", () => {
  const feedSurface = getFeedSurfaceByTab("recommended");
  const detailSurface = getShortSurfaceById("rooftop");
  const continueMainSurface = getShortSurfaceById("softlight");
  const directUnlockSurface = getShortSurfaceById("afterrain");
  const ownerPreviewSurface = getShortSurfaceById("balcony");
  const feedDialogTitle = "quiet rooftop preview の続きを見る";
  const detailDialogTitle = detailSurface ? "quiet rooftop preview の続きを見る" : "";
  const pinnedDetailOrigin = {
    from: "short" as const,
    shortFanTab: "pinned" as const,
    shortId: "rooftop",
  };

  beforeEach(() => {
    mockedUpdateCreatorFollow.mockReset();
    mockedUseFanAuthDialog.mockReset();
    openFanAuthDialog.mockReset();
    mockedUseFanAuthDialog.mockReturnValue({
      closeFanAuthDialog: vi.fn(),
      isFanAuthDialogOpen: false,
      openFanAuthDialog,
    });
  });

  afterEach(() => {
    push.mockReset();
    vi.unstubAllGlobals();
  });

  it("opens the mini paywall for setup-required feed content", async () => {
    const user = userEvent.setup();

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    expect(screen.getByRole("link", { name: /おすすめ/i })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "href",
      "/creators/creator_mina_rei?from=feed&tab=recommended",
    );
    expect(screen.queryByRole("link", { name: /Back/i })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pinned short" })).toBeInTheDocument();
    expect(screen.getByText("Follow")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(screen.getByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();
  });

  it("updates the feed follow CTA after an authenticated follow succeeds", async () => {
    const user = userEvent.setup();

    mockedUpdateCreatorFollow.mockResolvedValue({
      stats: {
        fanCount: 12,
      },
      viewer: {
        isFollowing: true,
      },
    });

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Follow" }));

    await waitFor(() => {
      expect(mockedUpdateCreatorFollow).toHaveBeenCalledWith({
        action: "follow",
        creatorId: feedSurface.creator.id,
      });
      expect(screen.getByRole("button", { name: "Following" })).toHaveAttribute("aria-pressed", "true");
    });
  });

  it("keeps the feed follow CTA pending and ignores duplicate clicks", async () => {
    const user = userEvent.setup();
    let resolveUpdate: (() => void) | undefined;

    mockedUpdateCreatorFollow.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveUpdate = () => {
            resolve({
              stats: {
                fanCount: 12,
              },
              viewer: {
                isFollowing: true,
              },
            });
          };
        }),
    );

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    const followButton = screen.getByRole("button", { name: "Follow" });

    await user.click(followButton);
    await user.click(screen.getByRole("button", { name: "Following..." }));

    expect(mockedUpdateCreatorFollow).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("button", { name: "Following..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Following..." })).toHaveAttribute("aria-busy", "true");

    resolveUpdate?.();

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Following" })).toBeInTheDocument();
    });
  });

  it("opens the shared auth dialog when an unauthenticated viewer presses feed follow", async () => {
    const user = userEvent.setup();

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: "Follow" }));

    expect(openFanAuthDialog).toHaveBeenCalledTimes(1);
    expect(mockedUpdateCreatorFollow).not.toHaveBeenCalled();
  });

  it("reopens the shared auth dialog when the feed follow request returns auth_required", async () => {
    const user = userEvent.setup();

    mockedUpdateCreatorFollow.mockRejectedValue(
      new CreatorFollowApiError("auth_required", "creator follow requires authentication", {
        requestId: "req_creator_follow_put_auth_required_001",
        status: 401,
      }),
    );

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Follow" }));

    await waitFor(() => {
      expect(openFanAuthDialog).toHaveBeenCalledTimes(1);
    });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("shows an inline error when feed follow fails", async () => {
    const user = userEvent.setup();

    mockedUpdateCreatorFollow.mockRejectedValue(new Error("boom"));

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Follow" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "フォロー状態を更新できませんでした。少し時間を置いてから再度お試しください。",
    );
  });

  it("redirects unauthenticated viewers to the login entry before opening paywall", async () => {
    const user = userEvent.setup();

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(push).toHaveBeenCalledWith("/login");
    expect(screen.queryByRole("dialog", { name: feedDialogTitle })).not.toBeInTheDocument();
  });

  it("renders detail mode with back navigation and the same creator block", async () => {
    if (!detailSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    renderWithViewerSession(
      <ImmersiveShortSurface backHref="/" creatorProfileOrigin={pinnedDetailOrigin} mode="detail" surface={detailSurface} />,
      { hasSession: true },
    );

    expect(screen.getByRole("link", { name: /Back/i })).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "href",
      "/creators/creator_mina_rei?from=short&shortFanTab=pinned&shortId=rooftop",
    );
    expect(screen.queryByRole("link", { name: /おすすめ/i })).not.toBeInTheDocument();
    expect(screen.getByText(detailSurface.short.caption)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Following" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(screen.getByRole("dialog", { name: detailDialogTitle })).toBeInTheDocument();
  });

  it("updates the detail follow CTA after an authenticated unfollow succeeds", async () => {
    if (!detailSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    mockedUpdateCreatorFollow.mockResolvedValue({
      stats: {
        fanCount: 11,
      },
      viewer: {
        isFollowing: false,
      },
    });

    renderWithViewerSession(
      <ImmersiveShortSurface backHref="/" creatorProfileOrigin={pinnedDetailOrigin} mode="detail" surface={detailSurface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Following" }));

    await waitFor(() => {
      expect(mockedUpdateCreatorFollow).toHaveBeenCalledWith({
        action: "unfollow",
        creatorId: detailSurface.creator.id,
      });
      expect(screen.getByRole("button", { name: "Follow" })).toHaveAttribute("aria-pressed", "false");
    });
  });

  it("updates the detail pin CTA after an authenticated unpin succeeds", async () => {
    if (!detailSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(
          JSON.stringify({
            data: {
              viewer: {
                isPinned: false,
              },
            },
            error: null,
            meta: {
              page: null,
              requestId: "req_short_pin_delete_success_001",
            },
          }),
          {
            headers: {
              "Content-Type": "application/json",
            },
            status: 200,
          },
        ),
      ),
    );

    renderWithViewerSession(
      <ImmersiveShortSurface backHref="/" creatorProfileOrigin={pinnedDetailOrigin} mode="detail" surface={detailSurface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Pinned short" }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Pin short" })).toHaveAttribute("aria-pressed", "false");
    });
  });

  it("redirects unauthenticated viewers to login when detail pin is tapped", async () => {
    if (!detailSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    renderWithViewerSession(
      <ImmersiveShortSurface backHref="/" creatorProfileOrigin={pinnedDetailOrigin} mode="detail" surface={detailSurface} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: "Pinned short" }));

    expect(push).toHaveBeenCalledWith("/login");
  });

  it("renders continue-main detail content as an action button", async () => {
    if (!continueMainSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: "auth_required",
              message: "login required",
            },
          }),
          {
            headers: {
              "Content-Type": "application/json",
            },
            status: 401,
          },
        ),
      ),
    );

    renderWithViewerSession(
      <ImmersiveShortSurface
        backHref="/"
        creatorProfileOrigin={{ from: "short", shortId: "softlight" }}
        mode="detail"
        surface={continueMainSurface}
      />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Continue main/i }));

    expect(push).toHaveBeenCalledWith("/login");
  });

  it("renders direct-unlock detail content as an action button", () => {
    if (!directUnlockSurface) {
      throw new Error("fixture missing");
    }

    renderWithViewerSession(
      <ImmersiveShortSurface
        backHref="/"
        creatorProfileOrigin={{ from: "short", shortId: "afterrain" }}
        mode="detail"
        surface={directUnlockSurface}
      />,
      { hasSession: true },
    );

    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
  });

  it("renders owner-preview detail content as an action button", () => {
    if (!ownerPreviewSurface) {
      throw new Error("fixture missing");
    }

    renderWithViewerSession(
      <ImmersiveShortSurface
        backHref="/"
        creatorProfileOrigin={{ from: "short", shortId: "balcony" }}
        mode="detail"
        surface={ownerPreviewSurface}
      />,
      { hasSession: true },
    );

    expect(screen.getByRole("button", { name: /Owner preview/i })).toBeInTheDocument();
  });

  it("renders creator initials when the creator has no custom avatar", () => {
    renderWithViewerSession(
      <ImmersiveShortSurface
        activeTab="recommended"
        mode="feed"
        surface={{ ...feedSurface, creator: { ...feedSurface.creator, avatar: null } }}
      />,
      { hasSession: true },
    );

    expect(screen.getByText("MR")).toBeInTheDocument();
  });

  it("falls back to a generic paywall title when the short has no caption", async () => {
    const user = userEvent.setup();
    const surface = {
      ...feedSurface,
      short: {
        ...feedSurface.short,
        caption: "",
        title: "",
      },
      unlock: {
        ...feedSurface.unlock,
        main: {
          ...feedSurface.unlock.main,
          title: "",
        },
        short: {
          ...feedSurface.unlock.short,
          caption: "",
          title: "",
        },
      },
    };

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(screen.getByRole("dialog", { name: "この short の続きを見る" })).toBeInTheDocument();
  });

  it("renders feed mode without short theme lookup for unknown short ids", () => {
    const surface = {
      ...feedSurface,
      short: {
        ...feedSurface.short,
        id: "short_dbcc1756d3d9406988e6860c7348609c",
      },
      unlock: {
        ...feedSurface.unlock,
        short: {
          ...feedSurface.unlock.short,
          id: "short_dbcc1756d3d9406988e6860c7348609c",
        },
      },
    };

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: true },
    );

    expect(screen.getByRole("link", { name: /おすすめ/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
  });
});
