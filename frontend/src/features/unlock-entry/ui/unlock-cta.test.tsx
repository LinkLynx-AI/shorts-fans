import { render, screen } from "@testing-library/react";

import { getShortById } from "@/entities/short";
import { UnlockCta } from "@/features/unlock-entry";

describe("UnlockCta", () => {
  const short = getShortById("rooftop");

  it("renders a locked CTA as a link", () => {
    if (!short) {
      throw new Error("fixture missing");
    }

    render(<UnlockCta href="/shorts/rooftop" short={short} />);

    expect(screen.getByRole("link", { name: /Unlock/i })).toHaveAttribute("href", "/shorts/rooftop");
    expect(screen.getByText("¥1,800 | 8m")).toBeInTheDocument();
  });

  it("renders an unlocked CTA without navigation", () => {
    if (!short) {
      throw new Error("fixture missing");
    }

    render(<UnlockCta short={short} state="unlocked" />);

    expect(screen.getByText("Continue main")).toBeInTheDocument();
    expect(screen.getByText("8:14 left")).toBeInTheDocument();
  });
});
