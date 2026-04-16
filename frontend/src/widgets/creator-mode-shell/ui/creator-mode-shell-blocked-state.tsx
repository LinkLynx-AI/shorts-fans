"use client";

import Link from "next/link";
import type { ReactNode } from "react";

import { Button, SurfacePanel } from "@/shared/ui";

import type { CreatorModeShellBlockedState } from "../model/creator-mode-shell";

function CreatorModeBlockedFrame({ children }: { children: ReactNode }) {
  return (
    <main className="min-h-svh bg-background text-foreground sm:px-4">
      {children}
    </main>
  );
}

export function CreatorModeWorkspaceFrame({ children }: { children: ReactNode }) {
  return (
    <main className="bg-background sm:px-4">
      <div className="mx-auto min-h-svh w-full max-w-[408px] overflow-hidden bg-white text-foreground sm:border sm:border-border sm:shadow-[var(--device-shadow)]">
        {children}
      </div>
    </main>
  );
}

export function CreatorShellBlockedState({ state }: { state: CreatorModeShellBlockedState }) {
  return (
    <CreatorModeBlockedFrame>
      <div className="mx-auto flex min-h-svh max-w-3xl items-center px-4 py-12 sm:px-6">
        <SurfacePanel className="w-full px-6 py-7 sm:px-8 sm:py-9">
          <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[#0f6172]">{state.eyebrow}</p>
          <h1 className="mt-4 font-display text-[30px] font-semibold tracking-[-0.05em] text-foreground">
            {state.title}
          </h1>
          <p className="mt-3 text-sm leading-7 text-muted">{state.description}</p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Button asChild>
              <Link href={state.ctaHref}>{state.ctaLabel}</Link>
            </Button>
          </div>
        </SurfacePanel>
      </div>
    </CreatorModeBlockedFrame>
  );
}
