import { render, screen } from "@testing-library/react";
import { usePathname } from "next/navigation";

import { ViewerSessionProvider } from "@/entities/viewer";
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

    render(
      <ViewerSessionProvider hasSession>
        <FanBottomNavigation />
      </ViewerSessionProvider>,
    );

    expect(screen.getByRole("link", { name: "検索" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: "フィード" })).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: "マイ" })).toHaveAttribute("href", "/fan");
  });

  it("routes the profile tab to login while the backend does not require auth yet", () => {
    vi.mocked(usePathname).mockReturnValue("/");

    render(
      <ViewerSessionProvider hasSession={false}>
        <FanBottomNavigation />
      </ViewerSessionProvider>,
    );

    expect(screen.getByRole("link", { name: "マイ" })).toHaveAttribute("href", "/login");
  });
});
