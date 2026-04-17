import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import type { CreatorReviewCase } from "@/entities/creator-review";

import { applyCreatorReviewDecision } from "@/entities/creator-review";
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

vi.mock("@/entities/creator-review", async () => {
  const actual = await vi.importActual<typeof import("@/entities/creator-review")>("@/entities/creator-review");

  return {
    ...actual,
    applyCreatorReviewDecision: vi.fn(),
  };
});

function createReviewCase(state: CreatorReviewCase["state"]): CreatorReviewCase {
  return {
    creatorBio: "quiet rooftop",
    evidences: [],
    intake: {
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      declaresNoProhibitedCategory: true,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
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
    vi.mocked(applyCreatorReviewDecision).mockReset();
  });

  it("submits the selected reject reason and refreshes the route", async () => {
    const user = userEvent.setup();

    vi.mocked(applyCreatorReviewDecision).mockResolvedValue(createReviewCase("rejected"));

    render(<CreatorReviewDecisionForm reviewCase={createReviewCase("submitted")} />);

    await user.click(screen.getAllByRole("button", { name: "却下する" })[0]!);
    await user.selectOptions(screen.getByRole("combobox"), "documents_blurry");
    await user.click(screen.getByRole("button", { name: /support review が必要/i }));
    await user.click(screen.getAllByRole("button", { name: "却下する" })[1]!);

    await waitFor(() => {
      expect(applyCreatorReviewDecision).toHaveBeenCalledWith({
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

    vi.mocked(applyCreatorReviewDecision).mockResolvedValue(createReviewCase("suspended"));

    render(<CreatorReviewDecisionForm reviewCase={createReviewCase("approved")} />);

    expect(screen.getByRole("button", { name: "停止する" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "却下する" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "停止する" }));

    await waitFor(() => {
      expect(applyCreatorReviewDecision).toHaveBeenCalledWith({
        decision: "suspended",
        isResubmitEligible: false,
        isSupportReviewRequired: false,
        reasonCode: "",
        userId: "11111111-1111-1111-1111-111111111111",
      });
    });
  });

  it("renders a read-only note when no decision is available", () => {
    render(<CreatorReviewDecisionForm reviewCase={createReviewCase("rejected")} />);

    expect(screen.getByText("この状態では追加の admin decision はありません。")).toBeInTheDocument();
  });
});
