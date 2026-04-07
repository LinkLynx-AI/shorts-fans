import userEvent from "@testing-library/user-event";
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

  it("opens a temporary auth dialog instead of routing to /login for unauthenticated profile opens", async () => {
    vi.mocked(usePathname).mockReturnValue("/");
    const user = userEvent.setup();

    render(
      <ViewerSessionProvider hasSession={false}>
        <FanBottomNavigation />
      </ViewerSessionProvider>,
    );

    expect(screen.getByRole("link", { name: "マイ" })).toHaveAttribute("href", "/fan");

    await user.click(screen.getByRole("link", { name: "マイ" }));

    expect(screen.getByRole("dialog", { name: "続けるにはログインが必要です" })).toBeInTheDocument();
  });
});
