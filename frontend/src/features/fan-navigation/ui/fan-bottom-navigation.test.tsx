import { render, screen } from "@testing-library/react";
import { usePathname } from "next/navigation";

import { FanBottomNavigation, resolveActiveFanNavigation } from "@/features/fan-navigation";

vi.mock("next/navigation", () => ({
  usePathname: vi.fn(),
}));

describe("fan navigation", () => {
  it("resolves the active tab from pathname", () => {
    expect(resolveActiveFanNavigation("/")).toBe("feed");
    expect(resolveActiveFanNavigation("/search")).toBe("search");
    expect(resolveActiveFanNavigation("/fan")).toBe("fan");
    expect(resolveActiveFanNavigation("/shorts/rooftop")).toBe("feed");
  });

  it("renders the bottom navigation with the current page", () => {
    vi.mocked(usePathname).mockReturnValue("/search");

    render(<FanBottomNavigation />);

    expect(screen.getByRole("link", { name: /search/i })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: /feed/i })).toHaveAttribute("href", "/");
  });
});
