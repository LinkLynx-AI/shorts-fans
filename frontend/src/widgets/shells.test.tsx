import { render, screen } from "@testing-library/react";

import { getFanHubState } from "@/entities/fan-profile";
import {
  CurrentViewerProvider,
  ViewerSessionProvider,
} from "@/entities/viewer";
import { FanAuthDialogProvider } from "@/features/fan-auth";
import type { CreatorSearchState } from "@/features/creator-search";
import { DetailShell } from "@/widgets/detail-shell";
import { FanHubShell } from "@/widgets/fan-hub-shell";
import { FeedShell, getFollowingFeedShellState, getMockFeedShellState } from "@/widgets/feed-shell";
import { SearchShell } from "@/widgets/search-shell";

const mockedRouter = vi.hoisted(() => ({
  back: vi.fn(),
  forward: vi.fn(),
  prefetch: vi.fn(),
  push: vi.fn(),
  refresh: vi.fn(),
  replace: vi.fn(),
}));

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");

  return {
    ...actual,
    useRouter: () => mockedRouter,
  };
});

function renderFanHubShell(activeTab: "following" | "library" | "pinned") {
  return render(
    <ViewerSessionProvider hasSession>
      <CurrentViewerProvider
        currentViewer={{
          activeMode: "fan",
          canAccessCreatorMode: false,
          id: "viewer_123",
        }}
      >
        <FanAuthDialogProvider>
          <FanHubShell
            headerProfile={{
              avatarUrl: null,
              displayName: "Alex_Fan",
              handle: "@alex_f",
            }}
            state={getFanHubState(activeTab)}
          />
        </FanAuthDialogProvider>
      </CurrentViewerProvider>
    </ViewerSessionProvider>,
  );
}

function renderFeedShell(state: ReturnType<typeof getMockFeedShellState>) {
  return render(
    <ViewerSessionProvider hasSession>
      <CurrentViewerProvider currentViewer={null}>
        <FanAuthDialogProvider>
          <FeedShell state={state} />
        </FanAuthDialogProvider>
      </CurrentViewerProvider>
    </ViewerSessionProvider>,
  );
}

describe("widgets", () => {
  it("renders the feed shell", () => {
    renderFeedShell(getMockFeedShellState("recommended"));

    expect(screen.getByRole("link", { name: /For You/i })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Mina Rei" })).toBeInTheDocument();
    expect(screen.getByText("@minarei")).toBeInTheDocument();
    expect(screen.getByText("quiet rooftop preview.")).toBeInTheDocument();
  });

  it("renders following empty state", () => {
    render(<FeedShell state={getFollowingFeedShellState("empty")} />);

    expect(screen.getByRole("link", { name: /Following/i })).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("フォロー中の creator はまだいません")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "creatorを探す" })).toHaveAttribute("href", "/search");
  });

  it("renders following auth-required state", () => {
    render(<FeedShell state={getFollowingFeedShellState("auth_required")} />);

    expect(screen.getByRole("link", { name: /Following/i })).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("フォロー中を見るにはログインが必要です")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ログインへ進む" })).toHaveAttribute("href", "/login");
  });

  it("renders the search UI and keeps query text", () => {
    const readyState: CreatorSearchState = {
      items: [
        {
          avatar: null,
          bio: "soft light と close framing の short を中心に更新中。",
          displayName: "Aoi N",
          handle: "@aoina",
          id: "creator_aoi_n",
        },
      ],
      kind: "ready",
      query: "",
    };
    const { rerender } = render(<SearchShell initialState={readyState} query="" />);

    expect(screen.getByRole("searchbox")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Aoi N/i })).toBeInTheDocument();

    rerender(
      <SearchShell
        initialState={{
          items: [
            {
              avatar: null,
              bio: "quiet rooftop と hotel light の preview を軸に投稿。",
              displayName: "Mina Rei",
              handle: "@minarei",
              id: "creator_mina_rei",
            },
          ],
          kind: "ready",
          query: "mina",
        }}
        query="mina"
      />,
    );

    expect(screen.getByDisplayValue("mina")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "href",
      "/creators/creator_mina_rei?from=search&q=mina",
    );
  });

  it("renders fan hub content and active tabs", () => {
    const { rerender } = renderFanHubShell("library");

    expect(screen.getByRole("heading", { name: "Profile" })).toBeInTheDocument();
    expect(screen.getByText("Alex_Fan")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Following" })).toHaveAttribute("href", "/fan?tab=following");
    expect(screen.getByRole("link", { name: "Library" })).toHaveAttribute("aria-current", "page");

    rerender(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider
          currentViewer={{
            activeMode: "fan",
            canAccessCreatorMode: false,
            id: "viewer_123",
          }}
        >
          <FanAuthDialogProvider>
            <FanHubShell
              headerProfile={{
                avatarUrl: null,
                displayName: "Alex_Fan",
                handle: "@alex_f",
              }}
              state={getFanHubState("pinned")}
            />
          </FanAuthDialogProvider>
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    expect(screen.getByRole("link", { name: "Pinned Shorts" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: /雨上がりの balcony preview/i })).toHaveAttribute(
      "href",
      "/shorts/afterrain?fanTab=pinned&from=fan",
    );

    rerender(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider
          currentViewer={{
            activeMode: "fan",
            canAccessCreatorMode: false,
            id: "viewer_123",
          }}
        >
          <FanAuthDialogProvider>
            <FanHubShell
              headerProfile={{
                avatarUrl: null,
                displayName: "Alex_Fan",
                handle: "@alex_f",
              }}
              state={getFanHubState("following")}
            />
          </FanAuthDialogProvider>
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    expect(screen.getByRole("link", { name: "Following" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByPlaceholderText("検索")).toBeInTheDocument();
    expect(screen.getByText("3 creators")).toBeInTheDocument();
  });

  it("renders the surface detail shell", () => {
    render(
      <DetailShell backHref="/" variant="surface">
        <div>creator profile shell</div>
      </DetailShell>,
    );

    expect(screen.getByRole("link", { name: /back/i })).toHaveAttribute("href", "/");
    expect(screen.getByText("creator profile shell")).toBeInTheDocument();
  });

  it("renders the immersive detail shell", () => {
    render(
      <DetailShell backHref="/" style={{ "--short-bg-start": "#a7e8ff" } as React.CSSProperties} variant="immersive">
        <div>hero slot</div>
        <div>cta slot</div>
      </DetailShell>,
    );

    expect(screen.getByText("hero slot")).toBeInTheDocument();
    expect(screen.getByText("cta slot")).toBeInTheDocument();
  });
});
