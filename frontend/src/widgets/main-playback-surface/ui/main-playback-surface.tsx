"use client";

import { useCallback, useEffect, useRef, useState, type ChangeEvent } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Pause, Play, Volume2 } from "lucide-react";

import { getShortThemeStyle } from "@/entities/short";
import { Button } from "@/shared/ui";

import { formatPlaybackTimestamp } from "../lib/format-playback-timestamp";
import type { MainPlaybackSurface as MainPlaybackSurfaceModel } from "../model/main-playback-surface";

export type MainPlaybackSurfaceProps = {
  fallbackHref: string;
  isActive?: boolean;
  surface: MainPlaybackSurfaceModel;
};

const MAIN_PLAYBACK_VIDEO_LABEL = "Main playback video";

function resolveDurationSeconds(video: HTMLVideoElement, fallbackDurationSeconds: number): number {
  if (Number.isFinite(video.duration) && video.duration > 0) {
    return video.duration;
  }

  return fallbackDurationSeconds;
}

/**
 * unlock 後の main 継続視聴 surface を表示する。
 */
export function MainPlaybackSurface({
  fallbackHref,
  isActive = true,
  surface,
}: MainPlaybackSurfaceProps) {
  const router = useRouter();
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const resumeAppliedRef = useRef<string | null>(null);
  const playbackHeading = surface.entryShort?.caption.trim() || "Main playback";
  const resumePositionSeconds = surface.resumePositionSeconds;
  const resumeKey =
    resumePositionSeconds === null ? null : `${surface.main.media.id}:${surface.main.media.url}:${resumePositionSeconds}`;
  const mediaStateKey = `${surface.main.media.id}:${surface.main.media.url}:${surface.main.durationSeconds}:${resumePositionSeconds ?? "none"}`;
  const [currentTimeState, setCurrentTimeState] = useState({
    mediaStateKey,
    seconds: surface.resumePositionSeconds ?? 0,
  });
  const [durationState, setDurationState] = useState({
    mediaStateKey,
    seconds: surface.main.durationSeconds,
  });
  const [playbackState, setPlaybackState] = useState({
    isPlaying: isActive,
    mediaStateKey,
  });
  const [audioState, setAudioState] = useState({
    isMuted: false,
    mediaStateKey,
  });
  const currentTimeSeconds =
    currentTimeState.mediaStateKey === mediaStateKey ? currentTimeState.seconds : surface.resumePositionSeconds ?? 0;
  const durationSeconds =
    durationState.mediaStateKey === mediaStateKey ? durationState.seconds : surface.main.durationSeconds;
  const isPlaying = isActive && (playbackState.mediaStateKey === mediaStateKey ? playbackState.isPlaying : true);
  const isMuted = audioState.mediaStateKey === mediaStateKey ? audioState.isMuted : false;
  const progressPercent = durationSeconds > 0 ? Math.min(100, (currentTimeSeconds / durationSeconds) * 100) : 0;
  const playbackButtonLabel = isMuted && isPlaying ? "Enable audio" : isPlaying ? "Pause playback" : "Play playback";

  const handleBack = () => {
    if (window.history.length > 1) {
      router.back();
      return;
    }

    router.push(fallbackHref);
  };

  const syncMutedState = useCallback((video: HTMLVideoElement) => {
    setAudioState({
      isMuted: video.muted,
      mediaStateKey,
    });
  }, [mediaStateKey]);

  useEffect(() => {
    const video = videoRef.current;

    if (!video) {
      return;
    }

    if (!isActive) {
      video.pause();
      return;
    }

    let cancelled = false;

    const syncDuration = () => {
      setDurationState({
        mediaStateKey,
        seconds: resolveDurationSeconds(video, surface.main.durationSeconds),
      });
    };

    const applyResumePosition = () => {
      if (resumePositionSeconds === null || resumeKey === null || resumeAppliedRef.current === resumeKey) {
        return;
      }

      video.currentTime = resumePositionSeconds;
      resumeAppliedRef.current = resumeKey;
      setCurrentTimeState({
        mediaStateKey,
        seconds: resumePositionSeconds,
      });
    };

    const attemptPlayback = async () => {
      applyResumePosition();
      syncDuration();
      video.muted = false;
      syncMutedState(video);

      try {
        await video.play();

        if (!cancelled) {
          syncMutedState(video);
          setPlaybackState({
            isPlaying: true,
            mediaStateKey,
          });
        }
      } catch {
        if (cancelled) {
          return;
        }

        video.muted = true;
        syncMutedState(video);

        try {
          await video.play();

          if (!cancelled) {
            syncMutedState(video);
            setPlaybackState({
              isPlaying: true,
              mediaStateKey,
            });
          }
        } catch {
          if (!cancelled) {
            syncMutedState(video);
            setPlaybackState({
              isPlaying: false,
              mediaStateKey,
            });
          }
        }
      }
    };

    const handleMetadataLoaded = () => {
      syncDuration();
      void attemptPlayback();
    };

    if (video.readyState >= 1) {
      void attemptPlayback();
    } else {
      video.addEventListener("loadedmetadata", handleMetadataLoaded);
    }

    return () => {
      cancelled = true;
      video.removeEventListener("loadedmetadata", handleMetadataLoaded);
    };
  }, [
    isActive,
    mediaStateKey,
    resumeKey,
    resumePositionSeconds,
    syncMutedState,
    surface.main.durationSeconds,
    surface.main.media.id,
    surface.main.media.url,
  ]);

  const handleLoadedMetadata = () => {
    const video = videoRef.current;

    if (!video) {
      return;
    }

    setDurationState({
      mediaStateKey,
      seconds: resolveDurationSeconds(video, surface.main.durationSeconds),
    });

    if (resumePositionSeconds === null || resumeKey === null || resumeAppliedRef.current === resumeKey) {
      return;
    }

    video.currentTime = resumePositionSeconds;
    resumeAppliedRef.current = resumeKey;
    setCurrentTimeState({
      mediaStateKey,
      seconds: resumePositionSeconds,
    });
  };

  const handleDurationChange = () => {
    const video = videoRef.current;

    if (!video) {
      return;
    }

    setDurationState({
      mediaStateKey,
      seconds: resolveDurationSeconds(video, surface.main.durationSeconds),
    });
  };

  const handleTimeUpdate = () => {
    const video = videoRef.current;

    if (!video) {
      return;
    }

    setCurrentTimeState({
      mediaStateKey,
      seconds: video.currentTime,
    });
  };

  const handleSeek = (event: ChangeEvent<HTMLInputElement>) => {
    const video = videoRef.current;

    if (!video) {
      return;
    }

    const nextSeconds = Number.isFinite(event.currentTarget.valueAsNumber) ? event.currentTarget.valueAsNumber : 0;

    video.currentTime = nextSeconds;
    setCurrentTimeState({
      mediaStateKey,
      seconds: nextSeconds,
    });
  };

  const handleTogglePlayback = async () => {
    const video = videoRef.current;

    if (!video) {
      return;
    }

    if (video.paused || video.ended) {
      try {
        video.muted = false;
        syncMutedState(video);
        await video.play();
        syncMutedState(video);
        setPlaybackState({
          isPlaying: true,
          mediaStateKey,
        });
      } catch {
        setPlaybackState({
          isPlaying: false,
          mediaStateKey,
        });
      }

      return;
    }

    if (video.muted) {
      video.muted = false;
      syncMutedState(video);
      return;
    }

    video.pause();
    setPlaybackState({
      isPlaying: false,
      mediaStateKey,
    });
  };

  return (
    <section className="absolute inset-0 overflow-hidden text-white" style={getShortThemeStyle(surface.themeShort)}>
      <div className="absolute inset-0 bg-[linear-gradient(180deg,var(--short-bg-start)_0%,var(--short-bg-accent)_22%,var(--short-bg-mid)_56%,var(--short-bg-end)_100%)]" />
      <video
        ref={videoRef}
        aria-label={MAIN_PLAYBACK_VIDEO_LABEL}
        autoPlay={isActive}
        className="absolute inset-0 size-full object-cover"
        onDurationChange={handleDurationChange}
        onLoadedMetadata={handleLoadedMetadata}
        onPause={() => {
          setPlaybackState({
            isPlaying: false,
            mediaStateKey,
          });
        }}
        onPlay={() => {
          setPlaybackState({
            isPlaying: true,
            mediaStateKey,
          });
        }}
        onTimeUpdate={handleTimeUpdate}
        onVolumeChange={() => {
          const video = videoRef.current;

          if (!video) {
            return;
          }

          syncMutedState(video);
        }}
        playsInline
        poster={surface.main.media.posterUrl ?? undefined}
        preload="metadata"
        src={surface.main.media.url}
      />
      <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(3,10,18,0.32)_0%,rgba(3,10,18,0.08)_24%,rgba(3,10,18,0.18)_58%,rgba(3,10,18,0.86)_100%)]" />

      <div className="relative h-full">
        <h1 className="sr-only">{playbackHeading}</h1>

        <div className="absolute top-0 z-10 w-full bg-gradient-to-b from-black/78 to-transparent px-4 pb-6 pt-14">
          <Button
            aria-label="Back"
            className="text-white hover:bg-white/16 hover:text-white"
            onClick={handleBack}
            size="icon"
            type="button"
            variant="ghost"
          >
            <ArrowLeft className="size-6" strokeWidth={2.2} />
          </Button>
        </div>

        <div className="absolute bottom-0 z-10 w-full bg-gradient-to-t from-black/92 via-black/72 to-transparent px-5 pb-10 pt-24">
          <div className="flex items-center gap-4">
            <button
              aria-label={playbackButtonLabel}
              className="inline-flex size-9 items-center justify-center rounded-full text-white transition hover:bg-white/10"
              onClick={() => {
                void handleTogglePlayback();
              }}
              type="button"
            >
              {isMuted && isPlaying ? (
                <Volume2 className="size-5" strokeWidth={2.2} />
              ) : isPlaying ? (
                <Pause className="size-5 fill-current" />
              ) : (
                <Play className="size-5 fill-current" />
              )}
            </button>
            <span className="text-[13px] font-bold tabular-nums tracking-wide text-white/92">
              {formatPlaybackTimestamp(currentTimeSeconds)} / {formatPlaybackTimestamp(durationSeconds)}
            </span>
          </div>

          <div className="mt-3 mb-[calc(40px+env(safe-area-inset-bottom,0px))] w-full">
            <div className="group relative h-2 w-full overflow-hidden rounded-full border border-white/18 bg-white/32 shadow-[0_2px_10px_rgba(0,0,0,0.24)]">
              <div
                className="absolute inset-y-0 left-0 rounded-full bg-[#78c9ff] shadow-[0_0_12px_rgba(120,201,255,0.48)]"
                data-testid="playback-progress-fill"
                style={{ width: `${progressPercent}%` }}
              />
              <div className="pointer-events-none absolute inset-0 rounded-full ring-0 ring-white/60 transition group-focus-within:ring-2" />
              <input
                aria-label="Playback progress"
                aria-valuetext={`${formatPlaybackTimestamp(currentTimeSeconds)} of ${formatPlaybackTimestamp(durationSeconds)}`}
                className="absolute inset-0 h-full w-full cursor-pointer appearance-none bg-transparent opacity-0"
                max={durationSeconds > 0 ? durationSeconds : 0}
                min={0}
                onChange={handleSeek}
                step="0.1"
                type="range"
                value={Math.min(currentTimeSeconds, durationSeconds || 0)}
              />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
