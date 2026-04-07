import userEvent from "@testing-library/user-event";
import { render, screen } from "@testing-library/react";

import { ViewerSessionProvider } from "@/entities/viewer";
import { getFeedSurfaceByTab, getShortSurfaceById } from "@/widgets/immersive-short-surface";

import { ImmersiveShortSurface } from "./immersive-short-surface";

const push = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push,
  }),
}));

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

    expect(screen.getByRole("dialog", { name: "quiet rooftop preview の続きを見る" })).toBeInTheDocument();
  });

  it("redirects unauthenticated viewers to the login entry before opening paywall", async () => {
    const user = userEvent.setup();

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(push).toHaveBeenCalledWith("/login");
    expect(screen.queryByRole("dialog", { name: "quiet rooftop preview の続きを見る" })).not.toBeInTheDocument();
  });

  it("renders detail mode with back navigation and the same creator block", async () => {
    if (!detailSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    renderWithViewerSession(
      <ImmersiveShortSurface backHref="/" mode="detail" surface={detailSurface} />,
      { hasSession: true },
    );

    expect(screen.getByRole("link", { name: /Back/i })).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "href",
      "/creators/creator_mina_rei?from=short&shortId=rooftop",
    );
    expect(screen.queryByRole("link", { name: /おすすめ/i })).not.toBeInTheDocument();
    expect(screen.getByText(detailSurface.short.caption)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
    expect(screen.getByText("Following")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(screen.getByRole("dialog", { name: "quiet rooftop preview の続きを見る" })).toBeInTheDocument();
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
      <ImmersiveShortSurface backHref="/" mode="detail" surface={continueMainSurface} />,
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
      <ImmersiveShortSurface backHref="/" mode="detail" surface={directUnlockSurface} />,
      { hasSession: true },
    );

    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
  });

  it("renders owner-preview detail content as an action button", () => {
    if (!ownerPreviewSurface) {
      throw new Error("fixture missing");
    }

    renderWithViewerSession(
      <ImmersiveShortSurface backHref="/" mode="detail" surface={ownerPreviewSurface} />,
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
});
