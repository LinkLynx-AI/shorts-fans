import {
  render,
  screen,
} from "@testing-library/react";

import type { CreatorReviewCase } from "@/entities/creator-review";

import { getCreatorReviewCase } from "@/entities/creator-review";

import AdminCreatorReviewCasePage from "./page";

const { notFound } = vi.hoisted(() => ({
  notFound: vi.fn(),
}));

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");

  return {
    ...actual,
    notFound,
  };
});

vi.mock("../../_lib/admin-ui-access", () => ({
  assertAdminUiEnabled: vi.fn(),
}));

vi.mock("@/features/creator-review-decision", () => ({
  CreatorReviewDecisionForm: ({ reviewCase }: { reviewCase: CreatorReviewCase }) => (
    <div data-testid="creator-review-decision-form">{reviewCase.userId}</div>
  ),
}));

vi.mock("@/entities/creator-review", async () => {
  const actual = await vi.importActual<typeof import("@/entities/creator-review")>("@/entities/creator-review");

  return {
    ...actual,
    getCreatorReviewCase: vi.fn(),
  };
});

function createReviewCase(): CreatorReviewCase {
  return {
    creatorBio: "quiet rooftop",
    evidences: [],
    intake: {
      acceptsConsentResponsibility: true,
      birthDate: "2000-01-01",
      declaresNoProhibitedCategory: true,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
    },
    rejection: null,
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
    userId: "11111111-1111-1111-1111-111111111111",
  };
}

describe("AdminCreatorReviewCasePage", () => {
  beforeEach(() => {
    notFound.mockReset();
    vi.mocked(getCreatorReviewCase).mockReset();
  });

  it("renders review navigation to video review from the creator detail", async () => {
    const reviewCase = createReviewCase();
    vi.mocked(getCreatorReviewCase).mockResolvedValue(reviewCase);

    render(await AdminCreatorReviewCasePage({
      params: Promise.resolve({ userId: reviewCase.userId }),
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
    expect(screen.getByRole("link", { name: "一覧へ戻る" })).toHaveAttribute(
      "data-prefetch",
      "false",
    );
  });
});
