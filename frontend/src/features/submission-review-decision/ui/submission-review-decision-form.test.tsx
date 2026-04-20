import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import type { SubmissionReviewCase } from "@/entities/submission-review";

import { applySubmissionReviewDecision } from "@/entities/submission-review";
import { SubmissionReviewDecisionForm } from "./submission-review-decision-form";

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

vi.mock("@/entities/submission-review", async () => {
  const actual = await vi.importActual<typeof import("@/entities/submission-review")>("@/entities/submission-review");

  return {
    ...actual,
    applySubmissionReviewDecision: vi.fn(),
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
      status: "pending_review",
      submitKind: "initial_submit",
      submittedAt: "2026-04-20T09:00:00Z",
    },
    main: {
      currencyCode: "JPY",
      decisionRequired: true,
      id: "22222222-2222-2222-2222-222222222222",
      intakeDecisionLog: null,
      media: {
        durationSeconds: 540,
        id: "asset_main_1",
        kind: "video",
        posterUrl: "https://cdn.example.com/mains/poster.jpg",
        url: "https://cdn.example.com/mains/video.mp4",
      },
      priceJpy: 1800,
      review: {
        decisionSource: null,
        decisionedAt: null,
        reasonCode: null,
        reviewNote: null,
      },
      state: "pending_review",
    },
    shorts: [
      {
        caption: "quiet rooftop cut",
        decisionRequired: true,
        id: "55555555-5555-5555-5555-555555555555",
        intakeDecisionLog: null,
        media: {
          durationSeconds: 18,
          id: "asset_short_1",
          kind: "video",
          posterUrl: "https://cdn.example.com/shorts/poster.jpg",
          url: "https://cdn.example.com/shorts/video.mp4",
        },
        review: {
          decisionSource: null,
          decisionedAt: null,
          reasonCode: null,
          reviewNote: null,
        },
        state: "pending_review",
      },
    ],
  };
}

describe("SubmissionReviewDecisionForm", () => {
  beforeEach(() => {
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    vi.mocked(applySubmissionReviewDecision).mockReset();
  });

  it("requires an explicit decision for every pending target", async () => {
    const user = userEvent.setup();

    render(<SubmissionReviewDecisionForm reviewCase={createReviewCase()} />);

    await user.click(screen.getByRole("button", { name: "decision を反映する" }));

    expect(screen.getByRole("alert")).toHaveTextContent("本編 の decision を選択してください。");
  });

  it("submits main and short decisions with reason and notes", async () => {
    const user = userEvent.setup();

    vi.mocked(applySubmissionReviewDecision).mockResolvedValue(undefined);

    render(<SubmissionReviewDecisionForm reviewCase={createReviewCase()} />);

    await user.click(screen.getAllByRole("button", { name: "承認する" })[0]!);
    await user.type(screen.getByLabelText("main-review-note"), "unlock ready");
    await user.click(screen.getAllByRole("button", { name: "修正依頼にする" })[1]!);
    await user.selectOptions(screen.getByLabelText("short 1-reason"), "quality_issue");
    await user.type(screen.getByLabelText("short 1-review-note"), "reframe intro");
    await user.click(screen.getByRole("button", { name: "decision を反映する" }));

    await waitFor(() => {
      expect(applySubmissionReviewDecision).toHaveBeenCalledWith({
        intakeId: "11111111-1111-1111-1111-111111111111",
        mainDecision: {
          decision: "approved",
          reasonCode: "",
          reviewNote: "unlock ready",
        },
        shortDecisions: [
          {
            decision: "revision_requested",
            reasonCode: "quality_issue",
            reviewNote: "reframe intro",
            shortId: "55555555-5555-5555-5555-555555555555",
          },
        ],
      });
      expect(mockedRouter.replace).toHaveBeenCalledWith("/admin/submission-reviews");
    });
  });

  it("requires an explicit reason selection for rejected or revision requested decisions", async () => {
    const user = userEvent.setup();

    render(<SubmissionReviewDecisionForm reviewCase={createReviewCase()} />);

    await user.click(screen.getAllByRole("button", { name: "却下する" })[0]!);
    await user.click(screen.getAllByRole("button", { name: "承認する" })[1]!);
    await user.click(screen.getByRole("button", { name: "decision を反映する" }));

    expect(screen.getByRole("alert")).toHaveTextContent("本編 の reason code を選択してください。");
    expect(applySubmissionReviewDecision).not.toHaveBeenCalled();
  });

  it("does not render an actionable form for a non-pending intake", () => {
    const reviewCase = createReviewCase();
    reviewCase.intake.status = "decision_applied";
    reviewCase.main.decisionRequired = true;
    reviewCase.shorts[0]!.decisionRequired = true;

    render(<SubmissionReviewDecisionForm reviewCase={reviewCase} />);

    expect(screen.getByText("この intake では追加の admin decision はありません。")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "decision を反映する" })).not.toBeInTheDocument();
  });
});
