import userEvent from "@testing-library/user-event";
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
    expect(screen.getByText(/1,800 \| 8分/)).toBeInTheDocument();
  });

  it("renders an action button when click behavior is provided", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();

    render(
      <UnlockCta
        cta={{
          mainDurationSeconds: null,
          priceJpy: null,
          resumePositionSeconds: 494,
          state: "continue_main",
        }}
        onClick={onClick}
      />,
    );

    await user.click(screen.getByRole("button", { name: /Continue main/i }));

    expect(onClick).toHaveBeenCalledTimes(1);
    expect(screen.getByText("8:14")).toBeInTheDocument();
  });
});
