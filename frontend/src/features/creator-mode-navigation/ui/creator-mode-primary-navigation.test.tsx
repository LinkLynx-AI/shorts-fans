import { render, screen } from "@testing-library/react";

import { CreatorModePrimaryNavigation } from "@/features/creator-mode-navigation";

describe("CreatorModePrimaryNavigation", () => {
  it("renders the current dashboard and upload routes as live creator navigation links", () => {
    render(<CreatorModePrimaryNavigation activeKey="dashboard" />);

    expect(screen.getByRole("link", { name: /Dashboard/i })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: "Upload" })).toHaveAttribute("href", "/creator/upload");
    expect(screen.queryByRole("link", { name: "Linkage" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Review" })).not.toBeInTheDocument();
    expect(screen.getByText("Linkage")).toBeInTheDocument();
    expect(screen.getByText("Review")).toBeInTheDocument();
    expect(screen.getAllByText("soon")).toHaveLength(2);
    expect(screen.getAllByText("live")).toHaveLength(2);
  });

  it("can render the shared creator navigation in compact mode", () => {
    render(<CreatorModePrimaryNavigation activeKey="dashboard" variant="compact" />);

    expect(screen.getByRole("link", { name: "Dashboard" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: "Upload" })).toHaveAttribute("href", "/creator/upload");
    expect(screen.getByText("Linkage")).toHaveAttribute("aria-disabled", "true");
    expect(screen.getByText("Review")).toHaveAttribute("aria-disabled", "true");
  });
});
