"use client";

import * as Dialog from "@radix-ui/react-dialog";

import { Button } from "@/shared/ui";

import { FanAuthEntryPanel } from "./fan-auth-entry-panel";

type FanAuthDialogProps = {
  onAuthenticated: () => void;
  onOpenChange: (open: boolean) => void;
  open: boolean;
};

/**
 * 共通 fan auth modal を表示する。
 */
export function FanAuthDialog({
  onAuthenticated,
  onOpenChange,
  open,
}: FanAuthDialogProps) {
  return (
    <Dialog.Root onOpenChange={onOpenChange} open={open}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-[#061521]/36 backdrop-blur-[2px]" />
        <Dialog.Content className="fixed inset-x-4 top-1/2 z-50 mx-auto w-full max-w-[376px] -translate-y-1/2">
          <Dialog.Title className="sr-only">続けるにはログインが必要です</Dialog.Title>
          <Dialog.Description className="sr-only">
            email で sign in または sign up を始める modal
          </Dialog.Description>
          <FanAuthEntryPanel
            dismissAction={
              <Dialog.Close asChild>
                <Button className="w-full" variant="secondary">
                  閉じる
                </Button>
              </Dialog.Close>
            }
            onAuthenticated={onAuthenticated}
          />
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
