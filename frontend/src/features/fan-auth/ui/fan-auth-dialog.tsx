"use client";

import * as Dialog from "@radix-ui/react-dialog";

import { Button } from "@/shared/ui";

import { useFanAuthEntry } from "../model/use-fan-auth-entry";
import { FanAuthEntryPanel } from "./fan-auth-entry-panel";

type FanAuthDialogProps = {
  onAuthenticated: () => Promise<string | null> | string | null;
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
  const {
    avatar,
    avatarInputKey,
    clearAvatarSelection,
    displayName,
    email,
    errorMessage,
    handle,
    isSubmitting,
    mode,
    selectAvatarFile,
    setDisplayName,
    setEmail,
    setHandle,
    submit,
    switchMode,
  } = useFanAuthEntry({
    onAuthenticated,
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
          className="fixed inset-x-4 top-1/2 z-50 mx-auto w-full max-w-[376px] -translate-y-1/2"
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
          <Dialog.Title className="sr-only">続けるにはログインが必要です</Dialog.Title>
          <Dialog.Description className="sr-only">
            email で sign in または sign up を始める modal
          </Dialog.Description>
          <FanAuthEntryPanel
            avatar={avatar}
            avatarInputKey={avatarInputKey}
            clearAvatarSelection={clearAvatarSelection}
            displayName={displayName}
            dismissAction={(
              <Dialog.Close asChild>
                <Button className="w-full" disabled={isSubmitting} variant="secondary">
                  閉じる
                </Button>
              </Dialog.Close>
            )}
            email={email}
            errorMessage={errorMessage}
            handle={handle}
            isSubmitting={isSubmitting}
            mode={mode}
            onAvatarSelect={selectAvatarFile}
            onDisplayNameChange={setDisplayName}
            onEmailChange={setEmail}
            onHandleChange={setHandle}
            onModeSwitch={switchMode}
            onSubmit={submit}
          />
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
