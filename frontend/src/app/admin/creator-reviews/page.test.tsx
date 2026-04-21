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
    vi.mocked(getCreatorReviewQueue).mockResolvedValue({
      items: [],
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
  });
});
