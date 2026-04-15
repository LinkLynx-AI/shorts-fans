import userEvent from "@testing-library/user-event";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import { getMainPlaybackSurfaceById } from "@/widgets/main-playback-surface";

import { MainPlaybackSurface } from "./main-playback-surface";

const back = vi.fn();
const push = vi.fn();
const pause = vi.fn();
const play = vi.fn<() => Promise<void>>();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    back,
    push,
  }),
}));

function formatPlaybackTimestamp(totalSeconds: number): string {
  const normalizedSeconds = Math.max(0, Math.floor(totalSeconds));
  const hours = Math.floor(normalizedSeconds / 3600);
  const minutes = Math.floor(normalizedSeconds / 60);
  const remainingSeconds = normalizedSeconds % 60;

  if (hours > 0) {
    return `${hours}:${String(Math.floor((normalizedSeconds % 3600) / 60)).padStart(2, "0")}:${String(remainingSeconds).padStart(2, "0")}`;
  }

  return `${String(minutes).padStart(2, "0")}:${String(remainingSeconds).padStart(2, "0")}`;
}

describe("MainPlaybackSurface", () => {
  beforeEach(() => {
    pause.mockReset();
    pause.mockImplementation(() => {});
    play.mockReset();
    play.mockResolvedValue(undefined);
    vi.spyOn(HTMLMediaElement.prototype, "pause").mockImplementation(pause);
    vi.spyOn(HTMLMediaElement.prototype, "play").mockImplementation(play);
  });

  afterEach(() => {
    back.mockReset();
    push.mockReset();
    vi.restoreAllMocks();
  });

  it("renders the refreshed playback chrome and falls back to the provided short when there is no browser history", async () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    render(<MainPlaybackSurface fallbackHref="/shorts/softlight" surface={surface} />);

    const video = screen.getByLabelText("Main playback video");

    expect(screen.getByRole("button", { name: "Back" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pause playback" })).toBeInTheDocument();
    expect(
      screen.getByText(
        `${formatPlaybackTimestamp(surface.resumePositionSeconds ?? 0)} / ${formatPlaybackTimestamp(surface.main.durationSeconds)}`,
      ),
    ).toBeInTheDocument();
    expect(screen.getByRole("slider", { name: "Playback progress" })).toBeInTheDocument();
    expect(screen.queryByText("Playing main")).not.toBeInTheDocument();
    expect(screen.queryByText("Owner preview")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Pin short" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /Aoi N/i })).not.toBeInTheDocument();
    expect(video).not.toHaveAttribute("controls");
    expect(video).toHaveAttribute("playsinline");
    expect(video).toHaveAttribute("poster", surface.main.media.posterUrl ?? undefined);
    expect(video).toHaveAttribute("src", surface.main.media.url);

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

    expect(screen.queryByText("Owner preview")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Back" }));

    expect(back).toHaveBeenCalled();
    expect(push).not.toHaveBeenCalled();
  });

  it("applies the resume position to actual playback after metadata loads", async () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    render(<MainPlaybackSurface fallbackHref="/shorts/softlight" surface={surface} />);

    const video = screen.getByLabelText("Main playback video") as HTMLVideoElement;

    expect(video.currentTime).toBe(0);

    await act(async () => {
      fireEvent(video, new Event("loadedmetadata"));
    });

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

  it("reapplies the resume position when the playback URL changes for the same media", async () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    const { rerender } = render(<MainPlaybackSurface fallbackHref="/shorts/softlight" surface={surface} />);

    const video = screen.getByLabelText("Main playback video") as HTMLVideoElement;

    await act(async () => {
      fireEvent(video, new Event("loadedmetadata"));
    });
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

    await act(async () => {
      fireEvent(screen.getByLabelText("Main playback video"), new Event("loadedmetadata"));
    });

    expect((screen.getByLabelText("Main playback video") as HTMLVideoElement).currentTime).toBe(
      surface.resumePositionSeconds,
    );
  });

  it("updates playback time, toggles pause and play, and seeks through the custom progress control", async () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    render(<MainPlaybackSurface fallbackHref="/shorts/softlight" surface={surface} />);

    const video = screen.getByLabelText("Main playback video") as HTMLVideoElement;
    const progress = screen.getByRole("slider", { name: "Playback progress" });
    const progressFill = screen.getByTestId("playback-progress-fill");
    let pausedState = false;

    Object.defineProperty(video, "paused", {
      configurable: true,
      get: () => pausedState,
    });
    Object.defineProperty(video, "duration", {
      configurable: true,
      value: surface.main.durationSeconds,
    });

    pause.mockImplementation(() => {
      pausedState = true;
    });
    play.mockImplementation(() => {
      pausedState = false;
      return Promise.resolve();
    });

    fireEvent(video, new Event("loadedmetadata"));
    await waitFor(() => {
      expect(play).toHaveBeenCalled();
    });

    video.currentTime = 210;
    fireEvent(video, new Event("timeupdate"));

    expect(
      screen.getByText(`03:30 / ${formatPlaybackTimestamp(surface.main.durationSeconds)}`),
    ).toBeInTheDocument();
    expect(progressFill).toHaveStyle({
      width: `${(210 / surface.main.durationSeconds) * 100}%`,
    });

    await user.click(screen.getByRole("button", { name: "Pause playback" }));

    expect(pause).toHaveBeenCalled();

    fireEvent.change(progress, {
      target: {
        value: "120",
      },
    });

    expect(video.currentTime).toBe(120);
    expect(
      screen.getByText(`02:00 / ${formatPlaybackTimestamp(surface.main.durationSeconds)}`),
    ).toBeInTheDocument();
    expect(progressFill).toHaveStyle({
      width: `${(120 / surface.main.durationSeconds) * 100}%`,
    });

    await user.click(screen.getByRole("button", { name: "Play playback" }));

    await waitFor(() => {
      expect(play).toHaveBeenCalledTimes(2);
    });
  });

  it("pauses playback while the panel is inactive", () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    render(<MainPlaybackSurface fallbackHref="/shorts/softlight" isActive={false} surface={surface} />);

    expect(pause).toHaveBeenCalled();
    expect(play).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Play playback" })).toBeInTheDocument();
  });
});
