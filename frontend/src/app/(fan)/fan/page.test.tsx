import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import { CurrentViewerProvider, ViewerSessionProvider } from "@/entities/viewer";
import {
  fetchFanProfileFollowingPage,
  fetchFanProfileLibraryPage,
  fetchFanProfileOverview,
  fetchFanProfilePinnedShortsPage,
} from "@/entities/fan-profile";
import { FanAuthDialogProvider } from "@/features/fan-auth";
import { getViewerProfile } from "@/features/viewer-profile";
import { ApiError } from "@/shared/api";

import FanPage from "./page";

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
    fetchFanProfileLibraryPage: vi.fn(),
    fetchFanProfileOverview: vi.fn(),
    fetchFanProfilePinnedShortsPage: vi.fn(),
  };
});

vi.mock("@/features/viewer-profile", async () => {
  const actual = await vi.importActual<typeof import("@/features/viewer-profile")>("@/features/viewer-profile");

  return {
    ...actual,
    getViewerProfile: vi.fn(),
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

describe("FanPage", () => {
  beforeEach(() => {
    cookiesMock.mockReset();
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    vi.mocked(fetchFanProfileFollowingPage).mockReset();
    vi.mocked(fetchFanProfileLibraryPage).mockReset();
    vi.mocked(fetchFanProfileOverview).mockReset();
    vi.mocked(fetchFanProfilePinnedShortsPage).mockReset();
    vi.mocked(getViewerProfile).mockReset();
    vi.mocked(getViewerProfile).mockResolvedValue({
      avatar: null,
      displayName: "Alex_Fan",
      handle: "@alex_f",
    });
  });

  it("fetches overview and pinned shorts with the viewer session when the pinned tab is active", async () => {
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    vi.mocked(fetchFanProfileOverview).mockResolvedValue({
      counts: {
        following: 9,
        library: 7,
        pinnedShorts: 8,
      },
      title: "Archive from API",
    });
    vi.mocked(fetchFanProfilePinnedShortsPage).mockResolvedValue({
      items: [
        {
          creator: {
            avatar: null,
            bio: "after rain と balcony mood の short をまとめています。",
            displayName: "Sora Vale",
            handle: "@soravale",
            id: "creator_sora_vale",
          },
          short: {
            caption: "after rain preview",
            canonicalMainId: "main_sora_after_rain",
            creatorId: "creator_sora_vale",
            id: "short_sora_afterrain",
            media: {
              durationSeconds: 17,
              id: "asset_short_sora_afterrain",
              kind: "video",
              posterUrl: "https://cdn.example.com/shorts/sora-after-rain-poster.jpg",
              url: "https://cdn.example.com/shorts/sora-after-rain.mp4",
            },
            previewDurationSeconds: 17,
          },
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_fan_profile_pinned_shorts_001",
    });

    renderWithFanAuthDialog(
      await FanPage({
        searchParams: Promise.resolve({}),
      }),
    );

    expect(fetchFanProfileOverview).toHaveBeenCalledWith({
      sessionToken: "valid-session",
    });
    expect(getViewerProfile).toHaveBeenCalledWith({
      sessionToken: "valid-session",
    });
    expect(fetchFanProfilePinnedShortsPage).toHaveBeenCalledWith({
      sessionToken: "valid-session",
    });
    expect(screen.getByRole("heading", { name: "Profile" })).toBeInTheDocument();
    expect(screen.getByText("Alex_Fan")).toBeInTheDocument();
    expect(screen.getByText("@alex_f")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Pinned Shorts" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: "Sora Vale after rain preview" })).toHaveAttribute(
      "href",
      "/shorts/short_sora_afterrain?fanTab=pinned&from=fan",
    );
    expect(screen.getByText("Following 9, Pinned Shorts 8, Library 7, Archive from API")).toBeInTheDocument();
  });

  it("fetches overview and following items when the following tab is active", async () => {
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    vi.mocked(fetchFanProfileOverview).mockResolvedValue({
      counts: {
        following: 3,
        library: 2,
        pinnedShorts: 4,
      },
      title: "Archive from API",
    });
    vi.mocked(fetchFanProfileFollowingPage).mockResolvedValue({
      items: [
        {
          creator: {
            avatar: null,
            bio: "after rain と balcony mood の short をまとめています。",
            displayName: "Mika Aoi",
            handle: "@mikaaoi",
            id: "creator_mika_aoi",
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
      requestId: "req_fan_profile_following_001",
    });

    renderWithFanAuthDialog(
      await FanPage({
        searchParams: Promise.resolve({
          tab: "following",
        }),
      }),
    );

    expect(fetchFanProfilePinnedShortsPage).not.toHaveBeenCalled();
    expect(fetchFanProfileLibraryPage).not.toHaveBeenCalled();
    expect(fetchFanProfileFollowingPage).toHaveBeenCalledWith({
      sessionToken: "valid-session",
    });
    expect(screen.getByRole("link", { name: "Following" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("Mika Aoi")).toBeInTheDocument();
    expect(screen.getByText("1 creators")).toBeInTheDocument();
  });

  it("fetches overview and library items when the library tab is active", async () => {
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    vi.mocked(fetchFanProfileOverview).mockResolvedValue({
      counts: {
        following: 1,
        library: 2,
        pinnedShorts: 3,
      },
      title: "Archive from API",
    });
    vi.mocked(fetchFanProfileLibraryPage).mockResolvedValue({
      items: [
        {
          access: {
            mainId: "main_mina_quiet_rooftop",
            reason: "session_unlocked",
            status: "unlocked",
          },
          creator: {
            avatar: null,
            bio: "quiet rooftop と hotel light の preview を軸に投稿。",
            displayName: "Mina Rei",
            handle: "@minarei",
            id: "creator_mina_rei",
          },
          entryShort: {
            caption: "quiet rooftop preview。",
            canonicalMainId: "main_mina_quiet_rooftop",
            creatorId: "creator_mina_rei",
            id: "short_mina_rooftop",
            media: {
              durationSeconds: 16,
              id: "asset_short_mina_rooftop",
              kind: "video",
              posterUrl: "https://cdn.example.com/shorts/mina-rooftop-poster.jpg",
              url: "https://cdn.example.com/shorts/mina-rooftop.mp4",
            },
            previewDurationSeconds: 16,
          },
          main: {
            durationSeconds: 480,
            id: "main_mina_quiet_rooftop",
          },
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_fan_profile_library_001",
    });

    renderWithFanAuthDialog(
      await FanPage({
        searchParams: Promise.resolve({
          tab: "library",
        }),
      }),
    );

    expect(fetchFanProfilePinnedShortsPage).not.toHaveBeenCalled();
    expect(getViewerProfile).toHaveBeenCalledWith({
      sessionToken: "valid-session",
    });
    expect(fetchFanProfileLibraryPage).toHaveBeenCalledWith({
      sessionToken: "valid-session",
    });
    expect(screen.getByRole("link", { name: "Library" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("Alex_Fan")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Mina Rei quiet rooftop preview。" })).toHaveAttribute(
      "href",
      "/mains/main_mina_quiet_rooftop?fanTab=library&from=fan&fromShortId=short_mina_rooftop",
    );
  });

  it("opens the shared auth dialog when the overview api responds with auth_required", async () => {
    cookiesMock.mockResolvedValue({
      get: () => undefined,
    });
    vi.mocked(fetchFanProfileOverview).mockRejectedValue(
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

    renderWithFanAuthDialog(
      await FanPage({
        searchParams: Promise.resolve({
          tab: "library",
        }),
      }),
    );

    expect(await screen.findByRole("dialog", { name: "続けるには認証が必要です" })).toBeInTheDocument();
  });

  it("opens the shared auth dialog when the viewer profile api responds with auth_required", async () => {
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    vi.mocked(fetchFanProfileOverview).mockResolvedValue({
      counts: {
        following: 9,
        library: 7,
        pinnedShorts: 8,
      },
      title: "Archive from API",
    });
    vi.mocked(fetchFanProfilePinnedShortsPage).mockResolvedValue({
      items: [],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_fan_profile_pinned_shorts_004",
    });
    vi.mocked(getViewerProfile).mockRejectedValue(
      new ApiError("unauthorized", {
        code: "http",
        details: JSON.stringify({
          error: {
            code: "auth_required",
            message: "fan profile requires authentication",
          },
        }),
        status: 401,
      }),
    );

    renderWithFanAuthDialog(
      await FanPage({
        searchParams: Promise.resolve({}),
      }),
    );

    expect(await screen.findByRole("dialog", { name: "続けるには認証が必要です" })).toBeInTheDocument();
  });

  it("opens the shared auth dialog when the pinned shorts api responds with auth_required", async () => {
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    vi.mocked(fetchFanProfileOverview).mockResolvedValue({
      counts: {
        following: 9,
        library: 7,
        pinnedShorts: 8,
      },
      title: "Archive from API",
    });
    vi.mocked(fetchFanProfilePinnedShortsPage).mockRejectedValue(
      new ApiError("unauthorized", {
        code: "http",
        details: JSON.stringify({
          error: {
            code: "auth_required",
            message: "fan profile requires authentication",
          },
        }),
        status: 401,
      }),
    );

    renderWithFanAuthDialog(
      await FanPage({
        searchParams: Promise.resolve({}),
      }),
    );

    expect(await screen.findByRole("dialog", { name: "続けるには認証が必要です" })).toBeInTheDocument();
  });

  it("opens the shared auth dialog when the library api responds with auth_required", async () => {
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    vi.mocked(fetchFanProfileOverview).mockResolvedValue({
      counts: {
        following: 9,
        library: 7,
        pinnedShorts: 8,
      },
      title: "Archive from API",
    });
    vi.mocked(fetchFanProfileLibraryPage).mockRejectedValue(
      new ApiError("unauthorized", {
        code: "http",
        details: JSON.stringify({
          error: {
            code: "auth_required",
            message: "fan profile requires authentication",
          },
        }),
        status: 401,
      }),
    );

    renderWithFanAuthDialog(
      await FanPage({
        searchParams: Promise.resolve({
          tab: "library",
        }),
      }),
    );

    expect(await screen.findByRole("dialog", { name: "続けるには認証が必要です" })).toBeInTheDocument();
  });

  it("returns to the previous route when the auth-required fan dialog is dismissed", async () => {
    const user = userEvent.setup();

    window.history.pushState({}, "", "/");
    window.history.pushState({}, "", "/fan");

    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    vi.mocked(fetchFanProfileOverview).mockRejectedValue(
      new ApiError("unauthorized", {
        code: "http",
        details: JSON.stringify({
          error: {
            code: "auth_required",
            message: "fan profile requires authentication",
          },
        }),
        status: 401,
      }),
    );
    vi.mocked(fetchFanProfilePinnedShortsPage).mockResolvedValue({
      items: [],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_fan_profile_pinned_shorts_002",
    });

    renderWithFanAuthDialog(
      await FanPage({
        searchParams: Promise.resolve({}),
      }),
    );

    await user.click(await screen.findByRole("button", { name: "閉じる" }));

    await waitFor(() => {
      expect(mockedRouter.back).toHaveBeenCalledTimes(1);
    });
  });

  it("pushes home when the auth-required fan dialog is dismissed without in-app history", async () => {
    const user = userEvent.setup();
    const historyLengthSpy = vi.spyOn(window.history, "length", "get").mockReturnValue(1);

    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    vi.mocked(fetchFanProfileOverview).mockRejectedValue(
      new ApiError("unauthorized", {
        code: "http",
        details: JSON.stringify({
          error: {
            code: "auth_required",
            message: "fan profile requires authentication",
          },
        }),
        status: 401,
      }),
    );
    vi.mocked(fetchFanProfilePinnedShortsPage).mockResolvedValue({
      items: [],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_fan_profile_pinned_shorts_003",
    });

    renderWithFanAuthDialog(
      await FanPage({
        searchParams: Promise.resolve({}),
      }),
    );

    await user.click(await screen.findByRole("button", { name: "閉じる" }));

    await waitFor(() => {
      expect(mockedRouter.push).toHaveBeenCalledWith("/");
    });

    historyLengthSpy.mockRestore();
  });
});
