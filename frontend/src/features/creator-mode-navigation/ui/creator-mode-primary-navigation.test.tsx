import { render, screen } from "@testing-library/react";

import { CreatorModePrimaryNavigation } from "@/features/creator-mode-navigation";

describe("CreatorModePrimaryNavigation", () => {
  it("renders the current dashboard route as the only live creator navigation link", () => {
    render(<CreatorModePrimaryNavigation activeKey="dashboard" />);

    expect(screen.getByRole("link", { name: /Dashboard/i })).toHaveAttribute("aria-current", "page");
    expect(screen.queryByRole("link", { name: "Upload" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Linkage" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Review" })).not.toBeInTheDocument();
    expect(screen.getByText("Upload")).toBeInTheDocument();
    expect(screen.getByText("Linkage")).toBeInTheDocument();
    expect(screen.getByText("Review")).toBeInTheDocument();
    expect(screen.getAllByText("soon")).toHaveLength(3);
  });
});
