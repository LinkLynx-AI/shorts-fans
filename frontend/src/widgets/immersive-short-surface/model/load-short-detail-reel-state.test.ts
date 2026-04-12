import { getCreatorProfileShortGrid } from "@/entities/creator";
import { fetchFanProfilePinnedShortsPage } from "@/entities/fan-profile";
import { getPublicShortDetail } from "@/entities/short";

import { loadShortDetailReelState } from "./load-short-detail-reel-state";

vi.mock("@/entities/creator", async () => {
  const actual = await vi.importActual<typeof import("@/entities/creator")>("@/entities/creator");

  return {
    ...actual,
    getCreatorProfileShortGrid: vi.fn(),
  };
});

vi.mock("@/entities/fan-profile", async () => {
  const actual = await vi.importActual<typeof import("@/entities/fan-profile")>("@/entities/fan-profile");

  return {
    ...actual,
    fetchFanProfilePinnedShortsPage: vi.fn(),
  };
});

vi.mock("@/entities/short", async () => {
  const actual = await vi.importActual<typeof import("@/entities/short")>("@/entities/short");

  return {
    ...actual,
    getPublicShortDetail: vi.fn(),
  };
});

const mockedFetchFanProfilePinnedShortsPage = vi.mocked(fetchFanProfilePinnedShortsPage);
const mockedGetCreatorProfileShortGrid = vi.mocked(getCreatorProfileShortGrid);
const mockedGetPublicShortDetail = vi.mocked(getPublicShortDetail);

describe("loadShortDetailReelState", () => {
  beforeEach(() => {
    mockedFetchFanProfilePinnedShortsPage.mockReset();
    mockedGetCreatorProfileShortGrid.mockReset();
    mockedGetPublicShortDetail.mockReset();
  });

  it("returns creator-source short ids without preloading every detail payload", async () => {
    mockedGetCreatorProfileShortGrid.mockResolvedValue({
      items: [
        {
          canonicalMainId: "main_1",
          creatorId: "creator_1",
          id: "short_1",
          media: {
            durationSeconds: 16,
            id: "asset_1",
            kind: "video",
            posterUrl: null,
            url: "https://cdn.example.com/shorts/1.mp4",
          },
          previewDurationSeconds: 16,
        },
        {
          canonicalMainId: "main_2",
          creatorId: "creator_1",
          id: "short_2",
          media: {
            durationSeconds: 18,
            id: "asset_2",
            kind: "video",
            posterUrl: null,
            url: "https://cdn.example.com/shorts/2.mp4",
          },
          previewDurationSeconds: 18,
        },
        {
          canonicalMainId: "main_3",
          creatorId: "creator_1",
          id: "short_3",
          media: {
            durationSeconds: 20,
            id: "asset_3",
            kind: "video",
            posterUrl: null,
            url: "https://cdn.example.com/shorts/3.mp4",
          },
          previewDurationSeconds: 20,
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_creator_short_grid_001",
    });

    await expect(
      loadShortDetailReelState({
        creatorId: "creator_1",
        kind: "creator",
        shortId: "short_2",
      }),
    ).resolves.toEqual({
      initialIndex: 1,
      shortIds: ["short_1", "short_2", "short_3"],
    });
    expect(mockedGetPublicShortDetail).not.toHaveBeenCalled();
  });

  it("returns pinned-source short ids and preserves the source tab", async () => {
    mockedFetchFanProfilePinnedShortsPage.mockResolvedValue({
      items: [
        {
          creator: {
            avatar: null,
            bio: "quiet rooftop と hotel light の preview を軸に投稿。",
            displayName: "Mina Rei",
            handle: "@minarei",
            id: "creator_1",
          },
          short: {
            caption: "first",
            canonicalMainId: "main_1",
            creatorId: "creator_1",
            id: "short_1",
            media: {
              durationSeconds: 16,
              id: "asset_1",
              kind: "video",
              posterUrl: null,
              url: "https://cdn.example.com/shorts/1.mp4",
            },
            previewDurationSeconds: 16,
          },
        },
        {
          creator: {
            avatar: null,
            bio: "quiet rooftop と hotel light の preview を軸に投稿。",
            displayName: "Mina Rei",
            handle: "@minarei",
            id: "creator_1",
          },
          short: {
            caption: "second",
            canonicalMainId: "main_2",
            creatorId: "creator_1",
            id: "short_2",
            media: {
              durationSeconds: 18,
              id: "asset_2",
              kind: "video",
              posterUrl: null,
              url: "https://cdn.example.com/shorts/2.mp4",
            },
            previewDurationSeconds: 18,
          },
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_fan_profile_pinned_001",
    } as Awaited<ReturnType<typeof fetchFanProfilePinnedShortsPage>>);

    await expect(
      loadShortDetailReelState({
        kind: "fan",
        shortId: "short_2",
        tab: "pinned",
      }),
    ).resolves.toEqual({
      initialIndex: 1,
      shortIds: ["short_1", "short_2"],
      sourceTab: "pinned",
    });
    expect(mockedGetPublicShortDetail).not.toHaveBeenCalled();
  });
});
