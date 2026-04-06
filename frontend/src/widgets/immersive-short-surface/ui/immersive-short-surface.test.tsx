import userEvent from "@testing-library/user-event";
import { render, screen } from "@testing-library/react";

import { getFeedSurfaceByTab, getShortSurfaceById } from "@/widgets/immersive-short-surface";

import { ImmersiveShortSurface } from "./immersive-short-surface";

describe("ImmersiveShortSurface", () => {
  const feedSurface = getFeedSurfaceByTab("recommended");
  const detailSurface = getShortSurfaceById("rooftop");
  const continueMainSurface = getShortSurfaceById("softlight");
  const directUnlockSurface = getShortSurfaceById("afterrain");
  const ownerPreviewSurface = getShortSurfaceById("balcony");

  it("opens the mini paywall for setup-required feed content", async () => {
    const user = userEvent.setup();

    render(<ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />);

    expect(screen.getByRole("link", { name: /おすすめ/i })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "href",
      "/creators/mina?from=feed&tab=recommended",
    );
    expect(screen.queryByRole("link", { name: /Back/i })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pinned short" })).toBeInTheDocument();
    expect(screen.getByText("Follow")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(screen.getByRole("dialog", { name: "quiet rooftop preview の続きを見る" })).toBeInTheDocument();
  });

  it("renders detail mode with back navigation and the same creator block", async () => {
    if (!detailSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    render(<ImmersiveShortSurface backHref="/" mode="detail" surface={detailSurface} />);

    expect(screen.getByRole("link", { name: /Back/i })).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "href",
      "/creators/mina?from=short&shortId=rooftop",
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

    render(<ImmersiveShortSurface backHref="/" mode="detail" surface={continueMainSurface} />);

    await user.click(screen.getByRole("button", { name: /Continue main/i }));

    expect(screen.getByRole("button", { name: /Continue main/i })).toBeInTheDocument();
  });

  it("renders direct-unlock detail content as an action button", async () => {
    if (!directUnlockSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    render(<ImmersiveShortSurface backHref="/" mode="detail" surface={directUnlockSurface} />);

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
  });

  it("renders owner-preview detail content as an action button", async () => {
    if (!ownerPreviewSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    render(<ImmersiveShortSurface backHref="/" mode="detail" surface={ownerPreviewSurface} />);

    await user.click(screen.getByRole("button", { name: /Owner preview/i }));

    expect(screen.getByRole("button", { name: /Owner preview/i })).toBeInTheDocument();
  });
});
