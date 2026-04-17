import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import { ApiError } from "@/shared/api";

import { CreatorRegistrationSuccessPanel } from "./creator-registration-success-panel";

const mockedRouter = vi.hoisted(() => ({
  replace: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => mockedRouter,
}));

vi.mock("../api/fetch-creator-registration", () => ({
  fetchCreatorRegistration: vi.fn(),
}));

describe("CreatorRegistrationSuccessPanel", () => {
  afterEach(() => {
    mockedRouter.replace.mockReset();
  });

  it("renders concise Japanese copy for submitted registrations", async () => {
    const { fetchCreatorRegistration } = await import("../api/fetch-creator-registration");

    vi.mocked(fetchCreatorRegistration).mockResolvedValue({
      actions: {
        canEnterCreatorMode: false,
        canResubmit: false,
        canSubmit: false,
      },
      creatorDraft: {
        bio: "",
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
    });

    render(<CreatorRegistrationSuccessPanel />);

    expect(await screen.findByRole("heading", { name: "申請を受け付けました" })).toBeInTheDocument();
    expect(screen.getByText("現在の状態")).toBeInTheDocument();
    expect(screen.getByText("確認中")).toBeInTheDocument();
    expect(screen.getByText("受付日時")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "プロフィール設定を開く" })).toHaveAttribute("href", "/fan/settings/profile");
    expect(screen.getByRole("link", { name: "ホームに戻る" })).toHaveAttribute("href", "/fan");
    expect(screen.queryByText("creator submitted")).not.toBeInTheDocument();
    expect(screen.queryByText("fan hub に戻る")).not.toBeInTheDocument();
  });

  it("redirects back to the register page when the latest status is not submitted", async () => {
    const { fetchCreatorRegistration } = await import("../api/fetch-creator-registration");

    vi.mocked(fetchCreatorRegistration).mockResolvedValue({
      actions: {
        canEnterCreatorMode: false,
        canResubmit: false,
        canSubmit: true,
      },
      creatorDraft: {
        bio: "",
      },
      rejection: null,
      review: {
        approvedAt: null,
        rejectedAt: null,
        submittedAt: null,
        suspendedAt: null,
      },
      sharedProfile: {
        avatar: null,
        displayName: "Mina",
        handle: "@mina",
      },
      state: "draft",
      surface: {
        kind: "read_only_onboarding",
        workspacePreview: "static_mock",
      },
    });

    render(<CreatorRegistrationSuccessPanel />);

    await waitFor(() => {
      expect(mockedRouter.replace).toHaveBeenCalledWith("/fan/creator/register");
    });
  });

  it("shows a Japanese profile guidance message when registration status is missing", async () => {
    const { fetchCreatorRegistration } = await import("../api/fetch-creator-registration");

    vi.mocked(fetchCreatorRegistration).mockRejectedValue(
      new ApiError("not found", {
        code: "http",
        details: JSON.stringify({
          error: {
            code: "not_found",
            message: "missing profile",
          },
          meta: {
            requestId: "req_creator_success_001",
          },
        }),
        status: 404,
      }),
    );

    render(<CreatorRegistrationSuccessPanel />);

    expect(await screen.findByRole("alert")).toHaveTextContent("プロフィール設定をご確認ください。");
    expect(screen.queryByText("profile settings")).not.toBeInTheDocument();
    expect(screen.queryByText("確認中")).not.toBeInTheDocument();
    expect(screen.queryByText("現在の状態")).not.toBeInTheDocument();
  });
});
