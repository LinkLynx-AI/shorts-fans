"use client";

import type {
  ComponentProps,
  ReactNode,
} from "react";

import { SharedViewerProfileFields } from "@/features/viewer-profile";
import { Button, SurfacePanel } from "@/shared/ui";

import {
  getFanAuthModeDescription,
  getFanAuthModeTitle,
  getFanAuthSubmitLabel,
  type FanAuthMode,
} from "../model/fan-auth";

type FanAuthEntryPanelProps = {
  avatar: ComponentProps<typeof SharedViewerProfileFields>["avatar"];
  avatarInputKey: number;
  canResend: boolean;
  clearAvatarSelection: () => void;
  confirmationCode: string;
  deliveryDestinationHint: string | null;
  dismissAction?: ReactNode;
  displayName: string;
  email: string;
  errorMessage: string | null;
  handle: string;
  hasConfirmedSignUp: boolean;
  infoMessage: string | null;
  isSubmitting: boolean;
  mode: FanAuthMode;
  newPassword: string;
  password: string;
  onAvatarSelect: (file: File | null) => void;
  onConfirmationCodeChange: (confirmationCode: string) => void;
  onDisplayNameChange: (displayName: string) => void;
  onEmailChange: (email: string) => void;
  onHandleChange: (handle: string) => void;
  onModeChange: (mode: FanAuthMode) => void;
  onNewPasswordChange: (newPassword: string) => void;
  onPasswordChange: (password: string) => void;
  onResend: () => void | Promise<void>;
  onSubmit: () => void | Promise<void>;
};

function ModeSwitchButton({
  disabled,
  label,
  onClick,
}: {
  disabled: boolean;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      className="font-semibold text-accent-strong transition hover:text-accent disabled:cursor-not-allowed disabled:opacity-60"
      disabled={disabled}
      onClick={onClick}
      type="button"
    >
      {label}
    </button>
  );
}

function FanAuthEmailField({
  disabled,
  onChange,
  value,
}: {
  disabled: boolean;
  onChange: (value: string) => void;
  value: string;
}) {
  return (
    <label className="grid gap-1.5">
      <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
        Email
      </span>
      <input
        autoComplete="email"
        className="min-h-12 rounded-[18px] border border-[#bae7ff]/90 bg-white/88 px-4 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-accent focus:ring-4 focus:ring-ring/60"
        disabled={disabled}
        inputMode="email"
        onChange={(event) => onChange(event.target.value)}
        placeholder="fan@example.com"
        type="email"
        value={value}
      />
    </label>
  );
}

function FanAuthPasswordField({
  autoComplete,
  disabled,
  label,
  onChange,
  placeholder,
  value,
}: {
  autoComplete?: string;
  disabled: boolean;
  label: string;
  onChange: (value: string) => void;
  placeholder: string;
  value: string;
}) {
  return (
    <label className="grid gap-1.5">
      <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
        {label}
      </span>
      <input
        autoComplete={autoComplete}
        className="min-h-12 rounded-[18px] border border-[#bae7ff]/90 bg-white/88 px-4 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-accent focus:ring-4 focus:ring-ring/60"
        disabled={disabled}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        type="password"
        value={value}
      />
    </label>
  );
}

function FanAuthConfirmationCodeField({
  disabled,
  onChange,
  value,
}: {
  disabled: boolean;
  onChange: (value: string) => void;
  value: string;
}) {
  return (
    <label className="grid gap-1.5">
      <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
        Confirmation code
      </span>
      <input
        autoComplete="one-time-code"
        className="min-h-12 rounded-[18px] border border-[#bae7ff]/90 bg-white/88 px-4 text-sm tracking-[0.24em] text-foreground outline-none transition placeholder:text-muted focus:border-accent focus:ring-4 focus:ring-ring/60"
        disabled={disabled}
        inputMode="numeric"
        onChange={(event) => onChange(event.target.value)}
        placeholder="123456"
        type="text"
        value={value}
      />
    </label>
  );
}

function FanAuthMessage({
  kind,
  message,
}: {
  kind: "error" | "info";
  message: string;
}) {
  return (
    <p
      aria-live="polite"
      className={
        kind === "error"
          ? "rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
          : "rounded-[18px] border border-[#b9e1f5] bg-[#f1faff] px-4 py-3 text-sm leading-6 text-[#21617f]"
      }
      role={kind === "error" ? "alert" : "status"}
    >
      {message}
    </p>
  );
}

function buildDeliveryHintText(
  mode: FanAuthMode,
  deliveryDestinationHint: string | null,
) {
  if (mode !== "confirm-sign-up" && mode !== "confirm-password-reset") {
    return null;
  }

  if (deliveryDestinationHint) {
    return `確認コードを ${deliveryDestinationHint} に送りました。メールを確認してください。`;
  }

  if (mode === "confirm-password-reset") {
    return "確認コードを送信しました。メールを確認してください。";
  }

  return null;
}

function buildPrimaryActionLabel(mode: FanAuthMode, hasConfirmedSignUp: boolean, isSubmitting: boolean) {
  if (mode === "confirm-sign-up" && hasConfirmedSignUp && !isSubmitting) {
    return "avatar を保存して閉じる";
  }

  return getFanAuthSubmitLabel(mode, isSubmitting);
}

/**
 * 共通 fan auth entry panel を表示する。
 */
export function FanAuthEntryPanel({
  avatar,
  avatarInputKey,
  canResend,
  clearAvatarSelection,
  confirmationCode,
  deliveryDestinationHint,
  dismissAction,
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
  onAvatarSelect,
  onConfirmationCodeChange,
  onDisplayNameChange,
  onEmailChange,
  onHandleChange,
  onModeChange,
  onNewPasswordChange,
  onPasswordChange,
  onResend,
  onSubmit,
}: FanAuthEntryPanelProps) {
  const title = getFanAuthModeTitle(mode);
  const description = getFanAuthModeDescription(mode);
  const deliveryHintText = buildDeliveryHintText(mode, deliveryDestinationHint);

  return (
    <SurfacePanel className="w-full px-5 py-6 text-foreground">
      <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">fan access</p>
      <h1 className="mt-3 font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em]">
        {title}
      </h1>
      <p className="mt-3 text-sm leading-6 text-muted">{description}</p>

      <form
        className="mt-5 grid gap-3"
        onSubmit={(event) => {
          event.preventDefault();
          void onSubmit();
        }}
      >
        {mode === "re-auth" ? null : (
          <FanAuthEmailField disabled={isSubmitting} onChange={onEmailChange} value={email} />
        )}

        {mode === "sign-in" || mode === "sign-up" || mode === "re-auth" ? (
          <FanAuthPasswordField
            autoComplete={
              mode === "sign-in"
                ? "current-password"
                : mode === "re-auth"
                  ? "current-password"
                  : "new-password"
            }
            disabled={isSubmitting}
            label="Password"
            onChange={onPasswordChange}
            placeholder={
              mode === "re-auth" ? "現在のパスワード" : "8文字以上のパスワード"
            }
            value={password}
          />
        ) : null}

        {mode === "sign-up" ? (
          <SharedViewerProfileFields
            avatar={avatar}
            avatarInputKey={avatarInputKey}
            displayName={displayName}
            handle={handle}
            isSubmitting={isSubmitting}
            onAvatarClear={clearAvatarSelection}
            onAvatarSelect={onAvatarSelect}
            onDisplayNameChange={onDisplayNameChange}
            onHandleChange={onHandleChange}
          />
        ) : null}

        {mode === "confirm-sign-up" || mode === "confirm-password-reset" ? (
          <FanAuthConfirmationCodeField
            disabled={isSubmitting}
            onChange={onConfirmationCodeChange}
            value={confirmationCode}
          />
        ) : null}

        {mode === "confirm-password-reset" ? (
          <FanAuthPasswordField
            autoComplete="new-password"
            disabled={isSubmitting}
            label="New password"
            onChange={onNewPasswordChange}
            placeholder="新しいパスワード"
            value={newPassword}
          />
        ) : null}

        {deliveryHintText ? <FanAuthMessage kind="info" message={deliveryHintText} /> : null}
        {infoMessage ? <FanAuthMessage kind="info" message={infoMessage} /> : null}
        {errorMessage ? <FanAuthMessage kind="error" message={errorMessage} /> : null}

        <Button className="w-full" disabled={isSubmitting} type="submit">
          {buildPrimaryActionLabel(mode, hasConfirmedSignUp, isSubmitting)}
        </Button>
      </form>

      <div className="mt-4 rounded-[18px] bg-white/72 px-4 py-3 text-sm leading-6 text-muted">
        {mode === "sign-in" ? (
          <>
            <p>アカウントがまだない場合</p>
            <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-2">
              <ModeSwitchButton
                disabled={isSubmitting}
                label="サインアップへ"
                onClick={() => onModeChange("sign-up")}
              />
              <ModeSwitchButton
                disabled={isSubmitting}
                label="パスワードを再設定"
                onClick={() => onModeChange("password-reset-request")}
              />
            </div>
          </>
        ) : null}

        {mode === "sign-up" ? (
          <>
            <p>すでに登録済みの場合</p>
            <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-2">
              <ModeSwitchButton
                disabled={isSubmitting}
                label="サインインへ"
                onClick={() => onModeChange("sign-in")}
              />
            </div>
          </>
        ) : null}

        {mode === "confirm-sign-up" ? (
          <>
            <p>確認コードが届かない場合</p>
            <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-2">
              {canResend ? (
                <ModeSwitchButton
                  disabled={isSubmitting}
                  label="コードを再送"
                  onClick={() => {
                    void onResend();
                  }}
                />
              ) : null}
              <ModeSwitchButton
                disabled={isSubmitting}
                label="サインインへ"
                onClick={() => onModeChange("sign-in")}
              />
              {!hasConfirmedSignUp ? (
                <ModeSwitchButton
                  disabled={isSubmitting}
                  label="登録情報を編集"
                  onClick={() => onModeChange("sign-up")}
                />
              ) : null}
            </div>
          </>
        ) : null}

        {mode === "password-reset-request" ? (
          <>
            <p>サインインへ戻る場合</p>
            <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-2">
              <ModeSwitchButton
                disabled={isSubmitting}
                label="サインインへ"
                onClick={() => onModeChange("sign-in")}
              />
            </div>
          </>
        ) : null}

        {mode === "confirm-password-reset" ? (
          <>
            <p>確認コードが届かない場合</p>
            <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-2">
              {canResend ? (
                <ModeSwitchButton
                  disabled={isSubmitting}
                  label="コードを再送"
                  onClick={() => {
                    void onResend();
                  }}
                />
              ) : null}
              <ModeSwitchButton
                disabled={isSubmitting}
                label="サインインへ"
                onClick={() => onModeChange("sign-in")}
              />
            </div>
          </>
        ) : null}

        {mode === "re-auth" ? (
          <p>現在の fan session を維持したまま、必要な操作だけを続行します。</p>
        ) : null}
      </div>

      {dismissAction ? <div className="mt-3">{dismissAction}</div> : null}
    </SurfacePanel>
  );
}
