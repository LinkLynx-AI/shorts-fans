import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";

import {
  CreatorFollowApiError,
  updateCreatorFollow,
} from "@/entities/creator";
import { listFollowingItems } from "@/entities/fan-profile";
import { useFanAuthDialog } from "@/features/fan-auth";
import { ApiError } from "@/shared/api";
import { FollowingShell } from "@/widgets/following-shell";

vi.mock("@/entities/creator", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/entities/creator")>();

  return {
    ...actual,
    updateCreatorFollow: vi.fn(),
  };
});

vi.mock("@/features/fan-auth", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/features/fan-auth")>();

  return {
    ...actual,
    useFanAuthDialog: vi.fn(),
  };
});

const mockedUpdateCreatorFollow = vi.mocked(updateCreatorFollow);
const mockedUseFanAuthDialog = vi.mocked(useFanAuthDialog);
const openFanAuthDialog = vi.fn();

function getFirstFollowingButton(): HTMLElement {
  const [firstFollowingButton] = screen.getAllByRole("button", { name: "フォロー中" });

  if (!firstFollowingButton) {
    throw new Error("フォロー中 CTA が見つかりません。");
  }

  return firstFollowingButton;
}

describe("FollowingShell", () => {
  beforeEach(() => {
    mockedUpdateCreatorFollow.mockReset();
    mockedUseFanAuthDialog.mockReset();
    openFanAuthDialog.mockReset();
    mockedUseFanAuthDialog.mockReturnValue({
      closeFanAuthDialog: vi.fn(),
      isFanAuthDialogOpen: false,
      openFanAuthDialog,
    });
  });

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

  it("keeps the creator row visible after unfollow so it can be followed again", async () => {
    const user = userEvent.setup();
    const firstCreator = listFollowingItems()[0];
    const firstCreatorId = firstCreator?.creator.id;
    const firstCreatorDisplayName = firstCreator?.creator.displayName;
    const updateFollowingCreatorRelation = vi.fn().mockResolvedValue(undefined);

    if (!firstCreatorId || !firstCreatorDisplayName) {
      throw new Error("最初の following creator が見つかりません。");
    }

    render(
      <FollowingShell
        items={listFollowingItems()}
        updateFollowingCreatorRelation={updateFollowingCreatorRelation}
      />,
    );

    expect(screen.getAllByRole("button", { name: "フォロー中" })).toHaveLength(3);

    const firstFollowingButton = getFirstFollowingButton();
    await user.click(firstFollowingButton);

    await waitFor(() => {
      expect(updateFollowingCreatorRelation).toHaveBeenCalledWith({
        action: "unfollow",
        creatorId: firstCreatorId,
      });
      expect(screen.getByRole("button", { name: "フォロー" })).toBeInTheDocument();
      expect(screen.getByText("3 creators")).toBeInTheDocument();
      expect(screen.getByText(firstCreatorDisplayName)).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: "フォロー" }));

    await waitFor(() => {
      expect(updateFollowingCreatorRelation).toHaveBeenLastCalledWith({
        action: "follow",
        creatorId: firstCreatorId,
      });
      expect(screen.getAllByRole("button", { name: "フォロー中" })).toHaveLength(3);
    });
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

    const firstFollowingButton = getFirstFollowingButton();
    await user.click(firstFollowingButton);

    expect(updateFollowingCreatorRelation).toHaveBeenCalledWith({
      action: "unfollow",
      creatorId: firstCreatorId,
    });
    expect(screen.getByRole("button", { name: "フォロー解除中..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "フォロー解除中..." })).toHaveAttribute("aria-busy", "true");

    resolveUpdate?.();

    await waitFor(() => {
      expect(screen.getAllByRole("button", { name: "フォロー中" })).toHaveLength(2);
      expect(screen.getByRole("button", { name: "フォロー" })).toBeEnabled();
      expect(screen.getByText("3 creators")).toBeInTheDocument();
    });
  });

  it("opens the shared auth dialog when unfollow returns auth_required", async () => {
    const user = userEvent.setup();
    const firstCreatorId = listFollowingItems()[0]?.creator.id;

    if (!firstCreatorId) {
      throw new Error("最初の following creator が見つかりません。");
    }

    mockedUpdateCreatorFollow.mockRejectedValue(
      new CreatorFollowApiError("auth_required", "creator follow requires authentication", {
        requestId: "req_creator_follow_delete_auth_required_001",
        status: 401,
      }),
    );

    render(<FollowingShell items={listFollowingItems()} />);

    await user.click(getFirstFollowingButton());

    await waitFor(() => {
      expect(mockedUpdateCreatorFollow).toHaveBeenCalledWith({
        action: "unfollow",
        creatorId: firstCreatorId,
      });
      expect(openFanAuthDialog).toHaveBeenCalledTimes(1);
    });

    expect(screen.getByText("3 creators")).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("shows a recoverable row error when unfollow fails", async () => {
    const user = userEvent.setup();

    mockedUpdateCreatorFollow.mockRejectedValue(
      new ApiError("network failed", {
        code: "network",
      }),
    );

    render(<FollowingShell items={listFollowingItems()} />);

    await user.click(getFirstFollowingButton());

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "フォロー状態を更新できませんでした。通信状態を確認してから再度お試しください。",
    );
    expect(openFanAuthDialog).not.toHaveBeenCalled();
    expect(screen.getByText("3 creators")).toBeInTheDocument();
    expect(getFirstFollowingButton()).toBeEnabled();
  });

  it("shows a generic retry message when unfollow fails with a non-network api error", async () => {
    const user = userEvent.setup();

    mockedUpdateCreatorFollow.mockRejectedValue(
      new ApiError("server failed", {
        code: "http",
        status: 500,
      }),
    );

    render(<FollowingShell items={listFollowingItems()} />);

    await user.click(getFirstFollowingButton());

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "フォロー状態を更新できませんでした。少し時間を置いてから再度お試しください。",
    );
    expect(openFanAuthDialog).not.toHaveBeenCalled();
  });
});
