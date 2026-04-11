import { ApiError } from "@/shared/api";
import { loadCreatorProfileShellState } from "@/widgets/creator-profile-shell";
import {
  getCreatorById,
  getCreatorProfileHeader,
  getCreatorProfileShortGrid,
} from "@/entities/creator";

vi.mock("@/entities/creator", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/entities/creator")>();

  return {
    ...actual,
    getCreatorProfileHeader: vi.fn(),
    getCreatorProfileShortGrid: vi.fn(),
  };
});

const mockedGetCreatorProfileHeader = vi.mocked(getCreatorProfileHeader);
const mockedGetCreatorProfileShortGrid = vi.mocked(getCreatorProfileShortGrid);

describe("loadCreatorProfileShellState", () => {
  beforeEach(() => {
    mockedGetCreatorProfileHeader.mockReset();
    mockedGetCreatorProfileShortGrid.mockReset();
  });

  it("builds the ready state from creator profile responses", async () => {
    const creator = getCreatorById("creator_mina_rei");

    if (!creator) {
      throw new Error("fixture missing");
    }

    mockedGetCreatorProfileHeader.mockResolvedValue({
      creator,
      stats: {
        fanCount: 24000,
        shortCount: 2,
      },
      viewer: {
        isFollowing: true,
      },
    });
    mockedGetCreatorProfileShortGrid.mockResolvedValue({
      items: [
        {
          canonicalMainId: "main_mina_quiet_rooftop",
          creatorId: "creator_mina_rei",
          id: "short_mina_rooftop",
          media: {
            durationSeconds: 16,
            id: "asset_short_mina_rooftop",
            kind: "video",
            posterUrl: null,
            url: "https://cdn.example.com/shorts/mina-rooftop.mp4",
          },
          previewDurationSeconds: 16,
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_creator_profile_shorts_001",
    });

    await expect(
      loadCreatorProfileShellState("creator_mina_rei", {
        sessionToken: "raw-session-token",
      }),
    ).resolves.toEqual({
      creator,
      kind: "ready",
      shorts: [
        {
          canonicalMainId: "main_mina_quiet_rooftop",
          creatorId: "creator_mina_rei",
          id: "short_mina_rooftop",
          media: {
            durationSeconds: 16,
            id: "asset_short_mina_rooftop",
            kind: "video",
            posterUrl: null,
            url: "https://cdn.example.com/shorts/mina-rooftop.mp4",
          },
          previewDurationSeconds: 16,
          routeShortId: "short_mina_rooftop",
        },
      ],
      stats: {
        fanCount: 24000,
        shortCount: 2,
      },
      viewer: {
        isFollowing: true,
      },
    });
    expect(mockedGetCreatorProfileHeader).toHaveBeenCalledWith({
      creatorId: "creator_mina_rei",
      sessionToken: "raw-session-token",
    });
  });

  it("builds the empty state when the profile has no public shorts", async () => {
    const creator = getCreatorById("creator_sora_vale");

    if (!creator) {
      throw new Error("fixture missing");
    }

    mockedGetCreatorProfileHeader.mockResolvedValue({
      creator,
      stats: {
        fanCount: 16000,
        shortCount: 0,
      },
      viewer: {
        isFollowing: false,
      },
    });
    mockedGetCreatorProfileShortGrid.mockResolvedValue({
      items: [],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_creator_profile_shorts_empty_001",
    });

    await expect(loadCreatorProfileShellState("creator_sora_vale")).resolves.toEqual({
      creator,
      kind: "empty",
      shorts: [],
      stats: {
        fanCount: 16000,
        shortCount: 0,
      },
      viewer: {
        isFollowing: false,
      },
    });
  });

  it("returns undefined when the backend reports not_found", async () => {
    mockedGetCreatorProfileHeader.mockRejectedValue(
      new ApiError("not found", {
        code: "http",
        status: 404,
      }),
    );

    await expect(loadCreatorProfileShellState("creator_missing")).resolves.toBeUndefined();
  });
});
