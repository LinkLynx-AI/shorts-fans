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
  } = useFanAuthEntry();

  return (
    <main className="mx-auto flex min-h-svh w-full max-w-[408px] items-center px-4 py-10">
      <FanAuthEntryPanel
        avatar={avatar}
        avatarInputKey={avatarInputKey}
        canResend={canResend}
        clearAvatarSelection={clearAvatarSelection}
        confirmationCode={confirmationCode}
        deliveryDestinationHint={deliveryDestinationHint}
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
        hasConfirmedSignUp={hasConfirmedSignUp}
        infoMessage={infoMessage}
        isSubmitting={isSubmitting}
        mode={mode}
        newPassword={newPassword}
        onAvatarSelect={selectAvatarFile}
        onConfirmationCodeChange={setConfirmationCode}
        onDisplayNameChange={setDisplayName}
        onEmailChange={setEmail}
        onHandleChange={setHandle}
        onModeChange={setMode}
        onNewPasswordChange={setNewPassword}
        onPasswordChange={setPassword}
        onResend={resend}
        onSubmit={submit}
        password={password}
      />
    </main>
  );
}
