import { render, screen } from "@testing-library/react";

import { CurrentViewerProvider, ViewerSessionProvider } from "@/entities/viewer";
import {
  fetchFanProfileFollowingPage,
  type FanProfileFollowingPage,
} from "@/entities/fan-profile";
import { FanAuthDialogProvider } from "@/features/fan-auth";
import { ApiError } from "@/shared/api";

import FollowingPage from "./page";

const { cookiesMock, mockedRouter } = vi.hoisted(() => ({
  cookiesMock: vi.fn(),
  mockedRouter: {
    back: vi.fn(),
    forward: vi.fn(),
    prefetch: vi.fn(),
    push: vi.fn(),
    refresh: vi.fn(),
    replace: vi.fn(),
  },
}));

vi.mock("next/headers", () => ({
  cookies: cookiesMock,
}));

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");

  return {
    ...actual,
    useRouter: () => mockedRouter,
  };
});

vi.mock("@/entities/fan-profile", async () => {
  const actual = await vi.importActual<typeof import("@/entities/fan-profile")>("@/entities/fan-profile");

  return {
    ...actual,
    fetchFanProfileFollowingPage: vi.fn(),
  };
});

function renderWithFanAuthDialog(node: React.ReactNode) {
  return render(
    <ViewerSessionProvider hasSession={false}>
      <CurrentViewerProvider currentViewer={null}>
        <FanAuthDialogProvider>{node}</FanAuthDialogProvider>
      </CurrentViewerProvider>
    </ViewerSessionProvider>,
  );
}

describe("FollowingPage", () => {
  beforeEach(() => {
    cookiesMock.mockReset();
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    vi.mocked(fetchFanProfileFollowingPage).mockReset();
  });

  it("fetches following data with the viewer session and renders the list", async () => {
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    vi.mocked(fetchFanProfileFollowingPage).mockResolvedValue({
      items: [
        {
          creator: {
            avatar: null,
            bio: "Public shorts から paid main へつながる creator mock profile.",
            displayName: "Mika Aoi",
            handle: "@mikaaoi",
            id: "creator_11111111111111111111111111111111",
          },
          viewer: {
            isFollowing: true,
          },
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_fan_profile_following_page_001",
    } satisfies FanProfileFollowingPage);

    renderWithFanAuthDialog(await FollowingPage());

    expect(fetchFanProfileFollowingPage).toHaveBeenCalledWith({
      sessionToken: "valid-session",
    });
    expect(screen.getByRole("heading", { name: "following" })).toBeInTheDocument();
    expect(screen.getByText("Mika Aoi")).toBeInTheDocument();
    expect(screen.getByText("1 creators")).toBeInTheDocument();
  });

  it("opens the shared auth dialog when the following api responds with auth_required", async () => {
    cookiesMock.mockResolvedValue({
      get: () => undefined,
    });
    vi.mocked(fetchFanProfileFollowingPage).mockRejectedValue(
      new ApiError("unauthorized", {
        code: "http",
        details: JSON.stringify({
          error: {
            code: "auth_required",
            message: "login required",
          },
        }),
        status: 401,
      }),
    );

    renderWithFanAuthDialog(await FollowingPage());

    expect(await screen.findByRole("dialog", { name: "続けるにはログインが必要です" })).toBeInTheDocument();
  });
});
