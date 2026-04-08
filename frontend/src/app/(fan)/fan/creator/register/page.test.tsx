import { render, screen } from "@testing-library/react";

import {
  CurrentViewerProvider,
  ViewerSessionProvider,
} from "@/entities/viewer";

import CreatorRegisterPage from "./page";

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

vi.mock("@/features/fan-auth-gate", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth-gate")>("@/features/fan-auth-gate");

  return {
    ...actual,
    getFanAuthGateState: vi.fn(),
  };
});

describe("CreatorRegisterPage", () => {
  afterEach(() => {
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
          {await CreatorRegisterPage()}
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    expect(screen.getByRole("heading", { name: "Creator登録を始める" })).toBeInTheDocument();
  });
});
