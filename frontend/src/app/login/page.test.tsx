import { render, screen } from "@testing-library/react";

import LoginPage from "./page";

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

vi.mock("next/navigation", () => ({
  redirect,
  useRouter: () => mockedRouter,
}));

vi.mock("@/features/fan-auth-gate", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth-gate")>("@/features/fan-auth-gate");

  return {
    ...actual,
    getFanAuthGateState: vi.fn(),
  };
});

describe("LoginPage", () => {
  afterEach(() => {
    redirect.mockReset();
  });

  it("renders the login entry for unauthenticated viewers", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: false,
    });

    render(await LoginPage());

    expect(screen.getByRole("heading", { name: "続けるにはログインが必要です" })).toBeInTheDocument();
  });

  it("redirects authenticated viewers away from the login entry", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_123",
      },
      hasSession: true,
    });

    await LoginPage();

    expect(redirect).toHaveBeenCalledWith("/");
  });
});
