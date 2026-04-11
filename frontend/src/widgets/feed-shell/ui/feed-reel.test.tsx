import { fireEvent, render, screen } from "@testing-library/react";

import { getFeedSurfaceByTab } from "@/widgets/immersive-short-surface";

import { FeedReel } from "./feed-reel";

describe("FeedReel", () => {
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

    const { container } = render(
      <FeedReel activeTab="recommended" surfaces={[firstSurface, secondSurface]} />,
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
});
