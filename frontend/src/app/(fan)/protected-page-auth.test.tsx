import { getFanAuthGateState } from "@/features/fan-auth-gate";

import FanPage from "./fan/page";
import FollowingPage from "./fan/following/page";
import RootPage from "./page";

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

describe("protected fan route auth gates", () => {
  afterEach(() => {
    redirect.mockReset();
  });

  it("redirects unauthenticated following feed access to the login entry", async () => {
    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: false,
    });

    await RootPage({
      searchParams: Promise.resolve({
        tab: "following",
      }),
    });

    expect(redirect).toHaveBeenCalledWith("/login");
  });

  it("redirects unauthenticated fan hub access to the login entry", async () => {
    vi.mocked(getFanAuthGateState).mockResolvedValue({
      currentViewer: null,
      hasSession: false,
    });

    await FanPage({
      searchParams: Promise.resolve({}),
    });
    await FollowingPage();

    expect(redirect).toHaveBeenNthCalledWith(1, "/login");
    expect(redirect).toHaveBeenNthCalledWith(2, "/login");
  });
});
