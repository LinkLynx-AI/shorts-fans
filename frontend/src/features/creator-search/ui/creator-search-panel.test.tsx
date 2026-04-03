import { act, fireEvent, render, screen } from "@testing-library/react";

import { CreatorSearchPanel } from "@/features/creator-search";

describe("CreatorSearchPanel", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("renders the initial query and recent creators", () => {
    render(<CreatorSearchPanel initialQuery="mina" />);

    expect(screen.getByDisplayValue("mina")).toBeInTheDocument();
    expect(screen.queryByText("最近")).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "href",
      "/creators/mina?from=search&q=mina",
    );
    expect(screen.queryByText("open")).not.toBeInTheDocument();
  });

  it("applies the filter after a short delay", () => {
    render(<CreatorSearchPanel initialQuery="" />);

    expect(screen.getByText("最近")).toBeInTheDocument();

    fireEvent.change(screen.getByRole("searchbox"), {
      target: { value: "sora" },
    });

    expect(screen.queryByText("最近")).not.toBeInTheDocument();
    expect(screen.queryByText("Sora Vale")).not.toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(250);
    });

    expect(screen.getByText("Sora Vale")).toBeInTheDocument();
    expect(screen.queryByText("Aoi N")).not.toBeInTheDocument();
  });
});
