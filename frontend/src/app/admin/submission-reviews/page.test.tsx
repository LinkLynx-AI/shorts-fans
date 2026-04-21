import {
  render,
  screen,
} from "@testing-library/react";

import { getSubmissionReviewQueue } from "@/entities/submission-review";

import AdminSubmissionReviewsPage from "./page";

vi.mock("../_lib/admin-ui-access", () => ({
  assertAdminUiEnabled: vi.fn(),
}));

vi.mock("@/entities/submission-review", async () => {
  const actual = await vi.importActual<typeof import("@/entities/submission-review")>("@/entities/submission-review");

  return {
    ...actual,
    getSubmissionReviewQueue: vi.fn(),
  };
});

describe("AdminSubmissionReviewsPage", () => {
  beforeEach(() => {
    vi.mocked(getSubmissionReviewQueue).mockReset();
  });

  it("renders review navigation to creator review from the video queue", async () => {
    vi.mocked(getSubmissionReviewQueue).mockResolvedValue({ items: [] });

    render(await AdminSubmissionReviewsPage());

    expect(screen.getByRole("link", { name: /Video 審査 main \/ short/i })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getByRole("link", { name: /Creator 審査 登録申請/i })).toHaveAttribute(
      "href",
      "/admin/creator-reviews",
    );
  });
});
