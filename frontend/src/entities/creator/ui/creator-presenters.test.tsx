import { render, screen } from "@testing-library/react";

import { CreatorAvatar, CreatorIdentity, CreatorStatList, getCreatorById, getCreatorProfileStatsById } from "@/entities/creator";

describe("creator presenters", () => {
  const creator = getCreatorById("mina");
  const stats = getCreatorProfileStatsById("mina");

  it("renders avatar, identity, and stats", () => {
    if (!creator || !stats) {
      throw new Error("fixture missing");
    }

    render(
      <div>
        <CreatorAvatar creator={creator} />
        <CreatorIdentity creator={creator} />
        <CreatorStatList stats={stats} />
      </div>,
    );

    expect(screen.getByText("Mina Rei")).toBeInTheDocument();
    expect(screen.getByText("@minarei")).toBeInTheDocument();
    expect(screen.getByText("24K")).toBeInTheDocument();
  });
});
