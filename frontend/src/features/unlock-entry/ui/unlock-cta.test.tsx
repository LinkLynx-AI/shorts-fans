import { render, screen } from "@testing-library/react";

import { UnlockCta } from "@/features/unlock-entry";

describe("UnlockCta", () => {
  it("renders an unlock-available CTA as a link", () => {
    render(
      <UnlockCta
        cta={{
          mainDurationSeconds: 480,
          priceJpy: 1800,
          resumePositionSeconds: null,
          state: "unlock_available",
        }}
        href="/shorts/rooftop"
      />,
    );

    expect(screen.getByRole("link", { name: /Unlock/i })).toHaveAttribute("href", "/shorts/rooftop");
    expect(screen.getByText(/1,800 \| 8m/)).toBeInTheDocument();
  });

  it("renders a continue-main CTA without navigation", () => {
    render(
      <UnlockCta
        cta={{
          mainDurationSeconds: null,
          priceJpy: null,
          resumePositionSeconds: 494,
          state: "continue_main",
        }}
      />,
    );

    expect(screen.getByText("Continue main")).toBeInTheDocument();
    expect(screen.getByText("8:14")).toBeInTheDocument();
  });
});
