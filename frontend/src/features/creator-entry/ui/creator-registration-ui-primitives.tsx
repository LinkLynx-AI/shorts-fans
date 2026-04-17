"use client";

import type { ReactNode } from "react";

import { cn } from "@/shared/lib";

export const creatorRegistrationShellClassName =
  "w-full rounded-[32px] bg-white px-5 py-8 shadow-[0_18px_40px_rgba(15,23,42,0.08)]";

export const creatorRegistrationSectionClassName =
  "rounded-[24px] bg-[#f7f9fc] px-4 py-4 text-foreground";

export const creatorRegistrationInlineSurfaceClassName =
  "rounded-[22px] bg-white px-4 py-4";

export const creatorRegistrationFieldLabelClassName =
  "block text-[12px] font-black tracking-[0.08em] text-[#a3adbc]";

export const creatorRegistrationInputClassName =
  "min-h-16 w-full rounded-[24px] border border-transparent bg-[#f7f9fc] px-5 text-[15px] font-bold text-foreground outline-none transition placeholder:text-[#9aa4b2] focus:border-[#d8ebfb] focus:bg-white focus:ring-4 focus:ring-ring/60 disabled:cursor-not-allowed disabled:opacity-60";

export const creatorRegistrationTextareaClassName =
  "min-h-[144px] w-full resize-none rounded-[24px] border border-transparent bg-[#f7f9fc] px-5 py-4 text-[15px] leading-7 text-foreground outline-none transition placeholder:text-[#9aa4b2] focus:border-[#d8ebfb] focus:bg-white focus:ring-4 focus:ring-ring/60 disabled:cursor-not-allowed disabled:opacity-60";

export const creatorRegistrationButtonClassName =
  "h-14 w-full text-[16px] font-bold";

export const creatorRegistrationFocusRingClassName =
  "focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-ring/60";

export function buildCreatorRegistrationAvatarFallback(displayName: string) {
  const trimmed = displayName.trim();
  if (trimmed === "") {
    return "ME";
  }

  return trimmed
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part.at(0)?.toUpperCase() ?? "")
    .join("");
}

export function CreatorRegistrationMessage({
  className,
  kind,
  message,
}: {
  className?: string;
  kind: "error" | "info" | "success";
  message: string;
}) {
  return (
    <p
      aria-live={kind === "error" ? "assertive" : "polite"}
      className={cn(
        kind === "error"
          ? "rounded-[22px] border border-[#f5c8d1] bg-[#fff6f8] px-4 py-3 text-[13px] leading-6 text-[#b2394f]"
          : kind === "success"
            ? "rounded-[22px] border border-[#dce8e1] bg-[#f7fbf8] px-4 py-3 text-[13px] leading-6 text-[#25664a]"
            : "rounded-[22px] border border-[#d7ebfb] bg-[#f4fafe] px-4 py-3 text-[13px] leading-6 text-[#256182]",
        className,
      )}
      role={kind === "error" ? "alert" : "status"}
    >
      {message}
    </p>
  );
}

export function CreatorRegistrationSectionHeading({
  children,
}: {
  children: ReactNode;
}) {
  return (
    <p className={creatorRegistrationFieldLabelClassName}>
      {children}
    </p>
  );
}
