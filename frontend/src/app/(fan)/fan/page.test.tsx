import { render, screen } from "@testing-library/react";

import { CurrentViewerProvider, ViewerSessionProvider } from "@/entities/viewer";
import {
  fetchFanProfileOverview,
  fetchFanProfilePinnedShortsPage,
} from "@/entities/fan-profile";
import { FanAuthDialogProvider } from "@/features/fan-auth";
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
    fetchFanProfileOverview: vi.fn(),
    fetchFanProfilePinnedShortsPage: vi.fn(),
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
    vi.mocked(fetchFanProfileOverview).mockReset();
    vi.mocked(fetchFanProfilePinnedShortsPage).mockReset();
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
    expect(fetchFanProfilePinnedShortsPage).toHaveBeenCalledWith({
      sessionToken: "valid-session",
    });
    expect(screen.getByRole("heading", { name: "Archive from API" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Pinned" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: "Sora Vale after rain preview" })).toHaveAttribute(
      "href",
      "/shorts/short_sora_afterrain?fanTab=pinned&from=fan",
    );
    expect(screen.getByText("9")).toBeInTheDocument();
  });

  it("does not fetch pinned shorts when the library tab is active", async () => {
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

    renderWithFanAuthDialog(
      await FanPage({
        searchParams: Promise.resolve({
          tab: "library",
        }),
      }),
    );

    expect(fetchFanProfilePinnedShortsPage).not.toHaveBeenCalled();
    expect(screen.getByRole("link", { name: "Library" })).toHaveAttribute("aria-current", "page");
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

    expect(await screen.findByRole("dialog", { name: "続けるにはログインが必要です" })).toBeInTheDocument();
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

    expect(await screen.findByRole("dialog", { name: "続けるにはログインが必要です" })).toBeInTheDocument();
  });
});
