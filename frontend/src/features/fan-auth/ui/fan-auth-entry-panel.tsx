"use client";

import type { ReactNode } from "react";

import { Button, SurfacePanel } from "@/shared/ui";

import {
  getFanAuthModeHint,
  getFanAuthModeSwitchLabel,
  getFanAuthSubmitLabel,
  type FanAuthMode,
} from "../model/fan-auth";

type FanAuthEntryPanelProps = {
  dismissAction?: ReactNode;
  email: string;
  errorMessage: string | null;
  isSubmitting: boolean;
  mode: FanAuthMode;
  onEmailChange: (email: string) => void;
  onModeSwitch: () => void;
  onSubmit: () => void | Promise<void>;
};

/**
 * 共通 fan auth entry panel を表示する。
 */
export function FanAuthEntryPanel({
  dismissAction,
  email,
  errorMessage,
  isSubmitting,
  mode,
  onEmailChange,
  onModeSwitch,
  onSubmit,
}: FanAuthEntryPanelProps) {
  return (
    <SurfacePanel className="w-full px-5 py-6 text-foreground">
      <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">fan access</p>
      <h1 className="mt-3 font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em]">
        続けるにはログインが必要です
      </h1>
      <p className="mt-3 text-sm leading-6 text-muted">
        fan session を開始すると、フォロー中、library、main 再生のような protected surface をそのまま続けられます。
      </p>

      <form
        className="mt-5 grid gap-3"
        onSubmit={(event) => {
          event.preventDefault();
          void onSubmit();
        }}
      >
        <label className="grid gap-1.5">
          <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
            Email
          </span>
          <input
            autoComplete="email"
            className="min-h-12 rounded-[18px] border border-[#bae7ff]/90 bg-white/88 px-4 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-accent focus:ring-4 focus:ring-ring/60"
            disabled={isSubmitting}
            inputMode="email"
            onChange={(event) => onEmailChange(event.target.value)}
            placeholder="fan@example.com"
            type="email"
            value={email}
          />
        </label>

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
          {getFanAuthSubmitLabel(mode, isSubmitting)}
        </Button>
      </form>

      <div className="mt-4 rounded-[18px] bg-white/72 px-4 py-3 text-sm leading-6 text-muted">
        <p>{getFanAuthModeHint(mode)}</p>
        <button
          className="mt-1 font-semibold text-accent-strong transition hover:text-accent"
          disabled={isSubmitting}
          onClick={onModeSwitch}
          type="button"
        >
          {getFanAuthModeSwitchLabel(mode)}
        </button>
      </div>

      {dismissAction ? <div className="mt-3">{dismissAction}</div> : null}
    </SurfacePanel>
  );
}
