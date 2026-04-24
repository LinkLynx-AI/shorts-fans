import {
  render,
  screen,
} from "@testing-library/react";

import type { SubmissionReviewCase } from "@/entities/submission-review";

import { getSubmissionReviewCase } from "@/entities/submission-review";

import { assertAdminUiAccess } from "../../_lib/admin-ui-access";
import AdminSubmissionReviewCasePage from "./page";

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
  assertAdminUiAccess: vi.fn(),
}));

vi.mock("@/features/submission-review-decision", () => ({
  SubmissionReviewDecisionForm: ({ reviewCase }: { reviewCase: SubmissionReviewCase }) => (
    <div data-testid="submission-review-decision-form">{reviewCase.intake.id}</div>
  ),
}));

vi.mock("@/entities/submission-review", async () => {
  const actual = await vi.importActual<typeof import("@/entities/submission-review")>("@/entities/submission-review");

  return {
    ...actual,
    getSubmissionReviewCase: vi.fn(),
  };
});

function createReviewCase(): SubmissionReviewCase {
  return {
    creator: {
      avatar: null,
      bio: "quiet rooftop",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_11111111111111111111111111111111",
    },
    intake: {
      canonicalMainId: "22222222-2222-2222-2222-222222222222",
      consentConfirmed: true,
      creatorUserId: "33333333-3333-3333-3333-333333333333",
      id: "11111111-1111-1111-1111-111111111111",
      mainMediaAssetId: "44444444-4444-4444-4444-444444444444",
      mainPriceJpy: 1800,
      ownershipConfirmed: true,
      previousIntakeId: null,
      status: "decision_applied",
      submitKind: "initial_submit",
      submittedAt: "2026-04-20T09:00:00Z",
    },
    main: {
      currencyCode: "JPY",
      decisionRequired: false,
      id: "22222222-2222-2222-2222-222222222222",
      intakeDecisionLog: {
        decisionSource: "manual",
        decisionedAt: "2026-04-20T10:00:00Z",
        reasonCode: "quality_issue",
        reviewNote: "main intake note",
        targetState: "revision_requested",
      },
      media: {
        durationSeconds: 540,
        id: "asset_main_1",
        kind: "video",
        posterUrl: "https://cdn.example.com/mains/poster.jpg",
        url: "https://cdn.example.com/mains/video.mp4",
      },
      priceJpy: 1800,
      review: {
        decisionSource: "manual_override",
        decisionedAt: "2026-04-20T11:00:00Z",
        reasonCode: "quality_issue",
        reviewNote: "main current note",
      },
      state: "revision_requested",
    },
    shorts: [
      {
        caption: "quiet rooftop cut",
        decisionRequired: false,
        id: "55555555-5555-5555-5555-555555555555",
        intakeDecisionLog: {
          decisionSource: "manual_override",
          decisionedAt: "2026-04-20T10:15:00Z",
          reasonCode: "metadata_incomplete",
          reviewNote: "short intake note",
          targetState: "rejected",
        },
        media: {
          durationSeconds: 18,
          id: "asset_short_1",
          kind: "video",
          posterUrl: "https://cdn.example.com/shorts/poster.jpg",
          url: "https://cdn.example.com/shorts/video.mp4",
        },
        review: {
          decisionSource: "auto",
          decisionedAt: "2026-04-20T11:15:00Z",
          reasonCode: "metadata_incomplete",
          reviewNote: "short current note",
        },
        state: "rejected",
      },
    ],
  };
}

function getFixtureShort(reviewCase: SubmissionReviewCase) {
  const [short] = reviewCase.shorts;
  if (!short) {
    throw new Error("submission review fixture must include a short");
  }

  return short;
}

describe("AdminSubmissionReviewCasePage", () => {
  beforeEach(() => {
    notFound.mockReset();
    vi.mocked(assertAdminUiAccess).mockReset();
    vi.mocked(getSubmissionReviewCase).mockReset();
  });

  it("renders current and intake review provenance for main and short detail", async () => {
    const reviewCase = createReviewCase();
    vi.mocked(getSubmissionReviewCase).mockResolvedValue(reviewCase);

    render(await AdminSubmissionReviewCasePage({
      params: Promise.resolve({ intakeId: reviewCase.intake.id }),
    }));

    expect(getSubmissionReviewCase).toHaveBeenCalledWith({
      fetcher: expect.any(Function),
      intakeId: reviewCase.intake.id,
    });
    expect(screen.getByRole("link", { name: /Video 審査 main \/ short/i })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getByRole("link", { name: /Creator 審査 登録申請/i })).toHaveAttribute(
      "href",
      "/admin/creator-reviews",
    );
    expect(screen.getByRole("link", { name: "一覧へ戻る" })).toHaveAttribute(
      "data-prefetch",
      "false",
    );
    expect(screen.queryByText("manual_override")).not.toBeInTheDocument();
    expect(screen.getAllByText("manual")).toHaveLength(3);
    expect(screen.getByText("auto")).toBeInTheDocument();
    expect(screen.getByText("2026/04/20 11:00 UTC")).toBeInTheDocument();
    expect(screen.getByText("2026/04/20 11:15 UTC")).toBeInTheDocument();
    expect(screen.getByText("2026/04/20 10:00 UTC")).toBeInTheDocument();
    expect(screen.getByText("2026/04/20 10:15 UTC")).toBeInTheDocument();
  });

  it("renders fallback copy when provenance is not recorded", async () => {
    const reviewCase = createReviewCase();
    const short = getFixtureShort(reviewCase);
    reviewCase.main.review = {
      decisionSource: null,
      decisionedAt: null,
      reasonCode: null,
      reviewNote: null,
    };
    reviewCase.main.intakeDecisionLog = null;
    short.review = {
      decisionSource: null,
      decisionedAt: null,
      reasonCode: null,
      reviewNote: null,
    };
    short.intakeDecisionLog = null;
    vi.mocked(getSubmissionReviewCase).mockResolvedValue(reviewCase);

    render(await AdminSubmissionReviewCasePage({
      params: Promise.resolve({ intakeId: reviewCase.intake.id }),
    }));

    expect(screen.getAllByText("未記録")).toHaveLength(8);
    expect(screen.getAllByText("この intake ではまだ decision log がありません。")).toHaveLength(2);
  });

  it("does not fetch the review case when admin access is rejected", async () => {
    const reviewCase = createReviewCase();
    vi.mocked(assertAdminUiAccess).mockRejectedValue(new Error("blocked"));

    await expect(AdminSubmissionReviewCasePage({
      params: Promise.resolve({ intakeId: reviewCase.intake.id }),
    })).rejects.toThrow("blocked");

    expect(getSubmissionReviewCase).not.toHaveBeenCalled();
  });
});
