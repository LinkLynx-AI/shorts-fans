import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import { updateCreatorWorkspaceProfile } from "@/features/viewer-profile";

import { useCreatorWorkspaceSummary } from "../model/use-creator-workspace-summary";
import { CreatorProfileSettingsShell } from "./creator-profile-settings-shell";

const mockedRouter = vi.hoisted(() => ({
  back: vi.fn(),
  forward: vi.fn(),
  prefetch: vi.fn(),
  push: vi.fn(),
  refresh: vi.fn(),
  replace: vi.fn(),
}));

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");

  return {
    ...actual,
    useRouter: () => mockedRouter,
  };
});

vi.mock("@/features/viewer-profile", async () => {
  const actual = await vi.importActual<typeof import("@/features/viewer-profile")>("@/features/viewer-profile");

  return {
    ...actual,
    updateCreatorWorkspaceProfile: vi.fn(),
  };
});

vi.mock("../model/use-creator-workspace-summary", () => ({
  useCreatorWorkspaceSummary: vi.fn(),
}));

describe("CreatorProfileSettingsShell", () => {
  beforeEach(() => {
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    vi.mocked(updateCreatorWorkspaceProfile).mockReset();
    vi.mocked(updateCreatorWorkspaceProfile).mockResolvedValue(undefined);
    vi.mocked(useCreatorWorkspaceSummary).mockReturnValue({
      blockedState: null,
      retry: vi.fn(),
      state: {
        kind: "ready",
        summary: {
          creator: {
            avatar: {
              durationSeconds: null,
              id: "asset_creator_mina_avatar",
              kind: "image",
              posterUrl: null,
              url: "https://cdn.example.com/creator/mina/avatar.jpg",
            },
            bio: "quiet rooftop と hotel light の preview を軸に投稿。",
            displayName: "Mina Rei",
            handle: "@minarei",
            id: "creator_mina_rei",
          },
          overviewMetrics: {
            grossUnlockRevenueJpy: 120000,
            unlockCount: 238,
            uniquePurchaserCount: 164,
          },
          revisionRequestedSummary: {
            mainCount: 0,
            shortCount: 1,
            totalCount: 1,
          },
        },
      },
    });
  });

  it("renders the fan-aligned editor and submits bio with the creator profile payload", async () => {
    const user = userEvent.setup();

    render(<CreatorProfileSettingsShell />);

    expect(screen.getByRole("link", { name: "ワークスペースへ戻る" })).toHaveAttribute("href", "/creator");
    expect(screen.getByRole("textbox", { name: /display name/i })).toHaveValue("Mina Rei");
    expect(screen.getByRole("textbox", { name: /handle/i })).toHaveValue("@minarei");
    expect(screen.getByRole("textbox", { name: "Bio" })).toHaveValue(
      "quiet rooftop と hotel light の preview を軸に投稿。",
    );

    await user.clear(screen.getByRole("textbox", { name: /display name/i }));
    await user.type(screen.getByRole("textbox", { name: /display name/i }), "Sabe");
    await user.clear(screen.getByRole("textbox", { name: /handle/i }));
    await user.type(screen.getByRole("textbox", { name: /handle/i }), "@sabe_123");
    await user.clear(screen.getByRole("textbox", { name: "Bio" }));
    await user.type(screen.getByRole("textbox", { name: "Bio" }), "new creator bio");
    await user.click(screen.getByRole("button", { name: "保存する" }));

    await waitFor(() => {
      expect(updateCreatorWorkspaceProfile).toHaveBeenCalledWith({
        bio: "new creator bio",
        displayName: "Sabe",
        handle: "@sabe_123",
      });
      expect(mockedRouter.replace).toHaveBeenCalledWith("/creator");
    });
  });
});
