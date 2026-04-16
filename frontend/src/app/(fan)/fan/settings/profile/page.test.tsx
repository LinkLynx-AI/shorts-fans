import { render, screen } from "@testing-library/react";

import {
  CurrentViewerProvider,
  ViewerSessionProvider,
} from "@/entities/viewer";
import { getViewerProfile } from "@/features/viewer-profile";

import FanProfileSettingsPage from "./page";

const { cookiesMock, redirect } = vi.hoisted(() => ({
  cookiesMock: vi.fn(),
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

vi.mock("next/headers", () => ({
  cookies: cookiesMock,
}));

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");

  return {
    ...actual,
    redirect,
    useRouter: () => mockedRouter,
  };
});

vi.mock("@/features/fan-auth-gate", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth-gate")>("@/features/fan-auth-gate");

  return {
    ...actual,
    getFanAuthGateState: vi.fn(),
  };
});

vi.mock("@/features/viewer-profile", async () => {
  const actual = await vi.importActual<typeof import("@/features/viewer-profile")>("@/features/viewer-profile");

  return {
    ...actual,
    getViewerProfile: vi.fn(),
  };
});

describe("FanProfileSettingsPage", () => {
  beforeEach(() => {
    cookiesMock.mockReset();
    redirect.mockReset();
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    vi.mocked(getViewerProfile).mockReset();
    cookiesMock.mockResolvedValue({
      get: vi.fn().mockReturnValue({
        value: "raw-session-token",
      }),
    });
    vi.mocked(getViewerProfile).mockResolvedValue({
      avatar: {
        durationSeconds: null,
        id: "asset_viewer_profile_avatar",
        kind: "image",
        posterUrl: null,
        url: "https://cdn.example.com/viewer/mina/avatar.jpg",
      },
      displayName: "Mina Rei",
      handle: "@minarei",
    });
  });

  it("redirects unauthenticated viewers to login", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: false,
    });

    await FanProfileSettingsPage();

    expect(redirect).toHaveBeenCalledWith("/login");
  });

  it("renders the shared profile editor with viewer profile values", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_123",
      },
      hasSession: true,
    });

    render(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider
          currentViewer={{
            activeMode: "fan",
            canAccessCreatorMode: false,
            id: "viewer_123",
          }}
        >
          {await FanProfileSettingsPage()}
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    expect(await screen.findByDisplayValue("Mina Rei")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "プロフィールを編集" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: /handle/i })).toHaveValue("@minarei");
    expect(screen.getByRole("link", { name: "fan hub に戻る" })).toHaveAttribute("href", "/fan");
    expect(screen.getByRole("button", { name: "保存する" })).toBeInTheDocument();
  });
});
