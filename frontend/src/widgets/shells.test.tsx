import { render, screen } from "@testing-library/react";

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
  });

  it("renders the search structure and keeps query text", () => {
    const { rerender } = render(<SearchShell query="" />);

    expect(screen.getByRole("heading", { name: "Creator search structure" })).toBeInTheDocument();

    rerender(<SearchShell query="mina" />);

    expect(screen.getByDisplayValue("mina")).toBeInTheDocument();
    expect(screen.getByText('query "mina" を保持できる route だけを先に定義しています。')).toBeInTheDocument();
  });

  it("renders fan hub structure and active tab copy", () => {
    const { rerender } = render(<FanHubShell activeTab="library" />);

    expect(screen.getByRole("heading", { name: "Fan hub structure" })).toBeInTheDocument();
    expect(screen.getByText("Library panel の中身を後続 task で差し込む")).toBeInTheDocument();

    rerender(<FanHubShell activeTab="pinned" />);

    expect(screen.getByText("Pinned panel の中身を後続 task で差し込む")).toBeInTheDocument();
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
