"use client";

import * as Dialog from "@radix-ui/react-dialog";

import { Button } from "@/shared/ui";

import { useCreatorWorkspaceShortCaption } from "../model/use-creator-workspace-short-caption";

type CreatorWorkspaceShortCaptionDialogProps = {
  initialCaption: string;
  onOpenChange: (open: boolean) => void;
  onSaved: () => void;
  open: boolean;
  shortId: string;
};

/**
 * creator workspace short caption 編集 modal を表示する。
 */
export function CreatorWorkspaceShortCaptionDialog({
  initialCaption,
  onOpenChange,
  onSaved,
  open,
  shortId,
}: CreatorWorkspaceShortCaptionDialogProps) {
  const {
    caption,
    errorMessage,
    isSubmitting,
    setCaption,
    submit,
  } = useCreatorWorkspaceShortCaption({
    initialCaption,
    onSaved,
    open,
    shortId,
  });

  return (
    <Dialog.Root
      onOpenChange={(nextOpen) => {
        if (isSubmitting && !nextOpen) {
          return;
        }

        onOpenChange(nextOpen);
      }}
      open={open}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-[#061521]/36 backdrop-blur-[2px]" />
        <Dialog.Content
          className="fixed inset-x-4 top-1/2 z-50 mx-auto w-full max-w-[408px] -translate-y-1/2"
          onEscapeKeyDown={(event) => {
            if (isSubmitting) {
              event.preventDefault();
            }
          }}
          onInteractOutside={(event) => {
            if (isSubmitting) {
              event.preventDefault();
            }
          }}
        >
          <div className="rounded-[30px] border border-white/72 bg-[rgba(255,255,255,0.9)] p-5 text-foreground shadow-[0_24px_60px_rgba(28,78,114,0.16)] backdrop-blur-xl">
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="m-0 text-[11px] font-bold uppercase tracking-[0.24em] text-accent">short</p>
                <Dialog.Title className="mt-2 font-display text-[26px] font-semibold leading-[1.08] tracking-[-0.04em]">
                  captionを変更
                </Dialog.Title>
                <Dialog.Description className="mt-2 text-sm leading-6 text-muted">
                  公開中のショートに表示するcaptionを更新します。空欄で保存するとcaptionを削除します。
                </Dialog.Description>
              </div>
            </div>

            <div className="mt-4 grid gap-2">
              <label
                className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-strong"
                htmlFor="creator-workspace-short-caption"
              >
                caption
              </label>
              <textarea
                aria-label="caption"
                className="min-h-[132px] w-full resize-none rounded-[22px] border border-[rgba(167,220,249,0.58)] bg-white/88 px-4 py-3 text-sm leading-6 text-foreground shadow-[inset_0_0_0_1px_rgba(255,255,255,0.42)] outline-none transition focus:border-accent focus:ring-4 focus:ring-[rgba(16,130,200,0.18)] disabled:cursor-default disabled:opacity-70"
                disabled={isSubmitting}
                id="creator-workspace-short-caption"
                onChange={(event) => {
                  setCaption(event.target.value);
                }}
                placeholder="captionを入力"
                value={caption}
              />
              <p className="m-0 text-xs leading-5 text-muted">
                空欄で保存すると、ショート詳細に表示するcaptionを削除します。
              </p>
              {errorMessage ? (
                <p className="m-0 text-sm leading-6 text-[#b2394f]" role="alert">
                  {errorMessage}
                </p>
              ) : null}
            </div>

            <div className="mt-5 flex gap-2.5">
              <Dialog.Close asChild>
                <Button className="flex-1" disabled={isSubmitting} variant="secondary">
                  閉じる
                </Button>
              </Dialog.Close>
              <Button
                className="flex-1"
                disabled={isSubmitting}
                onClick={() => {
                  void submit();
                }}
              >
                {isSubmitting ? "保存中..." : "保存する"}
              </Button>
            </div>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
