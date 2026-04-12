import { StrictMode } from "react";
import { render, screen, waitFor } from "@testing-library/react";

import { getPublicShortDetail } from "@/entities/short";

import type { DetailShortSurface } from "../model/short-surface";
import { ShortDetailReel } from "./short-detail-reel";

vi.mock("@/entities/short", async () => {
  const actual = await vi.importActual<typeof import("@/entities/short")>("@/entities/short");

  return {
    ...actual,
    getPublicShortDetail: vi.fn(),
  };
});

vi.mock("./immersive-short-surface", () => ({
  ImmersiveShortSurface: vi.fn(
    ({ surface }: { surface: { short: { id: string } } }) => (
      <div data-testid={`immersive-short-surface-${surface.short.id}`} />
    ),
  ),
}));

const mockedGetPublicShortDetail = vi.mocked(getPublicShortDetail);

function createDetailPayload(shortId: string) {
  return {
    creator: {
      avatar: null,
      bio: "quiet rooftop と hotel light の preview を軸に投稿。",
      displayName: "Mina Rei",
      handle: "@minarei" as const,
      id: "creator_1",
    },
    short: {
      caption: `${shortId} caption`,
      canonicalMainId: `main_${shortId}`,
      creatorId: "creator_1",
      id: shortId,
      media: {
        durationSeconds: 16,
        id: `asset_${shortId}`,
        kind: "video" as const,
        posterUrl: null,
        url: `https://cdn.example.com/shorts/${shortId}.mp4`,
      },
      previewDurationSeconds: 16,
    },
    unlockCta: {
      mainDurationSeconds: 480,
      priceJpy: 1800,
      resumePositionSeconds: null,
      state: "unlock_available" as const,
    },
    viewer: {
      isFollowingCreator: false,
      isPinned: false,
    },
  } satisfies Awaited<ReturnType<typeof getPublicShortDetail>>;
}

function createInitialSurface(shortId: string): DetailShortSurface {
  return {
    creator: {
      avatar: null,
      bio: "quiet rooftop と hotel light の preview を軸に投稿。",
      displayName: "Mina Rei",
      handle: "@minarei" as const,
      id: "creator_1",
    },
    mainEntryEnabled: false,
    short: {
      caption: `${shortId} caption`,
      canonicalMainId: `main_${shortId}`,
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
    unlock: {
      access: {
        mainId: `main_${shortId}`,
        reason: "unlock_required",
        status: "locked",
      },
      creator: {
        avatar: null,
        bio: "quiet rooftop と hotel light の preview を軸に投稿。",
        displayName: "Mina Rei",
        handle: "@minarei" as const,
        id: "creator_1",
      },
      main: {
        durationSeconds: 480,
        id: `main_${shortId}`,
        priceJpy: 1800,
      },
      mainAccessEntry: {
        routePath: `/api/fan/mains/main_${shortId}/access-entry`,
        token: `token_${shortId}`,
      },
      setup: {
        required: false,
        requiresAgeConfirmation: false,
        requiresTermsAcceptance: false,
      },
      short: {
        caption: `${shortId} caption`,
        canonicalMainId: `main_${shortId}`,
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
      unlockCta: {
        mainDurationSeconds: 480,
        priceJpy: 1800,
        resumePositionSeconds: null,
        state: "unlock_available",
      },
    },
    viewer: {
      isFollowingCreator: false,
      isPinned: false,
    },
  };
}

describe("ShortDetailReel", () => {
  beforeEach(() => {
    mockedGetPublicShortDetail.mockReset();
  });

  it("preloads only the active short and its immediate neighbors", async () => {
    mockedGetPublicShortDetail.mockImplementation(async ({ shortId }) => createDetailPayload(shortId));

    render(
      <ShortDetailReel
        backHref="/fan?tab=pinned"
        fanTab="pinned"
        initialIndex={1}
        initialSurface={createInitialSurface("short_2")}
        shortIds={["short_1", "short_2", "short_3", "short_4"]}
        source="fan"
      />,
    );

    await waitFor(() => {
      expect(mockedGetPublicShortDetail).toHaveBeenCalledTimes(2);
    });
    expect(mockedGetPublicShortDetail).toHaveBeenCalledWith({
      shortId: "short_1",
    });
    expect(mockedGetPublicShortDetail).toHaveBeenCalledWith({
      shortId: "short_3",
    });
  });

  it("commits prefetched shorts in StrictMode after the dev remount pass", async () => {
    mockedGetPublicShortDetail.mockImplementation(async ({ shortId }) => createDetailPayload(shortId));

    render(
      <StrictMode>
        <ShortDetailReel
          backHref="/fan?tab=pinned"
          fanTab="pinned"
          initialIndex={1}
          initialSurface={createInitialSurface("short_2")}
          shortIds={["short_1", "short_2", "short_3"]}
          source="fan"
        />
      </StrictMode>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("immersive-short-surface-short_3")).toBeInTheDocument();
    });
  });
});
