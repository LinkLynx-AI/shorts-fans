import Link from "next/link";
import type { ReactNode } from "react";

import { Separator } from "@/shared/ui";

import { BottomTabBar } from "./bottom-tab-bar";

type ConsumerShellProps = {
  children: ReactNode;
};

/**
 * consumer 向け route に共通する shell を表示する。
 */
export function ConsumerShell({ children }: ConsumerShellProps) {
  return (
    <div className="relative min-h-screen overflow-hidden">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_10%_20%,rgba(255,255,255,0.4),transparent_28%),radial-gradient(circle_at_90%_10%,rgba(209,92,24,0.18),transparent_30%)]" />
      <div className="relative mx-auto flex min-h-screen w-full max-w-6xl flex-col gap-6 px-6 py-6 pb-28 sm:px-10">
        <header className="flex flex-col gap-4 rounded-[2rem] border border-white/70 bg-white/56 px-5 py-4 shadow-[0_18px_48px_rgba(87,38,8,0.12)] backdrop-blur-lg sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.26em] text-accent">frontend foundation</p>
            <p className="mt-2 text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">shorts first, home keeps the map.</p>
          </div>
          <div className="flex items-center gap-3">
            <Link
              className="rounded-full border border-white/70 bg-white/82 px-4 py-2 text-sm font-semibold text-foreground shadow-[0_12px_24px_rgba(87,38,8,0.1)]"
              href="/shorts"
            >
              primary lane
            </Link>
            <Link
              className="rounded-full bg-[#1d120d] px-4 py-2 text-sm font-semibold text-white shadow-[0_14px_28px_rgba(35,16,8,0.22)]"
              href="/creator/atelier-rin"
            >
              creator sample
            </Link>
          </div>
        </header>
        <Separator />
        <main className="flex-1">{children}</main>
      </div>
      <BottomTabBar />
    </div>
  );
}
