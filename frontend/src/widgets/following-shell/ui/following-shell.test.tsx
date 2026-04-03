import { fireEvent, render, screen } from "@testing-library/react";

import { listFollowingItems } from "@/entities/fan-profile";
import { FollowingShell } from "@/widgets/following-shell";

describe("FollowingShell", () => {
  it("filters creators by query", () => {
    render(<FollowingShell items={listFollowingItems()} />);

    expect(screen.getByText("3 creators")).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText("検索"), {
      target: { value: "mina" },
    });

    expect(screen.getByText("1 creators")).toBeInTheDocument();
    expect(screen.getByText("Mina Rei")).toBeInTheDocument();
    expect(screen.queryByText("Aoi N")).not.toBeInTheDocument();
  });
});
