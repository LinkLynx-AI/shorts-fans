import userEvent from "@testing-library/user-event";
import { render, screen } from "@testing-library/react";

import { getMainPlaybackSurfaceById } from "@/widgets/main-playback-surface";

import { MainPlaybackSurface } from "./main-playback-surface";

const back = vi.fn();
const push = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    back,
    push,
  }),
}));

describe("MainPlaybackSurface", () => {
  afterEach(() => {
    back.mockReset();
    push.mockReset();
  });

  it("renders playback status and falls back to the provided short when there is no browser history", async () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "purchased");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    render(<MainPlaybackSurface fallbackHref="/shorts/softlight" surface={surface} />);

    expect(screen.getByRole("button", { name: "Back" })).toBeInTheDocument();
    expect(screen.getByText("Playing main")).toBeInTheDocument();
    expect(screen.getByText("resume without another confirmation")).toBeInTheDocument();
    expect(screen.getByText("3:18")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pin short" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Aoi N/i })).toHaveAttribute("href", "/creators/aoi");
    expect(screen.getByText("soft light preview の続き。")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Back" }));

    expect(push).toHaveBeenCalledWith("/shorts/softlight");
    expect(back).not.toHaveBeenCalled();
  });

  it("uses router.back when a previous history entry exists", async () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "balcony", "owner");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    window.history.pushState({}, "", "/before");
    window.history.pushState({}, "", "/mains/main_aoi_blue_balcony");

    const user = userEvent.setup();

    render(<MainPlaybackSurface fallbackHref="/shorts/balcony" surface={surface} />);

    expect(screen.getAllByText("Owner preview")).toHaveLength(2);
    expect(screen.getByText("purchase confirmation is skipped for your own main")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pinned short" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Back" }));

    expect(back).toHaveBeenCalled();
    expect(push).not.toHaveBeenCalled();
  });
});
