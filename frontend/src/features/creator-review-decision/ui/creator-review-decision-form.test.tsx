import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import type { CreatorReviewCase } from "@/entities/creator-review";

import { CreatorReviewDecisionForm } from "./creator-review-decision-form";

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

function createReviewCase(state: CreatorReviewCase["state"]): CreatorReviewCase {
  return {
    creatorBio: "quiet rooftop",
    evidences: [],
    intake: {
      acceptsAdultBusinessCompliance: true,
      acceptsAppearanceVerification: true,
      acceptsConsentResponsibility: true,
      acceptsCoPerformerConsentResponsibility: false,
      birthDate: "1999-04-02",
      confirmsInformationMatchesDocuments: true,
      declaresNoProhibitedCategory: true,
      hasCoPerformers: false,
      identityDocumentType: "driver_license",
      legalAddress: "Tokyo-to Shibuya-ku 1-2-3",
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      targetAudienceCategory: "general_adult",
    },
    rejection: null,
    review: {
      approvedAt: null,
      rejectedAt: null,
      submittedAt: "2026-04-18T09:00:00Z",
      suspendedAt: null,
    },
    sharedProfile: {
      avatar: null,
      displayName: "Mina Rei",
      handle: "@minarei",
    },
    state,
    userId: "11111111-1111-1111-1111-111111111111",
  };
}

describe("CreatorReviewDecisionForm", () => {
  beforeEach(() => {
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
  });

  it("submits the selected reject reason and refreshes the route", async () => {
    const user = userEvent.setup();
    const onSubmitDecision = vi.fn().mockResolvedValue(undefined);

    render(
      <CreatorReviewDecisionForm
        onSubmitDecision={onSubmitDecision}
        reviewCase={createReviewCase("submitted")}
      />,
    );

    await user.click(screen.getAllByRole("button", { name: "却下する" })[0]!);
    await user.selectOptions(screen.getByRole("combobox"), "documents_blurry");
    await user.click(screen.getByRole("button", { name: /support review が必要/i }));
    await user.click(screen.getAllByRole("button", { name: "却下する" })[1]!);

    await waitFor(() => {
      expect(onSubmitDecision).toHaveBeenCalledWith({
        decision: "rejected",
        isResubmitEligible: false,
        isSupportReviewRequired: true,
        reasonCode: "documents_blurry",
        userId: "11111111-1111-1111-1111-111111111111",
      });
      expect(mockedRouter.refresh).toHaveBeenCalledTimes(1);
    });
  });

  it("shows a single suspend action for approved cases", async () => {
    const user = userEvent.setup();
    const onSubmitDecision = vi.fn().mockResolvedValue(undefined);

    render(
      <CreatorReviewDecisionForm
        onSubmitDecision={onSubmitDecision}
        reviewCase={createReviewCase("approved")}
      />,
    );

    expect(screen.getByRole("button", { name: "停止する" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "却下する" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "停止する" }));

    await waitFor(() => {
      expect(onSubmitDecision).toHaveBeenCalledWith({
        decision: "suspended",
        isResubmitEligible: false,
        isSupportReviewRequired: false,
        reasonCode: "",
        userId: "11111111-1111-1111-1111-111111111111",
      });
    });
  });

  it("renders a read-only note when no decision is available", () => {
    render(
      <CreatorReviewDecisionForm
        onSubmitDecision={vi.fn()}
        reviewCase={createReviewCase("rejected")}
      />,
    );

    expect(screen.getByText("この状態では追加の admin decision はありません。")).toBeInTheDocument();
  });
});
