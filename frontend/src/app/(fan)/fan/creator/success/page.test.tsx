import { render, screen } from "@testing-library/react";

import CreatorSuccessPage from "./page";

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
  CreatorRegistrationSuccessPanel: () => <div>creator registration success panel</div>,
  fetchCreatorRegistration: vi.fn(),
  getCreatorEntryErrorCode: vi.fn(),
}));

describe("CreatorSuccessPage", () => {
  afterEach(() => {
    cookies.mockReset();
    redirect.mockReset();
  });

  it("renders the submitted receipt panel for authenticated fans without creator access", async () => {
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

    render(await CreatorSuccessPage());

    expect(screen.getByText("creator registration success panel")).toBeInTheDocument();
  });

  it("redirects creator-capable fans in fan mode back to fan hub", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: true,
        id: "viewer_123",
      },
      hasSession: true,
    });

    await CreatorSuccessPage();

    expect(redirect).toHaveBeenCalledWith("/fan");
  });

  it("redirects non-submitted viewers back to the register surface on the server", async () => {
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

    await CreatorSuccessPage();

    expect(redirect).toHaveBeenCalledWith("/fan/creator/register");
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

    await CreatorSuccessPage();

    expect(redirect).toHaveBeenCalledWith("/fan/settings/profile");
  });

  it("renders the success panel when the server-side status fetch fails for a non-not_found reason", async () => {
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
    vi.mocked(fetchCreatorRegistration).mockRejectedValue(new Error("boom"));
    vi.mocked(getCreatorEntryErrorCode).mockReturnValue(null);

    render(await CreatorSuccessPage());

    expect(screen.getByText("creator registration success panel")).toBeInTheDocument();
    expect(redirect).not.toHaveBeenCalledWith("/fan/creator/register");
  });
});
