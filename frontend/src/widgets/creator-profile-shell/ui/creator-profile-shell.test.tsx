import { render, screen } from "@testing-library/react";

import { CreatorProfileShell, getCreatorProfileShellState } from "@/widgets/creator-profile-shell";

describe("CreatorProfileShell", () => {
  it("renders the normal creator profile without unlock affordances", () => {
    const state = getCreatorProfileShellState("mina");

    if (!state) {
      throw new Error("fixture missing");
    }

    render(<CreatorProfileShell routeState={{ from: "search", q: "mina" }} state={state} />);

    expect(screen.getByText("minarei")).toBeInTheDocument();
    expect(screen.queryByText("@minarei")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Following" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /quiet rooftop preview/i })).toHaveAttribute(
      "href",
      "/shorts/rooftop?creatorId=mina&from=creator&profileFrom=search&profileQ=mina",
    );
    expect(screen.queryByText("Unlock")).not.toBeInTheDocument();
  });

  it("renders the empty creator profile state", () => {
    const state = getCreatorProfileShellState("sora");

    if (!state) {
      throw new Error("fixture missing");
    }

    render(<CreatorProfileShell routeState={{ from: "feed", tab: "recommended" }} state={state} />);

    expect(screen.getByText("まだ公開中の short はありません。")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /preview/i })).not.toBeInTheDocument();
  });
});
