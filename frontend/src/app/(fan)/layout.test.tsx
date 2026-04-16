import { render, screen } from "@testing-library/react";

import FanLayout from "./layout";

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
    useRouter: () => mockedRouter,
    usePathname: () => "/",
  };
});

vi.mock("@/features/fan-auth-gate", async () => {
  const actual = await vi.importActual<typeof import("@/features/fan-auth-gate")>("@/features/fan-auth-gate");

  return {
    ...actual,
    getFanAuthGateState: vi.fn(),
  };
});

describe("FanLayout", () => {
  it("keeps the width cap while removing the desktop height clamp", async () => {
    const { getFanAuthGateState } = await import("@/features/fan-auth-gate");

    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_fan_001",
      },
      hasSession: true,
    });

    render(await FanLayout({ children: <div>fan surface</div> }));

    const contentWrapper = screen.getByText("fan surface").parentElement;
    const frame = contentWrapper?.parentElement;
    const shell = frame?.parentElement;

    expect(contentWrapper?.className).not.toContain("sm:min-h-[calc(100svh-48px)]");
    expect(frame?.className).toContain("max-w-[408px]");
    expect(frame?.className).not.toContain("sm:min-h-[calc(100svh-48px)]");
    expect(frame?.className).not.toContain("sm:rounded-[36px]");
    expect(shell?.className).not.toContain("sm:py-6");
  });
});
