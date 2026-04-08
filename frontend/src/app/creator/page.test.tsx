import { render, screen } from "@testing-library/react";

import CreatorPage from "./page";

const { redirect } = vi.hoisted(() => ({
  redirect: vi.fn(),
}));

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");

  return {
    ...actual,
    redirect,
  };
});

vi.mock("@/features/fan-auth-gate", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth-gate")>("@/features/fan-auth-gate");

  return {
    ...actual,
    getFanAuthGateState: vi.fn(),
  };
});

describe("CreatorPage", () => {
  afterEach(() => {
    redirect.mockReset();
  });

  it("redirects creator-capable viewers in fan mode to the success surface", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: true,
        id: "viewer_123",
      },
      hasSession: true,
    });

    await CreatorPage();

    expect(redirect).toHaveBeenCalledWith("/fan/creator/success");
  });

  it("renders the placeholder creator route when creator mode is active", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_123",
      },
      hasSession: true,
    });

    render(await CreatorPage());

    expect(screen.getByRole("heading", { name: "Creator mode" })).toBeInTheDocument();
  });
});
