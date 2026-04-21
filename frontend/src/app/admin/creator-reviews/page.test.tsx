import {
  render,
  screen,
} from "@testing-library/react";

import { getCreatorReviewQueue } from "@/entities/creator-review";

import AdminCreatorReviewsPage from "./page";

vi.mock("../_lib/admin-ui-access", () => ({
  assertAdminUiEnabled: vi.fn(),
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
});
