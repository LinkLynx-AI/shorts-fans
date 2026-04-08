import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
} from "@testing-library/react";

import { getFanHubState } from "@/entities/fan-profile";
import { CurrentViewerProvider } from "@/entities/viewer";

import { FanHubShell } from "./fan-hub-shell";

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

describe("FanHubShell account menu", () => {
  it("shows the registration entry for viewers without creator access", async () => {
    const user = userEvent.setup();

    render(
      <CurrentViewerProvider
        currentViewer={{
          activeMode: "fan",
          canAccessCreatorMode: false,
          id: "viewer_123",
        }}
      >
        <FanHubShell state={getFanHubState("library")} />
      </CurrentViewerProvider>,
    );

    await user.click(screen.getByRole("button", { name: "Settings" }));

    expect(screen.getByRole("link", { name: "Creator登録を始める" })).toHaveAttribute(
      "href",
      "/fan/creator/register",
    );
  });

  it("shows the creator switch entry for creator-capable viewers", async () => {
    const user = userEvent.setup();

    render(
      <CurrentViewerProvider
        currentViewer={{
          activeMode: "fan",
          canAccessCreatorMode: true,
          id: "viewer_123",
        }}
      >
        <FanHubShell state={getFanHubState("library")} />
      </CurrentViewerProvider>,
    );

    await user.click(screen.getByRole("button", { name: "Settings" }));

    expect(screen.getByRole("button", { name: "Creator mode に切り替え" })).toBeInTheDocument();
  });
});
