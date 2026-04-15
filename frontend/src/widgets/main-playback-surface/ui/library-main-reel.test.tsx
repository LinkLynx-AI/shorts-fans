import { StrictMode } from "react";
import { render, screen, waitFor } from "@testing-library/react";

import type { FanLibraryItem } from "@/entities/fan-profile";

import { resolveLibraryMainPlaybackSurface } from "../api/resolve-library-main-playback-surface";
import type { MainPlaybackSurface as MainPlaybackSurfaceModel } from "../model/main-playback-surface";
import { LibraryMainReel } from "./library-main-reel";

vi.mock("../api/resolve-library-main-playback-surface", () => ({
  resolveLibraryMainPlaybackSurface: vi.fn(),
}));

vi.mock("./main-playback-surface", () => ({
  MainPlaybackSurface: vi.fn(
    ({
      creatorProfileHref,
      surface,
    }: {
      creatorProfileHref: string;
      surface: { main: { id: string } };
    }) => (
      <div
        data-creator-profile-href={creatorProfileHref}
        data-testid={`main-playback-surface-${surface.main.id}`}
      />
    ),
  ),
}));

vi.mock("./main-playback-locked-state", () => ({
  MainPlaybackLockedState: vi.fn(() => <div data-testid="main-playback-locked-state" />),
}));

const mockedResolveLibraryMainPlaybackSurface = vi.mocked(resolveLibraryMainPlaybackSurface);

function createLibraryItem(shortId: string, mainId: string): FanLibraryItem {
  return {
    access: {
      mainId,
      reason: "session_unlocked",
      status: "unlocked",
    },
    creator: {
      avatar: null,
      bio: "quiet rooftop と hotel light の preview を軸に投稿。",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_1",
    },
    entryShort: {
      caption: `${shortId} caption`,
      canonicalMainId: mainId,
      creatorId: "creator_1",
      id: shortId,
      media: {
        durationSeconds: 16,
        id: `asset_${shortId}`,
        kind: "video",
        posterUrl: null,
        url: `https://cdn.example.com/shorts/${shortId}.mp4`,
      },
      previewDurationSeconds: 16,
    },
    main: {
      durationSeconds: 480,
      id: mainId,
    },
  };
}

function createMainPlaybackSurface(mainId: string, shortId: string): MainPlaybackSurfaceModel {
  return {
    access: {
      mainId,
      reason: "session_unlocked",
      status: "unlocked",
    },
    creator: {
      avatar: null,
      bio: "quiet rooftop と hotel light の preview を軸に投稿。",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_1",
    },
    entryShort: {
      caption: `${shortId} caption`,
      canonicalMainId: mainId,
      creatorId: "creator_1",
      id: shortId,
      media: {
        durationSeconds: 16,
        id: `asset_${shortId}`,
        kind: "video",
        posterUrl: null,
        url: `https://cdn.example.com/shorts/${shortId}.mp4`,
      },
      previewDurationSeconds: 16,
    },
    main: {
      durationSeconds: 480,
      id: mainId,
      media: {
        durationSeconds: 480,
        id: `asset_${mainId}`,
        kind: "video",
        posterUrl: null,
        url: `https://cdn.example.com/mains/${mainId}.mp4`,
      },
    },
    resumePositionSeconds: null,
    themeShort: {
      caption: `${shortId} caption`,
      canonicalMainId: mainId,
      creatorId: "creator_1",
      id: shortId,
      media: {
        durationSeconds: 16,
        id: `asset_${shortId}`,
        kind: "video",
        posterUrl: null,
        url: `https://cdn.example.com/shorts/${shortId}.mp4`,
      },
      previewDurationSeconds: 16,
    },
    viewer: {
      isPinned: false,
    },
  };
}

describe("LibraryMainReel", () => {
  beforeEach(() => {
    mockedResolveLibraryMainPlaybackSurface.mockReset();
  });

  it("commits resolved main playback in StrictMode after the dev remount pass", async () => {
    mockedResolveLibraryMainPlaybackSurface.mockResolvedValue({
      kind: "ready",
      surface: createMainPlaybackSurface("main_short_1", "short_1"),
    });

    render(
      <StrictMode>
        <LibraryMainReel
          backHref="/fan?tab=library"
          initialIndex={0}
          items={[createLibraryItem("short_1", "main_short_1")]}
        />
      </StrictMode>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("main-playback-surface-main_short_1")).toBeInTheDocument();
    });

    expect(screen.getByTestId("main-playback-surface-main_short_1")).toHaveAttribute(
      "data-creator-profile-href",
      "/creators/creator_1?from=short&shortFanTab=library&shortId=short_1",
    );
  });
});
