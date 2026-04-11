import userEvent from "@testing-library/user-event";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { ViewerSessionProvider } from "@/entities/viewer";
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
  { hasSession }: { hasSession: boolean },
) {
  return render(
    <ViewerSessionProvider hasSession={hasSession}>
      {ui}
    </ViewerSessionProvider>,
  );
}

describe("FeedReel", () => {
  afterEach(() => {
    push.mockReset();
    vi.unstubAllGlobals();
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
      <FeedReel activeTab="following" surfaces={[getFeedSurfaceByTab("following")]} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Pin short" }));

    await waitFor(() => {
      expect(fetcher).toHaveBeenCalledWith(
        new URL("http://localhost:8080/api/fan/shorts/softlight/pin"),
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

  it("redirects unauthenticated viewers to login when pin is tapped", async () => {
    const user = userEvent.setup();

    renderWithViewerSession(
      <FeedReel activeTab="following" surfaces={[getFeedSurfaceByTab("following")]} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: "Pin short" }));

    expect(push).toHaveBeenCalledWith("/login");
  });

  it("redirects to login when the backend returns auth_required for pin", async () => {
    const user = userEvent.setup();
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
      <FeedReel activeTab="following" surfaces={[getFeedSurfaceByTab("following")]} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Pin short" }));

    await waitFor(() => {
      expect(push).toHaveBeenCalledWith("/login");
    });
  });

  it("shows an inline alert when pin mutation fails", async () => {
    const user = userEvent.setup();
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
      <FeedReel activeTab="following" surfaces={[getFeedSurfaceByTab("following")]} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Pin short" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "pin 状態を更新できませんでした。少し時間を置いてから再度お試しください。",
    );
  });

  it("ignores repeated pin taps while the request is pending", async () => {
    const user = userEvent.setup();
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
      <FeedReel activeTab="following" surfaces={[getFeedSurfaceByTab("following")]} />,
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
