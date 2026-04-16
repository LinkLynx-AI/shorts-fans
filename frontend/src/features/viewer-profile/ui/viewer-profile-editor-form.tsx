"use client";

import Link from "next/link";
import { useId, type ReactNode } from "react";
import {
  AlertCircle,
  Camera,
  ChevronLeft,
  ImagePlus,
  LoaderCircle,
} from "lucide-react";

import { getCreatorInitials } from "@/entities/creator";
import { cn } from "@/shared/lib";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@/shared/ui";

import type { ViewerProfileAvatarField } from "../model/use-viewer-profile-draft";

type ViewerProfileEditorFormProps = {
  avatar: ViewerProfileAvatarField;
  avatarInputKey: number;
  backHref: string;
  backLabel: string;
  children?: ReactNode;
  displayName: string;
  errorMessage: string | null;
  handle: string;
  isSubmitting: boolean;
  onAvatarClear: () => void;
  onAvatarSelect: (file: File | null) => void;
  onDisplayNameChange: (value: string) => void;
  onHandleChange: (value: string) => void;
  onSubmit: () => void | Promise<void>;
  submitLabel?: string;
  submittingLabel?: string;
};

type AvatarStatus = {
  className: string;
  icon: typeof AlertCircle;
  message: string;
};

function buildAvatarStatus(avatar: ViewerProfileAvatarField): AvatarStatus | null {
  switch (avatar.kind) {
    case "invalid":
    case "failed":
      return {
        className: "text-[#c45b72]",
        icon: AlertCircle,
        message: avatar.message,
      };
    case "uploading":
      return {
        className: "text-[#6b7a90]",
        icon: LoaderCircle,
        message: "画像をアップロードしています。",
      };
    case "selected":
      return {
        className: "text-[#7b8799]",
        icon: ImagePlus,
        message: "保存すると新しい画像に切り替わります。",
      };
    default:
      return null;
  }
}

function ViewerProfileAvatarFieldRow({
  avatar,
  avatarInputKey,
  displayName,
  isSubmitting,
  onAvatarClear,
  onAvatarSelect,
}: {
  avatar: ViewerProfileAvatarField;
  avatarInputKey: number;
  displayName: string;
  isSubmitting: boolean;
  onAvatarClear: () => void;
  onAvatarSelect: (file: File | null) => void;
}) {
  const avatarInputId = useId();
  const avatarStatus = buildAvatarStatus(avatar);
  const avatarFallbackLabel = displayName.trim() === "" ? "ME" : getCreatorInitials(displayName);
  const StatusIcon = avatarStatus?.icon;

  return (
    <section className="flex flex-col items-center pt-2">
      <div className="relative">
        <Avatar className="size-[122px] border border-[#eef1f5] bg-[linear-gradient(180deg,#f1f3f7_0%,#dfe6ee_100%)] text-[32px] font-semibold text-[#2d3340] shadow-none">
          {avatar.previewUrl ? <AvatarImage alt={`${displayName} avatar`} src={avatar.previewUrl} /> : null}
          <AvatarFallback className="bg-transparent text-inherit">{avatarFallbackLabel}</AvatarFallback>
        </Avatar>
        <label
          className={cn(
            "absolute bottom-0 right-0 inline-flex size-[46px] cursor-pointer items-center justify-center rounded-full border border-[#e5e8ef] bg-white text-[#5a6678] shadow-[0_10px_18px_rgba(15,23,42,0.12)] transition hover:bg-[#f9fafc]",
            isSubmitting && "pointer-events-none opacity-60",
          )}
          htmlFor={avatarInputId}
        >
          <Camera aria-hidden="true" className="size-5" strokeWidth={2.1} />
          <span className="sr-only">avatar を変更</span>
        </label>
        <input
          accept={avatar.inputAccept}
          className="sr-only"
          disabled={isSubmitting}
          id={avatarInputId}
          key={avatarInputKey}
          onChange={(event) => onAvatarSelect(event.target.files?.[0] ?? null)}
          type="file"
        />
      </div>

      {avatarStatus && StatusIcon ? (
        <p
          aria-live="polite"
          className={cn("mt-4 inline-flex items-center gap-2 text-center text-[13px] font-medium", avatarStatus.className)}
        >
          <StatusIcon
            aria-hidden="true"
            className={cn("size-4 shrink-0", avatar.kind === "uploading" && "animate-spin")}
            strokeWidth={2.2}
          />
          <span>{avatarStatus.message}</span>
        </p>
      ) : null}

      {avatar.canClear ? (
        <button
          className="mt-3 text-[13px] font-medium text-[#8a93a2] transition hover:text-[#5f6c7f] disabled:cursor-not-allowed disabled:opacity-60"
          disabled={isSubmitting}
          onClick={onAvatarClear}
          type="button"
        >
          選択を外す
        </button>
      ) : null}
    </section>
  );
}

function ViewerProfileTextField({
  autoCapitalize,
  autoComplete,
  autoCorrect,
  children,
  disabled,
  id,
  label,
  onChange,
  spellCheck,
  type,
  value,
}: {
  autoCapitalize?: string;
  autoComplete?: string;
  autoCorrect?: string;
  children?: ReactNode;
  disabled: boolean;
  id: string;
  label: string;
  onChange: (value: string) => void;
  spellCheck?: boolean;
  type: "text";
  value: string;
}) {
  return (
    <label className="block" htmlFor={id}>
      <span className="block pl-1 text-[12px] font-extrabold uppercase tracking-[0.12em] text-[#a9aeb9]">
        {label}
      </span>
      <input
        autoCapitalize={autoCapitalize}
        autoComplete={autoComplete}
        autoCorrect={autoCorrect}
        className="mt-2.5 h-[56px] w-full rounded-[20px] border border-transparent bg-[#f6f7fa] px-5 text-[15px] font-semibold text-[#1f2430] outline-none transition placeholder:text-[#b5bbc6] focus:border-[#d7e6f5] focus:bg-white focus:ring-4 focus:ring-[rgba(113,180,234,0.18)]"
        disabled={disabled}
        id={id}
        onChange={(event) => onChange(event.target.value)}
        spellCheck={spellCheck}
        type={type}
        value={value}
      />
      {children}
    </label>
  );
}

/**
 * fan 基準の viewer profile 編集フォームを表示する。
 */
export function ViewerProfileEditorForm({
  avatar,
  avatarInputKey,
  backHref,
  backLabel,
  children,
  displayName,
  errorMessage,
  handle,
  isSubmitting,
  onAvatarClear,
  onAvatarSelect,
  onDisplayNameChange,
  onHandleChange,
  onSubmit,
  submitLabel = "保存する",
  submittingLabel = "保存中...",
}: ViewerProfileEditorFormProps) {
  return (
    <main className="mx-auto min-h-full w-full max-w-[430px] bg-white text-[#1f2430]">
      <header className="grid grid-cols-[40px_1fr_40px] items-center border-b border-[#edf0f4] px-6 pb-5 pt-5">
        <Link
          aria-label={backLabel}
          className="inline-flex size-10 items-center justify-center rounded-full text-[#1f2430] transition hover:bg-[#f5f7fa]"
          href={backHref}
        >
          <ChevronLeft aria-hidden="true" className="size-7" strokeWidth={2.1} />
        </Link>
        <h1 className="text-center text-[17px] font-bold tracking-[-0.02em] text-[#1f2430]">
          プロフィールを編集
        </h1>
        <span aria-hidden="true" className="block size-10" />
      </header>

      <form
        className="px-9 pb-14 pt-10"
        onSubmit={(event) => {
          event.preventDefault();
          void onSubmit();
        }}
      >
        <ViewerProfileAvatarFieldRow
          avatar={avatar}
          avatarInputKey={avatarInputKey}
          displayName={displayName}
          isSubmitting={isSubmitting}
          onAvatarClear={onAvatarClear}
          onAvatarSelect={onAvatarSelect}
        />

        <div className="mt-12 grid gap-6">
          <ViewerProfileTextField
            autoComplete="nickname"
            disabled={isSubmitting}
            id="viewer-profile-display-name"
            label="Display Name"
            onChange={onDisplayNameChange}
            type="text"
            value={displayName}
          />
          <ViewerProfileTextField
            autoCapitalize="none"
            autoCorrect="off"
            autoComplete="username"
            disabled={isSubmitting}
            id="viewer-profile-handle"
            label="Handle"
            onChange={onHandleChange}
            spellCheck={false}
            type="text"
            value={handle}
          />
          {children}
        </div>

        {errorMessage ? (
          <p
            aria-live="polite"
            className="mt-6 rounded-[18px] border border-[#f2ccd5] bg-[#fff5f7] px-5 py-4 text-[14px] leading-6 text-[#b2445d]"
            role="alert"
          >
            {errorMessage}
          </p>
        ) : null}

        <button
          className="mt-8 inline-flex h-[54px] w-full items-center justify-center rounded-full bg-[#68a7dc] px-5 text-[15px] font-bold text-white shadow-[0_10px_22px_rgba(104,167,220,0.24)] transition hover:bg-[#5b9fd9] disabled:cursor-not-allowed disabled:opacity-70"
          disabled={isSubmitting}
          type="submit"
        >
          {isSubmitting ? submittingLabel : submitLabel}
        </button>
      </form>
    </main>
  );
}
