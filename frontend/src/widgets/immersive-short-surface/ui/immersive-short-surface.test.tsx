import { render, screen } from "@testing-library/react";

import { getFeedSurfaceByTab, getShortSurfaceById } from "@/widgets/immersive-short-surface";

import { ImmersiveShortSurface } from "./immersive-short-surface";

describe("ImmersiveShortSurface", () => {
  const feedSurface = getFeedSurfaceByTab("recommended");
  const detailSurface = getShortSurfaceById("rooftop");

  it("renders feed mode with tab navigation and a detail CTA link", () => {
    render(<ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />);

    expect(screen.getByRole("link", { name: /おすすめ/i })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: /Unlock/i })).toHaveAttribute("href", "/shorts/rooftop");
    expect(screen.queryByRole("link", { name: /Back/i })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pinned short" })).toBeInTheDocument();
    expect(screen.getByText("Following")).toBeInTheDocument();
  });

  it("renders detail mode with back navigation and the same creator block", () => {
    if (!detailSurface) {
      throw new Error("fixture missing");
    }

    render(<ImmersiveShortSurface backHref="/" mode="detail" surface={detailSurface} />);

    expect(screen.getByRole("link", { name: /Back/i })).toHaveAttribute("href", "/");
    expect(screen.queryByRole("link", { name: /おすすめ/i })).not.toBeInTheDocument();
    expect(screen.getByText(detailSurface.short.caption)).toBeInTheDocument();
    expect(screen.getByText("Unlock")).toBeInTheDocument();
  });
});
