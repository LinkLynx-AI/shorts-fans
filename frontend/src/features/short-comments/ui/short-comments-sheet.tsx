"use client";

import * as Dialog from "@radix-ui/react-dialog";
import type {
  FormEvent,
  ReactElement,
} from "react";
import {
  memo,
  useState,
} from "react";
import {
  LoaderCircle,
  Send,
  X,
} from "lucide-react";

import {
  createShortComment,
  getShortComments,
  type ShortComment,
} from "@/entities/short";
import { cn } from "@/shared/lib";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@/shared/ui";

import {
  countCommentBodyCharacters,
  limitCommentDraftInput,
  maxCommentBodyLength,
  useShortComments,
} from "../model/use-short-comments";

type CommentDraftState = {
  body: string;
  shortId: string;
};

type ShortCommentsSheetProps = {
  createCommentApi?: typeof createShortComment | undefined;
  fetchComments?: typeof getShortComments | undefined;
  hasViewerSession: boolean;
  onOpenChange?: ((open: boolean) => void) | undefined;
  onAuthRequired: () => void;
  open?: boolean | undefined;
  shortId: string;
  trigger: ReactElement;
};

const commentTimeFormatter = new Intl.DateTimeFormat("ja-JP", {
  day: "numeric",
  hour: "2-digit",
  minute: "2-digit",
  month: "numeric",
});

function getCommentAuthorInitials(displayName: string): string {
  const letters = Array.from(displayName.trim()).filter((letter) => letter.trim() !== "");

  if (letters.length === 0) {
    return "??";
  }

  return letters.slice(0, 2).join("").toUpperCase();
}

function formatCommentTime(createdAt: string): string {
  const date = new Date(createdAt);

  if (Number.isNaN(date.getTime())) {
    return "";
  }

  return commentTimeFormatter.format(date);
}

const ShortCommentRow = memo(function ShortCommentRow({ comment }: { comment: ShortComment }) {
  const createdAt = formatCommentTime(comment.createdAt);

  return (
    <article className="flex gap-3 py-3">
      <Avatar className="size-9 border-border bg-surface-subtle shadow-none">
        {comment.author.avatar ? (
          <AvatarImage alt="" src={comment.author.avatar.url} />
        ) : null}
        <AvatarFallback className="text-[11px] text-[#173252]">
          {getCommentAuthorInitials(comment.author.displayName)}
        </AvatarFallback>
      </Avatar>
      <div className="min-w-0 flex-1">
        <div className="flex min-w-0 items-center gap-2">
          <p className="truncate text-[13px] font-semibold text-foreground">{comment.author.displayName}</p>
          <p className="shrink-0 text-[12px] text-muted">{comment.author.handle}</p>
          {createdAt ? <time className="shrink-0 text-[12px] text-muted">{createdAt}</time> : null}
        </div>
        <p className="mt-1 whitespace-pre-wrap break-words text-[14px] leading-6 text-foreground">{comment.body}</p>
      </div>
    </article>
  );
});

function ShortCommentSkeleton() {
  return (
    <div aria-hidden="true" className="flex gap-3 py-3">
      <div className="size-9 shrink-0 rounded-full bg-surface-subtle" />
      <div className="min-w-0 flex-1 space-y-2 pt-1">
        <div className="h-3 w-32 rounded-full bg-surface-subtle" />
        <div className="h-3 w-full rounded-full bg-surface-subtle" />
        <div className="h-3 w-2/3 rounded-full bg-surface-subtle" />
      </div>
    </div>
  );
}

type ShortCommentListProps = {
  comments: readonly ShortComment[];
  errorMessage: string | null;
  hasNextPage: boolean;
  isLoadingInitial: boolean;
  isLoadingMore: boolean;
  onLoadMore: () => Promise<void>;
};

const ShortCommentList = memo(function ShortCommentList({
  comments,
  errorMessage,
  hasNextPage,
  isLoadingInitial,
  isLoadingMore,
  onLoadMore,
}: ShortCommentListProps) {
  return (
    <div className="min-h-0 flex-1 overflow-y-auto px-4">
      {isLoadingInitial ? (
        <div aria-label="コメントを読み込み中" role="status">
          {Array.from({ length: 5 }).map((_, index) => (
            <ShortCommentSkeleton key={index} />
          ))}
        </div>
      ) : comments.length > 0 ? (
        <div className="divide-y divide-border/70">
          {comments.map((comment) => (
            <ShortCommentRow comment={comment} key={comment.id} />
          ))}
        </div>
      ) : (
        <div className="flex h-full min-h-[220px] items-center justify-center text-center">
          <p className="text-sm font-medium text-muted">
            {errorMessage ?? "まだコメントはありません"}
          </p>
        </div>
      )}

      {comments.length > 0 && errorMessage ? (
        <p className="py-3 text-center text-[13px] font-medium text-[#b2394f]" role="alert">
          {errorMessage}
        </p>
      ) : null}

      {hasNextPage ? (
        <div className="flex justify-center py-3">
          <button
            className="inline-flex min-h-9 items-center rounded-full border border-border px-4 text-[13px] font-semibold text-foreground transition hover:bg-surface-subtle disabled:cursor-wait disabled:opacity-60"
            disabled={isLoadingMore}
            onClick={() => {
              void onLoadMore();
            }}
            type="button"
          >
            {isLoadingMore ? (
              <LoaderCircle className="mr-2 size-4 animate-spin" strokeWidth={2} />
            ) : null}
            さらに表示
          </button>
        </div>
      ) : null}
    </div>
  );
});

/**
 * short comment を TikTok/Reels 型の bottom sheet として表示する。
 */
export function ShortCommentsSheet({
  createCommentApi,
  fetchComments,
  hasViewerSession,
  onOpenChange,
  onAuthRequired,
  open,
  shortId,
  trigger,
}: ShortCommentsSheetProps) {
  const [draft, setDraft] = useState<CommentDraftState>({
    body: "",
    shortId,
  });
  const [uncontrolledOpen, setUncontrolledOpen] = useState(false);
  const isOpen = open ?? uncontrolledOpen;
  const setOpen = onOpenChange ?? setUncontrolledOpen;
  const {
    comments,
    errorMessage,
    hasNextPage,
    isLoadingInitial,
    isLoadingMore,
    isSubmitting,
    loadMore,
    submitComment,
    submitErrorMessage,
  } = useShortComments({
    ...(createCommentApi ? { createCommentApi } : {}),
    ...(fetchComments ? { fetchComments } : {}),
    hasViewerSession,
    onAuthRequired,
    open: isOpen,
    shortId,
  });
  const body = draft.shortId === shortId ? draft.body : "";
  const trimmedBodyLength = countCommentBodyCharacters(body);
  const canSubmit = trimmedBodyLength > 0 && trimmedBodyLength <= maxCommentBodyLength && !isSubmitting;

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const didSubmit = await submitComment(body);

    if (didSubmit) {
      setDraft({
        body: "",
        shortId,
      });
    }
  };

  return (
    <Dialog.Root onOpenChange={setOpen} open={isOpen}>
      <Dialog.Trigger asChild>{trigger}</Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-y-0 left-1/2 z-40 w-full max-w-[408px] -translate-x-1/2 bg-black/28 backdrop-blur-[2px]" />
        <Dialog.Content className="fixed bottom-0 left-1/2 z-50 flex h-[min(78vh,640px)] w-full max-w-[408px] -translate-x-1/2 flex-col overflow-hidden rounded-t-[26px] border border-border bg-white text-foreground shadow-[0_-20px_48px_rgba(15,23,42,0.2)]">
          <div className="px-4 pt-2">
            <div aria-hidden="true" className="mx-auto h-1 w-10 rounded-full bg-border-strong" />
            <div className="flex h-12 items-center justify-between">
              <div className="size-9" />
              <Dialog.Title className="text-[15px] font-bold text-foreground">コメント</Dialog.Title>
              <Dialog.Close asChild>
                <button
                  aria-label="Close comments"
                  className="inline-flex size-9 items-center justify-center rounded-full text-muted transition hover:bg-surface-subtle hover:text-foreground"
                  type="button"
                >
                  <X className="size-5" strokeWidth={2} />
                </button>
              </Dialog.Close>
            </div>
          </div>
          <Dialog.Description className="sr-only">
            short に投稿されたコメントを表示します。
          </Dialog.Description>

          <ShortCommentList
            comments={comments}
            errorMessage={errorMessage}
            hasNextPage={hasNextPage}
            isLoadingInitial={isLoadingInitial}
            isLoadingMore={isLoadingMore}
            onLoadMore={loadMore}
          />

          <form className="border-t border-border bg-white px-3 pb-[calc(12px+env(safe-area-inset-bottom,0px))] pt-3" onSubmit={handleSubmit}>
            {submitErrorMessage ? (
              <p className="mb-2 px-2 text-[12px] font-medium text-[#b2394f]" role="alert">
                {submitErrorMessage}
              </p>
            ) : null}
            <div className="flex items-end gap-2">
              <textarea
                aria-label="コメントを入力"
                className={cn(
                  "min-h-10 max-h-28 flex-1 resize-none rounded-[20px] border border-border bg-surface-subtle px-4 py-2.5 text-[14px] leading-5 text-foreground outline-none transition placeholder:text-muted",
                  "focus:border-accent focus:bg-white focus:ring-2 focus:ring-accent/20",
                )}
                onChange={(event) => {
                  setDraft({
                    body: limitCommentDraftInput(event.target.value),
                    shortId,
                  });
                }}
                placeholder="コメントを追加..."
                rows={1}
                value={body}
              />
              <button
                aria-label="コメントを投稿"
                className="inline-flex size-10 shrink-0 items-center justify-center rounded-full bg-foreground text-white transition hover:bg-foreground/88 disabled:cursor-default disabled:bg-muted/36"
                disabled={!canSubmit}
                type="submit"
              >
                {isSubmitting ? (
                  <LoaderCircle className="size-4 animate-spin" strokeWidth={2.2} />
                ) : (
                  <Send className="size-4" strokeWidth={2.2} />
                )}
              </button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
