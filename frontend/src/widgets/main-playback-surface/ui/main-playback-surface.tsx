"use client";

import { useEffect, useRef, useState, type ChangeEvent } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Pause, Play } from "lucide-react";

import { getShortThemeStyle } from "@/entities/short";
import { Button } from "@/shared/ui";

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
  const currentTimeSeconds =
    currentTimeState.mediaStateKey === mediaStateKey ? currentTimeState.seconds : surface.resumePositionSeconds ?? 0;
  const durationSeconds =
    durationState.mediaStateKey === mediaStateKey ? durationState.seconds : surface.main.durationSeconds;
  const isPlaying = isActive && (playbackState.mediaStateKey === mediaStateKey ? playbackState.isPlaying : true);
  const progressPercent = durationSeconds > 0 ? Math.min(100, (currentTimeSeconds / durationSeconds) * 100) : 0;
  const playbackButtonLabel = isPlaying ? "Pause playback" : "Play playback";

  const handleBack = () => {
    if (window.history.length > 1) {
      router.back();
      return;
    }

    router.push(fallbackHref);
  };

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

      try {
        await video.play();

        if (!cancelled) {
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

        try {
          await video.play();

          if (!cancelled) {
            setPlaybackState({
              isPlaying: true,
              mediaStateKey,
            });
          }
        } catch {
          if (!cancelled) {
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
        await video.play();
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
              {isPlaying ? <Pause className="size-5 fill-current" /> : <Play className="size-5 fill-current" />}
            </button>
            <span className="text-[13px] font-bold tabular-nums tracking-wide text-white/92">
              {formatPlaybackTimestamp(currentTimeSeconds)} / {formatPlaybackTimestamp(durationSeconds)}
            </span>
          </div>

          <div className="group relative mt-3 flex h-5 items-center">
            <div className="pointer-events-none absolute inset-x-0 top-1/2 h-1.5 -translate-y-1/2 overflow-hidden rounded-full bg-white/24 shadow-[inset_0_0_0_1px_rgba(255,255,255,0.08)]">
              <div
                className="h-full rounded-full bg-[#78c9ff] shadow-[0_0_12px_rgba(120,201,255,0.4)]"
                data-testid="playback-progress-fill"
                style={{ width: `${progressPercent}%` }}
              />
            </div>
            <div className="pointer-events-none absolute inset-0 rounded-full ring-0 ring-white/45 transition group-focus-within:ring-2" />
            <input
              aria-label="Playback progress"
              aria-valuetext={`${formatPlaybackTimestamp(currentTimeSeconds)} of ${formatPlaybackTimestamp(durationSeconds)}`}
              className="absolute inset-x-0 top-1/2 h-5 w-full -translate-y-1/2 cursor-pointer appearance-none bg-transparent opacity-0"
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
    </section>
  );
}
