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
    expect(screen.queryByRole("link", { name: /おすすめ/i })).not.toBeInTheDocument();
    expect(screen.getByText(detailSurface.short.caption)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
    expect(screen.getByText("Following")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(screen.getByRole("dialog", { name: "quiet rooftop preview の続きを見る" })).toBeInTheDocument();
  });

  it("links continue-main detail content directly to playback", () => {
    if (!continueMainSurface) {
      throw new Error("fixture missing");
    }

    render(<ImmersiveShortSurface backHref="/" mode="detail" surface={continueMainSurface} />);

    expect(screen.getByRole("link", { name: /Continue main/i })).toHaveAttribute(
      "href",
      "/mains/main_aoi_blue_balcony?fromShortId=softlight",
    );
  });

  it("links direct-unlock detail content straight to playback", () => {
    if (!directUnlockSurface) {
      throw new Error("fixture missing");
    }

    render(<ImmersiveShortSurface backHref="/" mode="detail" surface={directUnlockSurface} />);

    expect(screen.getByRole("link", { name: /Unlock/i })).toHaveAttribute(
      "href",
      "/mains/main_sora_after_rain?fromShortId=afterrain",
    );
  });

  it("links owner-preview detail content straight to playback", () => {
    if (!ownerPreviewSurface) {
      throw new Error("fixture missing");
    }

    render(<ImmersiveShortSurface backHref="/" mode="detail" surface={ownerPreviewSurface} />);

    expect(screen.getByRole("link", { name: /Owner preview/i })).toHaveAttribute(
      "href",
      "/mains/main_aoi_blue_balcony?fromShortId=balcony",
    );
  });
});
