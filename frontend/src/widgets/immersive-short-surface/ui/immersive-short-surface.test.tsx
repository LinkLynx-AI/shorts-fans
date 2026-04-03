import { render, screen } from "@testing-library/react";

import { getCreatorById } from "@/entities/creator";
import { getShortById } from "@/entities/short";

import { ImmersiveShortSurface } from "./immersive-short-surface";

describe("ImmersiveShortSurface", () => {
  const short = getShortById("rooftop");
  const creator = getCreatorById("mina");

  it("renders feed mode with tab navigation and a detail CTA link", () => {
    if (!short || !creator) {
      throw new Error("fixture missing");
    }

    render(<ImmersiveShortSurface activeTab="recommended" creator={creator} mode="feed" short={short} />);

    expect(screen.getByRole("link", { name: /おすすめ/i })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: /Unlock/i })).toHaveAttribute("href", "/shorts/rooftop");
    expect(screen.queryByRole("link", { name: /Back/i })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pinned short" })).toBeInTheDocument();
    expect(screen.getByText("Following")).toBeInTheDocument();
  });

  it("renders detail mode with back navigation and the same creator block", () => {
    if (!short || !creator) {
      throw new Error("fixture missing");
    }

    render(<ImmersiveShortSurface backHref="/" creator={creator} mode="detail" short={short} />);

    expect(screen.getByRole("link", { name: /Back/i })).toHaveAttribute("href", "/");
    expect(screen.queryByRole("link", { name: /おすすめ/i })).not.toBeInTheDocument();
    expect(screen.getByText(short.caption)).toBeInTheDocument();
    expect(screen.getByText("Unlock")).toBeInTheDocument();
  });
});
