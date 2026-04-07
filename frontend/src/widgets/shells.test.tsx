import { render, screen } from "@testing-library/react";

import { getFanHubState } from "@/entities/fan-profile";
import type { CreatorSearchState } from "@/features/creator-search";
import { DetailShell } from "@/widgets/detail-shell";
import { FanHubShell } from "@/widgets/fan-hub-shell";
import { FeedShell, getFollowingFeedShellState, getMockFeedShellState } from "@/widgets/feed-shell";
import { SearchShell } from "@/widgets/search-shell";

describe("widgets", () => {
  it("renders the feed shell", () => {
    render(<FeedShell state={getMockFeedShellState("recommended")} />);

    expect(screen.getByRole("link", { name: /おすすめ/i })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
    expect(screen.getByText("Mina Rei")).toBeInTheDocument();
    expect(screen.getByText("quiet rooftop preview.")).toBeInTheDocument();
  });

  it("renders following empty state", () => {
    render(<FeedShell state={getFollowingFeedShellState("empty")} />);

    expect(screen.getByRole("link", { name: /フォロー中/i })).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("フォロー中の creator はまだいません")).toBeInTheDocument();
  });

  it("renders following auth-required state", () => {
    render(<FeedShell state={getFollowingFeedShellState("auth_required")} />);

    expect(screen.getByRole("link", { name: /フォロー中/i })).toHaveAttribute("aria-current", "page");
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
    const { rerender } = render(<FanHubShell state={getFanHubState("library")} />);

    expect(screen.getByRole("heading", { name: "My archive" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Back" })).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: "Following" })).toHaveAttribute("href", "/fan/following");
    expect(screen.getByRole("link", { name: "Library" })).toHaveAttribute("aria-current", "page");

    rerender(<FanHubShell state={getFanHubState("pinned")} />);

    expect(screen.getByRole("link", { name: "Pinned" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: /after rain preview/i })).toHaveAttribute("href", "/shorts/afterrain");
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
