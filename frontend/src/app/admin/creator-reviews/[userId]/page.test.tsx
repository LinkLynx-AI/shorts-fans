import {
  render,
  screen,
} from "@testing-library/react";

import type { CreatorReviewCase } from "@/entities/creator-review";

import { getCreatorReviewCase } from "@/entities/creator-review";

import { assertAdminUiAccess } from "../../_lib/admin-ui-access";
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
  assertAdminUiAccess: vi.fn(),
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
      acceptsAdultBusinessCompliance: true,
      acceptsAppearanceVerification: true,
      acceptsConsentResponsibility: true,
      acceptsCoPerformerConsentResponsibility: false,
      birthDate: "2000-01-01",
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
    vi.mocked(assertAdminUiAccess).mockReset();
    vi.mocked(getCreatorReviewCase).mockReset();
  });

  it("renders review navigation to video review from the creator detail", async () => {
    const reviewCase = createReviewCase();
    vi.mocked(getCreatorReviewCase).mockResolvedValue(reviewCase);

    render(await AdminCreatorReviewCasePage({
      params: Promise.resolve({ userId: reviewCase.userId }),
      searchParams: Promise.resolve({}),
    }));

    expect(getCreatorReviewCase).toHaveBeenCalledWith({
      fetcher: expect.any(Function),
      userId: reviewCase.userId,
    });
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
    expect(screen.getByText("Tokyo-to Shibuya-ku 1-2-3")).toBeInTheDocument();
    expect(screen.getByText("運転免許証")).toBeInTheDocument();
    expect(screen.getByText("成人向け")).toBeInTheDocument();
    expect(screen.getByText("document match")).toBeInTheDocument();
    expect(screen.getByText("appearance verification")).toBeInTheDocument();
    expect(screen.getByText("adult business compliance")).toBeInTheDocument();
    expect(screen.getByText("co performers")).toBeInTheDocument();
  });

  it("does not fetch the review case when admin access is rejected", async () => {
    const reviewCase = createReviewCase();
    vi.mocked(assertAdminUiAccess).mockRejectedValue(new Error("blocked"));

    await expect(AdminCreatorReviewCasePage({
      params: Promise.resolve({ userId: reviewCase.userId }),
      searchParams: Promise.resolve({}),
    })).rejects.toThrow("blocked");

    expect(getCreatorReviewCase).not.toHaveBeenCalled();
  });
});
