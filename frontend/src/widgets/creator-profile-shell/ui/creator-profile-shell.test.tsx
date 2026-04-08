import { fireEvent, render, screen } from "@testing-library/react";

import { CreatorProfileShell } from "@/widgets/creator-profile-shell";
import { useHasViewerSession } from "@/entities/viewer";
import { useFanAuthDialog } from "@/features/fan-auth";

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

const mockedUseHasViewerSession = vi.mocked(useHasViewerSession);
const mockedUseFanAuthDialog = vi.mocked(useFanAuthDialog);
const openFanAuthDialog = vi.fn();

describe("CreatorProfileShell", () => {
  beforeEach(() => {
    openFanAuthDialog.mockReset();
    mockedUseFanAuthDialog.mockReturnValue({
      closeFanAuthDialog: vi.fn(),
      isFanAuthDialogOpen: false,
      openFanAuthDialog,
    });
  });

  it("renders the normal creator profile from contract-backed state without toggling follow", () => {
    mockedUseHasViewerSession.mockReturnValue(true);

    render(
      <CreatorProfileShell
        routeState={{ from: "search", q: "mina" }}
        state={{
          creator: {
            avatar: null,
            bio: "quiet rooftop と hotel light の preview を軸に投稿。",
            displayName: "Mina Rei",
            handle: "@minarei",
            id: "creator_mina_rei",
          },
          kind: "ready",
          shorts: [
            {
              canonicalMainId: "main_mina_quiet_rooftop",
              creatorId: "creator_mina_rei",
              id: "short_mina_rooftop",
              media: {
                durationSeconds: 16,
                id: "asset_short_mina_rooftop",
                kind: "video",
                posterUrl: null,
                url: "https://cdn.example.com/shorts/mina-rooftop.mp4",
              },
              previewDurationSeconds: 16,
              routeShortId: "rooftop",
            },
          ],
          stats: {
            fanCount: 24000,
            shortCount: 2,
          },
          viewer: {
            isFollowing: true,
          },
        }}
      />,
    );

    expect(screen.getByText("minarei")).toBeInTheDocument();
    expect(screen.queryByText("@minarei")).not.toBeInTheDocument();
    const followButton = screen.getByRole("button", { name: "Following" });
    expect(followButton).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("link", { name: /Mina Rei preview 0:16/i })).toHaveAttribute(
      "href",
      "/shorts/rooftop?creatorId=creator_mina_rei&from=creator&profileFrom=search&profileQ=mina",
    );

    fireEvent.click(followButton);

    expect(openFanAuthDialog).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Following" })).toHaveAttribute("aria-pressed", "true");
  });

  it("opens the shared auth modal when an anonymous viewer presses follow", () => {
    mockedUseHasViewerSession.mockReturnValue(false);

    render(
      <CreatorProfileShell
        routeState={{ from: "feed", tab: "recommended" }}
        state={{
          creator: {
            avatar: null,
            bio: "after rain と balcony mood の short をまとめています。",
            displayName: "Sora Vale",
            handle: "@soravale",
            id: "creator_sora_vale",
          },
          kind: "empty",
          shorts: [],
          stats: {
            fanCount: 16000,
            shortCount: 0,
          },
          viewer: {
            isFollowing: false,
          },
        }}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Follow" }));

    expect(openFanAuthDialog).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("button", { name: "Follow" })).toHaveAttribute("aria-pressed", "false");
  });

  it("renders the empty creator profile state", () => {
    mockedUseHasViewerSession.mockReturnValue(true);

    render(
      <CreatorProfileShell
        routeState={{ from: "feed", tab: "recommended" }}
        state={{
          creator: {
            avatar: null,
            bio: "after rain と balcony mood の short をまとめています。",
            displayName: "Sora Vale",
            handle: "@soravale",
            id: "creator_sora_vale",
          },
          kind: "empty",
          shorts: [],
          stats: {
            fanCount: 16000,
            shortCount: 0,
          },
          viewer: {
            isFollowing: false,
          },
        }}
      />,
    );

    expect(screen.getByText("まだ公開中の short はありません。")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Follow" })).toBeDisabled();
    expect(screen.queryByRole("link", { name: /preview/i })).not.toBeInTheDocument();
  });
});
