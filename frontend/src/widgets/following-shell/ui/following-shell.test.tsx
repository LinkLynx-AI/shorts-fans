import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";

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

  it("switches the row CTA from フォロー中 to フォロー", async () => {
    const user = userEvent.setup();

    render(<FollowingShell items={listFollowingItems()} />);

    expect(screen.getAllByRole("button", { name: "フォロー中" })).toHaveLength(3);

    const [firstFollowingButton] = screen.getAllByRole("button", { name: "フォロー中" });

    if (!firstFollowingButton) {
      throw new Error("フォロー中 CTA が見つかりません。");
    }

    await user.click(firstFollowingButton);

    expect(screen.getAllByRole("button", { name: "フォロー中" })).toHaveLength(2);
    expect(screen.getByRole("button", { name: "フォロー" })).toHaveAttribute("aria-pressed", "false");
  });

  it("keeps a row CTA disabled while the update is pending", async () => {
    const user = userEvent.setup();
    const firstCreatorId = listFollowingItems()[0]?.creator.id;
    let resolveUpdate: (() => void) | undefined;
    const updateFollowingCreatorRelation = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          resolveUpdate = resolve;
        }),
    );

    render(
      <FollowingShell
        items={listFollowingItems()}
        updateFollowingCreatorRelation={updateFollowingCreatorRelation}
      />,
    );

    const [firstFollowingButton] = screen.getAllByRole("button", { name: "フォロー中" });

    if (!firstFollowingButton) {
      throw new Error("フォロー中 CTA が見つかりません。");
    }

    await user.click(firstFollowingButton);

    expect(updateFollowingCreatorRelation).toHaveBeenCalledWith({
      action: "unfollow",
      creatorId: firstCreatorId,
    });
    expect(screen.getByRole("button", { name: "フォロー解除中..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "フォロー解除中..." })).toHaveAttribute("aria-busy", "true");

    resolveUpdate?.();

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "フォロー" })).toBeEnabled();
    });
  });
});
