import userEvent from "@testing-library/user-event";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import {
  CurrentViewerProvider,
  ViewerSessionProvider,
} from "@/entities/viewer";
import { FanAuthDialogProvider } from "@/features/fan-auth";
import { createApiUrl } from "@/shared/api";
import { getClientEnv } from "@/shared/config";
import { getFeedSurfaceByTab } from "@/widgets/immersive-short-surface";

import { FeedReel } from "./feed-reel";

const push = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push,
  }),
}));

function renderWithViewerSession(
  ui: React.ReactElement,
  {
    currentViewer = null,
    hasSession,
  }: {
    currentViewer?: {
      activeMode: "creator" | "fan";
      canAccessCreatorMode: boolean;
      id: string;
    } | null;
    hasSession: boolean;
  },
) {
  return render(
    <ViewerSessionProvider hasSession={hasSession}>
      <CurrentViewerProvider currentViewer={currentViewer}>
        <FanAuthDialogProvider>
          {ui}
        </FanAuthDialogProvider>
      </CurrentViewerProvider>
    </ViewerSessionProvider>,
  );
}

function buildPublicFeedSurface(tab: Parameters<typeof getFeedSurfaceByTab>[0], shortId: `short_${string}`) {
  const surface = getFeedSurfaceByTab(tab);

  return {
    ...surface,
    short: {
      ...surface.short,
      id: shortId,
    },
    unlock: {
      ...surface.unlock,
      mainAccessEntry: {
        ...surface.unlock.mainAccessEntry,
        token: `disabled-${shortId}`,
      },
      short: {
        ...surface.unlock.short,
        id: shortId,
      },
    },
  };
}

describe("FeedReel", () => {
  beforeEach(() => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "http://localhost:8080");
  });

  afterEach(() => {
    push.mockReset();
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it("scrolls by one viewport when wheeling to the next short", () => {
    const firstSurface = getFeedSurfaceByTab("recommended");
    const secondSurface = {
      ...getFeedSurfaceByTab("following"),
      short: {
        ...getFeedSurfaceByTab("following").short,
        id: "short_second_surface",
      },
      unlock: {
        ...getFeedSurfaceByTab("following").unlock,
        mainAccessEntry: {
          ...getFeedSurfaceByTab("following").unlock.mainAccessEntry,
          token: "disabled-short_second_surface",
        },
        short: {
          ...getFeedSurfaceByTab("following").unlock.short,
          id: "short_second_surface",
        },
      },
    };

    const { container } = renderWithViewerSession(
      <FeedReel activeTab="recommended" surfaces={[firstSurface, secondSurface]} />,
      { hasSession: true },
    );

    const reel = container.firstElementChild;

    if (!(reel instanceof HTMLDivElement)) {
      throw new Error("feed reel container missing");
    }

    Object.defineProperty(reel, "clientHeight", {
      configurable: true,
      value: 640,
    });

    const scrollTo = vi.fn();
    reel.scrollTo = scrollTo;

    fireEvent.wheel(reel, { deltaY: 120 });

    expect(scrollTo).toHaveBeenCalledWith({
      behavior: "smooth",
      top: 640,
    });
    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
  });

  it("pins a feed short and updates the local viewer state", async () => {
    const user = userEvent.setup();
    const surface = buildPublicFeedSurface("following", "short_softlight");
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            viewer: {
              isPinned: true,
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_short_pin_put_success_001",
          },
        }),
        {
          headers: {
            "Content-Type": "application/json",
          },
          status: 200,
        },
      ),
    );
    vi.stubGlobal("fetch", fetcher);

    renderWithViewerSession(
      <FeedReel activeTab="following" surfaces={[surface]} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Pin short" }));

    await waitFor(() => {
      expect(fetcher).toHaveBeenCalledWith(
        createApiUrl(getClientEnv().NEXT_PUBLIC_API_BASE_URL, "/api/fan/shorts/short_softlight/pin"),
        {
          credentials: "include",
          headers: {
            Accept: "application/json",
          },
          method: "PUT",
        },
      );
    });

    expect(await screen.findByRole("button", { name: "Pinned short" })).toHaveAttribute("aria-pressed", "true");
  });

  it("syncs the pin state when refreshed feed data changes the same short", () => {
    const initialSurface = buildPublicFeedSurface("following", "short_softlight");
    const refreshedSurface = {
      ...initialSurface,
      viewer: {
        ...initialSurface.viewer,
        isPinned: true,
      },
    };

    const { rerender } = renderWithViewerSession(
      <FeedReel activeTab="following" surfaces={[initialSurface]} />,
      { hasSession: true },
    );

    expect(screen.getByRole("button", { name: "Pin short" })).toHaveAttribute("aria-pressed", "false");

    rerender(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider currentViewer={null}>
          <FanAuthDialogProvider>
            <FeedReel activeTab="following" surfaces={[refreshedSurface]} />
          </FanAuthDialogProvider>
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    expect(screen.getByRole("button", { name: "Pinned short" })).toHaveAttribute("aria-pressed", "true");
  });

  it("keeps the successful local pin state until refreshed feed data catches up", async () => {
    const user = userEvent.setup();
    const initialSurface = buildPublicFeedSurface("following", "short_softlight");
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            viewer: {
              isPinned: true,
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_short_pin_put_success_003",
          },
        }),
        {
          headers: {
            "Content-Type": "application/json",
          },
          status: 200,
        },
      ),
    );
    vi.stubGlobal("fetch", fetcher);

    const { rerender } = renderWithViewerSession(
      <FeedReel activeTab="following" surfaces={[initialSurface]} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Pin short" }));

    expect(await screen.findByRole("button", { name: "Pinned short" })).toHaveAttribute("aria-pressed", "true");

    rerender(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider currentViewer={null}>
          <FanAuthDialogProvider>
            <FeedReel activeTab="following" surfaces={[initialSurface]} />
          </FanAuthDialogProvider>
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    expect(screen.getByRole("button", { name: "Pinned short" })).toHaveAttribute("aria-pressed", "true");

    rerender(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider currentViewer={null}>
          <FanAuthDialogProvider>
            <FeedReel
              activeTab="following"
              surfaces={[
                {
                  ...initialSurface,
                  viewer: {
                    ...initialSurface.viewer,
                    isPinned: true,
                  },
                },
              ]}
            />
          </FanAuthDialogProvider>
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    expect(screen.getByRole("button", { name: "Pinned short" })).toHaveAttribute("aria-pressed", "true");
  });

  it("drops stale local pin state after the short leaves the feed and re-enters", async () => {
    const user = userEvent.setup();
    const initialSurface = buildPublicFeedSurface("following", "short_softlight");
    const replacementSurface = {
      ...buildPublicFeedSurface("recommended", "short_rooftop"),
      viewer: {
        ...buildPublicFeedSurface("recommended", "short_rooftop").viewer,
        isPinned: false,
      },
    };
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            viewer: {
              isPinned: true,
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_short_pin_put_success_004",
          },
        }),
        {
          headers: {
            "Content-Type": "application/json",
          },
          status: 200,
        },
      ),
    );
    vi.stubGlobal("fetch", fetcher);

    const { rerender } = renderWithViewerSession(
      <FeedReel activeTab="following" surfaces={[initialSurface]} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Pin short" }));
    expect(await screen.findByRole("button", { name: "Pinned short" })).toHaveAttribute("aria-pressed", "true");

    rerender(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider currentViewer={null}>
          <FanAuthDialogProvider>
            <FeedReel activeTab="following" surfaces={[replacementSurface]} />
          </FanAuthDialogProvider>
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    expect(screen.getByRole("button", { name: "Pin short" })).toHaveAttribute("aria-pressed", "false");

    rerender(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider currentViewer={null}>
          <FanAuthDialogProvider>
            <FeedReel activeTab="following" surfaces={[initialSurface]} />
          </FanAuthDialogProvider>
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    expect(await screen.findByRole("button", { name: "Pin short" })).toHaveAttribute("aria-pressed", "false");
  });

  it("opens the shared auth dialog when an unauthenticated viewer pins a short", async () => {
    const user = userEvent.setup();
    const surface = buildPublicFeedSurface("following", "short_softlight");

    renderWithViewerSession(
      <FeedReel activeTab="following" surfaces={[surface]} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: "Pin short" }));

    expect(screen.getByRole("dialog", { name: "続けるにはログインが必要です" })).toBeInTheDocument();
  });

  it("opens the shared auth dialog when the backend returns auth_required for pin", async () => {
    const user = userEvent.setup();
    const surface = buildPublicFeedSurface("following", "short_softlight");
    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(
          JSON.stringify({
            data: null,
            error: {
              code: "auth_required",
              message: "short pin requires authentication",
            },
            meta: {
              page: null,
              requestId: "req_short_pin_put_auth_required_001",
            },
          }),
          {
            headers: {
              "Content-Type": "application/json",
            },
            status: 401,
          },
        ),
      ),
    );

    renderWithViewerSession(
      <FeedReel activeTab="following" surfaces={[surface]} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Pin short" }));

    await waitFor(() => {
      expect(screen.getByRole("dialog", { name: "続けるにはログインが必要です" })).toBeInTheDocument();
    });
  });

  it("shows an inline alert when pin mutation fails", async () => {
    const user = userEvent.setup();
    const surface = buildPublicFeedSurface("following", "short_softlight");
    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(
          JSON.stringify({
            data: null,
            error: {
              code: "internal_error",
              message: "unexpected failure",
            },
            meta: {
              page: null,
              requestId: "req_short_pin_put_internal_error_001",
            },
          }),
          {
            headers: {
              "Content-Type": "application/json",
            },
            status: 500,
          },
        ),
      ),
    );

    renderWithViewerSession(
      <FeedReel activeTab="following" surfaces={[surface]} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Pin short" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "pin 状態を更新できませんでした。少し時間を置いてから再度お試しください。",
    );
  });

  it("shows a generic retry alert when the error response cannot be parsed", async () => {
    const user = userEvent.setup();
    const surface = buildPublicFeedSurface("following", "short_softlight");
    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response("upstream unavailable", {
          headers: {
            "Content-Type": "text/plain",
          },
          status: 500,
        }),
      ),
    );

    renderWithViewerSession(
      <FeedReel activeTab="following" surfaces={[surface]} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Pin short" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "pin 状態を更新できませんでした。少し時間を置いてから再度お試しください。",
    );
  });

  it("ignores repeated pin taps while the request is pending", async () => {
    const user = userEvent.setup();
    const surface = buildPublicFeedSurface("following", "short_softlight");
    const pendingResponse = {
      resolve: null as ((value: Response) => void) | null,
    };
    const fetcher = vi.fn<typeof fetch>().mockImplementation(
      () =>
        new Promise<Response>((resolve) => {
          pendingResponse.resolve = resolve;
        }),
    );
    vi.stubGlobal("fetch", fetcher);

    renderWithViewerSession(
      <FeedReel activeTab="following" surfaces={[surface]} />,
      { hasSession: true },
    );

    const pinButton = screen.getByRole("button", { name: "Pin short" });

    await user.click(pinButton);
    await user.click(pinButton);

    expect(fetcher).toHaveBeenCalledTimes(1);

    if (pendingResponse.resolve === null) {
      throw new Error("pending short pin response was not captured");
    }

    pendingResponse.resolve(
      new Response(
        JSON.stringify({
          data: {
            viewer: {
              isPinned: true,
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_short_pin_put_success_002",
          },
        }),
        {
          headers: {
            "Content-Type": "application/json",
          },
          status: 200,
        },
      ),
    );

    expect(await screen.findByRole("button", { name: "Pinned short" })).toBeInTheDocument();
  });
});
