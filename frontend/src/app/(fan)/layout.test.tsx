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
  it("keeps the width cap while pinning the bottom navigation to the viewport", async () => {
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
    const navigation = screen.getByRole("navigation", { name: "Primary" });
    const navigationWidthWrapper = navigation.parentElement;
    const navigationViewportWrapper = navigationWidthWrapper?.parentElement;

    expect(contentWrapper?.className).not.toContain("sm:min-h-[calc(100svh-48px)]");
    expect(contentWrapper?.className).toContain("pb-[calc(76px+env(safe-area-inset-bottom,0px))]");
    expect(frame?.className).toContain("max-w-[408px]");
    expect(frame?.className).not.toContain("sm:min-h-[calc(100svh-48px)]");
    expect(frame?.className).not.toContain("sm:rounded-[36px]");
    expect(shell?.className).not.toContain("sm:py-6");
    expect(navigationWidthWrapper?.className).toContain("max-w-[408px]");
    expect(navigationViewportWrapper?.className).toContain("fixed");
    expect(navigationViewportWrapper?.className).toContain("bottom-0");
  });
});
