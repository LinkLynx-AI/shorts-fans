import userEvent from "@testing-library/user-event";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import {
  createShortComment,
  getShortComments,
  ShortCommentApiError,
  type ShortComment,
} from "@/entities/short";
import { ApiError } from "@/shared/api";

import { ShortCommentsSheet } from "./short-comments-sheet";

const baseComment: ShortComment = {
  author: {
    avatar: null,
    displayName: "Kana Mori",
    handle: "@kanamori",
  },
  body: "hello",
  createdAt: "2026-01-02T14:05:00Z",
  id: "comment_22222222222222222222222222222222",
  shortId: "short_11111111111111111111111111111111",
};
const baseShortId = "short_11111111111111111111111111111111";
const nextShortId = "short_99999999999999999999999999999999";

function createDeferred<T>() {
  let resolve: ((value: T) => void) | undefined;
  let reject: ((reason?: unknown) => void) | undefined;
  const promise = new Promise<T>((nextResolve, nextReject) => {
    resolve = nextResolve;
    reject = nextReject;
  });

  if (!resolve || !reject) {
    throw new Error("deferred callbacks were not initialized");
  }

  return {
    promise,
    reject,
    resolve,
  };
}

function renderSheet({
  createCommentApi = vi.fn<typeof createShortComment>().mockResolvedValue(baseComment),
  fetchComments = vi.fn<typeof getShortComments>().mockResolvedValue({
    items: [baseComment],
    page: {
      hasNext: false,
      nextCursor: null,
    },
    requestId: "req_short_comments_001",
  }),
  hasViewerSession = true,
  onAuthRequired = vi.fn(),
  shortId = baseShortId,
}: {
  createCommentApi?: typeof createShortComment;
  fetchComments?: typeof getShortComments;
  hasViewerSession?: boolean;
  onAuthRequired?: () => void;
  shortId?: string;
} = {}) {
  const renderSheetElement = (currentShortId: string) => (
    <ShortCommentsSheet
      createCommentApi={createCommentApi}
      fetchComments={fetchComments}
      hasViewerSession={hasViewerSession}
      onAuthRequired={onAuthRequired}
      shortId={currentShortId}
      trigger={<button type="button">open comments</button>}
    />
  );
  const renderResult = render(renderSheetElement(shortId));

  return {
    createCommentApi,
    fetchComments,
    onAuthRequired,
    rerenderShortId: (currentShortId: string) => {
      renderResult.rerender(renderSheetElement(currentShortId));
    },
  };
}

describe("ShortCommentsSheet", () => {
  it("loads comments when opened", async () => {
    const user = userEvent.setup();
    const { fetchComments } = renderSheet();

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(fetchComments).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "open comments" }));

    expect(await screen.findByText("hello")).toBeInTheDocument();
    expect(screen.getByText("@kanamori")).toBeInTheDocument();
    expect(fetchComments).toHaveBeenCalledWith({
      shortId: "short_11111111111111111111111111111111",
      signal: expect.any(AbortSignal),
    });
  });

  it("shows a not-found message when the initial comment load cannot display the short", async () => {
    const user = userEvent.setup();
    const fetchComments = vi.fn<typeof getShortComments>().mockRejectedValue(
      new ShortCommentApiError("not_found", "short was not found", {
        requestId: "req_short_comments_not_found_001",
        status: 404,
      }),
    );

    renderSheet({ fetchComments });

    await user.click(screen.getByRole("button", { name: "open comments" }));

    expect(await screen.findByText("この short は表示できません。")).toBeInTheDocument();
    expect(screen.queryByText("まだコメントはありません")).not.toBeInTheDocument();
  });

  it("shows a generic message when the initial comment load fails", async () => {
    const user = userEvent.setup();
    const fetchComments = vi.fn<typeof getShortComments>().mockRejectedValue(new Error("network down"));

    renderSheet({ fetchComments });

    await user.click(screen.getByRole("button", { name: "open comments" }));

    expect(await screen.findByText("コメントを読み込めませんでした。")).toBeInTheDocument();
    expect(screen.queryByText("まだコメントはありません")).not.toBeInTheDocument();
  });

  it("shows the empty state and keeps the composer usable when the public short has no comments", async () => {
    const user = userEvent.setup();
    const fetchComments = vi.fn<typeof getShortComments>().mockResolvedValue({
      items: [],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_short_comments_empty_001",
    });

    renderSheet({ fetchComments });

    await user.click(screen.getByRole("button", { name: "open comments" }));

    expect(await screen.findByText("まだコメントはありません")).toBeInTheDocument();
    expect(screen.getByLabelText("コメントを入力")).toBeEnabled();
  });

  it("ignores a stale initial load after the open sheet moves to another short", async () => {
    const user = userEvent.setup();
    const initialPage = createDeferred<Awaited<ReturnType<typeof getShortComments>>>();
    const nextShortComment = {
      ...baseComment,
      body: "next short comment",
      id: "comment_88888888888888888888888888888888",
      shortId: nextShortId,
    };
    const fetchComments = vi.fn<typeof getShortComments>()
      .mockReturnValueOnce(initialPage.promise)
      .mockResolvedValueOnce({
        items: [nextShortComment],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_comments_next_001",
      });
    const { rerenderShortId } = renderSheet({ fetchComments });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    await waitFor(() => {
      expect(fetchComments).toHaveBeenCalledWith({
        shortId: baseShortId,
        signal: expect.any(AbortSignal),
      });
    });

    rerenderShortId(nextShortId);
    expect(await screen.findByText("next short comment")).toBeInTheDocument();

    await act(async () => {
      initialPage.resolve({
        items: [baseComment],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_comments_stale_001",
      });
      await initialPage.promise;
    });

    expect(screen.queryByText("hello", { selector: "article p" })).not.toBeInTheDocument();
  });

  it("loads the next cursor page", async () => {
    const user = userEvent.setup();
    const secondComment = {
      ...baseComment,
      body: "second",
      id: "comment_33333333333333333333333333333333",
    };
    const fetchComments = vi.fn<typeof getShortComments>()
      .mockResolvedValueOnce({
        items: [baseComment],
        page: {
          hasNext: true,
          nextCursor: "cursor_next_001",
        },
        requestId: "req_short_comments_001",
      })
      .mockResolvedValueOnce({
        items: [secondComment],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_comments_002",
      });

    renderSheet({ fetchComments });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "さらに表示" }));

    expect(await screen.findByText("second")).toBeInTheDocument();
    expect(fetchComments).toHaveBeenLastCalledWith({
      cursor: "cursor_next_001",
      signal: expect.any(AbortSignal),
      shortId: "short_11111111111111111111111111111111",
    });
  });

  it("ignores duplicate load-more presses for the same cursor while a request is in flight", async () => {
    const user = userEvent.setup();
    const loadMorePage = createDeferred<Awaited<ReturnType<typeof getShortComments>>>();
    const fetchComments = vi.fn<typeof getShortComments>()
      .mockResolvedValueOnce({
        items: [baseComment],
        page: {
          hasNext: true,
          nextCursor: "cursor_next_001",
        },
        requestId: "req_short_comments_001",
      })
      .mockReturnValueOnce(loadMorePage.promise);

    renderSheet({ fetchComments });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    const loadMoreButton = screen.getByRole("button", { name: "さらに表示" });
    fireEvent.click(loadMoreButton);
    fireEvent.click(loadMoreButton);

    expect(fetchComments).toHaveBeenCalledTimes(2);
    expect(fetchComments).toHaveBeenLastCalledWith({
      cursor: "cursor_next_001",
      signal: expect.any(AbortSignal),
      shortId: "short_11111111111111111111111111111111",
    });

    await act(async () => {
      loadMorePage.resolve({
        items: [],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_comments_002",
      });
      await loadMorePage.promise;
    });
  });

  it("ignores a stale load-more success after the open sheet moves to another short", async () => {
    const user = userEvent.setup();
    const loadMorePage = createDeferred<Awaited<ReturnType<typeof getShortComments>>>();
    const staleComment = {
      ...baseComment,
      body: "stale load more comment",
      id: "comment_99999999999999999999999999999999",
    };
    const nextShortComment = {
      ...baseComment,
      body: "next short comment",
      id: "comment_88888888888888888888888888888888",
      shortId: nextShortId,
    };
    const fetchComments = vi.fn<typeof getShortComments>()
      .mockResolvedValueOnce({
        items: [baseComment],
        page: {
          hasNext: true,
          nextCursor: "cursor_next_001",
        },
        requestId: "req_short_comments_001",
      })
      .mockReturnValueOnce(loadMorePage.promise)
      .mockResolvedValueOnce({
        items: [nextShortComment],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_comments_next_001",
      });
    const { rerenderShortId } = renderSheet({ fetchComments });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "さらに表示" }));
    expect(fetchComments).toHaveBeenLastCalledWith({
      cursor: "cursor_next_001",
      signal: expect.any(AbortSignal),
      shortId: baseShortId,
    });

    rerenderShortId(nextShortId);
    expect(await screen.findByText("next short comment")).toBeInTheDocument();

    await act(async () => {
      loadMorePage.resolve({
        items: [staleComment],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_comments_stale_002",
      });
      await loadMorePage.promise;
    });

    expect(screen.queryByText("stale load more comment", { selector: "article p" })).not.toBeInTheDocument();
  });

  it("ignores stale load-more errors after the sheet is reopened", async () => {
    const user = userEvent.setup();
    const loadMorePage = createDeferred<Awaited<ReturnType<typeof getShortComments>>>();
    const fetchComments = vi.fn<typeof getShortComments>()
      .mockResolvedValueOnce({
        items: [baseComment],
        page: {
          hasNext: true,
          nextCursor: "cursor_next_001",
        },
        requestId: "req_short_comments_001",
      })
      .mockReturnValueOnce(loadMorePage.promise)
      .mockResolvedValueOnce({
        items: [baseComment],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_comments_003",
      });

    renderSheet({ fetchComments });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "さらに表示" }));
    expect(fetchComments).toHaveBeenLastCalledWith({
      cursor: "cursor_next_001",
      signal: expect.any(AbortSignal),
      shortId: "short_11111111111111111111111111111111",
    });
    const loadMoreSignal = fetchComments.mock.calls[1]?.[0].signal;

    await user.click(screen.getByRole("button", { name: "Close comments" }));
    expect(loadMoreSignal?.aborted).toBe(true);
    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    await act(async () => {
      loadMorePage.reject(new Error("late failure"));
      await expect(loadMorePage.promise).rejects.toThrow("late failure");
    });

    expect(screen.queryByText("コメントを読み込めませんでした。")).not.toBeInTheDocument();
  });

  it("suppresses aborted load-more errors when the sheet closes", async () => {
    const user = userEvent.setup();
    const fetchComments = vi.fn<typeof getShortComments>()
      .mockResolvedValueOnce({
        items: [baseComment],
        page: {
          hasNext: true,
          nextCursor: "cursor_next_001",
        },
        requestId: "req_short_comments_001",
      })
      .mockImplementationOnce(({ signal }) => new Promise((_, reject) => {
        signal?.addEventListener("abort", () => {
          reject(new ApiError("aborted", {
            cause: new DOMException("aborted", "AbortError"),
            code: "network",
          }));
        }, { once: true });
      }))
      .mockResolvedValueOnce({
        items: [baseComment],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_comments_003",
      });

    renderSheet({ fetchComments });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "さらに表示" }));
    await user.click(screen.getByRole("button", { name: "Close comments" }));

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();
    expect(screen.queryByText("コメントを読み込めませんでした。")).not.toBeInTheDocument();
  });

  it("keeps a created comment when the initial page resolves later", async () => {
    const user = userEvent.setup();
    const initialPage = createDeferred<Awaited<ReturnType<typeof getShortComments>>>();
    const createdComment = {
      ...baseComment,
      body: "new comment",
      id: "comment_44444444444444444444444444444444",
    };
    const fetchComments = vi.fn<typeof getShortComments>().mockReturnValueOnce(initialPage.promise);
    const createCommentApi = vi.fn<typeof createShortComment>().mockResolvedValue(createdComment);

    renderSheet({
      createCommentApi,
      fetchComments,
    });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    await user.type(screen.getByLabelText("コメントを入力"), "new comment");
    await user.click(screen.getByRole("button", { name: "コメントを投稿" }));
    await waitFor(() => {
      expect(createCommentApi).toHaveBeenCalledWith({
        body: "new comment",
        signal: expect.any(AbortSignal),
        shortId: "short_11111111111111111111111111111111",
      });
    });

    await act(async () => {
      initialPage.resolve({
        items: [baseComment],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_comments_001",
      });
      await initialPage.promise;
    });

    expect(await screen.findByText("new comment")).toBeInTheDocument();
    expect(screen.getByText("hello")).toBeInTheDocument();
  });

  it("submits a trimmed comment and prepends the created row", async () => {
    const user = userEvent.setup();
    const createdComment = {
      ...baseComment,
      body: "new comment",
      id: "comment_44444444444444444444444444444444",
    };
    const createCommentApi = vi.fn<typeof createShortComment>().mockResolvedValue(createdComment);

    renderSheet({ createCommentApi });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    await user.type(screen.getByLabelText("コメントを入力"), "  new comment  ");
    await user.click(screen.getByRole("button", { name: "コメントを投稿" }));

    expect(await screen.findByText("new comment")).toBeInTheDocument();
    expect(createCommentApi).toHaveBeenCalledWith({
      body: "new comment",
      signal: expect.any(AbortSignal),
      shortId: "short_11111111111111111111111111111111",
    });
    await waitFor(() => {
      expect(screen.getByLabelText("コメントを入力")).toHaveValue("");
    });
  });

  it("ignores a stale create response after the sheet closes and reopens", async () => {
    const user = userEvent.setup();
    const createResult = createDeferred<Awaited<ReturnType<typeof createShortComment>>>();
    const createdComment = {
      ...baseComment,
      body: "late comment",
      id: "comment_55555555555555555555555555555555",
    };
    const createCommentApi = vi.fn<typeof createShortComment>().mockReturnValueOnce(createResult.promise);

    renderSheet({ createCommentApi });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    await user.type(screen.getByLabelText("コメントを入力"), "late comment");
    await user.click(screen.getByRole("button", { name: "コメントを投稿" }));

    await waitFor(() => {
      expect(createCommentApi).toHaveBeenCalledWith({
        body: "late comment",
        signal: expect.any(AbortSignal),
        shortId: "short_11111111111111111111111111111111",
      });
    });
    const submitSignal = createCommentApi.mock.calls[0]?.[0].signal;

    await user.click(screen.getByRole("button", { name: "Close comments" }));
    expect(submitSignal?.aborted).toBe(true);
    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    await act(async () => {
      createResult.resolve(createdComment);
      await createResult.promise;
    });

    expect(screen.queryByText("late comment", { selector: "article p" })).not.toBeInTheDocument();
  });

  it("ignores a stale create response after the open sheet moves to another short", async () => {
    const user = userEvent.setup();
    const createResult = createDeferred<Awaited<ReturnType<typeof createShortComment>>>();
    const nextShortComment = {
      ...baseComment,
      body: "next short comment",
      id: "comment_66666666666666666666666666666666",
      shortId: nextShortId,
    };
    const staleCreatedComment = {
      ...baseComment,
      body: "stale short comment",
      id: "comment_77777777777777777777777777777777",
    };
    const fetchComments = vi.fn<typeof getShortComments>()
      .mockResolvedValueOnce({
        items: [baseComment],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_comments_001",
      })
      .mockResolvedValueOnce({
        items: [nextShortComment],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        requestId: "req_short_comments_002",
      });
    const createCommentApi = vi.fn<typeof createShortComment>().mockReturnValueOnce(createResult.promise);
    const { rerenderShortId } = renderSheet({
      createCommentApi,
      fetchComments,
    });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    await user.type(screen.getByLabelText("コメントを入力"), "stale short comment");
    await user.click(screen.getByRole("button", { name: "コメントを投稿" }));
    await waitFor(() => {
      expect(createCommentApi).toHaveBeenCalledWith({
        body: "stale short comment",
        signal: expect.any(AbortSignal),
        shortId: baseShortId,
      });
    });

    rerenderShortId(nextShortId);
    expect(await screen.findByText("next short comment")).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByLabelText("コメントを入力")).toHaveValue("");
    });

    await act(async () => {
      createResult.resolve(staleCreatedComment);
      await createResult.promise;
    });

    expect(screen.queryByText("stale short comment", { selector: "article p" })).not.toBeInTheDocument();
  });

  it("does not apply a native maxLength because the API limit is character-based", async () => {
    const user = userEvent.setup();

    renderSheet();

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();
    expect(screen.getByLabelText("コメントを入力")).not.toHaveAttribute("maxlength");
  });

  it("caps pasted draft input before it enters component state", async () => {
    const user = userEvent.setup();

    renderSheet();

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("コメントを入力"), {
      target: {
        value: "a".repeat(1_000),
      },
    });

    expect(screen.getByLabelText("コメントを入力")).toHaveValue("a".repeat(600));
  });

  it("opens auth instead of posting for signed-out viewers", async () => {
    const user = userEvent.setup();
    const createCommentApi = vi.fn<typeof createShortComment>().mockResolvedValue(baseComment);
    const onAuthRequired = vi.fn();

    renderSheet({
      createCommentApi,
      hasViewerSession: false,
      onAuthRequired,
    });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    await user.type(screen.getByLabelText("コメントを入力"), "new comment");
    await user.click(screen.getByRole("button", { name: "コメントを投稿" }));

    expect(onAuthRequired).toHaveBeenCalledTimes(1);
    expect(createCommentApi).not.toHaveBeenCalled();
  });

  it("opens auth when posting returns auth_required for an expired session", async () => {
    const user = userEvent.setup();
    const createCommentApi = vi.fn<typeof createShortComment>().mockRejectedValue(
      new ShortCommentApiError("auth_required", "sign in required", {
        requestId: "req_short_comment_auth_001",
        status: 401,
      }),
    );
    const onAuthRequired = vi.fn();

    renderSheet({
      createCommentApi,
      onAuthRequired,
    });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    await user.type(screen.getByLabelText("コメントを入力"), "new comment");
    await user.click(screen.getByRole("button", { name: "コメントを投稿" }));

    await waitFor(() => {
      expect(onAuthRequired).toHaveBeenCalledTimes(1);
    });
    expect(createCommentApi).toHaveBeenCalledWith({
      body: "new comment",
      signal: expect.any(AbortSignal),
      shortId: "short_11111111111111111111111111111111",
    });
  });

  it("keeps the draft and shows validation feedback when posting is rejected", async () => {
    const user = userEvent.setup();
    const createCommentApi = vi.fn<typeof createShortComment>().mockRejectedValue(
      new ShortCommentApiError("validation_error", "comment body is invalid", {
        requestId: "req_short_comment_validation_001",
        status: 400,
      }),
    );

    renderSheet({ createCommentApi });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    await user.type(screen.getByLabelText("コメントを入力"), "invalid draft");
    await user.click(screen.getByRole("button", { name: "コメントを投稿" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("1〜500文字で入力してください。");
    expect(screen.getByLabelText("コメントを入力")).toHaveValue("invalid draft");
  });

  it("keeps the draft and shows not-found feedback when the short can no longer accept comments", async () => {
    const user = userEvent.setup();
    const createCommentApi = vi.fn<typeof createShortComment>().mockRejectedValue(
      new ShortCommentApiError("not_found", "short was not found", {
        requestId: "req_short_comment_not_found_001",
        status: 404,
      }),
    );

    renderSheet({ createCommentApi });

    await user.click(screen.getByRole("button", { name: "open comments" }));
    expect(await screen.findByText("hello")).toBeInTheDocument();

    await user.type(screen.getByLabelText("コメントを入力"), "still here");
    await user.click(screen.getByRole("button", { name: "コメントを投稿" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("この short にはコメントできません。");
    expect(screen.getByLabelText("コメントを入力")).toHaveValue("still here");
  });
});
