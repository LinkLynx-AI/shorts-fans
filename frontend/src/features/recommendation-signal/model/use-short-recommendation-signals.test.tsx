import { act, renderHook } from "@testing-library/react";

import { useShortRecommendationSignals } from "./use-short-recommendation-signals";
import { createRecommendationSignalNonce, fireRecommendationSignal } from "./recommendation-signal";

vi.mock("./recommendation-signal", () => ({
  createRecommendationSignalIdempotencyKey: (...parts: string[]) => parts.join(":"),
  createRecommendationSignalNonce: vi.fn(),
  fireRecommendationSignal: vi.fn(),
  isRecommendationPublicCreatorId: vi.fn((value: string) => value.startsWith("creator_")),
  isRecommendationPublicShortId: vi.fn((value: string) => value.startsWith("short_")),
}));

const mockedCreateRecommendationSignalNonce = vi.mocked(createRecommendationSignalNonce);
const mockedFireRecommendationSignal = vi.mocked(fireRecommendationSignal);

describe("useShortRecommendationSignals", () => {
  beforeEach(() => {
    mockedCreateRecommendationSignalNonce.mockReset();
    mockedFireRecommendationSignal.mockReset();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("records impression, view start, completion, and rewatch loop signals for an active short session", () => {
    mockedCreateRecommendationSignalNonce.mockReturnValue("activation-1");

    const { result } = renderHook(() =>
      useShortRecommendationSignals({
        creatorId: "creator_11111111111111111111111111111111",
        isActive: true,
        shortId: "short_22222222222222222222222222222222",
        viewerId: "viewer_1",
      }),
    );

    expect(mockedFireRecommendationSignal).toHaveBeenNthCalledWith(1, {
      eventKind: "impression",
      idempotencyKey: "impression:short_22222222222222222222222222222222:activation-1",
      shortId: "short_22222222222222222222222222222222",
    });

    act(() => {
      result.current.handleVideoPlay();
      result.current.handleTimeUpdate(9.6, 10);
      result.current.handleTimeUpdate(0.5, 10);
    });

    expect(mockedFireRecommendationSignal).toHaveBeenNthCalledWith(2, {
      eventKind: "view_start",
      idempotencyKey: "view_start:short_22222222222222222222222222222222:activation-1",
      shortId: "short_22222222222222222222222222222222",
    });
    expect(mockedFireRecommendationSignal).toHaveBeenNthCalledWith(3, {
      eventKind: "view_completion",
      idempotencyKey: "view_completion:short_22222222222222222222222222222222:activation-1",
      shortId: "short_22222222222222222222222222222222",
    });
    expect(mockedFireRecommendationSignal).toHaveBeenNthCalledWith(4, {
      eventKind: "rewatch_loop",
      idempotencyKey: "rewatch_loop:short_22222222222222222222222222222222:activation-1:1",
      shortId: "short_22222222222222222222222222222222",
    });
  });

  it("suppresses loop recording immediately after a manual seek", () => {
    mockedCreateRecommendationSignalNonce.mockReturnValue("activation-1");
    const nowSpy = vi.spyOn(Date, "now");

    const { result } = renderHook(() =>
      useShortRecommendationSignals({
        creatorId: "creator_11111111111111111111111111111111",
        isActive: true,
        shortId: "short_22222222222222222222222222222222",
        viewerId: "viewer_1",
      }),
    );

    mockedFireRecommendationSignal.mockClear();
    nowSpy.mockReturnValue(1_000);
    act(() => {
      result.current.handleTimeUpdate(9.6, 10);
    });
    nowSpy.mockReturnValue(1_500);
    act(() => {
      result.current.markManualSeek();
    });
    nowSpy.mockReturnValue(2_000);
    act(() => {
      result.current.handleTimeUpdate(0.4, 10);
    });

    expect(mockedFireRecommendationSignal).toHaveBeenCalledWith({
      eventKind: "view_completion",
      idempotencyKey: "view_completion:short_22222222222222222222222222222222:activation-1",
      shortId: "short_22222222222222222222222222222222",
    });
    expect(
      mockedFireRecommendationSignal.mock.calls.some(
        ([input]) => input.eventKind === "rewatch_loop",
      ),
    ).toBe(false);
  });

  it("does not record signals when the viewer is signed out", () => {
    renderHook(() =>
      useShortRecommendationSignals({
        creatorId: "creator_11111111111111111111111111111111",
        isActive: true,
        shortId: "short_22222222222222222222222222222222",
        viewerId: null,
      }),
    );

    expect(mockedFireRecommendationSignal).not.toHaveBeenCalled();
  });

  it("waits for the surface exposure to be ready before starting a signed-in session", () => {
    mockedCreateRecommendationSignalNonce.mockReturnValue("activation-1");

    const { rerender } = renderHook(
      ({ isSurfaceReady }) =>
        useShortRecommendationSignals({
          creatorId: "creator_11111111111111111111111111111111",
          isActive: true,
          isSurfaceReady,
          shortId: "short_22222222222222222222222222222222",
          viewerId: "viewer_1",
        }),
      {
        initialProps: {
          isSurfaceReady: false,
        },
      },
    );

    expect(mockedFireRecommendationSignal).not.toHaveBeenCalled();

    rerender({
      isSurfaceReady: true,
    });

    expect(mockedFireRecommendationSignal).toHaveBeenCalledWith({
      eventKind: "impression",
      idempotencyKey: "impression:short_22222222222222222222222222222222:activation-1",
      shortId: "short_22222222222222222222222222222222",
    });
  });
});
