import { render, screen } from "@testing-library/react";

import { CreatorAvatar, CreatorIdentity, CreatorStatList, getCreatorById } from "@/entities/creator";

describe("creator presenters", () => {
  const creator = getCreatorById("mina");

  it("renders avatar, identity, and stats", () => {
    if (!creator) {
      throw new Error("fixture missing");
    }

    render(
      <div>
        <CreatorAvatar creator={creator} />
        <CreatorIdentity creator={creator} />
        <CreatorStatList stats={creator.stats} />
      </div>,
    );

    expect(screen.getByText("Mina Rei")).toBeInTheDocument();
    expect(screen.getByText("@minarei")).toBeInTheDocument();
    expect(screen.getByText("24K")).toBeInTheDocument();
  });
});
