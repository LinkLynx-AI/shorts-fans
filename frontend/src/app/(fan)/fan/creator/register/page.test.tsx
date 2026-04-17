import { render, screen } from "@testing-library/react";

import {
  CurrentViewerProvider,
  ViewerSessionProvider,
} from "@/entities/viewer";

import CreatorRegisterPage from "./page";

const { cookies } = vi.hoisted(() => ({
  cookies: vi.fn(),
}));
const { redirect } = vi.hoisted(() => ({
  redirect: vi.fn(),
}));
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
    redirect,
    useRouter: () => mockedRouter,
  };
});

vi.mock("next/headers", () => ({
  cookies,
}));

vi.mock("@/features/fan-auth-gate", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth-gate")>("@/features/fan-auth-gate");

  return {
    ...actual,
    getFanAuthGateState: vi.fn(),
  };
});

vi.mock("@/features/creator-entry", () => ({
  CreatorRegistrationPanel: () => <div>creator registration panel</div>,
  fetchCreatorRegistration: vi.fn(),
  getCreatorEntryErrorCode: vi.fn(),
}));

describe("CreatorRegisterPage", () => {
  afterEach(() => {
    cookies.mockReset();
    redirect.mockReset();
  });

  it("redirects unauthenticated viewers to login", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: false,
    });

    await CreatorRegisterPage();

    expect(redirect).toHaveBeenCalledWith("/login");
  });

  it("renders the registration form for authenticated fans without creator access", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");
    const { fetchCreatorRegistration, getCreatorEntryErrorCode } = await import("@/features/creator-entry");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_123",
      },
      hasSession: true,
    });
    cookies.mockResolvedValue({
      get: vi.fn().mockReturnValue({ value: "raw-session-token" }),
    });
    vi.mocked(fetchCreatorRegistration).mockResolvedValue(null);
    vi.mocked(getCreatorEntryErrorCode).mockReturnValue(null);

    render(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider
          currentViewer={{
            activeMode: "fan",
            canAccessCreatorMode: false,
            id: "viewer_123",
          }}
        >
          {await CreatorRegisterPage()}
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    expect(screen.getByText("creator registration panel")).toBeInTheDocument();
  });

  it("redirects submitted viewers to the success page on the server", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");
    const { fetchCreatorRegistration, getCreatorEntryErrorCode } = await import("@/features/creator-entry");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_123",
      },
      hasSession: true,
    });
    cookies.mockResolvedValue({
      get: vi.fn().mockReturnValue({ value: "raw-session-token" }),
    });
    vi.mocked(getCreatorEntryErrorCode).mockReturnValue(null);
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

    await CreatorRegisterPage();

    expect(redirect).toHaveBeenCalledWith("/fan/creator/success");
  });

  it("redirects missing shared profiles to profile settings", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");
    const { fetchCreatorRegistration, getCreatorEntryErrorCode } = await import("@/features/creator-entry");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_123",
      },
      hasSession: true,
    });
    cookies.mockResolvedValue({
      get: vi.fn().mockReturnValue({ value: "raw-session-token" }),
    });
    vi.mocked(fetchCreatorRegistration).mockRejectedValue(new Error("not_found"));
    vi.mocked(getCreatorEntryErrorCode).mockReturnValue("not_found");

    await CreatorRegisterPage();

    expect(redirect).toHaveBeenCalledWith("/fan/settings/profile");
  });
});
