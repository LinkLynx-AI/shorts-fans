"use client";

import Link from "next/link";
import type { ReactNode } from "react";

import { Button, SurfacePanel } from "@/shared/ui";

type ProfileEditorPanelProps = {
  backHref: string;
  backLabel: string;
  children: ReactNode;
  description: string;
  errorMessage: string | null;
  eyebrow: string;
  isSubmitting: boolean;
  onSubmit: () => void | Promise<void>;
  submitLabel: string;
  submittingLabel: string;
  title: string;
};

export function ProfileEditorPanel({
  backHref,
  backLabel,
  children,
  description,
  errorMessage,
  eyebrow,
  isSubmitting,
  onSubmit,
  submitLabel,
  submittingLabel,
  title,
}: ProfileEditorPanelProps) {
  return (
    <main className="mx-auto flex min-h-full w-full max-w-[408px] flex-col px-4 pb-28 pt-5">
      <SurfacePanel className="w-full px-5 py-6 text-foreground">
        <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">{eyebrow}</p>
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
          {children}

          {errorMessage ? (
            <p
              aria-live="polite"
              className="rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
              role="alert"
            >
              {errorMessage}
            </p>
          ) : null}

          <Button className="w-full" disabled={isSubmitting} type="submit">
            {isSubmitting ? submittingLabel : submitLabel}
          </Button>
        </form>

        <div className="mt-3">
          <Button asChild className="w-full" disabled={isSubmitting} variant="secondary">
            <Link href={backHref}>{backLabel}</Link>
          </Button>
        </div>
      </SurfacePanel>
    </main>
  );
}
