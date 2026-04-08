import { getFanAuthGateState } from "@/features/fan-auth-gate";

import RootPage from "./page";

const { cookiesMock, redirect } = vi.hoisted(() => ({
  cookiesMock: vi.fn(),
  redirect: vi.fn(),
}));

vi.mock("next/headers", () => ({
  cookies: cookiesMock,
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
    cookiesMock.mockReset();
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
});
