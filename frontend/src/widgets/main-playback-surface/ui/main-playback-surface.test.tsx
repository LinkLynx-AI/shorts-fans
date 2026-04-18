import userEvent from "@testing-library/user-event";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import { CurrentViewerProvider } from "@/entities/viewer";
import { buildCreatorProfileHref } from "@/features/creator-navigation";
import { fireRecommendationSignal } from "@/features/recommendation-signal";
import { getMainPlaybackSurfaceById } from "@/widgets/main-playback-surface";

import { formatPlaybackTimestamp } from "../lib/format-playback-timestamp";
import { MainPlaybackSurface } from "./main-playback-surface";

function getExpectedCreatorProfileHref(creatorId: string, shortId: string): string {
  return buildCreatorProfileHref(creatorId, {
    from: "short",
    shortId,
  });
}

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

vi.mock("@/features/recommendation-signal", () => ({
  createRecommendationSignalIdempotencyKey: vi.fn(() => "profile_click:creator:nonce"),
  createRecommendationSignalNonce: vi.fn(() => "nonce"),
  fireRecommendationSignal: vi.fn(),
  isRecommendationPublicCreatorId: vi.fn(() => true),
}));

const mockedFireRecommendationSignal = vi.mocked(fireRecommendationSignal);

describe("MainPlaybackSurface", () => {
  beforeEach(() => {
    mockedFireRecommendationSignal.mockReset();
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

    render(
      <MainPlaybackSurface
        creatorProfileHref={getExpectedCreatorProfileHref(surface.creator.id, "softlight")}
        fallbackHref="/shorts/softlight"
        surface={surface}
      />,
    );

    const video = screen.getByLabelText("Main playback video");

    expect(screen.getByRole("button", { name: "Back" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "More options" })).toBeInTheDocument();
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

  it("opens the bottom-sheet menu and exposes the creator profile action", async () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    render(
      <MainPlaybackSurface
        creatorProfileHref={getExpectedCreatorProfileHref(surface.creator.id, "softlight")}
        fallbackHref="/shorts/softlight"
        surface={surface}
      />,
    );

    await user.click(screen.getByRole("button", { name: "More options" }));

    expect(screen.getByRole("dialog", { name: "Main options" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "クリエイターのプロフィールへ" })).toHaveAttribute(
      "href",
      getExpectedCreatorProfileHref(surface.creator.id, "softlight"),
    );
  });

  it("records a profile click when the creator profile action is pressed by an authenticated viewer", async () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    render(
      <CurrentViewerProvider
        currentViewer={{
          activeMode: "fan",
          canAccessCreatorMode: false,
          id: "viewer_1",
        }}
      >
        <MainPlaybackSurface
          creatorProfileHref={getExpectedCreatorProfileHref(surface.creator.id, "softlight")}
          fallbackHref="/shorts/softlight"
          surface={surface}
        />
      </CurrentViewerProvider>,
    );

    await user.click(screen.getByRole("button", { name: "More options" }));
    await user.click(screen.getByRole("link", { name: "クリエイターのプロフィールへ" }));

    expect(mockedFireRecommendationSignal).toHaveBeenCalledWith({
      creatorId: surface.creator.id,
      eventKind: "profile_click",
      idempotencyKey: "profile_click:creator:nonce",
    });
  });

  it("does not record a profile click for owner preview playback", async () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "balcony", "owner");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    render(
      <CurrentViewerProvider
        currentViewer={{
          activeMode: "creator",
          canAccessCreatorMode: true,
          id: "viewer_1",
        }}
      >
        <MainPlaybackSurface
          creatorProfileHref={getExpectedCreatorProfileHref(surface.creator.id, "balcony")}
          fallbackHref="/shorts/balcony"
          surface={surface}
        />
      </CurrentViewerProvider>,
    );

    await user.click(screen.getByRole("button", { name: "More options" }));
    await user.click(screen.getByRole("link", { name: "クリエイターのプロフィールへ" }));

    expect(mockedFireRecommendationSignal).not.toHaveBeenCalled();
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

    render(
      <MainPlaybackSurface
        creatorProfileHref={getExpectedCreatorProfileHref(surface.creator.id, "balcony")}
        fallbackHref="/shorts/balcony"
        surface={surface}
      />,
    );

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

    render(
      <MainPlaybackSurface
        creatorProfileHref={getExpectedCreatorProfileHref(surface.creator.id, "softlight")}
        fallbackHref="/shorts/softlight"
        surface={surface}
      />,
    );

    const video = screen.getByLabelText("Main playback video") as HTMLVideoElement;

    expect(video.currentTime).toBe(0);

    await act(async () => {
      fireEvent(video, new Event("loadedmetadata"));
    });

    expect(video.currentTime).toBe(surface.resumePositionSeconds);
  });

  it("falls back to muted autoplay and lets the primary control enable audio", async () => {
    play.mockRejectedValueOnce(new Error("autoplay blocked")).mockResolvedValueOnce(undefined);

    const surface = getMainPlaybackSurfaceById("main_sora_after_rain", "afterrain", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    render(
      <MainPlaybackSurface
        creatorProfileHref={getExpectedCreatorProfileHref(surface.creator.id, "afterrain")}
        fallbackHref="/shorts/afterrain"
        surface={surface}
      />,
    );

    const video = screen.getByLabelText("Main playback video") as HTMLVideoElement;

    fireEvent(video, new Event("loadedmetadata"));

    await waitFor(() => {
      expect(play).toHaveBeenCalledTimes(2);
    });

    expect(video.muted).toBe(true);
    expect(screen.getByRole("button", { name: "Enable audio" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Enable audio" }));

    expect(video.muted).toBe(false);
    expect(pause).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Pause playback" })).toBeInTheDocument();
  });

  it("reapplies the resume position when the playback URL changes for the same media", async () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    const { rerender } = render(
      <MainPlaybackSurface
        creatorProfileHref={getExpectedCreatorProfileHref(surface.creator.id, "softlight")}
        fallbackHref="/shorts/softlight"
        surface={surface}
      />,
    );

    const video = screen.getByLabelText("Main playback video") as HTMLVideoElement;

    await act(async () => {
      fireEvent(video, new Event("loadedmetadata"));
    });
    expect(video.currentTime).toBe(surface.resumePositionSeconds);

    video.currentTime = 0;

    rerender(
      <MainPlaybackSurface
        creatorProfileHref={getExpectedCreatorProfileHref(surface.creator.id, "softlight")}
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

    render(
      <MainPlaybackSurface
        creatorProfileHref={getExpectedCreatorProfileHref(surface.creator.id, "softlight")}
        fallbackHref="/shorts/softlight"
        surface={surface}
      />,
    );

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

    render(
      <MainPlaybackSurface
        creatorProfileHref={getExpectedCreatorProfileHref(surface.creator.id, "softlight")}
        fallbackHref="/shorts/softlight"
        isActive={false}
        surface={surface}
      />,
    );

    expect(pause).toHaveBeenCalled();
    expect(play).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Play playback" })).toBeInTheDocument();
  });
});
