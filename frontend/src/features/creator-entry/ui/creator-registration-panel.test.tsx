import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import { CreatorRegistrationPanel } from "@/features/creator-entry";
import { ApiError } from "@/shared/api";

const mockedRouter = vi.hoisted(() => ({
  back: vi.fn(),
  forward: vi.fn(),
  prefetch: vi.fn(),
  push: vi.fn(),
  refresh: vi.fn(),
  replace: vi.fn(),
}));

const apiMocks = vi.hoisted(() => ({
  completeCreatorRegistrationEvidenceUpload: vi.fn(),
  createCreatorRegistrationEvidenceUpload: vi.fn(),
  fetchCreatorRegistration: vi.fn(),
  fetchCreatorRegistrationIntake: vi.fn(),
  registerCreator: vi.fn(),
  saveCreatorRegistrationIntake: vi.fn(),
  uploadCreatorRegistrationEvidenceTarget: vi.fn(),
}));

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");

  return {
    ...actual,
    useRouter: () => mockedRouter,
  };
});

vi.mock("@/features/creator-entry/api", () => apiMocks);

describe("CreatorRegistrationPanel", () => {
  beforeEach(() => {
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    apiMocks.completeCreatorRegistrationEvidenceUpload.mockReset();
    apiMocks.createCreatorRegistrationEvidenceUpload.mockReset();
    apiMocks.fetchCreatorRegistration.mockReset();
    apiMocks.fetchCreatorRegistrationIntake.mockReset();
    apiMocks.registerCreator.mockReset();
    apiMocks.saveCreatorRegistrationIntake.mockReset();
    apiMocks.uploadCreatorRegistrationEvidenceTarget.mockReset();
  });

  it("loads the intake draft and renders the shared profile preview", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: false,
      birthDate: null,
      canSubmit: false,
      creatorBio: "quiet rooftop",
      declaresNoProhibitedCategory: false,
      evidences: [],
      isReadOnly: false,
      legalName: "Mina Rei",
      payoutRecipientName: "",
      payoutRecipientType: null,
      registrationState: "draft",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });

    render(<CreatorRegistrationPanel initialRegistration={null} />);

    expect(await screen.findByRole("heading", { name: "クリエイター登録を始める" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "利用開始後に使える画面" })).toBeInTheDocument();
    expect(screen.getByText("Mina")).toBeInTheDocument();
    expect(screen.getByDisplayValue("quiet rooftop")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Mina Rei")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "編集する" })).toHaveAttribute("href", "/fan/settings/profile");
  });

  it("saves the current draft and routes to the success page on submit", async () => {
    const user = userEvent.setup();

    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: false,
      birthDate: "",
      canSubmit: false,
      creatorBio: "",
      declaresNoProhibitedCategory: false,
      evidences: [
        {
          fileName: "government-id.png",
          fileSizeBytes: 1024,
          kind: "government_id",
          mimeType: "image/png",
          uploadedAt: "2026-04-17T10:30:00.000Z",
        },
        {
          fileName: "bank-proof.pdf",
          fileSizeBytes: 2048,
          kind: "payout_proof",
          mimeType: "application/pdf",
          uploadedAt: "2026-04-17T10:32:00.000Z",
        },
      ],
      isReadOnly: false,
      legalName: "",
      payoutRecipientName: "",
      payoutRecipientType: null,
      registrationState: "draft",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });
    apiMocks.saveCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: true,
      creatorBio: "quiet rooftop",
      declaresNoProhibitedCategory: true,
      evidences: [
        {
          fileName: "government-id.png",
          fileSizeBytes: 1024,
          kind: "government_id",
          mimeType: "image/png",
          uploadedAt: "2026-04-17T10:30:00.000Z",
        },
        {
          fileName: "bank-proof.pdf",
          fileSizeBytes: 2048,
          kind: "payout_proof",
          mimeType: "application/pdf",
          uploadedAt: "2026-04-17T10:32:00.000Z",
        },
      ],
      isReadOnly: false,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "draft",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });
    apiMocks.registerCreator.mockResolvedValue(undefined);

    render(<CreatorRegistrationPanel initialRegistration={null} />);

    await screen.findByRole("heading", { name: "クリエイター登録を始める" });

    await user.type(screen.getByRole("textbox", { name: "紹介文" }), "quiet rooftop");
    await user.type(screen.getByRole("textbox", { name: "氏名" }), "Mina Rei");
    await user.type(screen.getByLabelText("生年月日"), "1999-04-02");
    await user.click(screen.getByLabelText("自分名義"));
    await user.type(screen.getByRole("textbox", { name: "受取名" }), "Mina Rei");
    await user.click(screen.getByRole("checkbox", { name: /禁止されている内容/ }));
    await user.click(screen.getByRole("checkbox", { name: /出演者の同意/ }));
    expect(screen.getByRole("button", { name: "申請を送る" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "下書きを保存する" }));
    await waitFor(() => {
      expect(apiMocks.saveCreatorRegistrationIntake).toHaveBeenCalledWith(
        {
          acceptsConsentResponsibility: true,
          birthDate: "1999-04-02",
          creatorBio: "quiet rooftop",
          declaresNoProhibitedCategory: true,
          legalName: "Mina Rei",
          payoutRecipientName: "Mina Rei",
          payoutRecipientType: "self",
        },
      );
    });
    await user.click(screen.getByRole("button", { name: "申請を送る" }));

    await waitFor(() => {
      expect(apiMocks.saveCreatorRegistrationIntake).toHaveBeenCalledTimes(2);
      expect(apiMocks.registerCreator).toHaveBeenCalledWith();
      expect(mockedRouter.push).toHaveBeenCalledWith("/fan/creator/success");
    });
  });

  it("keeps the save action disabled when the initial intake fetch fails", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockRejectedValue(new Error("boom"));

    render(<CreatorRegistrationPanel initialRegistration={null} />);

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "下書きを保存する" })).not.toBeInTheDocument();
  });

  it("redirects to the success page when the fetched registration is already submitted", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: false,
      creatorBio: "quiet rooftop",
      declaresNoProhibitedCategory: true,
      evidences: [],
      isReadOnly: true,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "submitted",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });

    render(
      <CreatorRegistrationPanel
        initialRegistration={{
          actions: {
            canEnterCreatorMode: false,
            canResubmit: false,
            canSubmit: false,
          },
          creatorDraft: {
            bio: "quiet rooftop",
          },
          rejection: null,
          review: {
            approvedAt: null,
            rejectedAt: null,
            submittedAt: "2026-04-17T10:30:00.000Z",
            suspendedAt: null,
          },
          sharedProfile: {
            avatar: null,
            displayName: "Mina",
            handle: "@mina",
          },
          state: "submitted",
          surface: {
            kind: "read_only_onboarding",
            workspacePreview: "static_mock",
          },
        }}
      />,
    );

    await waitFor(() => {
      expect(mockedRouter.replace).toHaveBeenCalledWith("/fan/creator/success");
    });
  });

  it("disables evidence upload buttons when the intake is read only", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: false,
      creatorBio: "quiet rooftop",
      declaresNoProhibitedCategory: true,
      evidences: [],
      isReadOnly: true,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "rejected",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });
    apiMocks.fetchCreatorRegistration.mockResolvedValue({
      actions: {
        canEnterCreatorMode: false,
        canResubmit: false,
        canSubmit: false,
      },
      creatorDraft: {
        bio: "quiet rooftop",
      },
      rejection: {
        isResubmitEligible: false,
        isSupportReviewRequired: true,
        reasonCode: "impersonation_suspected",
        selfServeResubmitCount: 1,
        selfServeResubmitRemaining: 1,
      },
      review: {
        approvedAt: null,
        rejectedAt: "2026-04-17T10:30:00.000Z",
        submittedAt: "2026-04-16T10:30:00.000Z",
        suspendedAt: null,
      },
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
      state: "rejected",
      surface: {
        kind: "read_only_onboarding",
        workspacePreview: "static_mock",
      },
    });

    render(
      <CreatorRegistrationPanel
        initialRegistration={{
          actions: {
            canEnterCreatorMode: false,
            canResubmit: false,
            canSubmit: false,
          },
          creatorDraft: {
            bio: "quiet rooftop",
          },
          rejection: {
            isResubmitEligible: false,
            isSupportReviewRequired: true,
            reasonCode: "impersonation_suspected",
            selfServeResubmitCount: 1,
            selfServeResubmitRemaining: 1,
          },
          review: {
            approvedAt: null,
            rejectedAt: "2026-04-17T10:30:00.000Z",
            submittedAt: "2026-04-16T10:30:00.000Z",
            suspendedAt: null,
          },
          sharedProfile: {
            avatar: null,
            displayName: "Mina",
            handle: "@mina",
          },
          state: "rejected",
          surface: {
            kind: "read_only_onboarding",
            workspacePreview: "static_mock",
          },
        }}
      />,
    );

    await screen.findByRole("heading", { name: "運営確認が必要です" });

    expect(screen.queryByRole("button", { name: "修正して再申請する" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "再申請する" })).not.toBeInTheDocument();
    expect(screen.queryByText("再申請できる残り回数: 1")).not.toBeInTheDocument();

    for (const button of screen.getAllByRole("button", { name: "書類をアップロード" })) {
      expect(button).toBeDisabled();
    }
  });

  it("shows the edit-and-resubmit path for eligible rejected registrations", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: true,
      creatorBio: "retry bio",
      declaresNoProhibitedCategory: true,
      evidences: [
        {
          fileName: "government-id.png",
          fileSizeBytes: 1024,
          kind: "government_id",
          mimeType: "image/png",
          uploadedAt: "2026-04-17T10:30:00.000Z",
        },
        {
          fileName: "bank-proof.pdf",
          fileSizeBytes: 2048,
          kind: "payout_proof",
          mimeType: "application/pdf",
          uploadedAt: "2026-04-17T10:32:00.000Z",
        },
      ],
      isReadOnly: false,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "rejected",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });

    render(
      <CreatorRegistrationPanel
        initialRegistration={{
          actions: {
            canEnterCreatorMode: false,
            canResubmit: true,
            canSubmit: false,
          },
          creatorDraft: {
            bio: "retry bio",
          },
          rejection: {
            isResubmitEligible: true,
            isSupportReviewRequired: false,
            reasonCode: "documents_incomplete",
            selfServeResubmitCount: 1,
            selfServeResubmitRemaining: 1,
          },
          review: {
            approvedAt: null,
            rejectedAt: "2026-04-17T10:30:00.000Z",
            submittedAt: "2026-04-16T10:30:00.000Z",
            suspendedAt: null,
          },
          sharedProfile: {
            avatar: null,
            displayName: "Mina",
            handle: "@mina",
          },
          state: "rejected",
          surface: {
            kind: "read_only_onboarding",
            workspacePreview: "static_mock",
          },
        }}
      />,
    );

    expect(await screen.findByRole("heading", { name: "申請が差し戻されました" })).toBeInTheDocument();
    expect(screen.getByText("再申請")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "編集する" })).toHaveAttribute("href", "/fan/settings/profile");
    expect(screen.getByRole("button", { name: "修正内容を保存する" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "再申請する" })).toBeInTheDocument();
    expect(screen.getByText("必要な書類または入力内容に不足があります。内容を見直して再度申請してください。")).toBeInTheDocument();
    expect(screen.getByText("残り申請回数：1回", { exact: false })).toBeInTheDocument();
    expect(screen.getByText("要修正")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新しい書類をアップロード" })).toBeInTheDocument();
  });

  it("recovers eligible rejected detail when the server-side status fetch was unavailable", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: true,
      creatorBio: "retry bio",
      declaresNoProhibitedCategory: true,
      evidences: [
        {
          fileName: "government-id.png",
          fileSizeBytes: 1024,
          kind: "government_id",
          mimeType: "image/png",
          uploadedAt: "2026-04-17T10:30:00.000Z",
        },
        {
          fileName: "bank-proof.pdf",
          fileSizeBytes: 2048,
          kind: "payout_proof",
          mimeType: "application/pdf",
          uploadedAt: "2026-04-17T10:32:00.000Z",
        },
      ],
      isReadOnly: false,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "rejected",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });
    apiMocks.fetchCreatorRegistration.mockResolvedValue({
      actions: {
        canEnterCreatorMode: false,
        canResubmit: true,
        canSubmit: false,
      },
      creatorDraft: {
        bio: "retry bio",
      },
      rejection: {
        isResubmitEligible: true,
        isSupportReviewRequired: false,
        reasonCode: "documents_incomplete",
        selfServeResubmitCount: 1,
        selfServeResubmitRemaining: 1,
      },
      review: {
        approvedAt: null,
        rejectedAt: "2026-04-17T10:30:00.000Z",
        submittedAt: "2026-04-16T10:30:00.000Z",
        suspendedAt: null,
      },
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
      state: "rejected",
      surface: {
        kind: "read_only_onboarding",
        workspacePreview: "static_mock",
      },
    });

    render(<CreatorRegistrationPanel initialRegistration={null} />);

    expect(await screen.findByRole("heading", { name: "申請が差し戻されました" })).toBeInTheDocument();
    expect(screen.getByText("残り申請回数：1回", { exact: false })).toBeInTheDocument();
    expect(screen.getByText("必要な書類または入力内容に不足があります。内容を見直して再度申請してください。")).toBeInTheDocument();
  });

  it("shows generic resubmit guidance when rejected detail cannot be recovered", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: true,
      creatorBio: "retry bio",
      declaresNoProhibitedCategory: true,
      evidences: [],
      isReadOnly: false,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "rejected",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });
    apiMocks.fetchCreatorRegistration.mockRejectedValue(new Error("status unavailable"));

    render(<CreatorRegistrationPanel initialRegistration={null} />);

    expect(await screen.findByRole("heading", { name: "申請が差し戻されました" })).toBeInTheDocument();
    expect(screen.getByText("表示された内容を見直し、必要な修正をしてから再度申請してください。")).toBeInTheDocument();
    expect(screen.queryByText("残り申請回数：", { exact: false })).not.toBeInTheDocument();
  });

  it("refreshes the surface after a registration state conflict closes resubmit", async () => {
    const user = userEvent.setup();

    apiMocks.fetchCreatorRegistrationIntake
      .mockResolvedValueOnce({
        acceptsConsentResponsibility: true,
        birthDate: "1999-04-02",
        canSubmit: true,
        creatorBio: "retry bio",
        declaresNoProhibitedCategory: true,
        evidences: [
          {
            fileName: "government-id.png",
            fileSizeBytes: 1024,
            kind: "government_id",
            mimeType: "image/png",
            uploadedAt: "2026-04-17T10:30:00.000Z",
          },
          {
            fileName: "bank-proof.pdf",
            fileSizeBytes: 2048,
            kind: "payout_proof",
            mimeType: "application/pdf",
            uploadedAt: "2026-04-17T10:32:00.000Z",
          },
        ],
        isReadOnly: false,
        legalName: "Mina Rei",
        payoutRecipientName: "Mina Rei",
        payoutRecipientType: "self",
        registrationState: "rejected",
        sharedProfile: {
          avatar: null,
          displayName: "Mina",
          handle: "@mina",
        },
      })
      .mockResolvedValueOnce({
        acceptsConsentResponsibility: true,
        birthDate: "1999-04-02",
        canSubmit: false,
        creatorBio: "retry bio",
        declaresNoProhibitedCategory: true,
        evidences: [
          {
            fileName: "government-id.png",
            fileSizeBytes: 1024,
            kind: "government_id",
            mimeType: "image/png",
            uploadedAt: "2026-04-17T10:30:00.000Z",
          },
          {
            fileName: "bank-proof.pdf",
            fileSizeBytes: 2048,
            kind: "payout_proof",
            mimeType: "application/pdf",
            uploadedAt: "2026-04-17T10:32:00.000Z",
          },
        ],
        isReadOnly: true,
        legalName: "Mina Rei",
        payoutRecipientName: "Mina Rei",
        payoutRecipientType: "self",
        registrationState: "rejected",
        sharedProfile: {
          avatar: null,
          displayName: "Mina",
          handle: "@mina",
        },
      });
    apiMocks.saveCreatorRegistrationIntake.mockRejectedValue(
      new ApiError("state conflict", {
        code: "http",
        details: JSON.stringify({
          error: {
            code: "registration_state_conflict",
            message: "state changed",
          },
          meta: {
            requestId: "req_creator_registration_conflict_001",
          },
        }),
        status: 409,
      }),
    );
    apiMocks.fetchCreatorRegistration.mockResolvedValue({
      actions: {
        canEnterCreatorMode: false,
        canResubmit: false,
        canSubmit: false,
      },
      creatorDraft: {
        bio: "retry bio",
      },
      rejection: {
        isResubmitEligible: false,
        isSupportReviewRequired: true,
        reasonCode: "impersonation_suspected",
        selfServeResubmitCount: 1,
        selfServeResubmitRemaining: 1,
      },
      review: {
        approvedAt: null,
        rejectedAt: "2026-04-17T10:30:00.000Z",
        submittedAt: "2026-04-16T10:30:00.000Z",
        suspendedAt: null,
      },
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
      state: "rejected",
      surface: {
        kind: "read_only_onboarding",
        workspacePreview: "static_mock",
      },
    });

    render(
      <CreatorRegistrationPanel
        initialRegistration={{
          actions: {
            canEnterCreatorMode: false,
            canResubmit: true,
            canSubmit: false,
          },
          creatorDraft: {
            bio: "retry bio",
          },
          rejection: {
            isResubmitEligible: true,
            isSupportReviewRequired: false,
            reasonCode: "documents_incomplete",
            selfServeResubmitCount: 1,
            selfServeResubmitRemaining: 1,
          },
          review: {
            approvedAt: null,
            rejectedAt: "2026-04-17T10:30:00.000Z",
            submittedAt: "2026-04-16T10:30:00.000Z",
            suspendedAt: null,
          },
          sharedProfile: {
            avatar: null,
            displayName: "Mina",
            handle: "@mina",
          },
          state: "rejected",
          surface: {
            kind: "read_only_onboarding",
            workspacePreview: "static_mock",
          },
        }}
      />,
    );

    await screen.findByRole("heading", { name: "申請が差し戻されました" });

    await user.click(screen.getByRole("button", { name: "再申請する" }));

    expect(await screen.findByRole("heading", { name: "運営確認が必要です" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "再申請する" })).not.toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent("現在の申請状態ではこの操作を実行できません。");
  });

  it("shows a non-support rejected surface when self-serve resubmit is exhausted", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: false,
      creatorBio: "retry bio",
      declaresNoProhibitedCategory: true,
      evidences: [],
      isReadOnly: true,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "rejected",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });
    apiMocks.fetchCreatorRegistration.mockResolvedValue({
      actions: {
        canEnterCreatorMode: false,
        canResubmit: false,
        canSubmit: false,
      },
      creatorDraft: {
        bio: "retry bio",
      },
      rejection: {
        isResubmitEligible: true,
        isSupportReviewRequired: false,
        reasonCode: "documents_incomplete",
        selfServeResubmitCount: 2,
        selfServeResubmitRemaining: 0,
      },
      review: {
        approvedAt: null,
        rejectedAt: "2026-04-17T10:30:00.000Z",
        submittedAt: "2026-04-16T10:30:00.000Z",
        suspendedAt: null,
      },
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
      state: "rejected",
      surface: {
        kind: "read_only_onboarding",
        workspacePreview: "static_mock",
      },
    });

    render(
      <CreatorRegistrationPanel
        initialRegistration={{
          actions: {
            canEnterCreatorMode: false,
            canResubmit: false,
            canSubmit: false,
          },
          creatorDraft: {
            bio: "retry bio",
          },
          rejection: {
            isResubmitEligible: true,
            isSupportReviewRequired: false,
            reasonCode: "documents_incomplete",
            selfServeResubmitCount: 2,
            selfServeResubmitRemaining: 0,
          },
          review: {
            approvedAt: null,
            rejectedAt: "2026-04-17T10:30:00.000Z",
            submittedAt: "2026-04-16T10:30:00.000Z",
            suspendedAt: null,
          },
          sharedProfile: {
            avatar: null,
            displayName: "Mina",
            handle: "@mina",
          },
          state: "rejected",
          surface: {
            kind: "read_only_onboarding",
            workspacePreview: "static_mock",
          },
        }}
      />,
    );

    expect(await screen.findByRole("heading", { name: "再申請は利用できません" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "修正して再申請する" })).not.toBeInTheDocument();
    expect(screen.queryByText("運営確認が必要です")).not.toBeInTheDocument();
    expect(screen.getByText("再申請できる残り回数: 0")).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 1, name: "再申請は利用できません" })).toBeInTheDocument();
  });

  it("shows a generic rejected fallback when status detail is unavailable", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: false,
      creatorBio: "retry bio",
      declaresNoProhibitedCategory: true,
      evidences: [],
      isReadOnly: true,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "rejected",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });
    apiMocks.fetchCreatorRegistration.mockRejectedValue(new Error("status unavailable"));

    render(<CreatorRegistrationPanel initialRegistration={null} />);

    expect(await screen.findByRole("heading", { name: "審査状態を再確認してください" })).toBeInTheDocument();
    expect(screen.queryByText("運営確認が必要です")).not.toBeInTheDocument();
    expect(screen.queryByText("再申請は利用できません")).not.toBeInTheDocument();
  });

  it("suppresses stale rejected detail when intake editability disagrees", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: true,
      creatorBio: "retry bio",
      declaresNoProhibitedCategory: true,
      evidences: [],
      isReadOnly: false,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "rejected",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });
    apiMocks.fetchCreatorRegistration.mockResolvedValue({
      actions: {
        canEnterCreatorMode: false,
        canResubmit: true,
        canSubmit: false,
      },
      creatorDraft: {
        bio: "retry bio",
      },
      rejection: {
        isResubmitEligible: true,
        isSupportReviewRequired: false,
        reasonCode: "documents_incomplete",
        selfServeResubmitCount: 1,
        selfServeResubmitRemaining: 1,
      },
      review: {
        approvedAt: null,
        rejectedAt: "2026-04-17T10:30:00.000Z",
        submittedAt: "2026-04-16T10:30:00.000Z",
        suspendedAt: null,
      },
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
      state: "rejected",
      surface: {
        kind: "read_only_onboarding",
        workspacePreview: "static_mock",
      },
    });

    render(
      <CreatorRegistrationPanel
        initialRegistration={{
          actions: {
            canEnterCreatorMode: false,
            canResubmit: false,
            canSubmit: false,
          },
          creatorDraft: {
            bio: "retry bio",
          },
          rejection: {
            isResubmitEligible: false,
            isSupportReviewRequired: true,
            reasonCode: "impersonation_suspected",
            selfServeResubmitCount: 1,
            selfServeResubmitRemaining: 1,
          },
          review: {
            approvedAt: null,
            rejectedAt: "2026-04-17T10:30:00.000Z",
            submittedAt: "2026-04-16T10:30:00.000Z",
            suspendedAt: null,
          },
          sharedProfile: {
            avatar: null,
            displayName: "Mina",
            handle: "@mina",
          },
          state: "rejected",
          surface: {
            kind: "read_only_onboarding",
            workspacePreview: "static_mock",
          },
        }}
      />,
    );

    expect(await screen.findByRole("heading", { name: "申請が差し戻されました" })).toBeInTheDocument();
    expect(screen.queryByText("運営確認が必要です")).not.toBeInTheDocument();
    expect(screen.queryByText("次の対応: 運営確認が必要です")).not.toBeInTheDocument();
    expect(screen.getByText("残り申請回数：1回", { exact: false })).toBeInTheDocument();
    expect(screen.getByText("必要な書類または入力内容に不足があります。内容を見直して再度申請してください。")).toBeInTheDocument();
  });

  it("refreshes read-only rejected detail before keeping a subtype-specific surface", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: false,
      creatorBio: "retry bio",
      declaresNoProhibitedCategory: true,
      evidences: [],
      isReadOnly: true,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "rejected",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });
    apiMocks.fetchCreatorRegistration.mockResolvedValue({
      actions: {
        canEnterCreatorMode: false,
        canResubmit: false,
        canSubmit: false,
      },
      creatorDraft: {
        bio: "retry bio",
      },
      rejection: {
        isResubmitEligible: true,
        isSupportReviewRequired: false,
        reasonCode: "documents_incomplete",
        selfServeResubmitCount: 2,
        selfServeResubmitRemaining: 0,
      },
      review: {
        approvedAt: null,
        rejectedAt: "2026-04-17T10:30:00.000Z",
        submittedAt: "2026-04-16T10:30:00.000Z",
        suspendedAt: null,
      },
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
      state: "rejected",
      surface: {
        kind: "read_only_onboarding",
        workspacePreview: "static_mock",
      },
    });

    render(
      <CreatorRegistrationPanel
        initialRegistration={{
          actions: {
            canEnterCreatorMode: false,
            canResubmit: false,
            canSubmit: false,
          },
          creatorDraft: {
            bio: "retry bio",
          },
          rejection: {
            isResubmitEligible: false,
            isSupportReviewRequired: true,
            reasonCode: "impersonation_suspected",
            selfServeResubmitCount: 1,
            selfServeResubmitRemaining: 1,
          },
          review: {
            approvedAt: null,
            rejectedAt: "2026-04-17T10:30:00.000Z",
            submittedAt: "2026-04-16T10:30:00.000Z",
            suspendedAt: null,
          },
          sharedProfile: {
            avatar: null,
            displayName: "Mina",
            handle: "@mina",
          },
          state: "rejected",
          surface: {
            kind: "read_only_onboarding",
            workspacePreview: "static_mock",
          },
        }}
      />,
    );

    expect(screen.getByRole("heading", { name: "運営確認が必要です" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "再申請は利用できません" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "運営確認が必要です" })).not.toBeInTheDocument();
  });

  it("falls back to the generic rejected surface when a read-only subtype cannot be refreshed", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: false,
      creatorBio: "retry bio",
      declaresNoProhibitedCategory: true,
      evidences: [],
      isReadOnly: true,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "rejected",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });
    apiMocks.fetchCreatorRegistration.mockRejectedValue(new Error("status drift"));

    render(
      <CreatorRegistrationPanel
        initialRegistration={{
          actions: {
            canEnterCreatorMode: false,
            canResubmit: false,
            canSubmit: false,
          },
          creatorDraft: {
            bio: "retry bio",
          },
          rejection: {
            isResubmitEligible: false,
            isSupportReviewRequired: true,
            reasonCode: "impersonation_suspected",
            selfServeResubmitCount: 1,
            selfServeResubmitRemaining: 1,
          },
          review: {
            approvedAt: null,
            rejectedAt: "2026-04-17T10:30:00.000Z",
            submittedAt: "2026-04-16T10:30:00.000Z",
            suspendedAt: null,
          },
          sharedProfile: {
            avatar: null,
            displayName: "Mina",
            handle: "@mina",
          },
          state: "rejected",
          surface: {
            kind: "read_only_onboarding",
            workspacePreview: "static_mock",
          },
        }}
      />,
    );

    expect(screen.getByRole("heading", { name: "運営確認が必要です" })).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "審査状態を再確認してください" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "運営確認が必要です" })).not.toBeInTheDocument();
  });

  it("uses the server registration on the first paint while intake is still loading", () => {
    apiMocks.fetchCreatorRegistrationIntake.mockReturnValue(new Promise(() => {}));

    render(
      <CreatorRegistrationPanel
        initialRegistration={{
          actions: {
            canEnterCreatorMode: false,
            canResubmit: true,
            canSubmit: false,
          },
          creatorDraft: {
            bio: "retry bio",
          },
          rejection: {
            isResubmitEligible: true,
            isSupportReviewRequired: false,
            reasonCode: "documents_incomplete",
            selfServeResubmitCount: 1,
            selfServeResubmitRemaining: 1,
          },
          review: {
            approvedAt: null,
            rejectedAt: "2026-04-17T10:30:00.000Z",
            submittedAt: "2026-04-16T10:30:00.000Z",
            suspendedAt: null,
          },
          sharedProfile: {
            avatar: null,
            displayName: "Mina",
            handle: "@mina",
          },
          state: "rejected",
          surface: {
            kind: "read_only_onboarding",
            workspacePreview: "static_mock",
          },
        }}
      />,
    );

    expect(screen.getByRole("heading", { name: "申請が差し戻されました" })).toBeInTheDocument();
    expect(screen.getByText("再申請")).toBeInTheDocument();
  });

  it("redirects away from the register surface when intake resolves to approved", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: false,
      creatorBio: "approved bio",
      declaresNoProhibitedCategory: true,
      evidences: [],
      isReadOnly: true,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "approved",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });

    render(<CreatorRegistrationPanel initialRegistration={null} />);

    await waitFor(() => {
      expect(mockedRouter.replace).toHaveBeenCalledWith("/fan");
    });
  });

  it("shows the suspended surface as read only", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockResolvedValue({
      acceptsConsentResponsibility: true,
      birthDate: "1999-04-02",
      canSubmit: false,
      creatorBio: "quiet rooftop",
      declaresNoProhibitedCategory: true,
      evidences: [],
      isReadOnly: true,
      legalName: "Mina Rei",
      payoutRecipientName: "Mina Rei",
      payoutRecipientType: "self",
      registrationState: "suspended",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });

    render(
      <CreatorRegistrationPanel
        initialRegistration={{
          actions: {
            canEnterCreatorMode: false,
            canResubmit: false,
            canSubmit: false,
          },
          creatorDraft: {
            bio: "quiet rooftop",
          },
          rejection: null,
          review: {
            approvedAt: null,
            rejectedAt: null,
            submittedAt: "2026-04-16T10:30:00.000Z",
            suspendedAt: "2026-04-17T10:30:00.000Z",
          },
          sharedProfile: {
            avatar: null,
            displayName: "Mina",
            handle: "@mina",
          },
          state: "suspended",
          surface: {
            kind: "read_only_onboarding",
            workspacePreview: "static_mock",
          },
        }}
      />,
    );

    expect(await screen.findByRole("heading", { name: "停止中のため再申請できません" })).toBeInTheDocument();
    expect(screen.getByText("利用停止中")).toBeInTheDocument();
    expect(screen.getByText("停止中のため再申請できません")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "修正して再申請する" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "申請を送る" })).not.toBeInTheDocument();
  });
});
