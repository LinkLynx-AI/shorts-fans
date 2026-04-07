const {
  cookiesMock,
  getCurrentViewerBootstrapMock,
} = vi.hoisted(() => ({
  cookiesMock: vi.fn(),
  getCurrentViewerBootstrapMock: vi.fn(),
}));

vi.mock("next/headers", () => ({
  cookies: cookiesMock,
}));

vi.mock("@/entities/viewer", async () => {
  const actual = await vi.importActual<typeof import("@/entities/viewer")>("@/entities/viewer");

  return {
    ...actual,
    getCurrentViewerBootstrap: getCurrentViewerBootstrapMock,
  };
});

describe("getFanAuthGateState", () => {
  beforeEach(() => {
    vi.resetModules();
    cookiesMock.mockReset();
    getCurrentViewerBootstrapMock.mockReset();
  });

  it("treats missing bootstrap viewer as unauthenticated even when the cookie exists", async () => {
    const { getFanAuthGateState } = await import("./request-fan-auth-gate-state");

    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "stale-session",
      }),
    });
    getCurrentViewerBootstrapMock.mockResolvedValue(null);

    await expect(getFanAuthGateState()).resolves.toEqual({
      currentViewer: null,
      hasSession: false,
    });
  });

  it("returns an authenticated gate state when bootstrap resolves a viewer", async () => {
    const { getFanAuthGateState } = await import("./request-fan-auth-gate-state");

    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    getCurrentViewerBootstrapMock.mockResolvedValue({
      activeMode: "fan",
      canAccessCreatorMode: false,
      id: "viewer_123",
    });

    await expect(getFanAuthGateState()).resolves.toEqual({
      currentViewer: {
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_123",
      },
      hasSession: true,
    });
  });

  it("falls back to unauthenticated when bootstrap fails unexpectedly", async () => {
    const { getFanAuthGateState } = await import("./request-fan-auth-gate-state");

    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "broken-session",
      }),
    });
    getCurrentViewerBootstrapMock.mockRejectedValue(new Error("bootstrap failed"));

    await expect(getFanAuthGateState()).resolves.toEqual({
      currentViewer: null,
      hasSession: false,
    });
  });
});
