import { render, screen } from "@testing-library/react";

import CreatorSuccessPage from "./page";

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

describe("CreatorSuccessPage", () => {
  afterEach(() => {
    redirect.mockReset();
  });

  it("redirects unregistered viewers back to the register page", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_123",
      },
      hasSession: true,
    });

    await CreatorSuccessPage();

    expect(redirect).toHaveBeenCalledWith("/fan/creator/register");
  });

  it("renders the success panel for creator-capable fans still in fan mode", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: true,
        id: "viewer_123",
      },
      hasSession: true,
    });

    render(await CreatorSuccessPage());

    expect(screen.getByRole("heading", { name: "Creator登録が完了しました" })).toBeInTheDocument();
  });
});
