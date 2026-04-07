import { render, screen } from "@testing-library/react";

import { RouteStructurePanel } from "./route-structure-panel";

describe("RouteStructurePanel", () => {
  it("renders the default eyebrow and route items", () => {
    render(
      <RouteStructurePanel
        description="route ごとの構成を見せる panel"
        items={[
          {
            description: "public feed route",
            key: "feed",
            label: "Feed",
          },
          {
            description: "protected fan route",
            key: "fan",
            label: "Fan hub",
          },
        ]}
        title="Fan routes"
      />,
    );

    expect(screen.getByText("Route structure")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Fan routes" })).toBeInTheDocument();
    expect(screen.getByText("route ごとの構成を見せる panel")).toBeInTheDocument();
    expect(screen.getByText("Feed")).toBeInTheDocument();
    expect(screen.getByText("public feed route")).toBeInTheDocument();
    expect(screen.getByText("Fan hub")).toBeInTheDocument();
    expect(screen.getByText("protected fan route")).toBeInTheDocument();
  });
});
