import {
  getCreatorWorkspacePreviewMains,
  getCreatorWorkspacePreviewShorts,
} from "../api/get-creator-workspace-preview-collections";
import { loadCreatorWorkspacePreviewCollections } from "./load-creator-workspace-preview-collections";

vi.mock("../api/get-creator-workspace-preview-collections", () => ({
  getCreatorWorkspacePreviewMains: vi.fn(),
  getCreatorWorkspacePreviewShorts: vi.fn(),
}));

describe("loadCreatorWorkspacePreviewCollections", () => {
  beforeEach(() => {
    vi.mocked(getCreatorWorkspacePreviewMains).mockReset();
    vi.mocked(getCreatorWorkspacePreviewShorts).mockReset();
  });

  it("loads every available page for shorts and mains", async () => {
    vi.mocked(getCreatorWorkspacePreviewShorts)
      .mockResolvedValueOnce({
        items: [
          {
            canonicalMainId: "main_1",
            id: "short_1",
            media: {
              durationSeconds: 16,
              id: "asset_short_1",
              kind: "video",
              posterUrl: "https://cdn.example.com/shorts/1.jpg",
            },
            previewDurationSeconds: 16,
          },
        ],
        page: {
          hasNext: true,
          nextCursor: "short_cursor_2",
        },
        requestId: "req_short_1",
      })
      .mockResolvedValueOnce({
        items: [
          {
            canonicalMainId: "main_2",
            id: "short_2",
            media: {
              durationSeconds: 18,
              id: "asset_short_2",
              kind: "video",
              posterUrl: "https://cdn.example.com/shorts/2.jpg",
            },
            previewDurationSeconds: 18,
          },
        ],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_2",
      });

    vi.mocked(getCreatorWorkspacePreviewMains)
      .mockResolvedValueOnce({
        items: [
          {
            durationSeconds: 720,
            id: "main_1",
            leadShortId: "short_1",
            media: {
              durationSeconds: 720,
              id: "asset_main_1",
              kind: "video",
              posterUrl: "https://cdn.example.com/mains/1.jpg",
            },
            priceJpy: 1800,
          },
        ],
        page: {
          hasNext: true,
          nextCursor: "main_cursor_2",
        },
        requestId: "req_main_1",
      })
      .mockResolvedValueOnce({
        items: [
          {
            durationSeconds: 840,
            id: "main_2",
            leadShortId: "short_2",
            media: {
              durationSeconds: 840,
              id: "asset_main_2",
              kind: "video",
              posterUrl: "https://cdn.example.com/mains/2.jpg",
            },
            priceJpy: 2200,
          },
        ],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_main_2",
      });

    await expect(loadCreatorWorkspacePreviewCollections()).resolves.toEqual({
      mains: {
        items: [
          expect.objectContaining({ id: "main_1" }),
          expect.objectContaining({ id: "main_2" }),
        ],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_main_1",
      },
      shorts: {
        items: [
          expect.objectContaining({ id: "short_1" }),
          expect.objectContaining({ id: "short_2" }),
        ],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_1",
      },
    });

    expect(getCreatorWorkspacePreviewShorts).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({
        credentials: "include",
      }),
    );
    expect(getCreatorWorkspacePreviewShorts).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({
        credentials: "include",
        cursor: "short_cursor_2",
      }),
    );
    expect(getCreatorWorkspacePreviewMains).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({
        credentials: "include",
      }),
    );
    expect(getCreatorWorkspacePreviewMains).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({
        credentials: "include",
        cursor: "main_cursor_2",
      }),
    );
  });
});
