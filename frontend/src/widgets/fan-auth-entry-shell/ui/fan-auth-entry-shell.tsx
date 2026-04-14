"use client";

import { useRouter } from "next/navigation";

import {
  FanAuthEntryPanel,
  useFanAuthEntry,
} from "@/features/fan-auth";
import { Button } from "@/shared/ui";

/**
 * protected fan surface から到達する fan login entry を表示する。
 */
export function FanAuthEntryShell() {
  const router = useRouter();
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
  } = useFanAuthEntry();

  return (
    <main className="mx-auto flex min-h-svh w-full max-w-[408px] items-center px-4 py-10">
      <FanAuthEntryPanel
        avatar={avatar}
        avatarInputKey={avatarInputKey}
        clearAvatarSelection={clearAvatarSelection}
        displayName={displayName}
        dismissAction={(
          <Button
            className="w-full"
            disabled={isSubmitting}
            onClick={() => router.push("/")}
            type="button"
            variant="secondary"
          >
            feed に戻る
          </Button>
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
    </main>
  );
}
