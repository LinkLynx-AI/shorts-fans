import userEvent from "@testing-library/user-event";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { getMainPlaybackSurfaceById } from "@/widgets/main-playback-surface";

import { MainPlaybackSurface } from "./main-playback-surface";

const back = vi.fn();
const push = vi.fn();
const play = vi.fn<() => Promise<void>>();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    back,
    push,
  }),
}));

describe("MainPlaybackSurface", () => {
  beforeEach(() => {
    play.mockReset();
    play.mockResolvedValue(undefined);
    vi.spyOn(HTMLMediaElement.prototype, "play").mockImplementation(play);
  });

  afterEach(() => {
    back.mockReset();
    push.mockReset();
    vi.restoreAllMocks();
  });

  it("renders playback status and falls back to the provided short when there is no browser history", async () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    render(<MainPlaybackSurface fallbackHref="/shorts/softlight" surface={surface} />);

    const video = screen.getByLabelText("Main playback video");

    expect(screen.getByRole("button", { name: "Back" })).toBeInTheDocument();
    expect(screen.getByText("Playing main")).toBeInTheDocument();
    expect(screen.getByText("resume without another unlock step")).toBeInTheDocument();
    expect(screen.getByText("3:18")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pin short" })).toBeInTheDocument();
    expect(video).toHaveAttribute("controls");
    expect(video).toHaveAttribute("playsinline");
    expect(video).toHaveAttribute("poster", surface.main.media.posterUrl ?? undefined);
    expect(video).toHaveAttribute("src", surface.main.media.url);
    expect(video.nextElementSibling).toHaveClass("pointer-events-none");
    expect(video.nextElementSibling?.nextElementSibling).toHaveClass("pointer-events-none");
    expect(screen.getByRole("link", { name: /Aoi N/i })).toHaveAttribute(
      "href",
      "/creators/creator_aoi_n",
    );
    expect(screen.getByText("soft light の preview の続き。")).toBeInTheDocument();

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
    expect(screen.getByText("unlock confirmation is skipped for your own main")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pinned short" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Back" }));

    expect(back).toHaveBeenCalled();
    expect(push).not.toHaveBeenCalled();
  });

  it("applies the resume position to actual playback after metadata loads", () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    render(<MainPlaybackSurface fallbackHref="/shorts/softlight" surface={surface} />);

    const video = screen.getByLabelText("Main playback video") as HTMLVideoElement;

    expect(video.currentTime).toBe(0);

    fireEvent(video, new Event("loadedmetadata"));

    expect(video.currentTime).toBe(surface.resumePositionSeconds);
  });

  it("falls back to muted autoplay when the first playback attempt is rejected", async () => {
    play.mockRejectedValueOnce(new Error("autoplay blocked")).mockResolvedValueOnce(undefined);

    const surface = getMainPlaybackSurfaceById("main_sora_after_rain", "afterrain", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    render(<MainPlaybackSurface fallbackHref="/shorts/afterrain" surface={surface} />);

    const video = screen.getByLabelText("Main playback video") as HTMLVideoElement;

    fireEvent(video, new Event("loadedmetadata"));

    await waitFor(() => {
      expect(play).toHaveBeenCalledTimes(2);
    });

    expect(video.muted).toBe(true);
  });

  it("reapplies the resume position when the playback URL changes for the same media", () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    const { rerender } = render(<MainPlaybackSurface fallbackHref="/shorts/softlight" surface={surface} />);

    const video = screen.getByLabelText("Main playback video") as HTMLVideoElement;

    fireEvent(video, new Event("loadedmetadata"));
    expect(video.currentTime).toBe(surface.resumePositionSeconds);

    video.currentTime = 0;

    rerender(
      <MainPlaybackSurface
        fallbackHref="/shorts/softlight"
        surface={{
          ...surface,
          main: {
            ...surface.main,
            media: {
              ...surface.main.media,
              url: "https://cdn.example.com/mains/aoi-blue-balcony-refreshed.mp4",
            },
          },
        }}
      />,
    );

    fireEvent(screen.getByLabelText("Main playback video"), new Event("loadedmetadata"));

    expect((screen.getByLabelText("Main playback video") as HTMLVideoElement).currentTime).toBe(
      surface.resumePositionSeconds,
    );
  });

  it("falls back to a generic continuation copy when the entry short caption is empty", () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    render(
      <MainPlaybackSurface
        fallbackHref="/shorts/softlight"
        surface={{
          ...surface,
          entryShort: surface.entryShort
            ? {
                ...surface.entryShort,
                caption: "   ",
              }
            : surface.entryShort,
        }}
      />,
    );

    expect(screen.getByText("short の続きから再生中。")).toBeInTheDocument();
  });
});
