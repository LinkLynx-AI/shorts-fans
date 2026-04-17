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
      <div className="w-full rounded-[32px] bg-white px-5 py-8 shadow-[0_18px_40px_rgba(15,23,42,0.08)]">
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
              className="h-14 w-full text-[16px] font-bold"
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
      </div>
    </main>
  );
}
