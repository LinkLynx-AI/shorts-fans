import {
  render,
  screen,
} from "@testing-library/react";

import { getCreatorReviewQueue } from "@/entities/creator-review";

import { assertAdminUiAccess } from "../_lib/admin-ui-access";
import AdminCreatorReviewsPage from "./page";

vi.mock("../_lib/admin-ui-access", () => ({
  assertAdminUiAccess: vi.fn(),
}));

vi.mock("@/entities/creator-review", async () => {
  const actual = await vi.importActual<typeof import("@/entities/creator-review")>("@/entities/creator-review");

  return {
    ...actual,
    getCreatorReviewQueue: vi.fn(),
  };
});

describe("AdminCreatorReviewsPage", () => {
  beforeEach(() => {
    vi.mocked(assertAdminUiAccess).mockReset();
    vi.mocked(getCreatorReviewQueue).mockReset();
  });

  it("renders review navigation to video review from the creator queue", async () => {
    const userId = "11111111-1111-1111-1111-111111111111";
    vi.mocked(getCreatorReviewQueue).mockResolvedValue({
      items: [
        {
          creatorBio: "quiet rooftop",
          legalName: "Mina Rei",
          review: {
            approvedAt: null,
            rejectedAt: null,
            submittedAt: "2026-04-20T09:00:00Z",
            suspendedAt: null,
          },
          sharedProfile: {
            avatar: null,
            displayName: "Mina Rei",
            handle: "@minarei",
          },
          state: "submitted",
          userId,
        },
      ],
      state: "submitted",
    });

    render(await AdminCreatorReviewsPage({
      searchParams: Promise.resolve({}),
    }));

    expect(getCreatorReviewQueue).toHaveBeenCalledWith({
      fetcher: expect.any(Function),
      state: "submitted",
    });
    expect(screen.getByRole("link", { name: /Creator 審査 登録申請/i })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getByRole("link", { name: /Video 審査 main \/ short/i })).toHaveAttribute(
      "href",
      "/admin/submission-reviews",
    );
    expect(screen.getByRole("link", { name: "承認済み" })).toHaveAttribute(
      "data-prefetch",
      "false",
    );
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "data-prefetch",
      "false",
    );
  });

  it("does not fetch the review queue when admin access is rejected", async () => {
    vi.mocked(assertAdminUiAccess).mockRejectedValue(new Error("blocked"));

    await expect(AdminCreatorReviewsPage({
      searchParams: Promise.resolve({}),
    })).rejects.toThrow("blocked");

    expect(getCreatorReviewQueue).not.toHaveBeenCalled();
  });
});
