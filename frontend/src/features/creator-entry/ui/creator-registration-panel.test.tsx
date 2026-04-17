import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import { CreatorRegistrationPanel } from "@/features/creator-entry";

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

    render(<CreatorRegistrationPanel />);

    expect(await screen.findByRole("heading", { name: "Creator審査申請を始める" })).toBeInTheDocument();
    expect(screen.getByText("Mina")).toBeInTheDocument();
    expect(screen.getByDisplayValue("quiet rooftop")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Mina Rei")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Profile settings を開く" })).toHaveAttribute("href", "/fan/settings/profile");
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

    render(<CreatorRegistrationPanel />);

    await screen.findByRole("heading", { name: "Creator審査申請を始める" });

    await user.type(screen.getByRole("textbox", { name: "Bio" }), "quiet rooftop");
    await user.type(screen.getByRole("textbox", { name: "Legal name" }), "Mina Rei");
    await user.type(screen.getByLabelText("Birth date"), "1999-04-02");
    await user.click(screen.getByLabelText("自分名義"));
    await user.type(screen.getByRole("textbox", { name: "Payout recipient name" }), "Mina Rei");
    await user.click(screen.getByRole("checkbox", { name: /prohibited category/i }));
    await user.click(screen.getByRole("checkbox", { name: /consent/i }));
    expect(screen.getByRole("button", { name: "審査申請を送信する" })).toBeDisabled();
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
    await user.click(screen.getByRole("button", { name: "審査申請を送信する" }));

    await waitFor(() => {
      expect(apiMocks.saveCreatorRegistrationIntake).toHaveBeenCalledTimes(2);
      expect(apiMocks.registerCreator).toHaveBeenCalledWith();
      expect(mockedRouter.push).toHaveBeenCalledWith("/fan/creator/success");
    });
  });

  it("keeps the save action disabled when the initial intake fetch fails", async () => {
    apiMocks.fetchCreatorRegistrationIntake.mockRejectedValue(new Error("boom"));

    render(<CreatorRegistrationPanel />);

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "下書きを保存する" })).toBeDisabled();
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
      registrationState: "submitted",
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
    });

    render(<CreatorRegistrationPanel />);

    await screen.findByRole("heading", { name: "Creator審査申請を始める" });

    for (const button of screen.getAllByRole("button", { name: "証跡を選択する" })) {
      expect(button).toBeDisabled();
    }
  });
});
