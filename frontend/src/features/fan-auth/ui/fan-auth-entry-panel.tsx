"use client";

import type { ReactNode } from "react";

import {
  AlertCircle,
  Camera,
  CheckCheck,
  ChevronRight,
  ImagePlus,
  LoaderCircle,
} from "lucide-react";

import { getCreatorInitials } from "@/entities/creator";
import type { ViewerProfileAvatarField } from "@/features/viewer-profile";
import { cn } from "@/shared/lib";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Button,
} from "@/shared/ui";

import {
  getFanAuthModeDescription,
  getFanAuthModeTitle,
  getFanAuthSubmitLabel,
  type FanAuthMode,
} from "../model/fan-auth";

type FanAuthEntryPanelProps = {
  avatar: ViewerProfileAvatarField;
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

const fieldLabelClassName =
  "block text-[12px] font-black uppercase tracking-[0.18em] text-[#a3adbc]";
const inputClassName =
  "min-h-16 w-full rounded-[24px] border border-transparent bg-[#f7f9fc] px-5 text-[15px] font-bold text-foreground outline-none transition placeholder:text-[#9aa4b2] focus:border-[#d8ebfb] focus:bg-white focus:ring-4 focus:ring-ring/60 disabled:cursor-not-allowed disabled:opacity-60";

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
      className="font-bold text-accent-strong transition hover:text-accent disabled:cursor-not-allowed disabled:opacity-60"
      disabled={disabled}
      onClick={onClick}
      type="button"
    >
      {label}
    </button>
  );
}

function FanAuthField({
  children,
  label,
}: {
  children: ReactNode;
  label: string;
}) {
  return (
    <label className="grid gap-2.5">
      <span className={fieldLabelClassName}>{label}</span>
      {children}
    </label>
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
    <FanAuthField label="Email">
      <input
        autoComplete="email"
        className={inputClassName}
        disabled={disabled}
        inputMode="email"
        onChange={(event) => onChange(event.target.value)}
        placeholder="fan@example.com"
        type="email"
        value={value}
      />
    </FanAuthField>
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
    <FanAuthField label={label}>
      <input
        autoComplete={autoComplete}
        className={inputClassName}
        disabled={disabled}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        type="password"
        value={value}
      />
    </FanAuthField>
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
    <FanAuthField label="Confirmation code">
      <input
        autoComplete="one-time-code"
        className={cn(inputClassName, "text-center text-[20px] tracking-[0.5em]")}
        disabled={disabled}
        inputMode="numeric"
        onChange={(event) => onChange(event.target.value)}
        placeholder="1 2 3 4 5 6"
        type="text"
        value={value}
      />
    </FanAuthField>
  );
}

function FanAuthTextField({
  autoCapitalize,
  autoCorrect,
  disabled,
  helpText,
  id,
  label,
  onChange,
  placeholder,
  spellCheck,
  value,
}: {
  autoCapitalize?: string;
  autoCorrect?: string;
  disabled: boolean;
  helpText?: ReactNode;
  id: string;
  label: string;
  onChange: (value: string) => void;
  placeholder: string;
  spellCheck?: boolean;
  value: string;
}) {
  return (
    <div className="grid gap-2.5">
      <label className="grid gap-2.5" htmlFor={id}>
        <span className={fieldLabelClassName}>{label}</span>
      </label>
      <input
        aria-describedby={helpText ? `${id}-help` : undefined}
        autoCapitalize={autoCapitalize}
        autoCorrect={autoCorrect}
        className={inputClassName}
        disabled={disabled}
        id={id}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        spellCheck={spellCheck}
        type="text"
        value={value}
      />
      {helpText ? (
        <span className="pl-1 text-[12px] leading-5 text-muted" id={`${id}-help`}>
          {helpText}
        </span>
      ) : null}
    </div>
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
          ? "rounded-[22px] border border-[#f5c8d1] bg-[#fff6f8] px-4 py-3 text-[13px] leading-6 text-[#b2394f]"
          : "rounded-[22px] border border-[#d7ebfb] bg-[#f4fafe] px-4 py-3 text-[13px] leading-6 text-[#256182]"
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

function buildPrimaryActionLabel(
  mode: FanAuthMode,
  hasConfirmedSignUp: boolean,
  isSubmitting: boolean,
) {
  if (mode === "confirm-sign-up" && hasConfirmedSignUp && !isSubmitting) {
    return "avatar を保存して閉じる";
  }

  return getFanAuthSubmitLabel(mode, isSubmitting);
}

function buildAvatarSurfaceClasses(kind: ViewerProfileAvatarField["kind"]) {
  switch (kind) {
    case "failed":
    case "invalid":
      return "border-[#f3cbd4] bg-[#fff7f8]";
    case "uploading":
      return "border-[#d8ebf9] bg-[#f6fbfe]";
    case "completed":
      return "border-[#dce8e1] bg-[#f7fbf8]";
    case "selected":
      return "border-[#ddeaf3] bg-[#f8fbfd]";
    case "empty":
    default:
      return "border-[#e6edf4] bg-[#f7f9fc]";
  }
}

function buildAvatarPreviewClasses(kind: ViewerProfileAvatarField["kind"]) {
  switch (kind) {
    case "failed":
    case "invalid":
      return "border-[#f1c4cf] bg-[#fdecef] text-[#9b3950]";
    case "uploading":
      return "border-[#d0e4ef] bg-[#edf7fb] text-[#2f6c82]";
    case "completed":
      return "border-[#d3e4da] bg-[#eef8f1] text-[#2b6c4d]";
    case "selected":
      return "border-[#d8e6ee] bg-[#eef6fa] text-[#31687e]";
    case "empty":
    default:
      return "border-[#dfe8ef] bg-white text-[#8290a1]";
  }
}

function buildAvatarActionLabel(fileName: string | null) {
  return fileName ? "画像を変更" : "画像を選択";
}

function renderAvatarStateIcon(kind: ViewerProfileAvatarField["kind"]) {
  switch (kind) {
    case "failed":
    case "invalid":
      return (
        <AlertCircle
          aria-hidden="true"
          className="size-3.5 shrink-0"
          strokeWidth={2.2}
        />
      );
    case "uploading":
      return (
        <LoaderCircle
          aria-hidden="true"
          className="size-3.5 shrink-0 animate-spin"
          strokeWidth={2.2}
        />
      );
    case "completed":
      return (
        <CheckCheck
          aria-hidden="true"
          className="size-3.5 shrink-0"
          strokeWidth={2.2}
        />
      );
    case "selected":
      return (
        <ImagePlus
          aria-hidden="true"
          className="size-3.5 shrink-0"
          strokeWidth={2.2}
        />
      );
    case "empty":
    default:
      return (
        <Camera
          aria-hidden="true"
          className="size-3.5 shrink-0"
          strokeWidth={2.2}
        />
      );
  }
}

function FanAuthAvatarField({
  avatar,
  avatarInputKey,
  clearAvatarSelection,
  displayName,
  isSubmitting,
  onAvatarSelect,
}: {
  avatar: ViewerProfileAvatarField;
  avatarInputKey: number;
  clearAvatarSelection: () => void;
  displayName: string;
  isSubmitting: boolean;
  onAvatarSelect: (file: File | null) => void;
}) {
  const avatarLabel = buildAvatarActionLabel(avatar.fileName);
  const avatarMonogram =
    displayName.trim() === "" ? "ME" : getCreatorInitials(displayName);

  return (
    <div className="grid gap-2.5">
      <div className="flex items-center gap-2">
        <span className={fieldLabelClassName}>Avatar</span>
        <span className="text-[10px] font-bold text-muted">(任意)</span>
      </div>

      <div
        className={cn(
          "rounded-[24px] border px-4 py-4 transition",
          buildAvatarSurfaceClasses(avatar.kind),
        )}
      >
        <div className="flex items-center justify-between gap-4">
          <div className="flex min-w-0 items-center gap-4">
            <Avatar
              className={cn(
                "size-14 rounded-full border shadow-none",
                buildAvatarPreviewClasses(avatar.kind),
              )}
            >
              {avatar.previewUrl ? (
                <AvatarImage alt="選択した avatar プレビュー" src={avatar.previewUrl} />
              ) : null}
              <AvatarFallback className="bg-transparent text-inherit">
                {avatar.previewUrl ? null : avatarMonogram}
              </AvatarFallback>
            </Avatar>

            <div className="min-w-0">
              <p className="truncate text-[14px] font-bold text-foreground">
                {avatarLabel}
              </p>
              <p className="mt-0.5 text-[11px] font-medium text-muted">
                JPEG / PNG / WebP, 5MBまで
              </p>
            </div>
          </div>

          <ChevronRight
            aria-hidden="true"
            className="size-5 shrink-0 text-[#b7c0cc]"
            strokeWidth={2.4}
          />
        </div>

        <div className="mt-3 flex flex-wrap items-center gap-x-3 gap-y-2">
          <label
            className={cn(
              "inline-flex min-h-10 cursor-pointer items-center justify-center rounded-full bg-white px-4 text-[13px] font-bold text-accent-strong shadow-[0_4px_12px_rgba(15,23,42,0.05)] transition hover:bg-white/90",
              isSubmitting && "cursor-not-allowed opacity-60",
            )}
            htmlFor="fan-auth-avatar"
          >
            {avatarLabel}
          </label>
          {avatar.canClear ? (
            <button
              className="text-[13px] font-bold text-muted transition hover:text-accent-strong disabled:cursor-not-allowed disabled:opacity-60"
              disabled={isSubmitting}
              onClick={clearAvatarSelection}
              type="button"
            >
              外す
            </button>
          ) : null}
        </div>

        <p
          aria-live="polite"
          className={cn(
            "mt-3 inline-flex items-start gap-1.5 text-[12px] leading-5",
            avatar.isError ? "text-[#a83853]" : "text-muted",
          )}
        >
          {renderAvatarStateIcon(avatar.kind)}
          {avatar.message}
        </p>

        <label className="sr-only" htmlFor="fan-auth-avatar">
          Avatar image
        </label>
        <input
          accept={avatar.inputAccept}
          className="sr-only"
          disabled={isSubmitting}
          id="fan-auth-avatar"
          key={avatarInputKey}
          onChange={(event) => onAvatarSelect(event.target.files?.[0] ?? null)}
          type="file"
        />
      </div>
    </div>
  );
}

function buildInlineLinks(mode: FanAuthMode) {
  switch (mode) {
    case "sign-in":
      return "grid justify-center gap-3 text-center";
    case "sign-up":
    case "password-reset-request":
    case "re-auth":
      return "grid justify-center gap-3 text-center";
    case "confirm-sign-up":
    case "confirm-password-reset":
      return "flex flex-wrap justify-center gap-x-4 gap-y-2 text-center";
  }
}

function buildTitleClassName(mode: FanAuthMode) {
  if (mode === "sign-in") {
    return "max-w-[11ch]";
  }

  return "max-w-full";
}

function buildInfoMessages({
  deliveryHintText,
  infoMessage,
}: {
  deliveryHintText: string | null;
  infoMessage: string | null;
}) {
  const messages: string[] = [];

  if (deliveryHintText) {
    messages.push(deliveryHintText);
  }

  if (infoMessage) {
    const shouldSuppressGenericDeliveryMessage =
      deliveryHintText !== null && infoMessage.startsWith("確認コードを送信しました");

    if (!shouldSuppressGenericDeliveryMessage) {
      messages.push(infoMessage);
    }
  }

  return messages;
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
  const infoMessages = buildInfoMessages({
    deliveryHintText,
    infoMessage,
  });
  const showDescription = mode === "re-auth";

  return (
    <div className="w-full text-foreground">
      <div className="mt-2">
        <h1
          className={cn(
            buildTitleClassName(mode),
            "font-display text-[32px] font-semibold leading-[1.12] tracking-[-0.05em] text-foreground",
          )}
        >
          {title}
        </h1>
        {showDescription ? (
          <p className="mt-3 max-w-[34ch] text-sm leading-6 text-muted">
            {description}
          </p>
        ) : null}
      </div>

      {infoMessages.length > 0 ? (
        <div className="mt-6 grid gap-3">
          {infoMessages.map((message) => (
            <FanAuthMessage key={message} kind="info" message={message} />
          ))}
        </div>
      ) : null}

      <form
        className="mt-6 grid gap-4"
        onSubmit={(event) => {
          event.preventDefault();
          void onSubmit();
        }}
      >
        {mode === "re-auth" ? null : (
          <FanAuthEmailField
            disabled={isSubmitting}
            onChange={onEmailChange}
            value={email}
          />
        )}

        {mode === "sign-in" || mode === "sign-up" || mode === "re-auth" ? (
          <FanAuthPasswordField
            autoComplete={
              mode === "sign-in" || mode === "re-auth"
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
          <>
            <FanAuthTextField
              disabled={isSubmitting}
              id="fan-auth-display-name"
              label="Display name"
              onChange={onDisplayNameChange}
              placeholder="Mina Rei"
              value={displayName}
            />
            <FanAuthTextField
              autoCapitalize="none"
              autoCorrect="off"
              disabled={isSubmitting}
              helpText={
                <>
                  先頭の
                  <code className="rounded bg-[#edf2f7] px-1.5 py-0.5 text-[11px] text-muted-strong">@</code>
                  は省略可、使える文字は英数字・
                  <code className="rounded bg-[#edf2f7] px-1.5 py-0.5 text-[11px] text-muted-strong">.</code>・
                  <code className="rounded bg-[#edf2f7] px-1.5 py-0.5 text-[11px] text-muted-strong">_</code>
                  です。
                </>
              }
              id="fan-auth-handle"
              label="Handle"
              onChange={onHandleChange}
              placeholder="@minarei"
              spellCheck={false}
              value={handle}
            />
            <FanAuthAvatarField
              avatar={avatar}
              avatarInputKey={avatarInputKey}
              clearAvatarSelection={clearAvatarSelection}
              displayName={displayName}
              isSubmitting={isSubmitting}
              onAvatarSelect={onAvatarSelect}
            />
          </>
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

        {errorMessage ? <FanAuthMessage kind="error" message={errorMessage} /> : null}

        <Button
          className="mt-2 h-14 w-full text-[16px] font-bold"
          disabled={isSubmitting}
          type="submit"
        >
          {buildPrimaryActionLabel(mode, hasConfirmedSignUp, isSubmitting)}
        </Button>
      </form>

      {mode === "re-auth" ? (
        <div className="mt-6 rounded-[22px] bg-[#f7f9fc] px-4 py-3 text-[13px] leading-6 text-muted">
          現在の fan session を維持したまま、必要な操作だけを続行します。
        </div>
      ) : (
        <div className="mt-6">
          {mode === "sign-up" ? (
            <p className="text-center text-[13px] font-medium text-muted">
              すでに登録済みの場合
            </p>
          ) : null}
          <div className={cn("mt-3", buildInlineLinks(mode))}>
            {mode === "sign-in" ? (
              <>
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
              </>
            ) : null}

            {mode === "sign-up" ? (
              <ModeSwitchButton
                disabled={isSubmitting}
                label="サインインへ"
                onClick={() => onModeChange("sign-in")}
              />
            ) : null}

            {mode === "confirm-sign-up" ? (
              <>
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
              </>
            ) : null}

            {mode === "password-reset-request" ? (
              <ModeSwitchButton
                disabled={isSubmitting}
                label="サインインへ戻る"
                onClick={() => onModeChange("sign-in")}
              />
            ) : null}

            {mode === "confirm-password-reset" ? (
              <>
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
              </>
            ) : null}
          </div>
        </div>
      )}

      {dismissAction ? <div className="mt-6">{dismissAction}</div> : null}
    </div>
  );
}
