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
    const intakeId = "11111111-1111-1111-1111-111111111111";
    vi.mocked(getSubmissionReviewQueue).mockResolvedValue({
      items: [
        {
          creator: {
            avatar: null,
            bio: "quiet rooftop",
            displayName: "Mina Rei",
            handle: "@minarei",
            id: "creator_11111111111111111111111111111111",
          },
          intakeId,
          mainDecisionRequired: true,
          pendingShortCount: 1,
          shortCount: 2,
          submitKind: "initial_submit",
          submittedAt: "2026-04-20T09:00:00Z",
        },
      ],
    });

    render(await AdminSubmissionReviewsPage());

    expect(screen.getByRole("link", { name: /Video 審査 main \/ short/i })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getByRole("link", { name: /Creator 審査 登録申請/i })).toHaveAttribute(
      "href",
      "/admin/creator-reviews",
    );
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "data-prefetch",
      "false",
    );
  });
});
