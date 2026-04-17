import { render } from "@testing-library/react";

import RootPage from "./page";

const {
  FeedShellMock,
  cookiesMock,
  getFollowingFeedShellStateMock,
  loadFeedShellStateMock,
} = vi.hoisted(() => ({
  FeedShellMock: vi.fn(({ state }: { state: unknown }) => <div data-testid="feed-shell">{JSON.stringify(state)}</div>),
  cookiesMock: vi.fn(),
  getFollowingFeedShellStateMock: vi.fn(() => ({
    kind: "auth_required",
    tab: "following",
  })),
  loadFeedShellStateMock: vi.fn(),
}));

vi.mock("next/headers", () => ({
  cookies: cookiesMock,
}));

vi.mock("@/widgets/feed-shell", () => ({
  FeedShell: FeedShellMock,
  getFollowingFeedShellState: getFollowingFeedShellStateMock,
  loadFeedShellState: loadFeedShellStateMock,
}));

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");
  return {
    ...actual,
  };
});

describe("protected fan route auth gates", () => {
  afterEach(() => {
    FeedShellMock.mockClear();
    cookiesMock.mockReset();
    getFollowingFeedShellStateMock.mockClear();
    loadFeedShellStateMock.mockReset();
  });

  it("short-circuits unauthenticated following access into the auth-required shell state", async () => {
    cookiesMock.mockResolvedValue({
      get: vi.fn().mockReturnValue(undefined),
    });

    render(await RootPage({
      searchParams: Promise.resolve({
        tab: "following",
      }),
    }));

    expect(loadFeedShellStateMock).not.toHaveBeenCalled();
    expect(getFollowingFeedShellStateMock).toHaveBeenCalledWith("auth_required");
    expect(FeedShellMock).toHaveBeenCalledWith(
      {
        state: {
          kind: "auth_required",
          tab: "following",
        },
      },
      undefined,
    );
  });
});
