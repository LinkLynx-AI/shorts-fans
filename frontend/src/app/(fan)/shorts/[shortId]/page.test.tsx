import { render, screen } from "@testing-library/react";

import { getPublicShortDetail } from "@/entities/short";
import {
  loadShortDetailReelState,
} from "@/widgets/immersive-short-surface";

import ShortDetailPage from "./page";

const { cookiesMock } = vi.hoisted(() => ({
  cookiesMock: vi.fn(),
}));

vi.mock("next/headers", () => ({
  cookies: cookiesMock,
}));

vi.mock("@/entities/short", async () => {
  const actual = await vi.importActual<typeof import("@/entities/short")>("@/entities/short");

  return {
    ...actual,
    getPublicShortDetail: vi.fn(),
  };
});

vi.mock("@/widgets/immersive-short-surface", async () => {
  const actual = await vi.importActual<typeof import("@/widgets/immersive-short-surface")>(
    "@/widgets/immersive-short-surface",
  );

  return {
    ...actual,
    ImmersiveShortSurface: vi.fn((props: { surface: { short: { id: string } } }) => (
      <div data-testid="single-short-detail">{props.surface.short.id}</div>
    )),
    ShortDetailReel: vi.fn((props: { initialIndex: number; source: string }) => (
      <div data-testid="short-detail-reel">{`${props.source}:${props.initialIndex}`}</div>
    )),
    loadShortDetailReelState: vi.fn(),
  };
});

const mockedGetPublicShortDetail = vi.mocked(getPublicShortDetail);
const mockedLoadShortDetailReelState = vi.mocked(loadShortDetailReelState);

describe("ShortDetailPage", () => {
  beforeEach(() => {
    cookiesMock.mockReset();
    mockedGetPublicShortDetail.mockReset();
    mockedLoadShortDetailReelState.mockReset();
  });

  it("renders the creator-source short reel when the source list is available", async () => {
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    mockedLoadShortDetailReelState.mockResolvedValue({
      initialIndex: 1,
      surfaces: [],
    });

    render(
      await ShortDetailPage({
        params: Promise.resolve({
          shortId: "short_mina_rooftop",
        }),
        searchParams: Promise.resolve({
          creatorId: "creator_mina_rei",
          from: "creator",
          profileFrom: "feed",
          profileTab: "recommended",
        }),
      }),
    );

    expect(mockedLoadShortDetailReelState).toHaveBeenCalledWith({
      creatorId: "creator_mina_rei",
      kind: "creator",
      sessionToken: "valid-session",
      shortId: "short_mina_rooftop",
    });
    expect(screen.getByTestId("short-detail-reel")).toHaveTextContent("creator:1");
    expect(mockedGetPublicShortDetail).not.toHaveBeenCalled();
  });

  it("falls back to a single short detail when the fan pinned reel cannot be built", async () => {
    cookiesMock.mockResolvedValue({
      get: () => ({
        value: "valid-session",
      }),
    });
    mockedLoadShortDetailReelState.mockResolvedValue(null);
    mockedGetPublicShortDetail.mockResolvedValue({
      creator: {
        avatar: null,
        bio: "quiet rooftop と hotel light の preview を軸に投稿。",
        displayName: "Mina Rei",
        handle: "@minarei",
        id: "creator_mina_rei",
      },
      short: {
        caption: "quiet rooftop preview",
        canonicalMainId: "main_mina_quiet_rooftop",
        creatorId: "creator_mina_rei",
        id: "short_mina_rooftop",
        media: {
          durationSeconds: 16,
          id: "asset_short_mina_rooftop",
          kind: "video",
          posterUrl: "https://cdn.example.com/shorts/mina-rooftop-poster.jpg",
          url: "https://cdn.example.com/shorts/mina-rooftop.mp4",
        },
        previewDurationSeconds: 16,
      },
      unlockCta: {
        mainDurationSeconds: 480,
        priceJpy: 1200,
        resumePositionSeconds: null,
        state: "unlock_available",
      },
      viewer: {
        isFollowingCreator: true,
        isPinned: true,
      },
    });

    render(
      await ShortDetailPage({
        params: Promise.resolve({
          shortId: "short_mina_rooftop",
        }),
        searchParams: Promise.resolve({
          fanTab: "pinned",
          from: "fan",
        }),
      }),
    );

    expect(mockedLoadShortDetailReelState).toHaveBeenCalledWith({
      kind: "fan",
      sessionToken: "valid-session",
      shortId: "short_mina_rooftop",
      tab: "pinned",
    });
    expect(mockedGetPublicShortDetail).toHaveBeenCalledWith({
      sessionToken: "valid-session",
      shortId: "short_mina_rooftop",
    });
    expect(screen.getByTestId("single-short-detail")).toHaveTextContent("short_mina_rooftop");
  });
});
