import userEvent from "@testing-library/user-event";
import { render, screen } from "@testing-library/react";
import { usePathname } from "next/navigation";

import {
  CurrentViewerProvider,
  ViewerSessionProvider,
} from "@/entities/viewer";
import { FanAuthDialogProvider } from "@/features/fan-auth";
import { FanBottomNavigation, resolveActiveFanNavigation } from "@/features/fan-navigation";

const mockedRouter = vi.hoisted(() => ({
  back: vi.fn(),
  forward: vi.fn(),
  prefetch: vi.fn(),
  push: vi.fn(),
  refresh: vi.fn(),
  replace: vi.fn(),
}));

vi.mock("next/navigation", async () => {
  return {
    usePathname: vi.fn(),
    useRouter: () => mockedRouter,
  };
});

describe("fan navigation", () => {
  function renderNavigation(hasSession: boolean) {
    return render(
      <ViewerSessionProvider hasSession={hasSession}>
        <CurrentViewerProvider currentViewer={null}>
          <FanAuthDialogProvider>
          <FanBottomNavigation />
          </FanAuthDialogProvider>
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );
  }

  it("resolves the active tab from pathname", () => {
    expect(resolveActiveFanNavigation("/")).toBe("feed");
    expect(resolveActiveFanNavigation("/search")).toBe("search");
    expect(resolveActiveFanNavigation("/fan")).toBe("fan");
    expect(resolveActiveFanNavigation("/shorts/rooftop")).toBe("feed");
  });

  it("renders the bottom navigation with the current page", () => {
    vi.mocked(usePathname).mockReturnValue("/search");

    renderNavigation(true);

    expect(screen.getByRole("link", { name: "検索" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: "フィード" })).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: "マイ" })).toHaveAttribute("href", "/fan");
  });

  it("does not open the shared auth dialog before /fan responds with auth_required", async () => {
    vi.mocked(usePathname).mockReturnValue("/");
    const user = userEvent.setup();

    renderNavigation(false);

    expect(screen.getByRole("link", { name: "マイ" })).toHaveAttribute("href", "/fan");

    await user.click(screen.getByRole("link", { name: "マイ" }));

    expect(screen.queryByRole("dialog", { name: "続けるには認証が必要です" })).not.toBeInTheDocument();
  });
});
