import { render, screen } from "@testing-library/react";

import { CurrentViewerProvider, ViewerSessionProvider } from "@/entities/viewer";
import { fetchFanProfileOverview } from "@/entities/fan-profile";
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
  });

  it("fetches overview data with the viewer session and renders the hub", async () => {
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

    renderWithFanAuthDialog(
      await FanPage({
        searchParams: Promise.resolve({
          tab: "library",
        }),
      }),
    );

    expect(fetchFanProfileOverview).toHaveBeenCalledWith({
      sessionToken: "valid-session",
    });
    expect(screen.getByRole("heading", { name: "Archive from API" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Library" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("9")).toBeInTheDocument();
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
        searchParams: Promise.resolve({}),
      }),
    );

    expect(await screen.findByRole("dialog", { name: "続けるにはログインが必要です" })).toBeInTheDocument();
  });
});
