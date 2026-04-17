"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { useState } from "react";

import { Button } from "@/shared/ui";

import { type FanAuthMode } from "../model/fan-auth";
import { useFanAuthEntry } from "../model/use-fan-auth-entry";
import { FanAuthEntryPanel } from "./fan-auth-entry-panel";

type FanAuthDialogOnAuthenticatedOptions = {
  afterViewerSynced?: () => Promise<string | null> | string | null;
  authenticatedMode?: FanAuthMode;
};

type FanAuthDialogProps = {
  allowClose?: boolean;
  initialMode?: FanAuthMode;
  onAuthenticated: (
    options?: FanAuthDialogOnAuthenticatedOptions,
  ) => Promise<string | null> | string | null;
  onFallbackToSignIn?: (() => void) | undefined;
  onOpenChange: (open: boolean) => void;
  open: boolean;
  sessionKey: number;
};

function FanAuthDialogBody({
  allowClose,
  initialMode,
  onAuthenticated,
  onFallbackToSignIn,
  onSubmittingChange,
  submitLockActive,
}: Pick<FanAuthDialogProps, "allowClose" | "initialMode" | "onAuthenticated" | "onFallbackToSignIn"> & {
  onSubmittingChange: (isSubmitting: boolean) => void;
  submitLockActive: boolean;
}) {
  const {
    avatar,
    avatarInputKey,
    canResend,
    clearAvatarSelection,
    confirmationCode,
    deliveryDestinationHint,
    displayName,
    email,
    errorMessage,
    handle,
    hasConfirmedSignUp,
    infoMessage,
    isSubmitting,
    mode,
    newPassword,
    password,
    resend,
    selectAvatarFile,
    setConfirmationCode,
    setDisplayName,
    setEmail,
    setHandle,
    setMode,
    setNewPassword,
    setPassword,
    submit,
  } = useFanAuthEntry({
    onFallbackToSignIn,
    onAuthenticated,
    ...(initialMode
      ? {
          initialMode,
        }
      : {}),
  });

  const handleSubmit = async () => {
    onSubmittingChange(true);

    try {
      await submit();
    } finally {
      onSubmittingChange(false);
    }
  };

  const handleResend = async () => {
    onSubmittingChange(true);

    try {
      await resend();
    } finally {
      onSubmittingChange(false);
    }
  };

  return (
    <FanAuthEntryPanel
      avatar={avatar}
      avatarInputKey={avatarInputKey}
      canResend={canResend}
      clearAvatarSelection={clearAvatarSelection}
      confirmationCode={confirmationCode}
      deliveryDestinationHint={deliveryDestinationHint}
      dismissAction={
        allowClose ? (
          <Dialog.Close asChild>
            <Button
              className="h-14 w-full text-[16px] font-bold"
              disabled={isSubmitting || submitLockActive}
              variant="secondary"
            >
              閉じる
            </Button>
          </Dialog.Close>
        ) : undefined
      }
      displayName={displayName}
      email={email}
      errorMessage={errorMessage}
      handle={handle}
      hasConfirmedSignUp={hasConfirmedSignUp}
      infoMessage={infoMessage}
      isSubmitting={isSubmitting}
      mode={mode}
      newPassword={newPassword}
      password={password}
      onAvatarSelect={selectAvatarFile}
      onConfirmationCodeChange={setConfirmationCode}
      onDisplayNameChange={setDisplayName}
      onEmailChange={setEmail}
      onHandleChange={setHandle}
      onModeChange={setMode}
      onNewPasswordChange={setNewPassword}
      onPasswordChange={setPassword}
      onResend={handleResend}
      onSubmit={handleSubmit}
    />
  );
}

/**
 * 共通 fan auth modal を表示する。
 */
export function FanAuthDialog({
  allowClose = true,
  initialMode = "sign-in",
  onAuthenticated,
  onFallbackToSignIn,
  onOpenChange,
  open,
  sessionKey,
}: FanAuthDialogProps) {
  const [isDismissLocked, setIsDismissLocked] = useState(false);
  const canClose = allowClose && !isDismissLocked;

  return (
    <Dialog.Root
      onOpenChange={(nextOpen) => {
        if (!canClose && !nextOpen) {
          return;
        }

        onOpenChange(nextOpen);
      }}
      open={open}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-y-0 left-1/2 z-40 w-full max-w-[408px] -translate-x-1/2 bg-[#061521]/52 backdrop-blur-[2px]" />
        <Dialog.Content
          className="fixed inset-x-0 bottom-0 left-1/2 z-50 flex w-full max-w-[408px] -translate-x-1/2 justify-center outline-none"
          onEscapeKeyDown={(event) => {
            if (!canClose) {
              event.preventDefault();
            }
          }}
          onInteractOutside={(event) => {
            if (!canClose) {
              event.preventDefault();
            }
          }}
        >
          <Dialog.Title className="sr-only">続けるには認証が必要です</Dialog.Title>
          <Dialog.Description className="sr-only">
            email と password を中心に fan auth を完了する shared modal
          </Dialog.Description>
          <div className="w-full">
            <div className="max-h-[90svh] w-full overflow-y-auto overscroll-contain rounded-t-[32px] bg-white px-5 pb-10 pt-4 shadow-[0_-18px_42px_rgba(15,23,42,0.18)]">
              <div className="mx-auto mb-5 h-1.5 w-12 rounded-full bg-[#dde5ef]" />
              <FanAuthDialogBody
                allowClose={allowClose}
                initialMode={initialMode}
                key={sessionKey}
                onAuthenticated={onAuthenticated}
                onFallbackToSignIn={onFallbackToSignIn}
                onSubmittingChange={setIsDismissLocked}
                submitLockActive={isDismissLocked}
              />
            </div>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
