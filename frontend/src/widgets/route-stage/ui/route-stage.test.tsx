import { render, screen } from "@testing-library/react";

import { RouteStage } from "@/widgets/route-stage";

describe("RouteStage", () => {
  it("renders route copy, highlights, and actions", () => {
    render(
      <RouteStage
        actions={[
          { href: "/", label: "shorts を開く" },
          { href: "/home", label: "home を確認", variant: "secondary" },
        ]}
        description="route description"
        eyebrow="eyebrow"
        highlights={["one", "two"]}
        title="route title"
      />,
    );

    expect(screen.getByText("route title")).toBeInTheDocument();
    expect(screen.getByText("one")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "shorts を開く" })).toHaveAttribute("href", "/");
  });
});
