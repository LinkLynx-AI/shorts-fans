import { render, screen } from "@testing-library/react";

import { SubscriptionCtaCard } from "@/features/subscription-cta";

describe("SubscriptionCtaCard", () => {
  it("renders creator subscription messaging", () => {
    render(
      <SubscriptionCtaCard
        creator={{
          displayName: "Atelier Rin",
          handle: "@atelier-rin",
          lockedPosts: 18,
          monthlyPriceLabel: "¥1,480 / month",
          publicShorts: 26,
          slug: "atelier-rin",
          teaser: "teaser",
        }}
      />,
    );

    expect(screen.getByText("Atelier Rin")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "購読導線の箱を確認" })).toHaveAttribute(
      "href",
      "/subscriptions",
    );
  });
});
