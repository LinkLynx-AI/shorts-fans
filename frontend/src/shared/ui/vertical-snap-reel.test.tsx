import { fireEvent, render, screen } from "@testing-library/react";

import { VerticalSnapReel } from "./vertical-snap-reel";

describe("VerticalSnapReel", () => {
  it("marks the configured initial item as active", () => {
    render(
      <VerticalSnapReel
        getKey={(item) => item}
        initialIndex={1}
        items={["first", "second", "third"]}
        renderItem={(item, { isActive }) => (
          <div data-testid={item}>{isActive ? "active" : "inactive"}</div>
        )}
      />,
    );

    expect(screen.getByTestId("first")).toHaveTextContent("inactive");
    expect(screen.getByTestId("second")).toHaveTextContent("active");
    expect(screen.getByTestId("third")).toHaveTextContent("inactive");
  });

  it("scrolls by one viewport when wheeling to the next item", () => {
    const { container } = render(
      <VerticalSnapReel
        getKey={(item) => item}
        items={["first", "second"]}
        renderItem={(item) => <div>{item}</div>}
      />,
    );

    const reel = container.firstElementChild;

    if (!(reel instanceof HTMLDivElement)) {
      throw new Error("vertical snap reel container missing");
    }

    Object.defineProperty(reel, "clientHeight", {
      configurable: true,
      value: 720,
    });

    const scrollTo = vi.fn();
    reel.scrollTo = scrollTo;

    fireEvent.wheel(reel, { deltaY: 120 });

    expect(scrollTo).toHaveBeenCalledWith({
      behavior: "smooth",
      top: 720,
    });
  });

  it("sizes each item to the reel container instead of the raw viewport", () => {
    const { container } = render(
      <VerticalSnapReel
        getKey={(item) => item}
        items={["first", "second"]}
        renderItem={(item) => <div>{item}</div>}
      />,
    );

    const reel = container.firstElementChild;

    if (!(reel instanceof HTMLDivElement)) {
      throw new Error("vertical snap reel container missing");
    }

    expect(reel).toHaveClass("h-full");

    const items = reel.querySelectorAll(":scope > div");

    expect(items).toHaveLength(2);
    expect(items[0]).toHaveClass("h-full");
    expect(items[0]).toHaveClass("min-h-full");
  });
});
